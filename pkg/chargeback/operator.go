package chargeback

import (
	"database/sql"
	"fmt"
	"time"

	ext_client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	"github.com/coreos-inc/kube-chargeback/pkg/cron"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

type Config struct {
	Namespace string

	HiveHost   string
	PrestoHost string
}

func New(cfg Config) (*Chargeback, error) {
	op := &Chargeback{
		namespace:  cfg.Namespace,
		hiveHost:   cfg.HiveHost,
		prestoHost: cfg.PrestoHost,
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	fmt.Println("setting up extensions client...")
	if op.extension, err = ext_client.NewForConfig(config); err != nil {
		return nil, err
	}

	fmt.Println("setting up chargeback client...")
	if op.charge, err = cb.NewForConfig(config); err != nil {
		return nil, err
	}

	if op.cronOp, err = cron.New(config); err != nil {
		return nil, err
	}

	op.reportInform = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  op.charge.Reports(cfg.Namespace).List,
			WatchFunc: op.charge.Reports(cfg.Namespace).Watch,
		},
		&cb.Report{}, 3*time.Minute, cache.Indexers{},
	)

	fmt.Println("configuring event listeners")
	op.reportInform.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: op.handleAddReport,
	})

	fmt.Println("All set up!")
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
}

func (c *Chargeback) Run(stopCh <-chan struct{}) error {
	// TODO: implement polling
	time.Sleep(15 * time.Second)

	go c.reportInform.Run(stopCh)
	go c.cronOp.Run(stopCh)

	fmt.Println("running")

	<-stopCh
	return nil
}

func (c *Chargeback) hiveConn() (*hive.Connection, error) {
	hive, err := hive.Connect(c.hiveHost)
	if err != nil {
		return nil, err
	}
	return hive, nil
}

func (c *Chargeback) prestoConn() (*sql.DB, error) {
	connStr := fmt.Sprintf("presto://%s/hive/default", c.prestoHost)
	db, err := sql.Open("prestgo", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to presto: %v", err)
	}
	return db, nil
}
