package operator

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"

	_ "github.com/prestodb/presto-go-client/presto"
	promapi "github.com/prometheus/client_golang/api"
	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/transport"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/db"
	cbClientset "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned"
	cbInformers "github.com/operator-framework/operator-metering/pkg/generated/informers/externalversions"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
	"github.com/operator-framework/operator-metering/pkg/presto"
	_ "github.com/operator-framework/operator-metering/pkg/util/reflector/prometheus" // for prometheus metric registration
	_ "github.com/operator-framework/operator-metering/pkg/util/workqueue/prometheus" // for prometheus metric registration
)

const (
	connBackoff         = time.Second * 15
	maxConnWaitTime     = time.Minute * 3
	defaultResyncPeriod = time.Minute * 15

	serviceServingCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt"

	DefaultPrometheusQueryInterval  = time.Minute * 5
	DefaultPrometheusQueryStepSize  = time.Minute
	DefaultPrometheusQueryChunkSize = time.Minute * 5
)

type TLSConfig struct {
	UseTLS  bool
	TLSCert string
	TLSKey  string
}

func (cfg *TLSConfig) Valid() error {
	if cfg.UseTLS {
		if cfg.TLSCert == "" {
			return fmt.Errorf("Must set TLS certificate if TLS is enabled")
		}
		if cfg.TLSKey == "" {
			return fmt.Errorf("Must set TLS private key if TLS is enabled")
		}
	}
	return nil
}

type Config struct {
	PodName    string
	Hostname   string
	Namespace  string
	Kubeconfig string

	HiveHost         string
	PrestoHost       string
	PromHost         string
	DisablePromsum   bool
	EnableFinalizers bool

	LogDMLQueries bool
	LogDDLQueries bool

	PrometheusQueryConfig cbTypes.PrometheusQueryConfig

	LeaderLeaseDuration time.Duration

	APITLSConfig     TLSConfig
	MetricsTLSConfig TLSConfig
}

type Reporting struct {
	cfg            Config
	kubeConfig     *rest.Config
	informers      cbInformers.SharedInformerFactory
	queues         queues
	meteringClient cbClientset.Interface
	kubeClient     corev1.CoreV1Interface

	prestoConn    *sql.DB
	prestoQueryer presto.ExecQueryer
	hiveQueryer   *hiveQueryer
	promConn      prom.API

	clock clock.Clock
	rand  *rand.Rand

	logger log.FieldLogger

	initializedMu sync.Mutex
	initialized   bool

	importersMu sync.Mutex
	importers   map[string]*prestostore.PrometheusImporter

	// ensures only at most a single testRead query is running against Presto
	// at one time
	healthCheckSingleFlight singleflight.Group
}

func New(logger log.FieldLogger, cfg Config, clock clock.Clock) (*Reporting, error) {
	op := &Reporting{
		cfg:       cfg,
		logger:    logger,
		clock:     clock,
		importers: make(map[string]*prestostore.PrometheusImporter),
	}
	logger.Debugf("Config: %+v", cfg)

	if err := cfg.APITLSConfig.Valid(); err != nil {
		return nil, err
	}
	if err := cfg.MetricsTLSConfig.Valid(); err != nil {
		return nil, err
	}

	op.rand = rand.New(rand.NewSource(clock.Now().Unix()))

	configOverrides := &clientcmd.ConfigOverrides{}
	var clientConfig clientcmd.ClientConfig
	if cfg.Kubeconfig == "" {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		clientConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	} else {
		apiCfg, err := clientcmd.LoadFromFile(cfg.Kubeconfig)
		if err != nil {
			return nil, err
		}
		clientConfig = clientcmd.NewDefaultClientConfig(*apiCfg, configOverrides)
	}

	var err error
	op.kubeConfig, err = clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Unable to get Kubernetes client config: %v", err)
	}

	logger.Debugf("setting up Kubernetes client...")
	op.kubeClient, err = corev1.NewForConfig(op.kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Kubernetes client: %v", err)
	}

	logger.Debugf("setting up Metering client...")
	op.meteringClient, err = cbClientset.NewForConfig(op.kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Metering client: %v", err)
	}

	op.setupInformers()
	op.setupQueues()
	op.setupEventHandlers()

	logger.Debugf("configuring event listeners...")
	return op, nil
}

