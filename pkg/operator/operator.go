package operator

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/prestodb/presto-go-client/presto"
	promapi "github.com/prometheus/client_golang/api"
	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/taozle/go-hive-driver"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/apimachinery/pkg/util/wait"

	meteringv1scheme "github.com/kube-reporting/metering-operator/pkg/generated/clientset/versioned/scheme"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	coordinatorv1 "k8s.io/client-go/kubernetes/typed/coordination/v1"
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

	metering "github.com/kube-reporting/metering-operator/pkg/apis/metering/v1"
	"github.com/kube-reporting/metering-operator/pkg/db"
	cbClientset "github.com/kube-reporting/metering-operator/pkg/generated/clientset/versioned"
	factory "github.com/kube-reporting/metering-operator/pkg/generated/informers/externalversions"
	listers "github.com/kube-reporting/metering-operator/pkg/generated/listers/metering/v1"
	"github.com/kube-reporting/metering-operator/pkg/operator/prestostore"
	"github.com/kube-reporting/metering-operator/pkg/operator/reporting"
	_ "github.com/kube-reporting/metering-operator/pkg/util/reflector/prometheus" // for prometheus metric registration
	_ "github.com/kube-reporting/metering-operator/pkg/util/workqueue/prometheus" // for prometheus metric registration
)

const (
	defaultResyncPeriod = time.Minute * 15
	prestoUsername      = "reporting-operator"

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
	Address         string
	SkipTLSVerify   bool
	BearerToken     string
	BearerTokenFile string
	CAFile          string
}

type Config struct {
	Hostname     string
	OwnNamespace string

	AllNamespaces    bool
	TargetNamespaces []string

	Kubeconfig string

	APIListen     string
	MetricsListen string
	PprofListen   string

	HiveHost                  string
	HiveUseTLS                bool
	HiveCAFile                string
	HiveTLSInsecureSkipVerify bool

	HiveUseClientCertAuth bool
	HiveClientCertFile    string
	HiveClientKeyFile     string

	PrestoHost                  string
	PrestoUseTLS                bool
	PrestoCAFile                string
	PrestoTLSInsecureSkipVerify bool

	PrestoUseClientCertAuth bool
	PrestoClientCertFile    string
	PrestoClientKeyFile     string

	PrestoMaxQueryLength int

	DisablePrometheusMetricsImporter bool
	EnableFinalizers                 bool

	LogDMLQueries bool
	LogDDLQueries bool

	PrometheusQueryConfig                         metering.PrometheusQueryConfig
	PrometheusDataSourceMaxQueryRangeDuration     time.Duration
	PrometheusDataSourceMaxBackfillImportDuration time.Duration
	PrometheusDataSourceGlobalImportFromTime      *time.Time

	ProxyTrustedCABundle string

	LeaderLeaseDuration time.Duration

	APITLSConfig     TLSConfig
	MetricsTLSConfig TLSConfig
	PrometheusConfig PrometheusConfig
}

