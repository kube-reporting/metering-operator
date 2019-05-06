package operator

import (
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
)

func (op *Reporting) runReportGenerationQueryWorker() {
	logger := op.logger.WithField("component", "reportGenerationQueryWorker")
	logger.Infof("ReportGenerationQuery worker started")
	// 10 requeues compared to the 5 others have because
	// ReportGenerationQueries can reference a lot of other resources, and it may
	// take time for them to all to finish setup
	const maxRequeues = 10
	for op.processResource(logger, op.syncReportGenerationQuery, "ReportGenerationQuery", op.reportGenerationQueryQueue, maxRequeues) {
	}
}

func (op *Reporting) syncReportGenerationQuery(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithFields(log.Fields{"reportGenerationQuery": name, "namespace": namespace})

	reportGenerationQueryLister := op.reportGenerationQueryLister
	reportGenerationQuery, err := reportGenerationQueryLister.ReportGenerationQueries(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("ReportGenerationQuery %s does not exist anymore", key)
			return nil
		}
		return err
	}
	q := reportGenerationQuery.DeepCopy()
	return op.handleReportGenerationQuery(logger, q)
}

func (op *Reporting) handleReportGenerationQuery(logger log.FieldLogger, generationQuery *cbTypes.ReportGenerationQuery) error {
	// queue any reportDataSources using this query to create views
	return op.queueDependentReportDataSourcesForQuery(generationQuery)
}

func (op *Reporting) uninitialiedDependendenciesHandler() *reporting.UninitialiedDependendenciesHandler {
	return &reporting.UninitialiedDependendenciesHandler{
		HandleUninitializedReportDataSource: op.enqueueReportDataSource,
	}
}

func (op *Reporting) queueDependentReportDataSourcesForQuery(generationQuery *cbTypes.ReportGenerationQuery) error {
	reportDataSourceLister := op.meteringClient.MeteringV1alpha1().ReportDataSources(generationQuery.Namespace)
	reportDataSources, err := reportDataSourceLister.List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, reportDataSource := range reportDataSources.Items {
		if reportDataSource.Spec.GenerationQueryView != nil && reportDataSource.Spec.GenerationQueryView.QueryName == generationQuery.Name {
			op.enqueueReportDataSource(reportDataSource)
		}
	}
	return nil
}