func (op *Reporting) Run(stopCh <-chan struct{}) error {
	var wg sync.WaitGroup
	// buffered big enough to hold the errs of each server we start.
	srvErrChan := make(chan error, 3)

	op.logger.Info("starting Metering operator")

	promServer := &http.Server{
		Addr:    ":8082",
		Handler: promhttp.Handler(),
	}
	pprofServer := newPprofServer()

	// start these servers at the beginning some pprof and metrics are
	// available before the reporting operator is ready
	op.logger.Info("starting Prometheus metrics & pprof servers")
	wg.Add(2)
	go func() {
		defer wg.Done()
		var srvErr error
		if op.cfg.MetricsTLSConfig.UseTLS {
			op.logger.Infof("Prometheus metrics server listening with TLS on 127.0.0.1:8082")
			srvErr = promServer.ListenAndServeTLS(op.cfg.MetricsTLSConfig.TLSCert, op.cfg.MetricsTLSConfig.TLSKey)
		} else {
			op.logger.Infof("Prometheus metrics server listening on 127.0.0.1:8082")
			srvErr = promServer.ListenAndServe()
		}
		op.logger.WithError(srvErr).Info("Prometheus metrics server exited")
		srvErrChan <- fmt.Errorf("Prometheus metrics server error: %v", srvErr)
	}()
	go func() {
		defer wg.Done()
		op.logger.Infof("pprof server started")
		srvErr := pprofServer.ListenAndServe()
		op.logger.WithError(srvErr).Info("pprof server exited")
		srvErrChan <- fmt.Errorf("pprof server error: %v", srvErr)
	}()

	go op.informers.Start(stopCh)

	op.logger.Infof("setting up DB connections")

	// Use errgroup to setup both hive and presto connections
	// at the sametime, waiting for both to be ready before continuing.
	// if either errors, we return the first error
	var g errgroup.Group
	g.Go(func() error {
		var err error
		op.prestoConn, err = op.newPrestoConn(stopCh)
		if err != nil {
			return err
		}
		prestoDB := db.New(op.prestoConn, op.logger, op.cfg.LogDMLQueries)
		op.prestoQueryer = presto.NewDB(prestoDB)
		return nil
	})
	g.Go(func() error {
		op.hiveQueryer = newHiveQueryer(op.logger, op.clock, op.cfg.HiveHost, op.cfg.LogDDLQueries, stopCh)
		_, err := op.hiveQueryer.getHiveConnection()
		return err
	})
	err := g.Wait()
	if err != nil {
		return err
	}

	defer op.prestoConn.Close()
	defer op.hiveQueryer.closeHiveConnection()

	transportConfig, err := op.kubeConfig.TransportConfig()
	if err != nil {
		return err
	}

	var roundTripper http.RoundTripper
	if _, err := os.Stat(serviceServingCAFile); err == nil {
		// use the service serving CA for prometheus
		transportConfig.TLS.CAFile = serviceServingCAFile
		roundTripper, err = transport.New(transportConfig)
		if err != nil {
			return err
		}
		op.logger.Infof("using %s as CA for Prometheus", serviceServingCAFile)
	}

	op.promConn, err = op.newPrometheusConn(promapi.Config{
		Address:      op.cfg.PromHost,
		RoundTripper: roundTripper,
	})
	if err != nil {
		return err
	}

	op.logger.Info("waiting for caches to sync")
	for t, synced := range op.informers.WaitForCacheSync(stopCh) {
		if !synced {
			return fmt.Errorf("cache for %s not synced in time", t)
		}
	}

	op.logger.Infof("starting HTTP server")
	listers := meteringListers{
		reports:                 op.informers.Metering().V1alpha1().Reports().Lister().Reports(op.cfg.Namespace),
		scheduledReports:        op.informers.Metering().V1alpha1().ScheduledReports().Lister().ScheduledReports(op.cfg.Namespace),
		reportGenerationQueries: op.informers.Metering().V1alpha1().ReportGenerationQueries().Lister().ReportGenerationQueries(op.cfg.Namespace),
		prestoTables:            op.informers.Metering().V1alpha1().PrestoTables().Lister().PrestoTables(op.cfg.Namespace),
	}

	apiRouter := newRouter(op.logger, op.prestoQueryer, op.rand, op.importPrometheusForTimeRange, listers)
	apiRouter.HandleFunc("/ready", op.readinessHandler)
	apiRouter.HandleFunc("/healthy", op.healthinessHandler)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: apiRouter,
	}

	// start the HTTP API server
	wg.Add(1)
	go func() {
		defer wg.Done()
		var srvErr error
		if op.cfg.APITLSConfig.UseTLS {
			op.logger.Infof("HTTP API server listening with TLS on 127.0.0.1:8080")
			srvErr = httpServer.ListenAndServeTLS(op.cfg.APITLSConfig.TLSCert, op.cfg.APITLSConfig.TLSKey)
		} else {
			op.logger.Infof("HTTP API server listening on 127.0.0.1:8080")
			srvErr = httpServer.ListenAndServe()
		}
		op.logger.WithError(srvErr).Info("HTTP API server exited")
		srvErrChan <- fmt.Errorf("HTTP API server error: %v", srvErr)
	}()

	// Poll until we can write to presto
	op.logger.Info("testing ability to write to Presto")
	err = wait.PollUntil(time.Second*5, func() (bool, error) {
		if op.testWriteToPresto(op.logger) {
			return true, nil
		}
		return false, nil
	}, stopCh)
	if err != nil {
		return err
	}
	op.logger.Info("writes to Presto are succeeding")

	op.logger.Info("basic initialization completed")
	op.setInitialized()

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(op.logger.Infof)
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: op.kubeClient.Events(op.cfg.Namespace)})
	eventRecorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: op.cfg.PodName})

	rl, err := resourcelock.New(resourcelock.ConfigMapsResourceLock,
		op.cfg.Namespace, "reporting-operator-leader-lease", op.kubeClient,
		resourcelock.ResourceLockConfig{
			Identity:      op.cfg.Hostname,
			EventRecorder: eventRecorder,
		})
	if err != nil {
		return fmt.Errorf("error creating lock %v", err)
	}

	stopWorkersCh := make(chan struct{})
	lostLeaderCh := make(chan struct{})

	leader, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: op.cfg.LeaderLeaseDuration,
		RenewDeadline: op.cfg.LeaderLeaseDuration / 2,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(leaderStopCh <-chan struct{}) {
				op.logger.Infof("became leader")
				op.logger.Info("starting Metering workers")
				op.startWorkers(wg, stopWorkersCh)
				op.logger.Infof("Metering workers started, watching for reports...")
			},
			OnStoppedLeading: func() {
				op.logger.Warn("leader election lost")
				close(lostLeaderCh)
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error creating leader elector: %v", err)
	}

	op.logger.Infof("starting leader election")
	go leader.Run()

	// wait for an shutdown signal to begin shutdown.
	// if we lose leadership or an error occurs from one of our server
	// processes exit immediately.
	select {
	case <-stopCh:
		op.logger.Info("got stop signal, shutting down Metering operator")
	case <-lostLeaderCh:
		op.logger.Warnf("lost leadership election, forcing shut down of Metering operator")
		return fmt.Errorf("lost leadership election")
	case err := <-srvErrChan:
		op.logger.WithError(err).Error("server process failed, shutting down Metering operator")
		return fmt.Errorf("server process failed, err: %v", err)
	}

	// if we stop being leader or get a shutdown signal, stop the workers
	op.logger.Infof("stopping workers and collectors")
	close(stopWorkersCh)

	// stop our running http servers
	wg.Add(3)
	go func() {
		op.logger.Infof("stopping HTTP API server")
		err := httpServer.Shutdown(context.TODO())
		if err != nil {
			op.logger.WithError(err).Warnf("got an error shutting down HTTP API server")
		}
		wg.Done()
	}()
	go func() {
		op.logger.Infof("stopping Prometheus metrics server")
		err := promServer.Shutdown(context.TODO())
		if err != nil {
			op.logger.WithError(err).Warnf("got an error shutting down Prometheus metrics server")
		}
		wg.Done()
	}()
	go func() {
		op.logger.Infof("stopping pprof server")
		err := pprofServer.Shutdown(context.TODO())
		if err != nil {
			op.logger.WithError(err).Warnf("got an error shutting down pprof server")
		}
		wg.Done()
	}()

	// shutdown queues so that they get drained, and workers can begin their
	// shutdown
	go op.queues.ShutdownQueues()

	// wait for our workers to stop
	wg.Wait()
	op.logger.Info("Metering workers and collectors stopped")
	return nil
}

