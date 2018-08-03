package chargeback

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
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/transport"
	"k8s.io/client-go/util/workqueue"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/db"
	cbClientset "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned"
	cbInformers "github.com/operator-framework/operator-metering/pkg/generated/informers/externalversions"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/presto"
	_ "github.com/operator-framework/operator-metering/pkg/util/workqueue/prometheus"
)

const (
	connBackoff         = time.Second * 15
	maxConnWaitTime     = time.Minute * 3
	defaultResyncPeriod = time.Minute

	serviceServingCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt"

	DefaultPrometheusQueryInterval  = time.Minute * 5
	DefaultPrometheusQueryStepSize  = time.Minute
	DefaultPrometheusQueryChunkSize = time.Minute * 5
)

type Config struct {
	PodName    string
	Hostname   string
	Namespace  string
	Kubeconfig string

	HiveHost       string
	PrestoHost     string
	PromHost       string
	DisablePromsum bool

	LogDMLQueries bool
	LogDDLQueries bool

	PrometheusQueryConfig cbTypes.PrometheusQueryConfig

	LeaderLeaseDuration time.Duration

	UseTLS  bool
	TLSCert string
	TLSKey  string
}

type Chargeback struct {
	cfg              Config
	kubeConfig       *rest.Config
	informers        cbInformers.SharedInformerFactory
	queues           queues
	chargebackClient cbClientset.Interface
	kubeClient       corev1.CoreV1Interface

	prestoConn    *sql.DB
	prestoQueryer presto.ExecQueryer
	hiveQueryer   *hiveQueryer
	promConn      prom.API

	scheduledReportRunner *scheduledReportRunner

	clock clock.Clock
	rand  *rand.Rand

	logger log.FieldLogger

	initializedMu sync.Mutex
	initialized   bool

	prestoTablePartitionQueue                    chan *cbTypes.ReportDataSource
	prometheusImporterNewDataSourceQueue         chan *cbTypes.ReportDataSource
	prometheusImporterDeletedDataSourceQueue     chan string
	prometheusImporterTriggerFromLastTimestampCh chan struct{}
	prometheusImporterTriggerForTimeRangeCh      chan prometheusImporterTimeRangeTrigger

	// ensures only at most a single testRead query is running against Presto
	// at one time
	healthCheckSingleFlight singleflight.Group
}

