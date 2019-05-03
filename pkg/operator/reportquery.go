package operator

import (
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
)

func (op *Reporting) runReportQueryWorker() {
	logger := op.logger.WithField("component", "reportQueryWorker")
	logger.Infof("ReportQuery worker started")
	// 10 requeues compared to the 5 others have because
	// ReportQueries can reference a lot of other resources, and it may
	// take time for them to all to finish setup
	const maxRequeues = 10
	for op.processResource(logger, op.syncReportQuery, "ReportQuery", op.reportQueryQueue, maxRequeues) {
	}
}

func (op *Reporting) syncReportQuery(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithFields(log.Fields{"reportQuery": name, "namespace": namespace})

	reportQueryLister := op.reportQueryLister
	reportQuery, err := reportQueryLister.ReportQueries(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("ReportQuery %s does not exist anymore", key)
			return nil
		}
		return err
	}
	q := reportQuery.DeepCopy()
	return op.handleReportQuery(logger, q)
}

func (op *Reporting) handleReportQuery(logger log.FieldLogger, query *cbTypes.ReportQuery) error {
	// queue any reportDataSources using this query to create views
	return op.queueDependentReportDataSourcesForQuery(query)
}

func (op *Reporting) uninitialiedDependendenciesHandler() *reporting.UninitialiedDependendenciesHandler {
	return &reporting.UninitialiedDependendenciesHandler{
		HandleUninitializedReportDataSource: op.enqueueReportDataSource,
	}
}

func (op *Reporting) queueDependentReportDataSourcesForQuery(query *cbTypes.ReportQuery) error {
	reportDataSourceLister := op.meteringClient.MeteringV1alpha1().ReportDataSources(query.Namespace)
	reportDataSources, err := reportDataSourceLister.List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, reportDataSource := range reportDataSources.Items {
		if reportDataSource.Spec.ReportQueryView != nil && reportDataSource.Spec.ReportQueryView.QueryName == query.Name {
			op.enqueueReportDataSource(reportDataSource)
		}
	}
	return nil
}
