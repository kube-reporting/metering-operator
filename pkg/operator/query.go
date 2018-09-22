package operator

import (
	"fmt"

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
	// take time for them to all to ***REMOVED***nish setup
	const maxRequeues = 10
	for op.processResource(logger, op.syncReportGenerationQuery, "ReportGenerationQuery", op.queues.reportGenerationQueryQueue, maxRequeues) {
	}
}

func (op *Reporting) syncReportGenerationQuery(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithField("ReportGenerationQuery", name)

	reportGenerationQueryLister := op.informers.Metering().V1alpha1().ReportGenerationQueries().Lister()
	reportGenerationQuery, err := reportGenerationQueryLister.ReportGenerationQueries(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("ReportGenerationQuery %s does not exist anymore", key)
			return nil
		}
		return err
	}

	return op.handleReportGenerationQuery(logger, reportGenerationQuery)
}

func (op *Reporting) handleReportGenerationQuery(logger log.FieldLogger, generationQuery *cbTypes.ReportGenerationQuery) error {
	generationQuery = generationQuery.DeepCopy()

	var viewName string
	if generationQuery.Spec.View.Disabled {
		logger.Infof("ReportGenerationQuery has spec.view.disabled=true, skipping processing")
		return nil
	} ***REMOVED*** if generationQuery.ViewName == "" {
		logger.Infof("new ReportGenerationQuery discovered")
		viewName = generationQueryViewName(generationQuery.Name)
	} ***REMOVED*** {
		logger.Infof("existing ReportGenerationQuery discovered, viewName: %s", generationQuery.ViewName)
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

	err = op.updateReportQueryViewName(logger, generationQuery, viewName)
	if err != nil {
		return err
	}

	// enqueue any queries depending on this one
	if err := op.queueDependentReportGeneratonQueries(generationQuery); err != nil {
		logger.WithError(err).Errorf("error queuing ReportGenerationQuery dependents of %s", generationQuery.Name)
	}

	return nil
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
			op.enqueueReportGenerationQuery(query)
		}

		for _, dataSource := range dependencyStatus.UninitializedReportDataSources {
			op.enqueueReportDataSource(dataSource)
		}
		return nil, err
	}
	return deps, nil
}

// queueDependentReportGeneratonQueries will queue all ReportGenerationQueries in the namespace which have a dependency on the generationQuery
func (op *Reporting) queueDependentReportGeneratonQueries(generationQuery *cbTypes.ReportGenerationQuery) error {
	queryLister := op.meteringClient.MeteringV1alpha1().ReportGenerationQueries(generationQuery.Namespace)
	queries, err := queryLister.List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, query := range queries.Items {
		// don't queue ourself
		if query.Name == generationQuery.Name {
			continue
		}
		// look at the list of dependencies
		depenencyNames := append(query.Spec.ReportQueries, query.Spec.DynamicReportQueries...)
		for _, dependency := range depenencyNames {
			if dependency == generationQuery.Name {
				// this query depends on the generationQuery passed in
				op.enqueueReportGenerationQuery(query)
				break
			}
		}
	}
	return nil
}