func New(logger log.FieldLogger, cfg Config, clock clock.Clock) (*Chargeback, error) {
	op := &Chargeback{
		cfg: cfg,
		prestoTablePartitionQueue:                    make(chan *cbTypes.ReportDataSource, 1),
		prometheusImporterNewDataSourceQueue:         make(chan *cbTypes.ReportDataSource),
		prometheusImporterDeletedDataSourceQueue:     make(chan string),
		prometheusImporterTriggerFromLastTimestampCh: make(chan struct{}),
		prometheusImporterTriggerForTimeRangeCh:      make(chan prometheusImporterTimeRangeTrigger),
		logger: logger,
		clock:  clock,
	}
	logger.Debugf("Config: %+v", cfg)

	if cfg.UseTLS {
		if cfg.TLSCert == "" {
			return nil, fmt.Errorf("Must set TLS certificate if TLS is enabled")
		}
		if cfg.TLSKey == "" {
			return nil, fmt.Errorf("Must set TLS private key if TLS is enabled")
		}
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
	op.chargebackClient, err = cbClientset.NewForConfig(op.kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Metering client: %v", err)
	}

	op.setupInformers()
	op.setupQueues()

	op.scheduledReportRunner = newScheduledReportRunner(op)

	logger.Debugf("configuring event listeners...")
	return op, nil
}

type queues struct {
	queueList                  []workqueue.RateLimitingInterface
	reportQueue                workqueue.RateLimitingInterface
	scheduledReportQueue       workqueue.RateLimitingInterface
	reportDataSourceQueue      workqueue.RateLimitingInterface
	reportGenerationQueryQueue workqueue.RateLimitingInterface
}

func (c *Chargeback) setupInformers() {
	c.informers = cbInformers.NewFilteredSharedInformerFactory(c.chargebackClient, defaultResyncPeriod, c.cfg.Namespace, nil)
	inf := c.informers.Chargeback().V1alpha1()
	// hacks to ensure these informers are created before we call
	// c.informers.Start()
	inf.PrestoTables().Informer()
	inf.StorageLocations().Informer()
	inf.ReportDataSources().Informer()
	inf.ReportGenerationQueries().Informer()
	inf.ReportPrometheusQueries().Informer()
	inf.Reports().Informer()
	inf.ScheduledReports().Informer()
}
func (c *Chargeback) setupQueues() {
	reportQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "reports")
	c.informers.Chargeback().V1alpha1().Reports().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				reportQueue.Add(key)
			}
		},
		UpdateFunc: func(old, current interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(current)
			if err == nil {
				reportQueue.Add(key)
			}
		},
	})

	scheduledReportQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "scheduledreports")
	c.informers.Chargeback().V1alpha1().ScheduledReports().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				scheduledReportQueue.Add(key)
			}
		},
		UpdateFunc: func(old, current interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(current)
			if err == nil {
				scheduledReportQueue.Add(key)
			}
		},
		DeleteFunc: c.handleScheduledReportDeleted,
	})

	reportDataSourceQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "reportdatasources")
	c.informers.Chargeback().V1alpha1().ReportDataSources().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				reportDataSourceQueue.Add(key)
			}
		},
		UpdateFunc: func(old, current interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(current)
			if err == nil {
				reportDataSourceQueue.Add(key)
			}
		},
		DeleteFunc: c.handleReportDataSourceDeleted,
	})

	reportGenerationQueryQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "reportgenerationqueries")
	c.informers.Chargeback().V1alpha1().ReportGenerationQueries().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				reportGenerationQueryQueue.Add(key)
			}
		},
		UpdateFunc: func(old, current interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(current)
			if err == nil {
				reportGenerationQueryQueue.Add(key)
			}
		},
	})
	c.queues = queues{
		queueList: []workqueue.RateLimitingInterface{
			reportQueue,
			scheduledReportQueue,
			reportDataSourceQueue,
			reportGenerationQueryQueue,
		},
		reportQueue:                reportQueue,
		scheduledReportQueue:       scheduledReportQueue,
		reportDataSourceQueue:      reportDataSourceQueue,
		reportGenerationQueryQueue: reportGenerationQueryQueue,
	}

}

func (qs queues) ShutdownQueues() {
	for _, queue := range qs.queueList {
		queue.ShutDown()
	}
}

