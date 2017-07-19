package operator

import (
	"errors"
	"fmt"

	"github.com/coreos-inc/kube-chargeback/pkg/aws"
	"github.com/coreos-inc/kube-chargeback/pkg/chargeback"
	"github.com/coreos-inc/kube-chargeback/pkg/hive"
	"github.com/coreos-inc/kube-chargeback/pkg/presto"
	"math/rand"
)

func (c *Chargeback) handleAddQuery(obj interface{}) {
	query := obj.(*chargeback.Query)

	// update status
	query.Status.Phase = chargeback.QueryPhaseStarted
	query, err := c.charge.Queries(query.Namespace).Update(query)
	if err != nil {
		fmt.Println("Failed to update: ", err)
	}

	rng := chargeback.Range{query.Spec.ReportingStart, query.Spec.ReportingEnd}
	results, err := aws.RetrieveManifests(query.Spec.AWS.Bucket, query.Spec.AWS.ReportPre***REMOVED***x, query.Spec.AWS.ReportName, rng)
	if err != nil {
		c.setError(query, err)
		return
	}

	if len(results) > 1 {
		c.setError(query, errors.New("currently only a single month can be reported on"))
		return
	} ***REMOVED*** if len(results) < 1 {
		c.setError(query, errors.New("no report data was returned for the given range"))
		return
	}

	hiveCon, err := c.hiveConn()
	if err != nil {
		c.setError(query, fmt.Errorf("Failed to con***REMOVED***gure Hive connection: %v", err))
		return
	}
	defer hiveCon.Close()

	prestoCon, err := c.prestoConn()
	if err != nil {
		c.setError(query, fmt.Errorf("Failed to con***REMOVED***gure Presto connection: %v", err))
		return
	}
	defer prestoCon.Close()

	reportTable := fmt.Sprintf("%s%d", query.Spec.Output.Pre***REMOVED***x, rand.Int31())
	bucket, pre***REMOVED***x := query.Spec.Output.Bucket, query.Spec.Output.Pre***REMOVED***x
	if err = hive.CreatePodCostTable(hiveCon, reportTable, bucket, pre***REMOVED***x); err != nil {
		c.setError(query, fmt.Errorf("Couldn't create table for output report: %v", err))
		return
	}

	promsumTable := fmt.Sprintf("%s%d", query.Spec.Chargeback.Pre***REMOVED***x, rand.Int31())
	bucket, pre***REMOVED***x = query.Spec.Chargeback.Bucket, query.Spec.Chargeback.Pre***REMOVED***x
	if err = hive.CreatePromsumTable(hiveCon, promsumTable, bucket, pre***REMOVED***x); err != nil {
		c.setError(query, fmt.Errorf("Couldn't create table for cluster usage metric data: %v", err))
		return
	}

	awsTable := fmt.Sprintf("%s%d", results[0].AssemblyID, rand.Int31())
	bucket = query.Spec.AWS.Bucket
	if err = hive.CreateAWSUsageTable(hiveCon, awsTable, bucket, results[0]); err != nil {
		c.setError(query, fmt.Errorf("Couldn't create table for AWS usage data: %v", err))
		return
	}

	if err = presto.RunAWSPodDollarReport(prestoCon, promsumTable, awsTable, reportTable, rng); err != nil {
		c.setError(query, fmt.Errorf("Failed to execute Pod Dollar report: %v", err))
		return
	}

	// update status
	query.Status.Phase = chargeback.QueryPhaseFinished
	query, err = c.charge.Queries(query.Namespace).Update(query)
	if err != nil {
		fmt.Println("Failed to update: ", err)
	}
}

func (c *Chargeback) setError(q *chargeback.Query, err error) {
	q.Status.Phase = chargeback.QueryPhaseError
	q.Status.Output = err.Error()
	_, err = c.charge.Queries(q.Namespace).Update(q)
	if err != nil {
		fmt.Println("FAILED TO REPORT ERROR: ", err)
	}
}
