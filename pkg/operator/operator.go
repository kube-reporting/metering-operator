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
	_ "github.com/prestodb/presto-go-client/presto"
	promapi "github.com/prometheus/client_golang/api"
	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	hive "github.com/taozle/go-hive-driver"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/db"
	cbClientset "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned"
	factory "github.com/operator-framework/operator-metering/pkg/generated/informers/externalversions"
	listers "github.com/operator-framework/operator-metering/pkg/generated/listers/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
	_ "github.com/operator-framework/operator-metering/pkg/util/reflector/prometheus" // for prometheus metric registration
	_ "github.com/operator-framework/operator-metering/pkg/util/workqueue/prometheus" // for prometheus metric registration
)

const (
	connBackoff         = time.Second * 15
	maxConnRetries      = 3
	defaultResyncPeriod = time.Minute * 15
	prestoUsername      = "reporting-operator"

	DefaultPrometheusQueryInterval                       = time.Minute * 5  // Query Prometheus every 5 minutes
	DefaultPrometheusQueryStepSize                       = time.Minute      // Query data from Prometheus at a 60 second resolution (one data point per minute max)
	DefaultPrometheusQueryChunkSize                      = 5 * time.Minute  // the default value for how much data we will insert into Presto per Prometheus query.
	DefaultPrometheusDataSourceMaxQueryRangeDuration     = 10 * time.Minute // how much data we will query from Prometheus at once
	DefaultPrometheusDataSourceMaxBack***REMOVED***llImportDuration = 2 * time.Hour    // how far we will query for backlogged data.
)

type TLSCon***REMOVED***g struct {
	UseTLS  bool
	TLSCert string
	TLSKey  string
}

func (cfg *TLSCon***REMOVED***g) Valid() error {
	if cfg.UseTLS {
		if cfg.TLSCert == "" {
			return fmt.Errorf("Must set TLS certi***REMOVED***cate if TLS is enabled")
		}
		if cfg.TLSKey == "" {
			return fmt.Errorf("Must set TLS private key if TLS is enabled")
		}
	}
	return nil
}

type PrometheusCon***REMOVED***g struct {
	Address         string
	SkipTLSVerify   bool
	BearerToken     string
	BearerTokenFile string
	CAFile          string
}

type Con***REMOVED***g struct {
	Hostname     string
	OwnNamespace string

	AllNamespaces    bool
	TargetNamespaces []string

	Kubecon***REMOVED***g string

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

	PrometheusQueryCon***REMOVED***g                         metering.PrometheusQueryCon***REMOVED***g
	PrometheusDataSourceMaxQueryRangeDuration     time.Duration
	PrometheusDataSourceMaxBack***REMOVED***llImportDuration time.Duration
	PrometheusDataSourceGlobalImportFromTime      *time.Time

	LeaderLeaseDuration time.Duration

	APITLSCon***REMOVED***g     TLSCon***REMOVED***g
	MetricsTLSCon***REMOVED***g TLSCon***REMOVED***g
	PrometheusCon***REMOVED***g PrometheusCon***REMOVED***g
}

