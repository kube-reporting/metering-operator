package chargeback

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/coreos-inc/kube-chargeback/pkg/aws"
	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
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
	report := obj.(*cb.Report)

	// update status
	report.Status.Phase = cb.ReportPhaseStarted
	report, err := c.charge.Reports().Update(report)
	if err != nil {
		fmt.Println("Failed to update: ", err)
	}

	rng := cb.Range{report.Spec.ReportingStart.Time, report.Spec.ReportingEnd.Time}
	results, err := aws.RetrieveManifests(report.Spec.AWSReport.Bucket, report.Spec.AWSReport.Prefix, rng)
	if err != nil {
		c.setError(report, err)
		return
	}

	if len(results) > 1 {
		c.setError(report, errors.New("currently only a single month can be reported on"))
		return
	} else if len(results) < 1 {
		c.setError(report, errors.New("no report data was returned for the given range"))
		return
	}

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

	reportTable := fmt.Sprintf("%s_%d", "cost_per_pod", rand.Int31())
	bucket, prefix := report.Spec.Output.Bucket, report.Spec.Output.Prefix
	fmt.Printf("Creating table for %s.", reportTable)
	if err = hive.CreatePodCostTable(hiveCon, reportTable, bucket, prefix); err != nil {
		c.setError(report, fmt.Errorf("Couldn't create table for output report: %v", err))
		return
	}

	promsumTable := fmt.Sprintf("%s_%d", "kube_usage", rand.Int31())
	bucket, prefix = report.Spec.Chargeback.Bucket, report.Spec.Chargeback.Prefix
	fmt.Printf("Creating table for %s.", promsumTable)
	if err = hive.CreatePromsumTable(hiveCon, promsumTable, bucket, prefix); err != nil {
		c.setError(report, fmt.Errorf("Couldn't create table for cluster usage metric data: %v", err))
		return
	}

	awsTable := fmt.Sprintf("%s_%d", "aws_usage", rand.Int31())
	bucket = report.Spec.AWSReport.Bucket
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