func (c *Chargeback) Run(stopCh <-chan struct{}) error {
	var wg sync.WaitGroup
	c.logger.Info("starting Metering operator")

	go c.informers.Start(stopCh)

	c.logger.Infof("setting up DB connections")

	// Use errgroup to setup both hive and presto connections
	// at the sametime, waiting for both to be ready before continuing.
	// if either errors, we return the first error
	var g errgroup.Group
	g.Go(func() error {
		var err error
		c.prestoConn, err = c.newPrestoConn(stopCh)
		if err != nil {
			return err
		}
		prestoDB := db.New(c.prestoConn, c.logger, c.cfg.LogDMLQueries)
		c.prestoQueryer = presto.NewDB(prestoDB)
		return nil
	})
	g.Go(func() error {
		c.hiveQueryer = newHiveQueryer(c.logger, c.clock, c.cfg.HiveHost, c.cfg.LogDDLQueries, stopCh)
		_, err := c.hiveQueryer.getHiveConnection()
		return err
	})
	err := g.Wait()
	if err != nil {
		return err
	}

	defer c.prestoConn.Close()
	defer c.hiveQueryer.closeHiveConnection()

	transportConfig, err := c.kubeConfig.TransportConfig()
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
		c.logger.Infof("using %s as CA for Prometheus", serviceServingCAFile)
	}

	c.promConn, err = c.newPrometheusConn(promapi.Config{
		Address:      c.cfg.PromHost,
		RoundTripper: roundTripper,
	})
	if err != nil {
		return err
	}

	c.logger.Info("waiting for caches to sync")
	for t, synced := range c.informers.WaitForCacheSync(stopCh) {
		if !synced {
			return fmt.Errorf("cache for %s not synced in time", t)
		}
	}

	c.logger.Infof("starting HTTP server")
	listers := meteringListers{
		reports:                 c.informers.Chargeback().V1alpha1().Reports().Lister().Reports(c.cfg.Namespace),
		scheduledReports:        c.informers.Chargeback().V1alpha1().ScheduledReports().Lister().ScheduledReports(c.cfg.Namespace),
		reportGenerationQueries: c.informers.Chargeback().V1alpha1().ReportGenerationQueries().Lister().ReportGenerationQueries(c.cfg.Namespace),
		prestoTables:            c.informers.Chargeback().V1alpha1().PrestoTables().Lister().PrestoTables(c.cfg.Namespace),
	}

	apiRouter := newRouter(c.logger, c.prestoQueryer, c.rand, c.triggerPrometheusImporterForTimeRange, listers)
	apiRouter.HandleFunc("/ready", c.readinessHandler)
	apiRouter.HandleFunc("/healthy", c.healthinessHandler)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: apiRouter,
	}
	promServer := &http.Server{
		Addr:    ":8082",
		Handler: promhttp.Handler(),
	}
	pprofServer := newPprofServer()

	// buffered big enough to hold the errs of each server.
	srvErrChan := make(chan error, 3)
	// start the HTTP servers
	wg.Add(3)
	go func() {
		defer wg.Done()
		var srvErr error
		if c.cfg.UseTLS {
			c.logger.Infof("HTTP API server listening with TLS on 127.0.0.1:8080")
			srvErr = httpServer.ListenAndServeTLS(c.cfg.TLSCert, c.cfg.TLSKey)
		} else {
			c.logger.Infof("HTTP API server listening on 127.0.0.1:8080")
			srvErr = httpServer.ListenAndServe()
		}
		c.logger.WithError(srvErr).Info("HTTP API server exited")
		srvErrChan <- fmt.Errorf("HTTP API server error: %v", srvErr)
	}()
	go func() {
		defer wg.Done()
		c.logger.Infof("Prometheus metrics server started")
		srvErr := promServer.ListenAndServe()
		c.logger.WithError(srvErr).Info("Prometheus metrics server exited")
		srvErrChan <- fmt.Errorf("Prometheus metrics server error: %v", srvErr)
	}()
	go func() {
		defer wg.Done()
		c.logger.Infof("pprof server started")
		srvErr := pprofServer.ListenAndServe()
		c.logger.WithError(srvErr).Info("pprof server exited")
		srvErrChan <- fmt.Errorf("pprof server error: %v", srvErr)
	}()

	// Poll until we can write to presto
	c.logger.Info("testing ability to write to Presto")
	err = wait.PollUntil(time.Second*5, func() (bool, error) {
		if c.testWriteToPresto(c.logger) {
			return true, nil
		}
		return false, nil
	}, stopCh)
	if err != nil {
		return err
	}
	c.logger.Info("writes to Presto are succeeding")

	c.logger.Info("basic initialization completed")
	c.setInitialized()

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(c.logger.Infof)
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: c.kubeClient.Events(c.cfg.Namespace)})
	eventRecorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: c.cfg.PodName})

	rl, err := resourcelock.New(resourcelock.ConfigMapsResourceLock,
		c.cfg.Namespace, "chargeback-operator-leader-lease", c.kubeClient,
		resourcelock.ResourceLockConfig{
			Identity:      c.cfg.Hostname,
			EventRecorder: eventRecorder,
		})
	if err != nil {
		return fmt.Errorf("error creating lock %v", err)
	}

	stopWorkersCh := make(chan struct{})
	lostLeaderCh := make(chan struct{})

	leader, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: c.cfg.LeaderLeaseDuration,
		RenewDeadline: c.cfg.LeaderLeaseDuration / 2,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(leaderStopCh <-chan struct{}) {
				c.logger.Infof("became leader")
				c.logger.Info("starting Metering workers")
				c.startWorkers(wg, stopWorkersCh)
				c.logger.Infof("Metering workers started, watching for reports...")
			},
			OnStoppedLeading: func() {
				c.logger.Warn("leader election lost")
				close(lostLeaderCh)
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error creating leader elector: %v", err)
	}

	c.logger.Infof("starting leader election")
	go leader.Run()

	// wait for an shutdown signal to begin shutdown.
	// if we lose leadership or an error occurs from one of our server
	// processes exit immediately.
	select {
	case <-stopCh:
		c.logger.Info("got stop signal, shutting down Metering operator")
	case <-lostLeaderCh:
		c.logger.Warnf("lost leadership election, forcing shut down of Metering operator")
		return fmt.Errorf("lost leadership election")
	case err := <-srvErrChan:
		c.logger.WithError(err).Error("server process failed, shutting down Metering operator")
		return fmt.Errorf("server process failed, err: %v", err)
	}

	// if we stop being leader or get a shutdown signal, stop the workers
	c.logger.Infof("stopping workers and collectors")
	close(stopWorkersCh)

	// stop our running http servers
	wg.Add(3)
	go func() {
		c.logger.Infof("stopping HTTP API server")
		err := httpServer.Shutdown(context.TODO())
		if err != nil {
			c.logger.WithError(err).Warnf("got an error shutting down HTTP API server")
		}
		wg.Done()
	}()
	go func() {
		c.logger.Infof("stopping Prometheus metrics server")
		err := promServer.Shutdown(context.TODO())
		if err != nil {
			c.logger.WithError(err).Warnf("got an error shutting down Prometheus metrics server")
		}
		wg.Done()
	}()
	go func() {
		c.logger.Infof("stopping pprof server")
		err := pprofServer.Shutdown(context.TODO())
		if err != nil {
			c.logger.WithError(err).Warnf("got an error shutting down pprof server")
		}
		wg.Done()
	}()

	// shutdown queues so that they get drained, and workers can begin their
	// shutdown
	go c.queues.ShutdownQueues()

	// wait for our workers to stop
	wg.Wait()
	c.logger.Info("Metering workers and collectors stopped")
	return nil
}

