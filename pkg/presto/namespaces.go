package presto

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/avct/prestgo"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
)

// RunNamespaceUsageReport creates a report with Namespace usage.
func RunNamespaceUsageReport(presto *sql.DB, promsumTable, outTable string, rng cb.Range) error {
	reportQuery := namespaceUsageQuery(promsumTable, rng.Start, rng.End)
	return executeInsertQuery(presto, outTable, reportQuery)
}

// RunAWSNamespaceDollarReport runs Presto queries to create a Namespace Cost summary with AWS billing and Kubernetes usage data.
func RunAWSNamespaceDollarReport(presto *sql.DB, promsumTable, awsTable, outTable string, rng cb.Range) error {
	reportQuery := namespaceDollarQuery(promsumTable, awsTable, rng.Start, rng.End)
	return executeInsertQuery(presto, outTable, reportQuery)
}

// namespaceUsageQuery is a Presto query calculating namespace usage based on request.
func namespaceUsageQuery(promsumTable string, startPeriod, endPeriod time.Time) string {
	query := `SELECT namespace, sum(amount) as usage, min(begin) as begin, max(stop) as stop
	FROM (
		SELECT kubeUsage.subject,
		kubeUsage.amount as amount,
		kubeUsage."start" as begin,
		kubeUsage."end" as stop,
		kubeUsage.labels['namespace'] as namespace,
		date_diff('millisecond', kubeUsage."start", kubeUsage."end") as duration
		FROM %s as kubeUsage
		WHERE kubeUsage."start" >= timestamp '%s'
		AND kubeUsage."end" <= timestamp '%s'
	)
GROUP BY namespace`
	return fmt.Sprintf(query, promsumTable, prestoTime(startPeriod), prestoTime(endPeriod))
}

// namespaceDollarQuery is a Presto query which calculates Cost Per Namespace over a period.
func namespaceDollarQuery(promsumTable, awsBillingTable string, startPeriod, endPeriod time.Time) string {
	query := `SELECT namespace, sum(amount * periodCost * percentPeriod) as cost, min(begin) as begin, max(stop) as stop
	FROM (
		SELECT kubeUsage.subject,
		kubeUsage.amount as amount,
		kubeUsage."start" as begin,
		kubeUsage."end" as stop,
		kubeUsage.labels['namespace'] as namespace,
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
		WHERE position('.csv' IN awsBilling."$path") != 0 -- This prevents JSON manifest files from being loaded.
		AND awsBilling.lineitem_operation = 'RunInstances'
		AND awsBilling.lineItem_UsageStartDate IS NOT NULL
		AND awsBilling.lineItem_UsageEndDate IS NOT NULL
		AND kubeUsage."start" >= timestamp '%s'
		AND kubeUsage."end" <= timestamp '%s'
		AND awsBilling.lineItem_resourceId = split_part(split_part(kubeUsage.labels['provider_id'], ':///', 2), '/', 2)
		AND awsBilling.lineItem_UsageStartDate <= kubeUsage."end"
		AND awsBilling.lineItem_UsageEndDate >= kubeUsage."start"
	)
	GROUP BY namespace`
	return fmt.Sprintf(query, promsumTable, awsBillingTable, prestoTime(startPeriod), prestoTime(endPeriod))
}
