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
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

const (
	connBackoff     = time.Second * 15
	maxConnWaitTime = time.Minute * 3
)

type Con***REMOVED***g struct {
	Namespace string

	HiveHost   string
	PrestoHost string
	LogReport  bool
}

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
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

	logger.Debugf("con***REMOVED***guring event listeners...")
	op.reportQueue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	op.reportInformer = cbInformers.NewReportInformer(op.chargebackClient, cfg.Namespace, time.Minute, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	op.reportInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				op.reportQueue.Add(key)
			}
		},
		UpdateFunc: func(old, current interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(current)
			if err == nil {
				op.reportQueue.Add(key)
			}
		},
	})

	return op, nil
}

type Chargeback struct {
	reportQueue      workqueue.RateLimitingInterface
	reportInformer   cache.SharedIndexInformer
	chargebackClient cbClientset.Interface

	hive   *hive.Connection
	presto *sql.DB

	logger log.FieldLogger

	namespace  string
	hiveHost   string
	prestoHost string
	logReport  bool
}

func (c *Chargeback) Run(stopCh <-chan struct{}) error {
	c.logger.Info("starting Chargeback operator")

	defer c.reportQueue.ShutDown()

	go c.reportInformer.Run(stopCh)
	go c.startHTTPServer()

	if !cache.WaitForCacheSync(stopCh, c.reportInformer.HasSynced) {
		return fmt.Errorf("cache for reports not synced in time")
	}

	c.logger.Infof("setting up DB connections")
	var err error
	c.hive, err = c.hiveConn()
	if err != nil {
		return err
	}
	c.presto, err = c.prestoConn()
	if err != nil {
		return err
	}

	c.logger.Infof("starting report worker")
	go wait.Until(c.runReportWorker, time.Second, stopCh)

	c.logger.Infof("chargeback successfully initialized, waiting for reports...")

	<-stopCh

	c.logger.Info("stopping Chargeback operator")
	return nil
}

func (c *Chargeback) hiveConn() (*hive.Connection, error) {
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

func (c *Chargeback) prestoConn() (*sql.DB, error) {
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