func (op *Reporting) startWorkers(wg sync.WaitGroup, stopCh <-chan struct{}) {
	wg.Add(1)
	go func() {
		op.logger.Infof("starting PrestoTable worker")
		op.runPrestoTableWorker(stopCh)
		wg.Done()
		op.logger.Infof("PrestoTable worker stopped")
	}()

	// We have a lot of ReportDataSources and we need to run more workers to
	// make sure we collect data quickly
	threadiness := 4
	for i := 0; i < threadiness; i++ {
		i := i

		wg.Add(1)
		go func() {
			op.logger.Infof("starting ReportDataSource worker #%d", i)
			wait.Until(op.runReportDataSourceWorker, time.Second, stopCh)
			wg.Done()
			op.logger.Infof("ReportDataSource worker #%d stopped", i)
		}()
	}

	// Reports and ScheduledReports we want to limit the number running
	// concurrently, and ReportGenerationQueries don't need many workers, so
	// these resources get less workers.
	threadiness = 2
	for i := 0; i < threadiness; i++ {
		i := i

		wg.Add(1)
		go func() {
			op.logger.Infof("starting ReportGenerationQuery worker #%d", i)
			wait.Until(op.runReportGenerationQueryWorker, time.Second, stopCh)
			wg.Done()
			op.logger.Infof("ReportGenerationQuery worker #%d stopped", i)
		}()

		wg.Add(1)
		go func() {
			op.logger.Infof("starting Report worker #%d", i)
			wait.Until(op.runReportWorker, time.Second, stopCh)
			wg.Done()
			op.logger.Infof("Report worker #%d stopped", i)
		}()

		wg.Add(1)
		go func() {
			op.logger.Infof("starting ScheduledReport worker #%d", i)
			wait.Until(op.runScheduledReportWorker, time.Second, stopCh)
			wg.Done()
			op.logger.Infof("ScheduledReport worker #%d stopped", i)
		}()
	}
}

