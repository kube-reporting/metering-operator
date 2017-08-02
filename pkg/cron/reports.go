package cron

import (
	"fmt"

	scheduler "github.com/robfig/cron"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
)

// createReportJob must implement the Job interface.
var _ scheduler.Job = createReportJob{}

func (o *Operator) createReportJob(report *cb.ReportTemplateSpec) *createReportJob {
	return &createReportJob{
		o:        o,
		template: report,
	}
}

type createReportJob struct {
	o        *Operator
	template *cb.ReportTemplateSpec
}

func (j createReportJob) Run() {
	report := &cb.Report{
		ObjectMeta: j.template.ObjectMeta,
		Spec:       j.template.Spec,
	}
	if _, err := j.o.charge.Reports().Create(report); err != nil {
		fmt.Printf("Failed to create scheduled report: %v", err)
	}
}
