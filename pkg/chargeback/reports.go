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

type scopeFuncStruct struct {
	RunUsageReport       func(*sql.DB, string, string, cb.Range) error
	RunAWSDollarReport   func(*sql.DB, string, string, string, cb.Range) error
	CreateAWSDollarTable func(*hive.Connection, string, string, string) error
	CreateUsageTable     func(*hive.Connection, string, string, string) error
}

var scopeFuncs = map[cb.ReportScope]scopeFuncStruct{
	cb.ReportScopePod: {
		RunUsageReport:       presto.RunPodUsageReport,
		RunAWSDollarReport:   presto.RunAWSPodDollarReport,
		CreateAWSDollarTable: hive.CreatePodCostTable,
		CreateUsageTable:     hive.CreatePodUsageTable,
	},
	cb.ReportScopeNamespace: {
		RunUsageReport:       presto.RunNamespaceUsageReport,
		RunAWSDollarReport:   presto.RunAWSNamespaceDollarReport,
		CreateAWSDollarTable: hive.CreateNamespaceCostTable,
		CreateUsageTable:     hive.CreateNamespaceUsageTable,
	},
}

func runAWSBillingReport(report *cb.Report, rng cb.Range, promsumTbl string, hiveCon *hive.Connection, prestoCon *sql.DB, reportScope cb.ReportScope) error {
	_, ok := scopeFuncs[reportScope]
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
	if err = scopeFuncs[reportScope].CreateAWSDollarTable(hiveCon, reportTable, bucket, prefix); err != nil {
		return fmt.Errorf("Couldn't create table for output report: %v", err)
	}

	awsTable := fmt.Sprintf("%s_%d", "aws_usage", rand.Int31())
	bucket = report.Spec.AWSReport.Bucket
	fmt.Printf("Creating table for %s.", awsTable)
	if err = hive.CreateAWSUsageTable(hiveCon, awsTable, bucket, results[0]); err != nil {
		return fmt.Errorf("Couldn't create table for AWS usage data: %v", err)
	}

	if err = scopeFuncs[reportScope].RunAWSDollarReport(prestoCon, promsumTbl, awsTable, reportTable, rng); err != nil {
		return fmt.Errorf("Failed to execute Pod Dollar report: %v", err)
	}
	return nil
}

func runUsageReport(report *cb.Report, rng cb.Range, promsumTbl string, hiveCon *hive.Connection, prestoCon *sql.DB, reportScope cb.ReportScope) error {
	_, ok := scopeFuncs[reportScope]
	if !ok {
		return fmt.Errorf("unknown report scope: %s", reportScope)
	}

	reportTable := fmt.Sprintf("%s_usage_%d", reportScope, rand.Int31())
	bucket, prefix := report.Spec.Output.Bucket, report.Spec.Output.Prefix
	if err := scopeFuncs[reportScope].CreateUsageTable(hiveCon, reportTable, bucket, prefix); err != nil {
		return fmt.Errorf("Couldn't create table for output report: %v", err)
	}

	if err := scopeFuncs[reportScope].RunUsageReport(prestoCon, promsumTbl, reportTable, rng); err != nil {
		return fmt.Errorf("Failed to execute %s usage report: %v", reportScope, err)
	}
	return nil
}