type Reporting struct {
	cfg        Config
	kubeConfig *rest.Config

	kubeClient        corev1.CoreV1Interface
	coordinatorClient coordinatorv1.CoordinationV1Interface
	meteringClient    cbClientset.Interface
	eventRecorder     record.EventRecorder

	informerFactory factory.SharedInformerFactory

	prestoTableLister      listers.PrestoTableLister
	hiveTableLister        listers.HiveTableLister
	reportDataSourceLister listers.ReportDataSourceLister
	reportQueryLister      listers.ReportQueryLister
	reportLister           listers.ReportLister
	storageLocationLister  listers.StorageLocationLister

	queueList             []workqueue.RateLimitingInterface
	reportQueue           workqueue.RateLimitingInterface
	reportDataSourceQueue workqueue.RateLimitingInterface
	reportQueryQueue      workqueue.RateLimitingInterface
	prestoTableQueue      workqueue.RateLimitingInterface
	hiveTableQueue        workqueue.RateLimitingInterface
	storageLocationQueue  workqueue.RateLimitingInterface

	reportResultsRepo     prestostore.ReportResultsRepo
	prometheusMetricsRepo prestostore.PrometheusMetricsRepo
	reportGenerator       reporting.ReportGenerator
	dependencyResolver    DependencyResolver

	prestoTableManager   reporting.PrestoTableManager
	hiveDatabaseManager  reporting.HiveDatabaseManager
	hiveTableManager     reporting.HiveTableManager
	hivePartitionManager reporting.HivePartitionManager

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

	if cfg.AllNamespaces {
		logger.Infof("watching all namespaces for metering.openshift.io resources")
	} else {
		logger.Infof("watching namespaces %q for metering.openshift.io resources", cfg.TargetNamespaces)
	}

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

	logger.Debugf("setting up Kubernetes coordinator client...")
	coordinatorClient, err := coordinatorv1.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Kubernetes client: %v", err)
	}

	logger.Debugf("setting up Metering client...")
	meteringClient, err := cbClientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Metering client: %v", err)
	}

	var informerNamespace string
	if cfg.AllNamespaces {
		informerNamespace = metav1.NamespaceAll
	} else if len(cfg.TargetNamespaces) == 1 {
		informerNamespace = cfg.TargetNamespaces[0]
	} else if len(cfg.TargetNamespaces) > 1 && !cfg.AllNamespaces {
		return nil, fmt.Errorf("must set --all-namespaces if more than one namespace is passed to --target-namespaces")
	} else {
		informerNamespace = cfg.OwnNamespace
	}

	clock := clock.RealClock{}
	rand := rand.New(rand.NewSource(clock.Now().Unix()))

	op := newReportingOperator(logger, clock, rand, cfg, kubeConfig, kubeClient, coordinatorClient, meteringClient, informerNamespace)

	return op, nil
}

