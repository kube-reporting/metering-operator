package chargeback

import (
	"database/sql"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	cbClientset "github.com/coreos-inc/kube-chargeback/pkg/generated/clientset/versioned"
	cbInformers "github.com/coreos-inc/kube-chargeback/pkg/generated/informers/externalversions/chargeback/v1alpha1"
	cbListers "github.com/coreos-inc/kube-chargeback/pkg/generated/listers/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

const (
	connBackoff         = time.Second * 15
	maxConnWaitTime     = time.Minute
	defaultResyncPeriod = time.Minute
)

type Con***REMOVED***g struct {
	Namespace string

	HiveHost   string
	PrestoHost string
	LogReport  bool
	LogQueries bool
}

type Chargeback struct {
	informers        informers
	chargebackClient cbClientset.Interface

	prestoConn  *sql.DB
	hiveQueryer *hiveQueryer

	logger log.FieldLogger

	namespace  string
	hiveHost   string
	prestoHost string
	logReport  bool
	logQueries bool
}

func New(logger log.FieldLogger, cfg Con***REMOVED***g) (*Chargeback, error) {
	op := &Chargeback{
		namespace:  cfg.Namespace,
		hiveHost:   cfg.HiveHost,
		prestoHost: cfg.PrestoHost,
		logReport:  cfg.LogReport,
		logQueries: cfg.LogQueries,
		logger:     logger,
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

	op.hiveQueryer = newHiveQueryer(cfg.HiveHost, logger, cfg.LogQueries)

	logger.Debugf("con***REMOVED***guring event listeners...")
	return op, nil
}

type informers struct {
	informerList []cache.SharedIndexInformer
	queueList    []workqueue.RateLimitingInterface

	reportQueue    workqueue.RateLimitingInterface
	reportInformer cache.SharedIndexInformer
	reportLister   cbListers.ReportLister

	reportDataStoreQueue    workqueue.RateLimitingInterface
	reportDataStoreInformer cache.SharedIndexInformer
	reportDataStoreLister   cbListers.ReportDataStoreLister

	reportGenerationQueryQueue    workqueue.RateLimitingInterface
	reportGenerationQueryInformer cache.SharedIndexInformer
	reportGenerationQueryLister   cbListers.ReportGenerationQueryLister
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

	reportDataStoreQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	reportDataStoreInformer := cbInformers.NewReportDataStoreInformer(chargebackClient, namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	reportDataStoreLister := cbListers.NewReportDataStoreLister(reportDataStoreInformer.GetIndexer())

	reportDataStoreInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				reportDataStoreQueue.Add(key)
			}
		},
		UpdateFunc: func(old, current interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(current)
			if err == nil {
				reportDataStoreQueue.Add(key)
			}
		},
	})

	reportGenerationQueryQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	reportGenerationQueryInformer := cbInformers.NewReportGenerationQueryInformer(chargebackClient, namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	reportGenerationQueryLister := cbListers.NewReportGenerationQueryLister(reportGenerationQueryInformer.GetIndexer())

	return informers{
		informerList: []cache.SharedIndexInformer{
			reportGenerationQueryInformer,
			reportDataStoreInformer,
			reportInformer,
		},
		queueList: []workqueue.RateLimitingInterface{
			reportGenerationQueryQueue,
			reportDataStoreQueue,
			reportQueue,
		},

		reportQueue:    reportQueue,
		reportInformer: reportInformer,
		reportLister:   reportLister,

		reportDataStoreQueue:    reportDataStoreQueue,
		reportDataStoreInformer: reportDataStoreInformer,
		reportDataStoreLister:   reportDataStoreLister,

		reportGenerationQueryQueue:    reportGenerationQueryQueue,
		reportGenerationQueryInformer: reportGenerationQueryInformer,
		reportGenerationQueryLister:   reportGenerationQueryLister,
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

func (inf informers) ShutdownQueues() {
	for _, queue := range inf.queueList {
		queue.ShutDown()
	}
}

func (c *Chargeback) Run(stopCh <-chan struct{}) error {
	c.logger.Info("starting Chargeback operator")

	defer c.informers.ShutdownQueues()
	go c.informers.Run(stopCh)
	go c.startHTTPServer()

	if !c.informers.WaitForCacheSync(stopCh) {
		return fmt.Errorf("cache for reports not synced in time")
	}

	c.logger.Infof("setting up DB connections")
	var err error
	c.prestoConn, err = c.newPrestoConn()
	if err != nil {
		return err
	}
	defer c.prestoConn.Close()

	_, err = c.hiveQueryer.getHiveConnection()
	if err != nil {
		return err
	}
	defer c.hiveQueryer.closeHiveConnection()

	c.logger.Infof("starting report worker")
	go wait.Until(c.runReportWorker, time.Second, stopCh)
	go wait.Until(c.runReportDataStoreWorker, time.Second, stopCh)

	c.logger.Infof("chargeback successfully initialized, waiting for reports...")

	<-stopCh

	c.logger.Info("stopping Chargeback operator")
	return nil
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

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Chargeback) handleErr(err error, objType string, key interface{}, queue workqueue.RateLimitingInterface) {
	if err == nil {
		queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if queue.NumRequeues(key) < 5 {
		c.logger.WithError(err).Errorf("Error syncing %s %q, adding back to queue", objType, key)

		queue.AddRateLimited(key)
		return
	}

	queue.Forget(key)
	c.logger.WithError(err).Infof("Dropping %s %q out of the queue", objType, key)
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
