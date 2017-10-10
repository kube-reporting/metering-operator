package chargeback

import (
	"database/sql"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	ext_client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	cbClientset "github.com/coreos-inc/kube-chargeback/pkg/generated/clientset/versioned"
	cbInformers "github.com/coreos-inc/kube-chargeback/pkg/generated/informers/externalversions/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

const (
	connBackoff     = time.Second * 15
	maxConnWaitTime = time.Minute * 3
)

type Config struct {
	Namespace string

	HiveHost   string
	PrestoHost string
	LogReport  bool
}

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
}

func New(cfg Config) (*Chargeback, error) {
	log.Debugf("Config: %+v", cfg)

	op := &Chargeback{
		namespace:  cfg.Namespace,
		hiveHost:   cfg.HiveHost,
		prestoHost: cfg.PrestoHost,
		logReport:  cfg.LogReport,
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	log.Debugf("setting up extensions client...")
	if op.extension, err = ext_client.NewForConfig(config); err != nil {
		return nil, err
	}

	log.Debugf("setting up chargeback client...")

	op.chargebackClient, err = cbClientset.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Debugf("configuring event listeners...")

	op.reportInformer = cbInformers.NewReportInformer(op.chargebackClient, cfg.Namespace, time.Minute, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	op.reportInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: op.handleAddReport,
	})

	return op, nil
}

type Chargeback struct {
	extension *ext_client.Clientset

	reportInformer   cache.SharedIndexInformer
	chargebackClient cbClientset.Interface

	namespace  string
	hiveHost   string
	prestoHost string
	logReport  bool
}

func (c *Chargeback) Run(stopCh <-chan struct{}) error {
	// TODO: implement polling
	time.Sleep(15 * time.Second)

	go c.reportInformer.Run(stopCh)
	go c.startHTTPServer()

	if !cache.WaitForCacheSync(stopCh, c.reportInformer.HasSynced) {
		return fmt.Errorf("cache for reports not synced in time")
	}

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
		} else if time.Since(startTime) > maxConnWaitTime {
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
		} else if time.Since(startTime) > maxConnWaitTime {
			log.Debugf("attempts timed out, failed to get presto connection")
			return nil, fmt.Errorf("failed to connect to presto: %v", err)
		}
		log.Debugf("error encountered, backing off and trying again: %v", err)
		time.Sleep(connBackoff)
	}
}
