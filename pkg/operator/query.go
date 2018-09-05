package operator

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
)

func (op *Reporting) runReportGenerationQueryWorker() {
	logger := op.logger.WithField("component", "reportGenerationQueryWorker")
	logger.Infof("ReportGenerationQuery worker started")
	for op.processReportGenerationQuery(logger) {

	}
}

func (op *Reporting) processReportGenerationQuery(logger log.FieldLogger) bool {
	obj, quit := op.queues.reportGenerationQueryQueue.Get()
	if quit {
		logger.Infof("queue is shutting down, exiting ReportGenerationQuery worker")
		return false
	}
	defer op.queues.reportGenerationQueryQueue.Done(obj)

	logger = logger.WithFields(newLogIdenti***REMOVED***er(op.rand))
	if key, ok := op.getKeyFromQueueObj(logger, "ReportGenerationQuery", obj, op.queues.reportGenerationQueryQueue); ok {
		err := op.syncReportGenerationQuery(logger, key)
		op.handleErr(logger, err, "ReportGenerationQuery", key, op.queues.reportGenerationQueryQueue)
	}
	return true
}

func (op *Reporting) syncReportGenerationQuery(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithField("generationQuery", name)

	reportGenerationQueryLister := op.informers.Metering().V1alpha1().ReportGenerationQueries().Lister()
	reportGenerationQuery, err := reportGenerationQueryLister.ReportGenerationQueries(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("ReportGenerationQuery %s does not exist anymore", key)
			return nil
		}
		return err
	}

	logger.Infof("syncing reportGenerationQuery %s", reportGenerationQuery.GetName())
	err = op.handleReportGenerationQuery(logger, reportGenerationQuery)
	if err != nil {
		logger.WithError(err).Errorf("error syncing reportGenerationQuery %s", reportGenerationQuery.GetName())
		return err
	}
	logger.Infof("successfully synced reportGenerationQuery %s", reportGenerationQuery.GetName())
	return nil
}

func (op *Reporting) handleReportGenerationQuery(logger log.FieldLogger, generationQuery *cbTypes.ReportGenerationQuery) error {
	generationQuery = generationQuery.DeepCopy()

	var viewName string
	if generationQuery.ViewName == "" {
		logger.Infof("new reportGenerationQuery discovered")
		if generationQuery.Spec.View.Disabled {
			logger.Infof("reportGenerationQuery has spec.view.disabled=true, skipping processing")
			return nil
		}
		viewName = generationQueryViewName(generationQuery.Name)
	} ***REMOVED*** {
		logger.Infof("existing reportGenerationQuery discovered, viewName: %s", generationQuery.ViewName)
		viewName = generationQuery.ViewName
	}

	reportDataSourceLister := op.informers.Metering().V1alpha1().ReportDataSources().Lister()
	reportGenerationQueryLister := op.informers.Metering().V1alpha1().ReportGenerationQueries().Lister()

	depsStatus, err := reporting.GetGenerationQueryDependenciesStatus(
		reporting.NewReportGenerationQueryListerGetter(reportGenerationQueryLister),
		reporting.NewReportDataSourceListerGetter(reportDataSourceLister),
		generationQuery,
	)
	if err != nil {
		return fmt.Errorf("unable to create view for ReportGenerationQuery %s, failed to retrieve dependencies: %v", generationQuery.Name, err)
	}
	validateResults, err := op.validateDependencyStatus(depsStatus)
	if err != nil {
		return fmt.Errorf("unable to create view for ReportGenerationQuery %s, failed to validate dependencies %v", generationQuery.Name, err)
	}

	templateInfo := &templateInfo{
		DynamicDependentQueries: validateResults.DynamicReportGenerationQueries,
		Report:                  nil,
	}

	qr := queryRenderer{templateInfo: templateInfo}
	renderedQuery, err := qr.Render(generationQuery.Spec.Query)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("CREATE OR REPLACE VIEW %s AS %s", viewName, renderedQuery)
	_, err = op.prestoConn.Query(query)
	if err != nil {
		return err
	}

	return op.updateReportQueryViewName(logger, generationQuery, viewName)
}

func (op *Reporting) updateReportQueryViewName(logger log.FieldLogger, generationQuery *cbTypes.ReportGenerationQuery, viewName string) error {
	generationQuery.ViewName = viewName
	_, err := op.meteringClient.MeteringV1alpha1().ReportGenerationQueries(generationQuery.Namespace).Update(generationQuery)
	if err != nil {
		logger.WithError(err).Errorf("failed to update ReportGenerationQuery view name for %q", generationQuery.Name)
		return err
	}
	return nil
}

// validateDependencyStatus runs
// reporting.ValidateGenerationQueryDependenciesStatus and requeues any
// uninitialized dependencies
func (op *Reporting) validateDependencyStatus(dependencyStatus *reporting.GenerationQueryDependenciesStatus) (*reporting.ReportGenerationQueryDependencies, error) {
	deps, err := reporting.ValidateGenerationQueryDependenciesStatus(dependencyStatus)
	if err != nil {
		for _, query := range dependencyStatus.UninitializedReportGenerationQueries {
			key, err := cache.MetaNamespaceKeyFunc(query)
			if err == nil {
				op.queues.reportGenerationQueryQueue.Add(key)
			}
		}

		for _, dataSource := range dependencyStatus.UninitializedReportDataSources {
			key, err := cache.MetaNamespaceKeyFunc(dataSource)
			if err == nil {
				op.queues.reportDataSourceQueue.Add(key)
			}
		}
		return nil, err
	}
	return deps, nil
}
