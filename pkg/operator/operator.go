package operator

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	_ "github.com/prestodb/presto-go-client/presto"
	promapi "github.com/prometheus/client_golang/api"
	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
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

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/db"
	cbClientset "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned"
	factory "github.com/operator-framework/operator-metering/pkg/generated/informers/externalversions"
	listers "github.com/operator-framework/operator-metering/pkg/generated/listers/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/hive"
	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
	"github.com/operator-framework/operator-metering/pkg/presto"
	_ "github.com/operator-framework/operator-metering/pkg/util/reflector/prometheus" // for prometheus metric registration
	_ "github.com/operator-framework/operator-metering/pkg/util/workqueue/prometheus" // for prometheus metric registration
)

const (
	connBackoff         = time.Second * 15
	maxConnRetries      = 3
	defaultResyncPeriod = time.Minute * 15

	serviceServingCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt"
	prestoUsername       = "reporting-operator"

	DefaultPrometheusQueryInterval                       = time.Minute * 5  // Query Prometheus every 5 minutes
	DefaultPrometheusQueryStepSize                       = time.Minute      // Query data from Prometheus at a 60 second resolution (one data point per minute max)
	DefaultPrometheusQueryChunkSize                      = 5 * time.Minute  // the default value for how much data we will insert into Presto per Prometheus query.
	DefaultPrometheusDataSourceMaxQueryRangeDuration     = 10 * time.Minute // how much data we will query from Prometheus at once
	DefaultPrometheusDataSourceMaxBackfillImportDuration = 2 * time.Hour    // how far we will query for backlogged data.
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

type PrometheusConfig struct {
	Address       string
	SkipTLSVerify bool
	BearerToken   string
}

type Config struct {
	Hostname   string
	Namespace  string
	Kubeconfig string

	HiveHost                string
	PrestoHost              string
	DisablePromsum          bool
	DisableWriteHealthCheck bool
	EnableFinalizers        bool

	PrestoMaxQueryLength int

	LogDMLQueries bool
	LogDDLQueries bool

	PrometheusQueryConfig                         cbTypes.PrometheusQueryConfig
	PrometheusDataSourceMaxQueryRangeDuration     time.Duration
	PrometheusDataSourceMaxBackfillImportDuration time.Duration
	PrometheusDataSourceGlobalImportFromTime      *time.Time

	LeaderLeaseDuration time.Duration

	APITLSConfig     TLSConfig
	MetricsTLSConfig TLSConfig
	PrometheusConfig PrometheusConfig
}

type Reporting struct {
	cfg        Config
	kubeConfig *rest.Config

	meteringClient cbClientset.Interface
	kubeClient     corev1.CoreV1Interface

	informerFactory factory.SharedInformerFactory

	prestoTableLister           listers.PrestoTableLister
	reportDataSourceLister      listers.ReportDataSourceLister
	reportGenerationQueryLister listers.ReportGenerationQueryLister
	reportPrometheusQueryLister listers.ReportPrometheusQueryLister
	reportLister       listers.ReportLister
	storageLocationLister       listers.StorageLocationLister

	queueList                  []workqueue.RateLimitingInterface
	reportQueue       workqueue.RateLimitingInterface
	reportDataSourceQueue      workqueue.RateLimitingInterface
	reportGenerationQueryQueue workqueue.RateLimitingInterface
	prestoTableQueue           workqueue.RateLimitingInterface

	reportResultsRepo     prestostore.ReportResultsRepo
	prometheusMetricsRepo prestostore.PrometheusMetricsRepo
	reportGenerator       reporting.ReportGenerator

	prestoViewCreator        PrestoViewCreator
	tableManager             reporting.TableManager
	awsTablePartitionManager reporting.AWSTablePartitionManager

	testWriteToPrestoFunc  func() bool
	testReadFromPrestoFunc func() bool

	promConn prom.API

	clock clock.Clock
	rand  *rand.Rand

	logger log.FieldLogger

	initializedMu sync.Mutex
	initialized   bool

	importersMu sync.Mutex
	importers   map[string]*prestostore.PrometheusImporter
}

func New(logger log.FieldLogger, cfg Config) (*Reporting, error) {
	if err := cfg.APITLSConfig.Valid(); err != nil {
		return nil, err
	}
	if err := cfg.MetricsTLSConfig.Valid(); err != nil {
		return nil, err
	}

	logger.Debugf("config: %s", spew.Sprintf("%+v", cfg))

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
	kubeConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Unable to get Kubernetes client config: %v", err)
	}

	logger.Debugf("setting up Kubernetes client...")
	kubeClient, err := corev1.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Kubernetes client: %v", err)
	}

	logger.Debugf("setting up Metering client...")
	meteringClient, err := cbClientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Metering client: %v", err)
	}

	clock := clock.RealClock{}
	rand := rand.New(rand.NewSource(clock.Now().Unix()))
	op := newReportingOperator(logger, clock, rand, cfg, kubeConfig, kubeClient, meteringClient)

	return op, nil
}

