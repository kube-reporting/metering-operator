package chargeback

import (
	"database/sql"
	"fmt"
	"time"

	ext_client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

type Con***REMOVED***g struct {
	HiveHost   string
	PrestoHost string
}

func New(cfg Con***REMOVED***g) (*Chargeback, error) {
	op := &Chargeback{
		hiveHost:   cfg.HiveHost,
		prestoHost: cfg.PrestoHost,
	}
	con***REMOVED***g, err := rest.InClusterCon***REMOVED***g()
	if err != nil {
		return nil, err
	}

	fmt.Println("setting up extensions client...")
	if op.extension, err = ext_client.NewForCon***REMOVED***g(con***REMOVED***g); err != nil {
		return nil, err
	}

	fmt.Println("setting up chargeback client...")
	if op.charge, err = cb.NewForCon***REMOVED***g(con***REMOVED***g); err != nil {
		return nil, err
	}

	op.reportInform = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  op.charge.Reports().List,
			WatchFunc: op.charge.Reports().Watch,
		},
		&cb.Report{}, 3*time.Minute, cache.Indexers{},
	)

	fmt.Println("con***REMOVED***guring event listeners")
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

	hiveHost   string
	prestoHost string
}

func (c *Chargeback) Run() error {
	err := c.createResources()
	if err != nil {
		panic(err)
	}

	// TODO: implement polling
	time.Sleep(15 * time.Second)

	stopCh := make(<-chan struct{})
	go c.reportInform.Run(stopCh)

	fmt.Println("running")

	<-stopCh
	return nil
}

func (c *Chargeback) createResources() error {
	cdrClient := c.extension.CustomResourceDe***REMOVED***nitions()
	for _, cdr := range cb.Resources {
		if _, err := cdrClient.Create(cdr); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}
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