func newReportingOperator(
	logger log.FieldLogger,
	clock clock.Clock,
	rand *rand.Rand,
	cfg Config,
	kubeConfig *rest.Config,
	kubeClient corev1.CoreV1Interface,
	coordinatorClient coordinatorv1.CoordinationV1Interface,
	meteringClient cbClientset.Interface,
	informerNamespace string,
) *Reporting {
	informerFactory := factory.NewSharedInformerFactoryWithOptions(meteringClient, defaultResyncPeriod, factory.WithNamespace(informerNamespace), factory.WithTweakListOptions(nil))

	prestoTableInformer := informerFactory.Metering().V1().PrestoTables()
	hiveTableInformer := informerFactory.Metering().V1().HiveTables()
	reportDataSourceInformer := informerFactory.Metering().V1().ReportDataSources()
	reportQueryInformer := informerFactory.Metering().V1().ReportQueries()
	reportInformer := informerFactory.Metering().V1().Reports()
	storageLocationInformer := informerFactory.Metering().V1().StorageLocations()

	reportQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "reports")
	reportDataSourceQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "reportdatasources")
	reportQueryQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "reportqueries")
	prestoTableQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "prestotables")
	hiveTableQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "hivetables")
	storageLocationQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "storagelocation")

	queueList := []workqueue.RateLimitingInterface{
		reportQueue,
		reportDataSourceQueue,
		reportQueryQueue,
		prestoTableQueue,
		hiveTableQueue,
		storageLocationQueue,
	}

	depResolver := reporting.NewDependencyResolver(
		reporting.NewReportQueryListerGetter(reportQueryInformer.Lister()),
		reporting.NewReportDataSourceListerGetter(reportDataSourceInformer.Lister()),
		reporting.NewReportListerGetter(reportInformer.Lister()),
	)

	logger.Infof("setting up event broadcasters")
	utilruntime.Must(meteringv1scheme.AddToScheme(scheme.Scheme))
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(logger.Infof)
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: kubeClient.Events(cfg.OwnNamespace)})
	eventRecorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: cfg.Hostname})

	op := &Reporting{
		logger:            logger,
		cfg:               cfg,
		kubeConfig:        kubeConfig,
		kubeClient:        kubeClient,
		coordinatorClient: coordinatorClient,
		meteringClient:    meteringClient,
		eventRecorder:     eventRecorder,

		informerFactory: informerFactory,

		prestoTableLister:      prestoTableInformer.Lister(),
		hiveTableLister:        hiveTableInformer.Lister(),
		reportDataSourceLister: reportDataSourceInformer.Lister(),
		reportQueryLister:      reportQueryInformer.Lister(),
		reportLister:           reportInformer.Lister(),
		storageLocationLister:  storageLocationInformer.Lister(),

		dependencyResolver: depResolver,

		queueList:             queueList,
		reportQueue:           reportQueue,
		reportDataSourceQueue: reportDataSourceQueue,
		reportQueryQueue:      reportQueryQueue,
		prestoTableQueue:      prestoTableQueue,
		hiveTableQueue:        hiveTableQueue,
		storageLocationQueue:  storageLocationQueue,

		rand:      rand,
		clock:     clock,
		importers: make(map[string]*prestostore.PrometheusImporter),
	}

	op.logger.Info("setting the informers")
	// all eventHandlers are wrapped in an
	// inTargetNamespaceResourceEventHandler which verifies the resources
	// passed to the eventHandler functions have a metadata.namespace contained
	// in the list of TargetNamespaces, and if so, runs the eventHandler func.
	// If TargetNamespaces is empty, it will no-op and return the original
	// eventHandler.
	reportInformer.Informer().AddEventHandler(newInTargetNamespaceEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addReport,
		UpdateFunc: op.updateReport,
		DeleteFunc: op.deleteReport,
	}, op.cfg.TargetNamespaces))

	reportDataSourceInformer.Informer().AddEventHandler(newInTargetNamespaceEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addReportDataSource,
		UpdateFunc: op.updateReportDataSource,
		DeleteFunc: op.deleteReportDataSource,
	}, op.cfg.TargetNamespaces))

	reportQueryInformer.Informer().AddEventHandler(newInTargetNamespaceEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addReportQuery,
		UpdateFunc: op.updateReportQuery,
	}, op.cfg.TargetNamespaces))

	prestoTableInformer.Informer().AddEventHandler(newInTargetNamespaceEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addPrestoTable,
		UpdateFunc: op.updatePrestoTable,
		DeleteFunc: op.deletePrestoTable,
	}, op.cfg.TargetNamespaces))

	hiveTableInformer.Informer().AddEventHandler(newInTargetNamespaceEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addHiveTable,
		UpdateFunc: op.updateHiveTable,
		DeleteFunc: op.deleteHiveTable,
	}, op.cfg.TargetNamespaces))

	storageLocationInformer.Informer().AddEventHandler(newInTargetNamespaceEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.addStorageLocation,
		UpdateFunc: op.updateStorageLocation,
		DeleteFunc: op.deleteStorageLocation,
	}, op.cfg.TargetNamespaces))

	return op
}