func (c *Chargeback) startWorkers(wg sync.WaitGroup, stopCh <-chan struct{}) {
	wg.Add(1)
	go func() {
		c.logger.Infof("starting PrestoTable worker")
		c.runPrestoTableWorker(stopCh)
		wg.Done()
		c.logger.Infof("PrestoTable worker stopped")
	}()

	threadiness := 2
	for i := 0; i < threadiness; i++ {
		i := i

		wg.Add(1)
		go func() {
			c.logger.Infof("starting ReportDataSource worker #%d", i)
			wait.Until(c.runReportDataSourceWorker, time.Second, stopCh)
			wg.Done()
			c.logger.Infof("ReportDataSource worker #%d stopped", i)
		}()

		wg.Add(1)
		go func() {
			c.logger.Infof("starting ReportGenerationQuery worker #%d", i)
			wait.Until(c.runReportGenerationQueryWorker, time.Second, stopCh)
			wg.Done()
			c.logger.Infof("ReportGenerationQuery worker #%d stopped", i)
		}()

		wg.Add(1)
		go func() {
			c.logger.Infof("starting Report worker #%d", i)
			wait.Until(c.runReportWorker, time.Second, stopCh)
			wg.Done()
			c.logger.Infof("Report worker #%d stopped", i)
		}()

		wg.Add(1)
		go func() {
			c.logger.Infof("starting ScheduledReport worker #%d", i)
			wait.Until(c.runScheduledReportWorker, time.Second, stopCh)
			wg.Done()
			c.logger.Infof("ScheduledReport worker #%d stopped", i)
		}()
	}

	wg.Add(1)
	go func() {
		c.logger.Debugf("starting ScheduledReportRunner")
		c.scheduledReportRunner.Run(stopCh)
		wg.Done()
		c.logger.Debugf("ScheduledReportRunner stopped")
	}()

	wg.Add(1)
	go func() {
		c.logger.Debugf("starting PrometheusImport worker")
		c.runPrometheusImporterWorker(stopCh)
		wg.Done()
		c.logger.Debugf("PrometheusImport worker stopped")
	}()
}

