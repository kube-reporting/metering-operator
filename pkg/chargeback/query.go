package chargeback

import (
	"fmt"
	"math/rand"
	"time"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (c *Chargeback) handleAddReport(obj interface{}) {
	if obj == nil {
		fmt.Println("received nil object!")
		return
	}

	fmt.Println("New object added!")
	report := obj.(*cb.Report)

	switch report.Status.Phase {
	case cb.ReportPhaseFinished:
		fallthrough
	case cb.ReportPhaseError:
		fmt.Printf("Ignoring %s, status: %s", report.GetSelfLink(), report.Status.Phase)
		return
	}

	// update status
	report.Status.Phase = cb.ReportPhaseStarted
	report, err := c.charge.Reports().Update(report)
	if err != nil {
		fmt.Println("Failed to update: ", err)
	}

	rng := cb.Range{report.Spec.ReportingStart.Time, report.Spec.ReportingEnd.Time}

	hiveCon, err := c.hiveConn()
	if err != nil {
		c.setError(report, fmt.Errorf("Failed to con***REMOVED***gure Hive connection: %v", err))
		return
	}
	defer hiveCon.Close()

	prestoCon, err := c.prestoConn()
	if err != nil {
		c.setError(report, fmt.Errorf("Failed to con***REMOVED***gure Presto connection: %v", err))
		return
	}
	defer prestoCon.Close()

	promsumTable := fmt.Sprintf("%s_%d", "kube_usage", rand.Int31())
	bucket, pre***REMOVED***x := report.Spec.Chargeback.Bucket, report.Spec.Chargeback.Pre***REMOVED***x
	fmt.Printf("Creating table for %s.", promsumTable)
	if err = hive.CreatePromsumTable(hiveCon, promsumTable, bucket, pre***REMOVED***x); err != nil {
		c.setError(report, fmt.Errorf("Couldn't create table for cluster usage metric data: %v", err))
		return
	}

	if report.Spec.AWSReport != nil {
		err = runAWSBillingReport(report, rng, promsumTable, hiveCon, prestoCon)
	} ***REMOVED*** {
		err = runPodUsageReport(report, rng, promsumTable, hiveCon, prestoCon)
	}

	if err != nil {
		c.setError(report, fmt.Errorf("Report execution failed: %v", err))
		return
	}

	// update status
	report.Status.Phase = cb.ReportPhaseFinished
	report, err = c.charge.Reports().Update(report)
	if err != nil {
		fmt.Println("Failed to update: ", err)
	}
}

func (c *Chargeback) setError(q *cb.Report, err error) {
	q.Status.Phase = cb.ReportPhaseError
	q.Status.Output = err.Error()
	_, err = c.charge.Reports().Update(q)
	if err != nil {
		fmt.Println("FAILED TO REPORT ERROR: ", err)
	}
}