func (op *Reporting) setInitialized() {
	op.initializedMu.Lock()
	op.initialized = true
	op.initializedMu.Unlock()
}

func (op *Reporting) isInitialized() bool {
	op.initializedMu.Lock()
	initialized := op.initialized
	op.initializedMu.Unlock()
	return initialized
}

func (op *Reporting) getDefaultReportGracePeriod() time.Duration {
	if op.cfg.PrometheusQueryConfig.QueryInterval.Duration > op.cfg.PrometheusQueryConfig.ChunkSize.Duration {
		return op.cfg.PrometheusQueryConfig.QueryInterval.Duration
	} else {
		return op.cfg.PrometheusQueryConfig.ChunkSize.Duration
	}
}

func (op *Reporting) newPrestoConn(stopCh <-chan struct{}) (*sql.DB, error) {
	// Presto may take longer to start than reporting-operator, so keep
	// attempting to connect in a loop in case we were just started and presto
	// is still coming up.
	connStr := fmt.Sprintf("http://root@%s?catalog=hive&schema=default", op.cfg.PrestoHost)
	startTime := op.clock.Now()
	op.logger.Debugf("getting Presto connection")
	for {
		db, err := sql.Open("presto", connStr)
		if err == nil {
			return db, nil
		} else if op.clock.Since(startTime) > maxConnWaitTime {
			op.logger.Debugf("attempts timed out, failed to get Presto connection")
			return nil, fmt.Errorf("failed to connect to presto: %v", err)
		}
		op.logger.Debugf("error encountered, backing off and trying again: %v", err)
		select {
		case <-op.clock.Tick(connBackoff):
		case <-stopCh:
			return nil, fmt.Errorf("got shutdown signal, closing Presto connection")
		}
	}
}