func newReportingOperator(
	logger log.FieldLogger,
	clock clock.Clock,
	rand *rand.Rand,
	cfg Config,
	kubeConfig *rest.Config,
	kubeClient corev1.CoreV1Interface,
	meteringClient cbClientset.Interface,
) *Reporting {

	informerFactory := factory.NewFilteredSharedInformerFactory(meteringClient, defaultResyncPeriod, cfg.Namespace, nil)

	prestoTableInformer := informerFactory.Metering().V1alpha1().PrestoTables()
	reportDataSourceInformer := informerFactory.Metering().V1alpha1().ReportDataSources()
	reportGenerationQueryInformer := informerFactory.Metering().V1alpha1().ReportGenerationQueries()
	reportPrometheusQueryInformer := informerFactory.Metering().V1alpha1().ReportPrometheusQueries()
	reportInformer := informerFactory.Metering().V1alpha1().Reports()
	storageLocationInformer := informerFactory.Metering().V1alpha1().StorageLocations()

	reportQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "reports")
	reportDataSourceQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "reportdatasources")
	reportGenerationQueryQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "reportgenerationqueries")
	prestoTableQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "prestotables")

	queueList := []workqueue.RateLimitingInterface{
		reportQueue,
		reportDataSourceQueue,
		reportGenerationQueryQueue,
		prestoTableQueue,
	}

	op := &Reporting{
		logger:         logger,
		cfg:            cfg,
		kubeConfig:     kubeConfig,
		meteringClient: meteringClient,
		kubeClient:     kubeClient,

		informerFactory: informerFactory,

		prestoTableLister:           prestoTableInformer.Lister(),
		reportDataSourceLister:      reportDataSourceInformer.Lister(),
		reportGenerationQueryLister: reportGenerationQueryInformer.Lister(),
		reportPrometheusQueryLister: reportPrometheusQueryInformer.Lister(),
		reportLister:       reportInformer.Lister(),
		storageLocationLister:       storageLocationInformer.Lister(),

		queueList:                  queueList,
		reportQueue:       reportQueue,
		reportDataSourceQueue:      reportDataSourceQueue,
		reportGenerationQueryQueue: reportGenerationQueryQueue,
		prestoTableQueue:           prestoTableQueue,

		rand:      rand,
		clock:     clock,
		importers: make(map[string]*prestostore.PrometheusImporter),
	}

	reportInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addReport,
		UpdateFunc: op.updateReport,
		DeleteFunc: op.deleteReport,
	})

	reportDataSourceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addReportDataSource,
		UpdateFunc: op.updateReportDataSource,
		DeleteFunc: op.deleteReportDataSource,
	})

	reportGenerationQueryInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addReportGenerationQuery,
		UpdateFunc: op.updateReportGenerationQuery,
	})

	prestoTableInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addPrestoTable,
		UpdateFunc: op.updatePrestoTable,
		DeleteFunc: op.deletePrestoTable,
	})

	return op
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

	go op.informerFactory.Start(stopCh)

	shutdownCtx, cancel := context.WithCancel(context.Background())
	// wait for stopChn to be closed, then cancel our context
	go func() {
		<-stopCh
		cancel()
	}()

	op.logger.Infof("setting up DB connections")

	var (
		prestoQueryer db.Queryer
		hiveQueryer   db.Queryer
	)

	// Use errgroup to setup both hive and presto connections
	// at the sametime, waiting for both to be ready before continuing.
	// if either errors, we return the first error
	var g errgroup.Group
	g.Go(func() error {
		var err error
		connStr := fmt.Sprintf("http://%s@%s?catalog=hive&schema=default", prestoUsername, op.cfg.PrestoHost)
		prestoConn, err := presto.NewPrestoConnWithRetry(shutdownCtx, op.logger, connStr, connBackoff, maxConnRetries)
		if err != nil {
			return err
		}
		prestoQueryer = db.NewLoggingQueryer(prestoConn, op.logger, op.cfg.LogDMLQueries)
		return nil
	})
	g.Go(func() error {
		var err error
		reconnectingHiveQueryer := hive.NewReconnectingQueryer(shutdownCtx, op.logger, op.cfg.HiveHost, connBackoff, maxConnRetries)
		if err != nil {
			return err
		}
		hiveQueryer = db.NewLoggingQueryer(reconnectingHiveQueryer, op.logger, op.cfg.LogDDLQueries)
		return nil
	})
	err := g.Wait()
	if err != nil {
		return err
	}

	defer prestoQueryer.Close()
	defer hiveQueryer.Close()

	op.promConn, err = op.newPrometheusConnFromURL(op.cfg.PrometheusConfig.Address)
	if err != nil {
		return err
	}

	op.logger.Info("waiting for caches to sync")
	for t, synced := range op.informerFactory.WaitForCacheSync(stopCh) {
		if !synced {
			return fmt.Errorf("cache for %s not synced in time", t)
		}
	}

	var prestoQueryBufferPool *sync.Pool
	if op.cfg.PrestoMaxQueryLength > 0 {
		bufferPool := prestostore.NewBufferPool(op.cfg.PrestoMaxQueryLength)
		prestoQueryBufferPool = &bufferPool
	}
	op.reportResultsRepo = prestostore.NewReportResultsRepo(prestoQueryer)
	op.reportGenerator = reporting.NewReportGenerator(op.logger, op.reportResultsRepo)
	op.prometheusMetricsRepo = prestostore.NewPrometheusMetricsRepo(prestoQueryer, prestoQueryBufferPool)
	op.prestoViewCreator = &prestoViewCreator{queryer: prestoQueryer}

	hiveTableManager := reporting.NewHiveTableManager(hiveQueryer)
	op.tableManager = hiveTableManager
	op.awsTablePartitionManager = hiveTableManager

	tableProperties, err := op.getHiveTableProperties(op.logger, nil, "health_check")
	if err != nil {
		return fmt.Errorf("no default storage configured, unable to setup health checker: %v", err)
	}
	newTableProperties, err := addTableNameToLocation(*tableProperties, "metering_health_check")
	if err != nil {
		return err
	}

	prestoHealthChecker := reporting.NewPrestoHealthChecker(op.logger, prestoQueryer, hiveTableManager, newTableProperties)
	if op.cfg.DisableWriteHealthCheck {
		op.testWriteToPrestoFunc = func() bool {
			op.logger.Debugf("configured to skip checking ability to write to presto")
			return true
		}
	} else {
		op.testWriteToPrestoFunc = func() bool {
			return prestoHealthChecker.TestWriteToPrestoSingleFlight()
		}
	}
	op.testReadFromPrestoFunc = func() bool {
		return prestoHealthChecker.TestReadFromPrestoSingleFlight()
	}

	op.logger.Infof("starting HTTP server")
	apiRouter := newRouter(
		op.logger, op.rand, op.prometheusMetricsRepo, op.reportResultsRepo, op.importPrometheusForTimeRange, op.cfg.Namespace,
		op.reportLister, op.reportGenerationQueryLister, op.prestoTableLister,
	)
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

	if op.cfg.DisableWriteHealthCheck {
		op.logger.Info("configured to skip checking ability to write to presto")
	} else {
		// Poll until we can write to presto
		op.logger.Info("testing ability to write to Presto")
		err = wait.PollUntil(time.Second*5, func() (bool, error) {
			if op.testWriteToPrestoFunc() {
				return true, nil
			}
			return false, nil
		}, stopCh)
		if err != nil {
			return err
		}
		op.logger.Info("writes to Presto are succeeding")
	}

	op.logger.Info("basic initialization completed")
	op.setInitialized()

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(op.logger.Infof)
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: op.kubeClient.Events(op.cfg.Namespace)})
	eventRecorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: op.cfg.Hostname})

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
	go op.shutdownQueues()

	// wait for our workers to stop
	wg.Wait()
	op.logger.Info("Metering workers and collectors stopped")
	return nil
}

