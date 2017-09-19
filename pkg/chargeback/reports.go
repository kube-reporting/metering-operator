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

// Maps cb.ReportScopes to matching presto usage report generating functions.
var scopeToUsageFunc = map[cb.ReportScope]func(*sql.DB, string, string, cb.Range) error{
	cb.ReportScopePod:       presto.RunPodUsageReport,
}

// Maps cb.ReportScopes to matching presto AWS dollar report generating
// functions.
var scopeToAWSDollarFunc = map[cb.ReportScope]func(*sql.DB, string, string, string, cb.Range) error{
	cb.ReportScopePod:       presto.RunAWSPodDollarReport,
}

func runAWSBillingReport(report *cb.Report, rng cb.Range, promsumTbl string, hiveCon *hive.Connection, prestoCon *sql.DB, reportScope cb.ReportScope) error {
	runAWSDollarReportFunc, ok := scopeToAWSDollarFunc[reportScope]
	if !ok {
		return fmt.Errorf("unknown report scope: %s", reportScope)
	}

	results, err := aws.RetrieveManifests(report.Spec.AWSReport.Bucket, report.Spec.AWSReport.Prefix, rng)
	if err != nil {
		return err
	}

	if len(results) > 1 {
		return errors.New("currently only a single month can be reported on")
	} else if len(results) < 1 {
		return errors.New("no report data was returned for the given range")
	}

	reportTable := fmt.Sprintf("%s_per_pod_%d", reportScope, rand.Int31())
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

	if err = runAWSDollarReportFunc(prestoCon, promsumTbl, awsTable, reportTable, rng); err != nil {
		return fmt.Errorf("Failed to execute Pod Dollar report: %v", err)
	}
	return nil
}

func runUsageReport(report *cb.Report, rng cb.Range, promsumTbl string, hiveCon *hive.Connection, prestoCon *sql.DB, reportScope cb.ReportScope) error {
	runUsageReportFunc, ok := scopeToUsageFunc[reportScope]
	if !ok {
		return fmt.Errorf("unknown report scope: %s", reportScope)
	}

	reportTable := fmt.Sprintf("%s_usage_%d", reportScope, rand.Int31())
	bucket, prefix := report.Spec.Output.Bucket, report.Spec.Output.Prefix
	fmt.Printf("Creating table for %s.", reportTable)
	if err := hive.CreatePodUsageTable(hiveCon, reportTable, bucket, prefix); err != nil {
		return fmt.Errorf("Couldn't create table for output report: %v", err)
	}

	if err := runUsageReportFunc(prestoCon, promsumTbl, reportTable, rng); err != nil {
		return fmt.Errorf("Failed to execute %s usage report: %v", reportScope, err)
	}
	return nil
}
