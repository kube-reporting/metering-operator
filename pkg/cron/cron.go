package cron

import (
	"errors"
	"fmt"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	cron "github.com/coreos-inc/kube-chargeback/pkg/cron/v1"
)

func (o *Operator) updateSchedule(c *cron.Cron) error {
	_, ok := o.uidToEntry[c.GetUID()]
	if ok {
		if err := o.removeSchedule(c); err != nil {
			fmt.Printf("Failed to remove scheduler for UID '%s': %v", c.GetUID(), err)
		}
	}

	if c.Spec.Suspend != nil && *c.Spec.Suspend {
		fmt.Printf("Not creating schedule for %s because is suspended.", c.GetSelfLink())
		return nil
	}

	return o.createSchedule(c)
}

type job struct {
	r func()
}

func (j job) Run() {
	j.r()
}

func (o *Operator) createSchedule(c *cron.Cron) error {
	if c == nil {
		return errors.New("cron object can't be nil")
	}

	job := job{func() {
		report := &cb.Report{
			ObjectMeta: c.Spec.ReportTemplate.ObjectMeta,
			Spec:       c.Spec.ReportTemplate.Spec,
		}
		if _, err := o.charge.Reports(c.GetNamespace()).Create(report); err != nil {
			fmt.Printf("Failed to create scheduled report: %v", err)
		}
	}}
	entryID, err := o.schedule.AddJob(c.Spec.Schedule, job)
	if err != nil {
		return fmt.Errorf("couldn't add report to scheduler: %v", err)
	}

	o.uidToEntry[c.GetUID()] = entryID
	return nil
}

func (o *Operator) removeSchedule(c *cron.Cron) error {
	if entryID, ok := o.uidToEntry[c.GetUID()]; ok {
		o.schedule.Remove(entryID)
	}
	return nil
}
