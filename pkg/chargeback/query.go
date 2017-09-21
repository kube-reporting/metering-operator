package chargeback

import (
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (c *Chargeback) handleAddReport(obj interface{}) {
	if obj == nil {
		log.Debugf("received nil object!")
		return
	}

	report := obj.(*cb.Report)

	log.WithFields(log.Fields{
		"name":  report.Name,
		"scope": report.Spec.Scope,
		"start": report.Spec.ReportingStart,
		"end":   report.Spec.ReportingEnd,
	}).Infof("new report discovered")

	switch report.Status.Phase {
	case cb.ReportPhaseFinished:
		fallthrough
	case cb.ReportPhaseError:
		log.Warnf("ignoring %s, status: %s", report.GetSelfLink(), report.Status.Phase)
		return
	}

	// update status
	report.Status.Phase = cb.ReportPhaseStarted
	report, err := c.charge.Reports(c.namespace).Update(report)
	if err != nil {
		log.Warnf("failed to update report status for %q: %v", report.Name, err)
	}

	rng := cb.Range{report.Spec.ReportingStart.Time, report.Spec.ReportingEnd.Time}

	hiveCon, err := c.hiveConn()
	if err != nil {
		c.setError(report, fmt.Errorf("Failed to configure Hive connection: %v", err))
		return
	}
	defer hiveCon.Close()

	prestoCon, err := c.prestoConn()
	if err != nil {
		c.setError(report, fmt.Errorf("Failed to configure Presto connection: %v", err))
		return
	}
	defer prestoCon.Close()

	promsumTable := fmt.Sprintf("%s_%d", "kube_usage", rand.Int31())
	bucket, prefix := report.Spec.Chargeback.Bucket, report.Spec.Chargeback.Prefix
	log.Debugf("Creating table for promsum: %q.", promsumTable)
	if err = hive.CreatePromsumTable(hiveCon, promsumTable, bucket, prefix); err != nil {
		c.setError(report, fmt.Errorf("Couldn't create table for cluster usage metric data: %v", err))
		return
	}

	if report.Spec.AWSReport != nil {
		err = runAWSBillingReport(report, rng, promsumTable, hiveCon, prestoCon, report.Spec.Scope)
	} else {
		err = runUsageReport(report, rng, promsumTable, hiveCon, prestoCon, report.Spec.Scope)
	}

	if err != nil {
		c.setError(report, fmt.Errorf("Report execution failed: %v", err))
		return
	}

	// update status
	report.Status.Phase = cb.ReportPhaseFinished
	report, err = c.charge.Reports(c.namespace).Update(report)
	if err != nil {
		log.Warnf("failed to update report status for %q: ", report.Name, err)
	} else {
		log.Infof("finished report %q", report.Name)
	}
}

func (c *Chargeback) setError(q *cb.Report, err error) {
	log.Warnf("%v", err)
	q.Status.Phase = cb.ReportPhaseError
	q.Status.Output = err.Error()
	_, err = c.charge.Reports(c.namespace).Update(q)
	if err != nil {
		log.Warnf("FAILED TO REPORT ERROR: %v", err)
	}
}