type Reporting struct {
	cfg        Con***REMOVED***g
	kubeCon***REMOVED***g *rest.Con***REMOVED***g

	meteringClient cbClientset.Interface
	kubeClient     corev1.CoreV1Interface

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

func New(logger log.FieldLogger, cfg Con***REMOVED***g) (*Reporting, error) {
	if err := cfg.APITLSCon***REMOVED***g.Valid(); err != nil {
		return nil, err
	}
	if err := cfg.MetricsTLSCon***REMOVED***g.Valid(); err != nil {
		return nil, err
	}

	logger.Debugf("con***REMOVED***g: %s", spew.Sprintf("%+v", cfg))

	if cfg.AllNamespaces {
		logger.Infof("watching all namespaces for metering.openshift.io resources")
	} ***REMOVED*** {
		logger.Infof("watching namespaces %q for metering.openshift.io resources", cfg.TargetNamespaces)
	}

	con***REMOVED***gOverrides := &clientcmd.Con***REMOVED***gOverrides{}
	var clientCon***REMOVED***g clientcmd.ClientCon***REMOVED***g
	if cfg.Kubecon***REMOVED***g == "" {
		loadingRules := clientcmd.NewDefaultClientCon***REMOVED***gLoadingRules()
		clientCon***REMOVED***g = clientcmd.NewNonInteractiveDeferredLoadingClientCon***REMOVED***g(loadingRules, con***REMOVED***gOverrides)
	} ***REMOVED*** {
		apiCfg, err := clientcmd.LoadFromFile(cfg.Kubecon***REMOVED***g)
		if err != nil {
			return nil, err
		}
		clientCon***REMOVED***g = clientcmd.NewDefaultClientCon***REMOVED***g(*apiCfg, con***REMOVED***gOverrides)
	}

	var err error
	kubeCon***REMOVED***g, err := clientCon***REMOVED***g.ClientCon***REMOVED***g()
	if err != nil {
		return nil, fmt.Errorf("Unable to get Kubernetes client con***REMOVED***g: %v", err)
	}

	logger.Debugf("setting up Kubernetes client...")
	kubeClient, err := corev1.NewForCon***REMOVED***g(kubeCon***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Kubernetes client: %v", err)
	}

	logger.Debugf("setting up Metering client...")
	meteringClient, err := cbClientset.NewForCon***REMOVED***g(kubeCon***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Metering client: %v", err)
	}

	var informerNamespace string
	if cfg.AllNamespaces {
		informerNamespace = metav1.NamespaceAll
	} ***REMOVED*** if len(cfg.TargetNamespaces) == 1 {
		informerNamespace = cfg.TargetNamespaces[0]
	} ***REMOVED*** if len(cfg.TargetNamespaces) > 1 && !cfg.AllNamespaces {
		return nil, fmt.Errorf("must set --all-namespaces if more than one namespace is passed to --target-namespaces")
	} ***REMOVED*** {
		informerNamespace = cfg.OwnNamespace
	}

	clock := clock.RealClock{}
	rand := rand.New(rand.NewSource(clock.Now().Unix()))

	op := newReportingOperator(logger, clock, rand, cfg, kubeCon***REMOVED***g, kubeClient, meteringClient, informerNamespace)

	return op, nil
}

func newReportingOperator(
	logger log.FieldLogger,
	clock clock.Clock,
	rand *rand.Rand,
	cfg Con***REMOVED***g,
	kubeCon***REMOVED***g *rest.Con***REMOVED***g,
	kubeClient corev1.CoreV1Interface,
	meteringClient cbClientset.Interface,
	informerNamespace string,
) *Reporting {

	informerFactory := factory.NewFilteredSharedInformerFactory(meteringClient, defaultResyncPeriod, informerNamespace, nil)

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

	op := &Reporting{
		logger:         logger,
		cfg:            cfg,
		kubeCon***REMOVED***g:     kubeCon***REMOVED***g,
		meteringClient: meteringClient,
		kubeClient:     kubeClient,

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

	// all eventHandlers are wrapped in an
	// inTargetNamespaceResourceEventHandler which veri***REMOVED***es the resources
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

	op.logger.Info("starting Metering operator")

	promServer := &http.Server{
		Addr:    op.cfg.MetricsListen,
		Handler: promhttp.Handler(),
	}
	pprofServer := newPprofServer(op.cfg.PprofListen)

	// start these servers at the beginning some pprof and metrics are
	// available before the reporting operator is ready
	op.logger.Info("starting Prometheus metrics & pprof servers")
	wg.Add(2)
	go func() {
		defer wg.Done()
		var srvErr error
		if op.cfg.MetricsTLSCon***REMOVED***g.UseTLS {
			op.logger.Infof("Prometheus metrics server listening with TLS on %s", op.cfg.MetricsListen)
			srvErr = promServer.ListenAndServeTLS(op.cfg.MetricsTLSCon***REMOVED***g.TLSCert, op.cfg.MetricsTLSCon***REMOVED***g.TLSKey)
		} ***REMOVED*** {
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

	op.logger.Infof("setting up DB clients")
	var hiveDialer hive.Dialer = hive.DialWrapper{}
	if op.cfg.HiveUseTLS {
		rootCert, err := ioutil.ReadFile(op.cfg.HiveCAFile)
		if err != nil {
			return fmt.Errorf("error loading Hive CA File: %s", err)
		}

		rootCertPool := x509.NewCertPool()
		rootCertPool.AppendCertsFromPEM(rootCert)

		hiveTLSCon***REMOVED***g := &tls.Con***REMOVED***g{
			RootCAs:            rootCertPool,
			InsecureSkipVerify: op.cfg.HiveTLSInsecureSkipVerify,
		}

		if op.cfg.HiveUseClientCertAuth {
			clientCert, err := tls.LoadX509KeyPair(op.cfg.HiveClientCertFile, op.cfg.HiveClientKeyFile)
			if err != nil {
				return fmt.Errorf("error loading Hive client cert/key: %v", err)
			}

			// mutate the necessary structure ***REMOVED***elds in hiveTLSCon***REMOVED***g to work with client certi***REMOVED***cates
			hiveTLSCon***REMOVED***g.Certi***REMOVED***cates = []tls.Certi***REMOVED***cate{clientCert}
		}

		hiveDialer = hive.TLSDialer{
			Con***REMOVED***g: hiveTLSCon***REMOVED***g,
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

		prestoTLSCon***REMOVED***g := &tls.Con***REMOVED***g{
			RootCAs:            rootCertPool,
			InsecureSkipVerify: op.cfg.PrestoTLSInsecureSkipVerify,
		}

		if op.cfg.PrestoUseClientCertAuth {
			clientCert, err := tls.LoadX509KeyPair(op.cfg.PrestoClientCertFile, op.cfg.PrestoClientKeyFile)
			if err != nil {
				return fmt.Errorf("presto: Error loading SSL Client cert/key ***REMOVED***le: %v", err)
			}
			// mutate the necessary structure ***REMOVED***elds in prestoTLSCon***REMOVED***g to work with client certi***REMOVED***cates
			prestoTLSCon***REMOVED***g.Certi***REMOVED***cates = []tls.Certi***REMOVED***cate{clientCert}
		}

		// build up the http client structure to use the correct certi***REMOVED***cates when TLS/auth is enabled/disabled
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientCon***REMOVED***g: prestoTLSCon***REMOVED***g,
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

	op.promConn, err = op.newPrometheusConnFromURL(op.cfg.PrometheusCon***REMOVED***g.Address)
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
		if op.cfg.APITLSCon***REMOVED***g.UseTLS {
			op.logger.Infof("HTTP API server listening with TLS on 127.0.0.1:8080")
			srvErr = httpServer.ListenAndServeTLS(op.cfg.APITLSCon***REMOVED***g.TLSCert, op.cfg.APITLSCon***REMOVED***g.TLSKey)
		} ***REMOVED*** {
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

	op.logger.Info("basic initialization completed")
	op.setInitialized()

	op.logger.Info("starting informers")
	go op.informerFactory.Start(ctx.Done())

	op.logger.Info("waiting for caches to sync")
	for t, synced := range op.informerFactory.WaitForCacheSync(ctx.Done()) {
		if !synced {
			return fmt.Errorf("cache for %s not synced in time", t)
		}
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(op.logger.Infof)
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: op.kubeClient.Events(op.cfg.OwnNamespace)})
	eventRecorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: op.cfg.Hostname})

	rl, err := resourcelock.New(resourcelock.Con***REMOVED***gMapsResourceLock,
		op.cfg.OwnNamespace, "reporting-operator-leader-lease", op.kubeClient,
		resourcelock.ResourceLockCon***REMOVED***g{
			Identity:      op.cfg.Hostname,
			EventRecorder: eventRecorder,
		})
	if err != nil {
		return fmt.Errorf("error creating lock %v", err)
	}

	// We use a new context instead of the parent because we want shutdown to
	// be different than leader election loss and do not want shutdown to
	// trigger the leader election lost case.
	lostLeaderCtx, leaderCancel := context.WithCancel(context.Background())
	defer leaderCancel()

	leader, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionCon***REMOVED***g{
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
	transportCon***REMOVED***g := &transport.Con***REMOVED***g{}
	if op.cfg.PrometheusCon***REMOVED***g.CAFile != "" {
		if _, err := os.Stat(op.cfg.PrometheusCon***REMOVED***g.CAFile); err == nil {
			// Use the con***REMOVED***gured CA for communicating to Prometheus
			transportCon***REMOVED***g.TLS.CAFile = op.cfg.PrometheusCon***REMOVED***g.CAFile
			op.logger.Infof("using %s as CA for Prometheus", op.cfg.PrometheusCon***REMOVED***g.CAFile)
		} ***REMOVED*** {
			return nil, err
		}
	} ***REMOVED*** {
		op.logger.Infof("using system CAs for Prometheus")
		transportCon***REMOVED***g.TLS.CAData = nil
		transportCon***REMOVED***g.TLS.CAFile = ""
	}

	if op.cfg.PrometheusCon***REMOVED***g.SkipTLSVerify {
		transportCon***REMOVED***g.TLS.Insecure = op.cfg.PrometheusCon***REMOVED***g.SkipTLSVerify
		transportCon***REMOVED***g.TLS.CAData = nil
		transportCon***REMOVED***g.TLS.CAFile = ""
	}
	if op.cfg.PrometheusCon***REMOVED***g.BearerToken != "" {
		transportCon***REMOVED***g.BearerToken = op.cfg.PrometheusCon***REMOVED***g.BearerToken
	}
	if op.cfg.PrometheusCon***REMOVED***g.BearerTokenFile != "" {
		transportCon***REMOVED***g.BearerTokenFile = op.cfg.PrometheusCon***REMOVED***g.BearerTokenFile
	}

	roundTripper, err := transport.New(transportCon***REMOVED***g)
	if err != nil {
		return nil, err
	}

	return op.newPrometheusConn(promapi.Con***REMOVED***g{
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

func (op *Reporting) newPrometheusConn(promCon***REMOVED***g promapi.Con***REMOVED***g) (prom.API, error) {
	client, err := promapi.NewClient(promCon***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("can't connect to prometheus: %v", err)
	}
	return prom.NewAPI(client), nil
}

type DependencyResolver interface {
	ResolveDependencies(namespace string, inputDefs []metering.ReportQueryInputDe***REMOVED***nition, inputVals []metering.ReportQueryInputValue) (*reporting.DependencyResolutionResult, error)
}