func (c *Chargeback) setInitialized() {
	c.initializedMu.Lock()
	c.initialized = true
	c.initializedMu.Unlock()
}

func (c *Chargeback) isInitialized() bool {
	c.initializedMu.Lock()
	initialized := c.initialized
	c.initializedMu.Unlock()
	return initialized
}

// getKeyFromQueueObj tries to convert the object from the queue into a string,
// and if it isn't, it forgets the key from the queue, and logs an error.
//
// We expect strings to come off the workqueue. These are of the
// form namespace/name. We do this as the delayed nature of the
// workqueue means the items in the informer cache may actually be
// more up to date that when the item was initially put onto the
// workqueue.
func (c *Chargeback) getKeyFromQueueObj(logger log.FieldLogger, objType string, obj interface{}, queue workqueue.RateLimitingInterface) (string, bool) {
	if key, ok := obj.(string); ok {
		return key, ok
	}
	queue.Forget(obj)
	logger.WithField(objType, obj).Errorf("expected string in work queue but got %#v", obj)
	return "", false
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Chargeback) handleErr(logger log.FieldLogger, err error, objType string, obj interface{}, queue workqueue.RateLimitingInterface) {
	if err == nil {
		queue.Forget(obj)
		return
	}

	logger = logger.WithField(objType, obj)

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if queue.NumRequeues(obj) < 5 {
		logger.WithError(err).Errorf("Error syncing %s %q, adding back to queue", objType, obj)
		queue.AddRateLimited(obj)
		return
	}

	queue.Forget(obj)
	logger.WithError(err).Infof("Dropping %s %q out of the queue", objType, obj)
}

func (c *Chargeback) getDefaultReportGracePeriod() time.Duration {
	if c.cfg.PrometheusQueryConfig.QueryInterval.Duration > c.cfg.PrometheusQueryConfig.ChunkSize.Duration {
		return c.cfg.PrometheusQueryConfig.QueryInterval.Duration
	} else {
		return c.cfg.PrometheusQueryConfig.ChunkSize.Duration
	}
}

func (c *Chargeback) newPrestoConn(stopCh <-chan struct{}) (*sql.DB, error) {
	// Presto may take longer to start than chargeback, so keep attempting to
	// connect in a loop in case we were just started and presto is still coming
	// up.
	connStr := fmt.Sprintf("http://root@%s?catalog=hive&schema=default", c.cfg.PrestoHost)
	startTime := c.clock.Now()
	c.logger.Debugf("getting Presto connection")
	for {
		db, err := sql.Open("presto", connStr)
		if err == nil {
			return db, nil
		} else if c.clock.Since(startTime) > maxConnWaitTime {
			c.logger.Debugf("attempts timed out, failed to get Presto connection")
			return nil, fmt.Errorf("failed to connect to presto: %v", err)
		}
		c.logger.Debugf("error encountered, backing off and trying again: %v", err)
		select {
		case <-c.clock.Tick(connBackoff):
		case <-stopCh:
			return nil, fmt.Errorf("got shutdown signal, closing Presto connection")
		}
	}
}

func (c *Chargeback) newPrometheusConn(promConfig promapi.Config) (prom.API, error) {
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
	// Hive may take longer to start than chargeback, so keep attempting to
	// connect in a loop in case we were just started and hive is still coming
	// up.
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
