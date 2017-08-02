package cron

import (
	"fmt"
	"time"

	scheduler "github.com/rob***REMOVED***g/cron"
	ext_client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	cron "github.com/coreos-inc/kube-chargeback/pkg/cron/v1"
)

func New(cfg *rest.Con***REMOVED***g) (op *Operator, err error) {
	op = new(Operator)
	if op.extension, err = ext_client.NewForCon***REMOVED***g(cfg); err != nil {
		return
	}

	if op.cron, err = cron.NewForCon***REMOVED***g(cfg); err != nil {
		return
	}

	if op.charge, err = cb.NewForCon***REMOVED***g(cfg); err != nil {
		return
	}

	// setup informer for cron
	op.cronInform = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc:  op.cron.Crons().List,
			WatchFunc: op.cron.Crons().Watch,
		},
		&cron.Cron{}, 3*time.Minute, cache.Indexers{},
	)

	// con***REMOVED***gure event listeners
	op.cronInform.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.handleAddCron,
		UpdateFunc: op.handleUpdateCron,
		DeleteFunc: op.handleDeleteCron,
	})

	// setup scheduler
	op.schedule = scheduler.New()
	op.uidToEntry = map[types.UID]scheduler.EntryID{}

	return op, nil
}

// Operator creates reports based on Cron schedules.
type Operator struct {
	extension *ext_client.Clientset
	charge    *cb.ChargebackClient
	cron      *cron.CronClient

	cronInform cache.SharedIndexInformer

	schedule *scheduler.Cron
	// uidToEntry maps Kubernetes UIDs to scheduler entries
	uidToEntry map[types.UID]scheduler.EntryID
}

func (o *Operator) Run(stopCh <-chan struct{}) error {
	if err := o.createResources(); err != nil {
		return err
	}

	// TODO: implement polling
	time.Sleep(15 * time.Second)

	go o.cronInform.Run(stopCh)
	go o.schedule.Start()
	defer o.schedule.Stop()

	fmt.Println("running")

	<-stopCh
	return nil
}

func (o *Operator) createResources() error {
	cdrClient := o.extension.CustomResourceDe***REMOVED***nitions()
	for _, cdr := range cron.Resources {
		if _, err := cdrClient.Create(cdr); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}
