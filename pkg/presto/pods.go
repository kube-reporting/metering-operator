package presto

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/avct/prestgo"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
)

// RunAWSPodDollarReport runs Presto queries to create a Pod Cost summary with AWS billing and Kubernetes usage data.
func RunAWSPodDollarReport(presto *sql.DB, promsumTable, awsTable, outTable string, rng cb.Range) error {
	if presto == nil {
		return errors.New("presto instance of DB cannot be nil")
	}

	reportQuery := podDollarQuery(promsumTable, awsTable, rng.Start, rng.End)
	insert := fmt.Sprintf("INSERT INTO %s %s", outTable, reportQuery)
	result, err := presto.Query(insert)
	if err == nil {
		cols, err := result.Columns()
		if err != nil {
			return fmt.Errorf("could not get columns: %v", err)
		}
		fmt.Println(cols)
	}
	return err
}

// podDollarQuery is a Presto query which calculates Cost Per Pod over a period.
func podDollarQuery(promsumTable, awsBillingTable string, startPeriod, endPeriod time.Time) string {
	query := `SELECT pod, namespace, node, sum(amount * periodCost * percentPeriod) as cost, min(begin) as begin, max(stop) as stop, labels
	FROM (
	    SELECT kubeUsage.subject,
		kubeUsage.amount as amount,
		kubeUsage."start" as begin,
		kubeUsage."end" as stop,
		kubeUsage.labels as labels,
		kubeUsage.labels['pod'] as pod,
		kubeUsage.labels['namespace'] as namespace,
		awsBilling.lineItem_resourceId as node,
		date_diff('millisecond', kubeUsage."start", kubeUsage."end") as duration,
		awsBilling.lineItem_BlendedCost as periodCost,
		CASE
		    WHEN (awsBilling.lineItem_UsageStartDate <= kubeUsage."start") AND (kubeUsage."end" <= awsBilling.lineItem_UsageEndDate) -- AWS data covers entire reporting period
			THEN cast(date_diff('millisecond', kubeUsage."start", kubeUsage."end") as double) / cast(date_diff('millisecond', awsBilling.lineItem_UsageStartDate, awsBilling.lineItem_UsageEndDate) as double)
		    WHEN (awsBilling.lineItem_UsageStartDate <= kubeUsage."start") -- AWS data covers start to middle
			    THEN cast(date_diff('millisecond', kubeUsage."start", awsBilling.lineItem_UsageEndDate) as double) / cast(date_diff('millisecond', awsBilling.lineItem_UsageStartDate, awsBilling.lineItem_UsageEndDate) as double)
		    WHEN (kubeUsage."end" <= awsBilling.lineItem_UsageEndDate) -- AWS data covers middle to end
			    THEN cast(date_diff('millisecond', awsBilling.lineItem_UsageStartDate, kubeUsage."end") as double) / cast(date_diff('millisecond', awsBilling.lineItem_UsageStartDate, awsBilling.lineItem_UsageEndDate) as double)
		    ELSE 1
		END as percentPeriod
	    FROM %s as kubeUsage, %s as awsBilling
	    WHERE position('.csv' IN awsBilling."$path") != 0 -- This prevents JSON manifest ***REMOVED***les from being loaded.
	    AND awsBilling.lineitem_operation = 'RunInstances'
	    AND awsBilling.lineItem_UsageStartDate IS NOT NULL
    	    AND awsBilling.lineItem_UsageEndDate IS NOT NULL
	    AND kubeUsage."start" >= timestamp '%s'
	    AND kubeUsage."end" <= timestamp '%s'
	    AND awsBilling.lineItem_resourceId = split_part(split_part(kubeUsage.labels['provider_id'], ':///', 2), '/', 2)
	    AND awsBilling.lineItem_UsageStartDate <= kubeUsage."end"
	    AND awsBilling.lineItem_UsageEndDate >= kubeUsage."start"
	)
	GROUP BY pod, namespace, node, labels`
	return fmt.Sprintf(query, promsumTable, awsBillingTable, prestoTime(startPeriod), prestoTime(endPeriod))
}
