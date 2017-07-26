package operator

import (
	"database/sql"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/coreos-inc/kube-chargeback/pkg/chargeback"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

type Config struct {
	HiveHost   string
	PrestoHost string
}

func New(cfg Config) (*Chargeback, error) {
	cb := &Chargeback{
		hiveHost:   cfg.HiveHost,
		prestoHost: cfg.PrestoHost,
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	if cb.kube, err = kubernetes.NewForConfig(config); err != nil {
		return nil, err
	}

	if cb.charge, err = chargeback.NewForConfig(config); err != nil {
		return nil, err
	}

	cb.reportInform = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  cb.charge.Reports(api.NamespaceAll).List,
			WatchFunc: cb.charge.Reports(api.NamespaceAll).Watch,
		},
		&chargeback.Report{}, 3*time.Minute, cache.Indexers{},
	)

	cb.reportInform.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: cb.handleAddReport,
	})

	return cb, nil
}

type Chargeback struct {
	kube   *kubernetes.Clientset
	charge *chargeback.ChargebackClient

	reportInform cache.SharedIndexInformer

	hiveHost   string
	prestoHost string
}

func (c *Chargeback) Run() error {
	err := c.createTPRs()
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

func (c *Chargeback) createTPRs() error {
	tprs := []*extensions.ThirdPartyResource{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "report." + chargeback.Group,
			},
			Versions: []extensions.APIVersion{
				{Name: chargeback.Version},
			},
			Description: "Billing report",
		},
	}
	tprClient := c.kube.ThirdPartyResources()

	for _, tpr := range tprs {
		if _, err := tprClient.Create(tpr); err != nil && !apierrors.IsAlreadyExists(err) {
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
