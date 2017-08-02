package cron

import (
	"fmt"

	cron "github.com/coreos-inc/kube-chargeback/pkg/cron/v1"
)

func (o *Operator) handleAddCron(obj interface{}) {
	c := castCron(obj)
	if err := o.updateSchedule(c); err != nil {
		fmt.Errorf("Failed to add Cron '%s': %v", c.GetSelfLink(), err)
	}
}

func (o *Operator) handleUpdateCron(oldObj, newObj interface{}) {
	c := castCron(newObj)
	if err := o.updateSchedule(c); err != nil {
		fmt.Errorf("Failed to update Cron '%s': %v", c.GetSelfLink(), err)
	}
}

func (o *Operator) handleDeleteCron(obj interface{}) {
	c := castCron(obj)
	if err := o.removeSchedule(c); err != nil {
		fmt.Errorf("Failed to remove Cron '%s': %v", c.GetSelfLink(), err)
	}
}

func castCron(obj interface{}) *cron.Cron {
	cron, ok := obj.(*cron.Cron)
	if !ok {
		fmt.Println("Error: could not cast to Cron")
		return nil
	}
	return cron
}
