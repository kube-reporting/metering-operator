package cron

import (
	"errors"
	"fmt"
	"time"

	scheduler "github.com/robfig/cron"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	cron "github.com/coreos-inc/kube-chargeback/pkg/cron/v1"
)

// ReportDelay is the amount of time to wait before a report is performed so data can be prepared.
var ReportDelay = -18 * time.Hour

// FrequencySchedules has the CRON expressions used with each frequency.
var FrequencySchedules = map[cron.CronFrequency]string{
	cron.CronFrequencyHourly: "0 * * * *",   // every hour at :00
	cron.CronFrequencyDaily:  "0 18 * * *",  // every day at 18:00
	cron.CronFrequencyWeekly: "0 18 * * 6 ", // every sunday at 18:00
}

// FrequencyDurations are the length of a report of each frequency is generated.
var FrequencyDurations = map[cron.CronFrequency]time.Duration{
	cron.CronFrequencyHourly: time.Hour,          // every hour at :00
	cron.CronFrequencyDaily:  24 * time.Hour,     // every day at 18:00
	cron.CronFrequencyWeekly: 7 * 24 * time.Hour, // every sunday at 18:00
}

// createReportJob must implement the Job interface.
var _ scheduler.Job = createReportJob{}

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

func (o *Operator) createSchedule(c *cron.Cron) error {
	if c == nil {
		return errors.New("cron object can't be nil")
	}

	spec := &c.Spec
	if spec.ReportTemplate.Spec.ReportingStart != nil || spec.ReportTemplate.Spec.ReportingEnd != nil {
		return errors.New("cron can't set ReportingStart or ReportingEnd. The Frequency set determines this.")
	}

	job := &createReportJob{
		o:    o,
		spec: spec,
	}

	schedule := FrequencySchedules[c.Spec.Frequency]
	entryID, err := o.schedule.AddJob(schedule, job)
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

type createReportJob struct {
	o    *Operator
	spec *cron.CronSpec
}

func (j createReportJob) Run() {
	// TODO: ideally the actual schedule time is used here
	end := v1.NewTime(time.Now().UTC())

	// apply standard delay
	end.Time = end.Add(ReportDelay)

	// setup report spec
	reportSpec := j.spec.ReportTemplate.Spec
	reportSpec.ReportingEnd = &end

	// determine beginning of period
	dur := FrequencyDurations[j.spec.Frequency]
	start := v1.NewTime(reportSpec.ReportingEnd.Add(-dur))
	reportSpec.ReportingStart = &start

	report := &cb.Report{
		ObjectMeta: j.spec.ReportTemplate.ObjectMeta,
		Spec:       reportSpec,
	}
	if _, err := j.o.charge.Reports().Create(report); err != nil {
		fmt.Printf("Failed to create scheduled report: %v", err)
	}
}
