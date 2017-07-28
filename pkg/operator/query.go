package operator

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/coreos-inc/kube-chargeback/pkg/aws"
	"github.com/coreos-inc/kube-chargeback/pkg/chargeback"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
	"github.com/coreos-inc/kube-chargeback/pkg/presto"
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
	report := obj.(*chargeback.Report)

	// update status
	report.Status.Phase = chargeback.ReportPhaseStarted
	report, err := c.charge.Reports().Update(report)
	if err != nil {
		fmt.Println("Failed to update: ", err)
	}

	rng := chargeback.Range{report.Spec.ReportingStart.Time, report.Spec.ReportingEnd.Time}
	results, err := aws.RetrieveManifests(report.Spec.AWS.Bucket, report.Spec.AWS.ReportPre***REMOVED***x, report.Spec.AWS.ReportName, rng)
	if err != nil {
		c.setError(report, err)
		return
	}

	if len(results) > 1 {
		c.setError(report, errors.New("currently only a single month can be reported on"))
		return
	} ***REMOVED*** if len(results) < 1 {
		c.setError(report, errors.New("no report data was returned for the given range"))
		return
	}

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

	reportTable := fmt.Sprintf("%s_%d", "cost_per_pod", rand.Int31())
	bucket, pre***REMOVED***x := report.Spec.Output.Bucket, report.Spec.Output.Pre***REMOVED***x
	fmt.Printf("Creating table for %s.", reportTable)
	if err = hive.CreatePodCostTable(hiveCon, reportTable, bucket, pre***REMOVED***x); err != nil {
		c.setError(report, fmt.Errorf("Couldn't create table for output report: %v", err))
		return
	}

	promsumTable := fmt.Sprintf("%s_%d", "kube_usage", rand.Int31())
	bucket, pre***REMOVED***x = report.Spec.Chargeback.Bucket, report.Spec.Chargeback.Pre***REMOVED***x
	fmt.Printf("Creating table for %s.", promsumTable)
	if err = hive.CreatePromsumTable(hiveCon, promsumTable, bucket, pre***REMOVED***x); err != nil {
		c.setError(report, fmt.Errorf("Couldn't create table for cluster usage metric data: %v", err))
		return
	}

	awsTable := fmt.Sprintf("%s_%d", "aws_usage", rand.Int31())
	bucket = report.Spec.AWS.Bucket
	fmt.Printf("Creating table for %s.", awsTable)
	if err = hive.CreateAWSUsageTable(hiveCon, awsTable, bucket, results[0]); err != nil {
		c.setError(report, fmt.Errorf("Couldn't create table for AWS usage data: %v", err))
		return
	}

	if err = presto.RunAWSPodDollarReport(prestoCon, promsumTable, awsTable, reportTable, rng); err != nil {
		c.setError(report, fmt.Errorf("Failed to execute Pod Dollar report: %v", err))
		return
	}

	// update status
	report.Status.Phase = chargeback.ReportPhaseFinished
	report, err = c.charge.Reports().Update(report)
	if err != nil {
		fmt.Println("Failed to update: ", err)
	}
}

func (c *Chargeback) setError(q *chargeback.Report, err error) {
	q.Status.Phase = chargeback.ReportPhaseError
	q.Status.Output = err.Error()
	_, err = c.charge.Reports().Update(q)
	if err != nil {
		fmt.Println("FAILED TO REPORT ERROR: ", err)
	}
}