func (op *Reporting) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	// buffered big enough to hold the errs of each server we start.
	srvErrChan := make(chan error, 3)

	op.logger.Info("starting the Metering operator")

	promServer := &http.Server{
		Addr:    op.cfg.MetricsListen,
		Handler: promhttp.Handler(),
	}
	pprofServer := newPprofServer(op.cfg.PprofListen)

	// start these servers at the beginning some pprof and metrics are
	// available before the reporting operator is ready
	op.logger.Info("starting the Prometheus metrics & pprof servers")
	wg.Add(2)
	go func() {
		defer wg.Done()
		var srvErr error
		if op.cfg.MetricsTLSConfig.UseTLS {
			op.logger.Infof("Prometheus metrics server listening with TLS on %s", op.cfg.MetricsListen)
			srvErr = promServer.ListenAndServeTLS(op.cfg.MetricsTLSConfig.TLSCert, op.cfg.MetricsTLSConfig.TLSKey)
		} else {
			op.logger.Infof("Prometheus metrics server listening on %s", op.cfg.MetricsListen)
			srvErr = promServer.ListenAndServe()
		}
		op.logger.WithError(srvErr).Info("Prometheus metrics server exited")
		srvErrChan <- fmt.Errorf("Prometheus metrics server error: %v", srvErr)
	}()
	go func() {
		defer wg.Done()
		op.logger.Infof("pprof server listening on %s", op.cfg.PprofListen)
		srvErr := pprofServer.ListenAndServe()
		op.logger.WithError(srvErr).Info("pprof server exited")
		srvErrChan <- fmt.Errorf("pprof server error: %v", srvErr)
	}()

	op.logger.Infof("setting up Hive and Presto database clients")
	var hiveDialer hive.Dialer = hive.DialWrapper{}
	if op.cfg.HiveUseTLS {
		rootCert, err := ioutil.ReadFile(op.cfg.HiveCAFile)
		if err != nil {
			return fmt.Errorf("error loading Hive CA File: %s", err)
		}

		rootCertPool := x509.NewCertPool()
		rootCertPool.AppendCertsFromPEM(rootCert)

		hiveTLSConfig := &tls.Config{
			RootCAs:            rootCertPool,
			InsecureSkipVerify: op.cfg.HiveTLSInsecureSkipVerify,
		}

		if op.cfg.HiveUseClientCertAuth {
			clientCert, err := tls.LoadX509KeyPair(op.cfg.HiveClientCertFile, op.cfg.HiveClientKeyFile)
			if err != nil {
				return fmt.Errorf("error loading Hive client cert/key: %v", err)
			}

			// mutate the necessary structure fields in hiveTLSConfig to work with client certificates
			hiveTLSConfig.Certificates = []tls.Certificate{clientCert}
		}

		hiveDialer = hive.TLSDialer{
			Config: hiveTLSConfig,
		}
	}
	hiveDB, err := hive.NewConnectorWithDialer(hiveDialer, fmt.Sprintf("hive://%s?batch=500", op.cfg.HiveHost))
	if err != nil {
		return err
	}

	hiveQueryer := sql.OpenDB(hiveDB)
	hiveQueryer.SetConnMaxLifetime(time.Minute)
	hiveQueryer.SetMaxOpenConns(2)
	hiveQueryer.SetMaxIdleConns(2)
	defer hiveQueryer.Close()

	// start building up the presto query endpoint, default value is http
	prestoURL, err := url.Parse(fmt.Sprintf("http://%s@%s?", prestoUsername, op.cfg.PrestoHost))
	if err != nil {
		return err
	}

	val := prestoURL.Query()
	val.Set("catalog", "hive")
	val.Set("schema", "default")

	// check if the PrestoUseTLS flag is set to true
	if op.cfg.PrestoUseTLS {
		rootCert, err := ioutil.ReadFile(op.cfg.PrestoCAFile)
		if err != nil {
			return fmt.Errorf("presto: Error loading SSL Cert File: %v", err)
		}

		rootCertPool := x509.NewCertPool()
		rootCertPool.AppendCertsFromPEM(rootCert)

		prestoTLSConfig := &tls.Config{
			RootCAs:            rootCertPool,
			InsecureSkipVerify: op.cfg.PrestoTLSInsecureSkipVerify,
		}

		if op.cfg.PrestoUseClientCertAuth {
			clientCert, err := tls.LoadX509KeyPair(op.cfg.PrestoClientCertFile, op.cfg.PrestoClientKeyFile)
			if err != nil {
				return fmt.Errorf("presto: Error loading SSL Client cert/key file: %v", err)
			}
			// mutate the necessary structure fields in prestoTLSConfig to work with client certificates
			prestoTLSConfig.Certificates = []tls.Certificate{clientCert}
		}

		// build up the http client structure to use the correct certificates when TLS/auth is enabled/disabled
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: prestoTLSConfig,
			},
		}

		prestoURL.Scheme = "https"
		val.Set("custom_client", "httpClient")
		err = presto.RegisterCustomClient("httpClient", httpClient)
		if err != nil {
			return fmt.Errorf("presto: Failed to register a custom HTTP client: %v", err)
		}
	}

	prestoURL.RawQuery = val.Encode()
	prestoQueryer, err := sql.Open("presto", prestoURL.String())
	if err != nil {
		return err
	}
	defer prestoQueryer.Close()

	op.promConn, err = op.newPrometheusConnFromURL(op.cfg.PrometheusConfig.Address)
	if err != nil {
		return err
	}

	var prestoQueryBufferPool *sync.Pool
	if op.cfg.PrestoMaxQueryLength > 0 {
		bufferPool := prestostore.NewBufferPool(op.cfg.PrestoMaxQueryLength)
		prestoQueryBufferPool = &bufferPool
	}

	loggingDMLPrestoQueryer := db.NewLoggingQueryer(prestoQueryer, op.logger, op.cfg.LogDMLQueries)
	loggingDDLPrestoQueryer := db.NewLoggingQueryer(prestoQueryer, op.logger, op.cfg.LogDDLQueries)
	loggingDDLHiveQueryer := db.NewLoggingExecer(hiveQueryer, op.logger, op.cfg.LogDDLQueries)

	op.reportResultsRepo = prestostore.NewReportResultsRepo(loggingDMLPrestoQueryer)
	op.reportGenerator = reporting.NewReportGenerator(op.logger, op.reportResultsRepo)
	op.prometheusMetricsRepo = prestostore.NewPrometheusMetricsRepo(loggingDMLPrestoQueryer, prestoQueryBufferPool)

	prestoTableManager := reporting.NewPrestoTableManager(loggingDDLPrestoQueryer)
	hiveManager := reporting.NewHiveManager(loggingDDLHiveQueryer)

	op.prestoTableManager = prestoTableManager
	op.hiveTableManager = hiveManager
	op.hiveDatabaseManager = hiveManager
	op.hivePartitionManager = hiveManager

	prestoHealthChecker := reporting.NewPrestoHealthChecker(op.logger, prestoQueryer, hiveManager, "", "operator_health_check")
	op.testReadFromPrestoFunc = func() bool {
		return prestoHealthChecker.TestReadFromPrestoSingleFlight()
	}

	op.logger.Infof("starting HTTP server")
	apiRouter := newRouter(
		op.logger, op.rand, op.prometheusMetricsRepo, op.reportResultsRepo, op.dependencyResolver, op.importPrometheusForTimeRange,
		op.reportLister, op.reportDataSourceLister, op.reportQueryLister, op.prestoTableLister,
	)
	apiRouter.HandleFunc("/ready", op.readinessHandler)
	apiRouter.HandleFunc("/healthy", op.readinessHandler)

	httpServer := &http.Server{
		Addr:    op.cfg.APIListen,
		Handler: apiRouter,
	}

	// start the HTTP API server
	wg.Add(1)
	go func() {
		defer wg.Done()
		var srvErr error
		if op.cfg.APITLSConfig.UseTLS {
			op.logger.Infof("HTTP API server listening with TLS on %s", op.cfg.APIListen)
			srvErr = httpServer.ListenAndServeTLS(op.cfg.APITLSConfig.TLSCert, op.cfg.APITLSConfig.TLSKey)
		} else {
			op.logger.Infof("HTTP API server listening on %s", op.cfg.APIListen)
			srvErr = httpServer.ListenAndServe()
		}
		op.logger.WithError(srvErr).Info("HTTP API server exited")
		srvErrChan <- fmt.Errorf("HTTP API server error: %v", srvErr)
	}()

	// Poll until we can read from presto
	op.logger.Info("testing ability to read to Presto")
	err = wait.PollUntil(time.Second*5, func() (bool, error) {
		return op.testReadFromPrestoFunc(), nil
	}, ctx.Done())
	if err != nil {
		return err
	}
	op.logger.Info("writes to Presto are succeeding")

	// In the case where the Prometheus metrics importer configuration has not
	// been disabled, try to fire off a test Prometheus query against the `up`
	// metric. This should weed out any invalid Prometheus configurations, e.g.
	// the service account token or CA passed to the reporting-operator is invalid,
	// and we should complain about that, but not require this as part of the
	// reporting-operator health checks. Note: in the case where we encounter
	// something like a timeout error while polling, just log the message
	// indicating there was a problem and ignore the error.
	if !op.cfg.DisablePrometheusMetricsImporter {
		op.logger.Info("testing the ability to query from Prometheus")

		err = wait.Poll(3*time.Second, 15*time.Second, func() (bool, error) {
			_, err := op.promConn.Query(context.TODO(), "up", time.Now())
			if err != nil {
				op.logger.Warnf("failed to succesfully query Prometheus: %v", err)
				return false, nil
			}

			op.logger.Infof("queries from Prometheus are succeeding")
			return true, nil
		})
		if err != nil {
			op.logger.Warnf("queries from Prometheus are failing: %v", err)
		}
	}

	op.logger.Info("basic initialization completed")
	op.setInitialized()

	op.logger.Info("starting the informers")
	go op.informerFactory.Start(ctx.Done())

	op.logger.Info("waiting for caches to sync")
	for t, synced := range op.informerFactory.WaitForCacheSync(ctx.Done()) {
		if !synced {
			return fmt.Errorf("cache for %s not synced in time", t)
		}
	}

	rl, err := resourcelock.New(resourcelock.ConfigMapsResourceLock, op.cfg.OwnNamespace, "reporting-operator-leader-lease", op.kubeClient, op.coordinatorClient,
		resourcelock.ResourceLockConfig{
			Identity:      op.cfg.Hostname,
			EventRecorder: op.eventRecorder,
		},
	)
	if err != nil {
		return fmt.Errorf("error creating lock %v", err)
	}

	// We use a new context instead of the parent because we want shutdown to
	// be different than leader election loss and do not want shutdown to
	// trigger the leader election lost case.
	lostLeaderCtx, leaderCancel := context.WithCancel(context.Background())
	defer leaderCancel()

	leader, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: op.cfg.LeaderLeaseDuration,
		RenewDeadline: op.cfg.LeaderLeaseDuration / 2,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				op.logger.Infof("became leader")
				op.logger.Info("starting Metering workers")
				op.startWorkers(&wg, ctx)
				op.logger.Infof("Metering workers started, watching for reports...")
			},
			OnStoppedLeading: func() {
				op.logger.Warn("leader election lost")
				leaderCancel()
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error creating leader elector: %v", err)
	}

	op.logger.Infof("starting leader election")
	go leader.Run(ctx)

	// wait for an shutdown signal to begin shutdown.
	// if we lose leadership or an error occurs from one of our server
	// processes exit immediately.
	select {
	case <-ctx.Done():
		op.logger.Info("got stop signal, shutting down Metering operator")
	case <-lostLeaderCtx.Done():
		op.logger.Warnf("lost leadership election, forcing shut down of Metering operator")
		return fmt.Errorf("lost leadership election")
	case err := <-srvErrChan:
		op.logger.WithError(err).Error("server process failed, shutting down Metering operator")
		return fmt.Errorf("server process failed, err: %v", err)
	}

	// if we stop being leader or get a shutdown signal, stop the workers
	op.logger.Infof("stopping workers and collectors")

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
	transportConfig := &transport.Config{}
	if op.cfg.PrometheusConfig.CAFile != "" {
		if _, err := os.Stat(op.cfg.PrometheusConfig.CAFile); err == nil {
			// Use the configured CA for communicating to Prometheus
			transportConfig.TLS.CAFile = op.cfg.PrometheusConfig.CAFile
			op.logger.Infof("using %s as CA for Prometheus", op.cfg.PrometheusConfig.CAFile)
		} else {
			return nil, err
		}
	} else {
		op.logger.Infof("using system CAs for Prometheus")
		transportConfig.TLS.CAData = nil
		transportConfig.TLS.CAFile = ""
	}

	if op.cfg.PrometheusConfig.SkipTLSVerify {
		transportConfig.TLS.Insecure = op.cfg.PrometheusConfig.SkipTLSVerify
		transportConfig.TLS.CAData = nil
		transportConfig.TLS.CAFile = ""
	}
	if op.cfg.PrometheusConfig.BearerToken != "" {
		transportConfig.BearerToken = op.cfg.PrometheusConfig.BearerToken
	}
	if op.cfg.PrometheusConfig.BearerTokenFile != "" {
		transportConfig.BearerTokenFile = op.cfg.PrometheusConfig.BearerTokenFile
	}

	roundTripper, err := transport.New(transportConfig)
	if err != nil {
		return nil, err
	}

	return op.newPrometheusConn(promapi.Config{
		Address:      url,
		RoundTripper: roundTripper,
	})
}