func (op *Reporting) newPrometheusConnFromURL(url string) (prom.API, error) {
	kubeTransportConfig, err := op.kubeConfig.TransportConfig()
	if err != nil {
		return nil, err
	}

	transportConfig := *kubeTransportConfig

	if _, err := os.Stat(serviceServingCAFile); err == nil {
		// use the service serving CA for prometheus
		transportConfig.TLS.CAFile = serviceServingCAFile
		op.logger.Infof("using %s as CA for Prometheus", serviceServingCAFile)
	}

	if op.cfg.PrometheusConfig.SkipTLSVerify {
		transportConfig.TLS.Insecure = op.cfg.PrometheusConfig.SkipTLSVerify
		transportConfig.TLS.CAData = nil
		transportConfig.TLS.CAFile = ""
	}
	if op.cfg.PrometheusConfig.BearerToken != "" {
		transportConfig.BearerToken = op.cfg.PrometheusConfig.BearerToken
	}

	roundTripper, err := transport.New(&transportConfig)
	if err != nil {
		return nil, err
	}

	return op.newPrometheusConn(promapi.Config{
		Address:      url,
		RoundTripper: roundTripper,
	})
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

	// Reports and Reports we want to limit the number running
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

func (op *Reporting) newPrometheusConn(promConfig promapi.Config) (prom.API, error) {
	client, err := promapi.NewClient(promConfig)
	if err != nil {
		return nil, fmt.Errorf("can't connect to prometheus: %v", err)
	}
	return prom.NewAPI(client), nil
}
