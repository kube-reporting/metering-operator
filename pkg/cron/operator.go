package cron

import (
	"fmt"
	"time"

	ext_client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	cron "github.com/coreos-inc/kube-chargeback/pkg/cron/v1"
)

func New(cfg *rest.Config) (op *Operator, err error) {
	op = new(Operator)
	if op.extension, err = ext_client.NewForConfig(cfg); err != nil {
		return
	}

	if op.cron, err = cron.NewForConfig(cfg); err != nil {
		return
	}

	if op.charge, err = cb.NewForConfig(cfg); err != nil {
		return
	}

	// setup informer for cron
	op.cronInform = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  op.cron.Crons(api.NamespaceAll).List,
			WatchFunc: op.cron.Crons(api.NamespaceAll).Watch,
		},
		&cron.Cron{}, 3*time.Minute, cache.Indexers{},
	)

	return op, nil
}

// Operator creates reports based on Cron schedules.
type Operator struct {
	extension *ext_client.Clientset
	charge    *cb.ChargebackClient
	cron      *cron.CronClient

	cronInform cache.SharedIndexInformer
}

func (o *Operator) Run() error {
	if err := o.createResources(); err != nil {
		return err
	}

	stopCh := make(<-chan struct{})
	go o.cronInform.Run(stopCh)

	// TODO: implement polling
	time.Sleep(15 * time.Second)

	fmt.Println("running")

	<-stopCh
	return nil
}

func (o *Operator) createResources() error {
	cdrClient := o.extension.CustomResourceDefinitions()
	for _, cdr := range cron.Resources {
		if _, err := cdrClient.Create(cdr); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}
