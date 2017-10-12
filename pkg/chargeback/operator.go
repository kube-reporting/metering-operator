package chargeback

import (
	"database/sql"
	"fmt"
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
	maxConnWaitTime     = time.Minute * 3
	defaultResyncPeriod = time.Minute
)

type Con***REMOVED***g struct {
	Namespace string

	HiveHost   string
	PrestoHost string
	LogReport  bool
}

func New(logger log.FieldLogger, cfg Con***REMOVED***g) (*Chargeback, error) {
	op := &Chargeback{
		namespace:  cfg.Namespace,
		hiveHost:   cfg.HiveHost,
		prestoHost: cfg.PrestoHost,
		logReport:  cfg.LogReport,
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

type Chargeback struct {
	informers        informers
	chargebackClient cbClientset.Interface

	hiveConn   *hive.Connection
	prestoConn *sql.DB

	logger log.FieldLogger

	namespace  string
	hiveHost   string
	prestoHost string
	logReport  bool
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
	c.hiveConn, err = c.newHiveConn()
	if err != nil {
		return err
	}
	defer c.hiveConn.Close()
	c.prestoConn, err = c.newPrestoConn()
	if err != nil {
		return err
	}
	defer c.prestoConn.Close()

	c.logger.Infof("starting report worker")
	go wait.Until(c.runReportWorker, time.Second, stopCh)
	go wait.Until(c.runReportDataStoreWorker, time.Second, stopCh)

	c.logger.Infof("chargeback successfully initialized, waiting for reports...")

	<-stopCh

	c.logger.Info("stopping Chargeback operator")
	return nil
}

func (c *Chargeback) newHiveConn() (*hive.Connection, error) {
	// Hive may take longer to start than chargeback, so keep attempting to
	// connect in a loop in case we were just started and hive is still coming
	// up.
	startTime := time.Now()
	c.logger.Debugf("getting hive connection")
	for {
		hive, err := hive.Connect(c.hiveHost)
		if err == nil {
			return hive, nil
		} ***REMOVED*** if time.Since(startTime) > maxConnWaitTime {
			c.logger.Debugf("attempts timed out, failed to get hive connection")
			return nil, err
		}
		c.logger.Debugf("error encountered, backing off and trying again: %v", err)
		time.Sleep(connBackoff)
	}
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
		c.logger.WithError(err).Error("Error syncing %s %q", objType, key)

		queue.AddRateLimited(key)
		return
	}

	queue.Forget(key)
	c.logger.WithError(err).Infof("Dropping %s %q out of the queue", objType, key)
}
