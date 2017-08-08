package chargeback

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"

	"github.com/coreos-inc/kube-chargeback/pkg/aws"
	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
	"github.com/coreos-inc/kube-chargeback/pkg/presto"
)

func runAWSBillingReport(report *cb.Report, rng cb.Range, promsumTbl string, hiveCon *hive.Connection, prestoCon *sql.DB) error {
	results, err := aws.RetrieveManifests(report.Spec.AWSReport.Bucket, report.Spec.AWSReport.Prefix, rng)
	if err != nil {
		return err
	}

	if len(results) > 1 {
		return errors.New("currently only a single month can be reported on")
	} else if len(results) < 1 {
		return errors.New("no report data was returned for the given range")
	}

	reportTable := fmt.Sprintf("%s_%d", "cost_per_pod", rand.Int31())
	bucket, prefix := report.Spec.Output.Bucket, report.Spec.Output.Prefix
	fmt.Printf("Creating table for %s.", reportTable)
	if err = hive.CreatePodCostTable(hiveCon, reportTable, bucket, prefix); err != nil {
		return fmt.Errorf("Couldn't create table for output report: %v", err)
	}

	awsTable := fmt.Sprintf("%s_%d", "aws_usage", rand.Int31())
	bucket = report.Spec.AWSReport.Bucket
	fmt.Printf("Creating table for %s.", awsTable)
	if err = hive.CreateAWSUsageTable(hiveCon, awsTable, bucket, results[0]); err != nil {
		return fmt.Errorf("Couldn't create table for AWS usage data: %v", err)
	}

	if err = presto.RunAWSPodDollarReport(prestoCon, promsumTbl, awsTable, reportTable, rng); err != nil {
		return fmt.Errorf("Failed to execute Pod Dollar report: %v", err)
	}
	return nil
}

func runPodUsageReport(report *cb.Report, rng cb.Range, promsumTbl string, hiveCon *hive.Connection, prestoCon *sql.DB) error {
	reportTable := fmt.Sprintf("%s_%d", "pod_usage", rand.Int31())
	bucket, prefix := report.Spec.Output.Bucket, report.Spec.Output.Prefix
	fmt.Printf("Creating table for %s.", reportTable)
	if err := hive.CreatePodUsageTable(hiveCon, reportTable, bucket, prefix); err != nil {
		return fmt.Errorf("Couldn't create table for output report: %v", err)
	}

	if err := presto.RunPodUsageReport(prestoCon, promsumTbl, reportTable, rng); err != nil {
		return fmt.Errorf("Failed to execute Pod Usage report: %v", err)
	}
	return nil
}