func (op *Reporting) newPrometheusConn(promConfig promapi.Config) (prom.API, error) {
	client, err := promapi.NewClient(promConfig)
	if err != nil {
		return nil, fmt.Errorf("can't connect to prometheus: %v", err)
	}
	return prom.NewAPI(client), nil
}

type hiveQueryer struct {
	hiveHost   string
	logger     log.FieldLogger
	logQueries bool

	clock    clock.Clock
	mu       sync.Mutex
	hiveConn *hive.Connection
	stopCh   <-chan struct{}
}

func newHiveQueryer(logger log.FieldLogger, clock clock.Clock, hiveHost string, logQueries bool, stopCh <-chan struct{}) *hiveQueryer {
	return &hiveQueryer{
		clock:      clock,
		hiveHost:   hiveHost,
		logger:     logger,
		logQueries: logQueries,
	}
}

func (q *hiveQueryer) Query(query string, args ...interface{}) (*sql.Rows, error) {
	const maxRetries = 3
	for retries := 0; retries < maxRetries; retries++ {
		hiveConn, err := q.getHiveConnection()
		if err != nil {
			if err == io.EOF || isErrBrokenPipe(err) {
				q.logger.WithError(err).Debugf("error occurred while getting connection, attempting to create new connection and retry")
				q.closeHiveConnection()
				continue
			}
			// We don't close the connection here because we got an error while
			// getting it
			return nil, err
		}
		rows, err := hiveConn.Query(query)
		if err != nil {
			if err == io.EOF || isErrBrokenPipe(err) {
				q.logger.WithError(err).Debugf("error occurred while making query, attempting to create new connection and retry")
				q.closeHiveConnection()
				continue
			}
			// We don't close the connection here because we got a good
			// connection, and made the query, but the query itself had an
			// error.
			return nil, err
		}
		return rows, nil
	}

	// We've tries 3 times, so close any connection and return an error
	q.closeHiveConnection()
	return nil, fmt.Errorf("unable to create new hive connection after existing hive connection closed")
}

func (q *hiveQueryer) getHiveConnection() (*hive.Connection, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	var err error
	if q.hiveConn == nil {
		q.hiveConn, err = q.newHiveConn()
	}
	return q.hiveConn, err
}

func (q *hiveQueryer) closeHiveConnection() {
	q.mu.Lock()
	if q.hiveConn != nil {
		q.hiveConn.Close()
	}
	// Discard our connection so we create a new one in getHiveConnection
	q.hiveConn = nil
	q.mu.Unlock()
}

func (q *hiveQueryer) newHiveConn() (*hive.Connection, error) {
	// Hive may take longer to start than reporting-operator, so keep
	// attempting to connect in a loop in case we were just started and hive is
	// still coming up.
	startTime := q.clock.Now()
	q.logger.Debugf("getting hive connection")
	for {
		select {
		case <-q.stopCh:
			// check stopCh once before connecting in case the last select loop
			// was on a tick and we got a cancellation since then
			return nil, fmt.Errorf("got shutdown signal, closing hive connection")
		default:
			// try connecting again
		}
		hive, err := hive.Connect(q.hiveHost)
		if err == nil {
			hive.SetLogQueries(q.logQueries)
			return hive, nil
		} else if q.clock.Since(startTime) > maxConnWaitTime {
			q.logger.WithError(err).Error("attempts timed out, failed to get hive connection")
			return nil, err
		}
		q.logger.WithError(err).Debugf("error encountered when connecting to hive, backing off and trying again")
		select {
		case <-q.clock.Tick(connBackoff):
		case <-q.stopCh:
			return nil, fmt.Errorf("got shutdown signal, closing hive connection")
		}
	}
}

func isErrBrokenPipe(err error) bool {
	if netErr, ok := err.(*net.OpError); ok {
		return netErr.Err == syscall.EPIPE
	}
	return false
}
