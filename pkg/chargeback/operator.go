package chargeback

import (
	"database/sql"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/coreos-inc/kube-chargeback/pkg/db"
	promapi "github.com/prometheus/client_golang/api"
	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	cbClientset "github.com/coreos-inc/kube-chargeback/pkg/generated/clientset/versioned"
	cbInformers "github.com/coreos-inc/kube-chargeback/pkg/generated/informers/externalversions/chargeback/v1alpha1"
	cbListers "github.com/coreos-inc/kube-chargeback/pkg/generated/listers/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

const (
	connBackoff         = time.Second * 15
	maxConnWaitTime     = time.Minute * 3
	defaultResyncPeriod = time.Minute
)

type Con***REMOVED***g struct {
	Namespace string

	HiveHost       string
	PrestoHost     string
	PromHost       string
	DisablePromsum bool

	LogReport     bool
	LogDMLQueries bool
	LogDDLQueries bool

	PromsumInterval  time.Duration
	PromsumStepSize  time.Duration
	PromsumChunkSize time.Duration
}

type Chargeback struct {
	informers        informers
	chargebackClient cbClientset.Interface

	prestoConn  db.Queryer
	prestoDB    *sql.DB
	hiveQueryer *hiveQueryer
	promConn    prom.API

	logger log.FieldLogger

	initializedMu sync.Mutex
	initialized   bool

	namespace      string
	hiveHost       string
	prestoHost     string
	promHost       string
	disablePromsum bool
	logReport      bool

	prestoTablePartitionQueue chan *cbTypes.ReportDataSource

	logDMLQueries bool
	logDDLQueries bool

	promsumInterval  time.Duration
	promsumStepSize  time.Duration
	promsumChunkSize time.Duration
}

func New(logger log.FieldLogger, cfg Con***REMOVED***g) (*Chargeback, error) {
	op := &Chargeback{
		namespace:                 cfg.Namespace,
		hiveHost:                  cfg.HiveHost,
		prestoHost:                cfg.PrestoHost,
		promHost:                  cfg.PromHost,
		disablePromsum:            cfg.DisablePromsum,
		logReport:                 cfg.LogReport,
		logDDLQueries:             cfg.LogDDLQueries,
		logDMLQueries:             cfg.LogDMLQueries,
		promsumInterval:           cfg.PromsumInterval,
		promsumStepSize:           cfg.PromsumStepSize,
		promsumChunkSize:          cfg.PromsumChunkSize,
		prestoTablePartitionQueue: make(chan *cbTypes.ReportDataSource, 1),
		logger: logger,
	}
	logger.Debugf("Con***REMOVED***g: %+v", cfg)

	con***REMOVED***g, err := rest.InClusterCon***REMOVED***g()
	if err != nil {
		return nil, err
	}

	logger.Debugf("setting up chargeback client...")
	op.chargebackClient, err = cbClientset.NewForCon***REMOVED***g(con***REMOVED***g)
	if err != nil {
		logger.Fatal(err)
	}

	op.informers = setupInformers(op.chargebackClient, cfg.Namespace, defaultResyncPeriod)

	op.hiveQueryer = newHiveQueryer(cfg.HiveHost, logger, cfg.LogDDLQueries)

	logger.Debugf("con***REMOVED***guring event listeners...")
	return op, nil
}

type informers struct {
	informerList []cache.SharedIndexInformer
	queueList    []workqueue.RateLimitingInterface

	reportQueue    workqueue.RateLimitingInterface
	reportInformer cache.SharedIndexInformer
	reportLister   cbListers.ReportLister

	reportDataSourceQueue    workqueue.RateLimitingInterface
	reportDataSourceInformer cache.SharedIndexInformer
	reportDataSourceLister   cbListers.ReportDataSourceLister

	reportGenerationQueryQueue    workqueue.RateLimitingInterface
	reportGenerationQueryInformer cache.SharedIndexInformer
	reportGenerationQueryLister   cbListers.ReportGenerationQueryLister

	reportPrometheusQueryQueue    workqueue.RateLimitingInterface
	reportPrometheusQueryInformer cache.SharedIndexInformer
	reportPrometheusQueryLister   cbListers.ReportPrometheusQueryLister

	storageLocationQueue    workqueue.RateLimitingInterface
	storageLocationInformer cache.SharedIndexInformer
	storageLocationLister   cbListers.StorageLocationLister

	prestoTableQueue    workqueue.RateLimitingInterface
	prestoTableInformer cache.SharedIndexInformer
	prestoTableLister   cbListers.PrestoTableLister
}

func setupInformers(chargebackClient cbClientset.Interface, namespace string, resyncPeriod time.Duration) informers {
	reportQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	reportInformer := cbInformers.NewReportInformer(chargebackClient, namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	reportLister := cbListers.NewReportLister(reportInformer.GetIndexer())

	reportInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
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

	reportDataSourceQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	reportDataSourceInformer := cbInformers.NewReportDataSourceInformer(chargebackClient, namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	reportDataSourceLister := cbListers.NewReportDataSourceLister(reportDataSourceInformer.GetIndexer())

	reportDataSourceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
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
	})

	reportGenerationQueryQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	reportGenerationQueryInformer := cbInformers.NewReportGenerationQueryInformer(chargebackClient, namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	reportGenerationQueryLister := cbListers.NewReportGenerationQueryLister(reportGenerationQueryInformer.GetIndexer())

	reportGenerationQueryInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
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

	reportPrometheusQueryQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	reportPrometheusQueryInformer := cbInformers.NewReportPrometheusQueryInformer(chargebackClient, namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	reportPrometheusQueryLister := cbListers.NewReportPrometheusQueryLister(reportPrometheusQueryInformer.GetIndexer())

	storageLocationQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	storageLocationInformer := cbInformers.NewStorageLocationInformer(chargebackClient, namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	storageLocationLister := cbListers.NewStorageLocationLister(storageLocationInformer.GetIndexer())

	prestoTableQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	prestoTableInformer := cbInformers.NewPrestoTableInformer(chargebackClient, namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	prestoTableLister := cbListers.NewPrestoTableLister(prestoTableInformer.GetIndexer())

	return informers{
		informerList: []cache.SharedIndexInformer{
			storageLocationInformer,
			reportPrometheusQueryInformer,
			reportGenerationQueryInformer,
			reportDataSourceInformer,
			prestoTableInformer,
			reportInformer,
		},
		queueList: []workqueue.RateLimitingInterface{
			storageLocationQueue,
			reportPrometheusQueryQueue,
			reportGenerationQueryQueue,
			reportDataSourceQueue,
			prestoTableQueue,
			reportQueue,
		},

		reportQueue:    reportQueue,
		reportInformer: reportInformer,
		reportLister:   reportLister,

		reportDataSourceQueue:    reportDataSourceQueue,
		reportDataSourceInformer: reportDataSourceInformer,
		reportDataSourceLister:   reportDataSourceLister,

		reportGenerationQueryQueue:    reportGenerationQueryQueue,
		reportGenerationQueryInformer: reportGenerationQueryInformer,
		reportGenerationQueryLister:   reportGenerationQueryLister,

		reportPrometheusQueryQueue:    reportPrometheusQueryQueue,
		reportPrometheusQueryInformer: reportPrometheusQueryInformer,
		reportPrometheusQueryLister:   reportPrometheusQueryLister,

		storageLocationQueue:    storageLocationQueue,
		storageLocationInformer: storageLocationInformer,
		storageLocationLister:   storageLocationLister,

		prestoTableQueue:    prestoTableQueue,
		prestoTableInformer: prestoTableInformer,
		prestoTableLister:   prestoTableLister,
	}
}

func (inf informers) Run(stopCh <-chan struct{}) {
	for _, informer := range inf.informerList {
		go informer.Run(stopCh)
	}
}

func (inf informers) WaitForCacheSync(stopCh <-chan struct{}) bool {
	for _, informer := range inf.informerList {
		if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
			return false
		}
	}
	return true
}

func (inf informers) HasSynced() bool {
	for _, informer := range inf.informerList {
		if informer.HasSynced() {
			continue
		}
		return false
	}
	return true
}

func (inf informers) ShutdownQueues() {
	for _, queue := range inf.queueList {
		queue.ShutDown()
	}
}

func (c *Chargeback) Run(stopCh <-chan struct{}) error {
	c.logger.Info("starting Chargeback operator")

	go c.informers.Run(stopCh)

	c.logger.Infof("starting HTTP server")
	httpSrv := newServer(c, c.logger)
	go httpSrv.start()

	c.logger.Infof("setting up DB connections")

	// Use errgroup to setup both hive and presto connections
	// at the sametime, waiting for both to be ready before continuing.
	// if either errors, we return the ***REMOVED***rst error
	var g errgroup.Group
	g.Go(func() error {
		var err error
		c.prestoDB, err = c.newPrestoConn()
		if err != nil {
			return err
		}
		c.prestoConn = db.New(c.prestoDB, c.logger, c.logDMLQueries)
		return nil
	})
	g.Go(func() error {
		_, err := c.hiveQueryer.getHiveConnection()
		return err
	})
	err := g.Wait()
	if err != nil {
		return err
	}

	defer c.prestoDB.Close()
	defer c.hiveQueryer.closeHiveConnection()

	if !c.disablePromsum {
		c.promConn, err = c.newPrometheusConn(promapi.Con***REMOVED***g{
			Address: c.promHost,
		})
		if err != nil {
			return err
		}
	}

	c.logger.Info("waiting for caches to sync")
	if !c.informers.WaitForCacheSync(stopCh) {
		return fmt.Errorf("cache for reports not synced in time")
	}

	// Poll until we can write to presto
	c.logger.Info("testing ability to write to Presto")
	wait.PollUntil(time.Second*5, func() (bool, error) {
		if c.testWriteToPresto(c.logger) {
			return true, nil
		}
		return false, nil
	}, stopCh)
	c.logger.Info("writes to Presto are succeeding")

	var wg sync.WaitGroup
	c.logger.Info("starting Chargeback workers")
	c.startWorkers(wg, stopCh)

	c.logger.Infof("Chargeback successfully initialized, waiting for reports...")
	c.setInitialized()

	<-stopCh
	c.logger.Info("got stop signal, shutting down Chargeback operator")
	httpSrv.stop()
	go c.informers.ShutdownQueues()
	wg.Wait()
	c.logger.Info("Chargeback workers and collectors stopped")

	return nil
}

func (c *Chargeback) startWorkers(wg sync.WaitGroup, stopCh <-chan struct{}) {
	wg.Add(1)
	go func() {
		c.logger.Infof("starting PrestoTable worker")
		c.runPrestoTableWorker(stopCh)
		wg.Done()
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
	}

	if !c.disablePromsum {
		wg.Add(1)
		go func() {
			c.logger.Debugf("starting Promsum collector")
			c.runPromsumWorker(stopCh)
			wg.Done()
			c.logger.Debugf("Promsum collector stopped")
		}()
	}
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

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Chargeback) handleErr(logger log.FieldLogger, err error, objType string, key interface{}, queue workqueue.RateLimitingInterface) {
	if err == nil {
		queue.Forget(key)
		return
	}

	logger = logger.WithField(objType, key)

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if queue.NumRequeues(key) < 5 {
		logger.WithError(err).Errorf("Error syncing %s %q, adding back to queue", objType, key)
		queue.AddRateLimited(key)
		return
	}

	queue.Forget(key)
	logger.WithError(err).Infof("Dropping %s %q out of the queue", objType, key)
}

func (c *Chargeback) newPrestoConn() (*sql.DB, error) {
	// Presto may take longer to start than chargeback, so keep attempting to
	// connect in a loop in case we were just started and presto is still coming
	// up.
	connStr := fmt.Sprintf("presto://%s/hive/default", c.prestoHost)
	startTime := time.Now()
	c.logger.Debugf("getting presto connection")
	for {
		db, err := sql.Open("prestgo", connStr)
		if err == nil {
			return db, nil
		} ***REMOVED*** if time.Since(startTime) > maxConnWaitTime {
			c.logger.Debugf("attempts timed out, failed to get presto connection")
			return nil, fmt.Errorf("failed to connect to presto: %v", err)
		}
		c.logger.Debugf("error encountered, backing off and trying again: %v", err)
		time.Sleep(connBackoff)
	}
}

func (c *Chargeback) newPrometheusConn(promCon***REMOVED***g promapi.Con***REMOVED***g) (prom.API, error) {
	client, err := promapi.NewClient(promCon***REMOVED***g)
	if err != nil {
		return nil, fmt.Errorf("can't connect to prometheus: %v", err)
	}
	return prom.NewAPI(client), nil
}

type hiveQueryer struct {
	hiveHost   string
	logger     log.FieldLogger
	logQueries bool

	mu       sync.Mutex
	hiveConn *hive.Connection
}

func newHiveQueryer(hiveHost string, logger log.FieldLogger, logQueries bool) *hiveQueryer {
	return &hiveQueryer{
		hiveHost:   hiveHost,
		logger:     logger,
		logQueries: logQueries,
	}
}

func (q *hiveQueryer) Query(query string) error {
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
			return err
		}
		err = hiveConn.Query(query)
		if err != nil {
			if err == io.EOF || isErrBrokenPipe(err) {
				q.logger.WithError(err).Debugf("error occurred while making query, attempting to create new connection and retry")
				q.closeHiveConnection()
				continue
			}
			// We don't close the connection here because we got a good
			// connection, and made the query, but the query itself had an
			// error.
			return err
		}
		return nil
	}

	// We've tries 3 times, so close any connection and return an error
	q.closeHiveConnection()
	return fmt.Errorf("unable to create new hive connection after existing hive connection closed")
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
	startTime := time.Now()
	q.logger.Debugf("getting hive connection")
	for {
		hive, err := hive.Connect(q.hiveHost)
		if err == nil {
			hive.SetLogQueries(q.logQueries)
			return hive, nil
		} ***REMOVED*** if time.Since(startTime) > maxConnWaitTime {
			q.logger.WithError(err).Error("attempts timed out, failed to get hive connection")
			return nil, err
		}
		q.logger.WithError(err).Debugf("error encountered when connecting to hive, backing off and trying again")
		time.Sleep(connBackoff)
	}
}

func isErrBrokenPipe(err error) bool {
	if netErr, ok := err.(*net.OpError); ok {
		return netErr.Err == syscall.EPIPE
	}
	return false
}
