package chargeback

import (
	"database/sql"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	ext_client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1/types"
	"github.com/coreos-inc/kube-chargeback/pkg/cron"
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

func New(cfg Con***REMOVED***g) (*Chargeback, error) {
	log.Debugf("Con***REMOVED***g: %+v", cfg)

	op := &Chargeback{
		namespace:  cfg.Namespace,
		hiveHost:   cfg.HiveHost,
		prestoHost: cfg.PrestoHost,
		logReport:  cfg.LogReport,
	}
	con***REMOVED***g, err := rest.InClusterCon***REMOVED***g()
	if err != nil {
		return nil, err
	}

	log.Debugf("setting up extensions client...")
	if op.extension, err = ext_client.NewForCon***REMOVED***g(con***REMOVED***g); err != nil {
		return nil, err
	}

	log.Debugf("setting up chargeback client...")
	if op.charge, err = cb.NewForCon***REMOVED***g(con***REMOVED***g); err != nil {
		return nil, err
	}

	if op.cronOp, err = cron.New(con***REMOVED***g); err != nil {
		return nil, err
	}

	op.reportInform = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  op.charge.Reports(cfg.Namespace).List,
			WatchFunc: op.charge.Reports(cfg.Namespace).Watch,
		},
		&cbTypes.Report{}, 3*time.Minute, cache.Indexers{},
	)

	log.Debugf("con***REMOVED***guring event listeners...")
	op.reportInform.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: op.handleAddReport,
	})

	return op, nil
}

type Chargeback struct {
	extension *ext_client.Clientset
	charge    *cb.ChargebackClient

	reportInform cache.SharedIndexInformer

	cronOp *cron.Operator

	namespace  string
	hiveHost   string
	prestoHost string
	logReport  bool
}

func (c *Chargeback) Run(stopCh <-chan struct{}) error {
	// TODO: implement polling
	time.Sleep(15 * time.Second)

	go c.reportInform.Run(stopCh)
	go c.cronOp.Run(stopCh)
	go c.startHTTPServer()

	log.Infof("chargeback successfully initialized, waiting for reports...")

	<-stopCh
	return nil
}

func (c *Chargeback) hiveConn() (*hive.Connection, error) {
	// Hive may take longer to start than chargeback, so keep attempting to
	// connect in a loop in case we were just started and hive is still coming
	// up.
	startTime := time.Now()
	log.Debugf("getting hive connection")
	for {
		hive, err := hive.Connect(c.hiveHost)
		if err == nil {
			return hive, nil
		} ***REMOVED*** if time.Since(startTime) > maxConnWaitTime {
			log.Debugf("attempts timed out, failed to get hive connection")
			return nil, err
		}
		log.Debugf("error encountered, backing off and trying again: %v", err)
		time.Sleep(connBackoff)
	}
}

func (c *Chargeback) prestoConn() (*sql.DB, error) {
	// Presto may take longer to start than chargeback, so keep attempting to
	// connect in a loop in case we were just started and presto is still coming
	// up.
	connStr := fmt.Sprintf("presto://%s/hive/default", c.prestoHost)
	startTime := time.Now()
	log.Debugf("getting presto connection")
	for {
		db, err := sql.Open("prestgo", connStr)
		if err == nil {
			return db, nil
		} ***REMOVED*** if time.Since(startTime) > maxConnWaitTime {
			log.Debugf("attempts timed out, failed to get presto connection")
			return nil, fmt.Errorf("failed to connect to presto: %v", err)
		}
		log.Debugf("error encountered, backing off and trying again: %v", err)
		time.Sleep(connBackoff)
	}
}