func (op *Reporting) startWorkers(wg *sync.WaitGroup, ctx context.Context) {
	stopCh := ctx.Done()

	startWorker := func(threads int, workerFunc func(id int)) {
		for i := 0; i < threads; i++ {
			i := i

			wg.Add(1)
			go func() {
				workerFunc(i)
				wg.Done()
			}()
		}
	}

	startWorker(4, func(i int) {
		op.logger.Infof("starting StorageLocation worker #%d", i)
		wait.Until(op.runStorageLocationWorker, time.Second, stopCh)
		op.logger.Infof("StorageLocation worker #%d stopped", i)
	})

	startWorker(10, func(i int) {
		op.logger.Infof("starting HiveTable worker #%d", i)
		wait.Until(op.runHiveTableWorker, time.Second, stopCh)
		op.logger.Infof("HiveTable worker #%d stopped", i)
	})

	startWorker(10, func(i int) {
		op.logger.Infof("starting PrestoTable worker")
		wait.Until(op.runPrestoTableWorker, time.Second, stopCh)
		op.logger.Infof("PrestoTable worker stopped")
	})

	startWorker(8, func(i int) {
		op.logger.Infof("starting ReportDataSource worker #%d", i)
		wait.Until(op.runReportDataSourceWorker, time.Second, stopCh)
		op.logger.Infof("ReportDataSource worker #%d stopped", i)
	})

	startWorker(4, func(i int) {
		op.logger.Infof("starting ReportQuery worker #%d", i)
		wait.Until(op.runReportQueryWorker, time.Second, stopCh)
		op.logger.Infof("ReportQuery worker #%d stopped", i)
	})

	startWorker(6, func(i int) {
		op.logger.Infof("starting Report worker #%d", i)
		wait.Until(op.runReportWorker, time.Second, stopCh)
		op.logger.Infof("Report worker #%d stopped", i)
	})
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

func (op *Reporting) newPrometheusConn(promConfig promapi.Config) (prom.API, error) {
	client, err := promapi.NewClient(promConfig)
	if err != nil {
		return nil, fmt.Errorf("can't connect to prometheus: %v", err)
	}
	return prom.NewAPI(client), nil
}

func (op *Reporting) handleExpiredReport(report *metering.Report, now time.Time) error {
	// check if the Report is being deleted already
	if report.DeletionTimestamp != nil {
		op.logger.Warnf("report was already marked for deletion")
		return nil
	}
	// check if the Report is past its retention time
	if reportExpired := isReportExpired(op.logger, report, now); !reportExpired {
		op.logger.Debugf("the %s Report in the %s namespace has not yet reached the expiration date", report.Name, report.Namespace)
		return nil
	}
	// check if the Report is used by any other Report or ReportQuery, if not delete it
	if reportIsNotInput := isReportNotUsedAsInput(report, op); reportIsNotInput {
		newDeletionTimestamp := metav1.Now()
		report.SetDeletionTimestamp(&newDeletionTimestamp)
		err := op.meteringClient.MeteringV1().Reports(report.Namespace).
			Delete(context.TODO(), report.Name, metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			op.logger.Infof("report: %s, not deleted because it was not found at time delete attempted", report.Name)
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to delete the expired report: %s. err: %s", report.Name, err)
		}
		op.logger.Infof("deleted Report: %s in the Namespace: %s because it reached the expiration time",
			report.Name, report.Namespace)
		op.eventRecorder.Event(report, v1.EventTypeNormal, "ExpiredReportHasBeenDeleted",
			fmt.Sprintf("Deleted the %s Report as the configured expiration date has passed", report.Name))
		return nil
	}
	// if here, warn about dependency before return
	op.logger.Warnf("report: %s, would be deleted because expired, but is depended on", report.Name)
	op.eventRecorder.Event(report, v1.EventTypeWarning, "ExpiredReportHasDependencies",
		fmt.Sprintf("Skipping the deletion of the %s Report as other resources are dependent on it, "+
			"despite reaching the desired expiration date.", report.Name))
	return nil
}

type DependencyResolver interface {
	ResolveDependencies(namespace string, inputDefs []metering.ReportQueryInputDefinition, inputVals []metering.ReportQueryInputValue) (*reporting.DependencyResolutionResult, error)
}
