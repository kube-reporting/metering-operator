package operator

import (
	"fmt"
	"strings"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
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

	logger = logger.WithFields(newLogIdentifier(op.rand))
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

	reportGenerationQuery, err := op.informers.Metering().V1alpha1().ReportGenerationQueries().Lister().ReportGenerationQueries(namespace).Get(name)
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
	} else {
		logger.Infof("existing reportGenerationQuery discovered, viewName: %s", generationQuery.ViewName)
		viewName = generationQuery.ViewName
	}

	if valid, err := op.validateGenerationQuery(logger, generationQuery, true); err != nil {
		return err
	} else if !valid {
		logger.Warnf("cannot create view for reportGenerationQuery, it has uninitialized dependencies")
		return nil
	}

	dependentQueries, err := op.getDependentGenerationQueries(generationQuery, true)
	if err != nil {
		return err
	}
	templateInfo := &templateInfo{
		DynamicDependentQueries: dependentQueries,
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

// validateGenerationQuery takes a generationQuery and checks if all of it's
// dependencies have been initialized. If the generationQuery has a dependency
// on another generationQuery with spec.view.disabled, the validation will
// return an error. Returns true if there are no invalid dependencies and all
// dependencies have a viewName or tableName set in the custom resource.
// Returns false if there is a dependency that is uninitialized.
func (op *Reporting) validateGenerationQuery(logger log.FieldLogger, generationQuery *cbTypes.ReportGenerationQuery, queueUninitialized bool) (bool, error) {
	// Validate ReportGenerationQuery's that should be views
	generationQueries, err := op.getDependentGenerationQueries(generationQuery, false)
	if err != nil {
		return false, err
	}

	// Validate dynamic generationQuery dependencies
	_, err = op.getDependentGenerationQueries(generationQuery, true)
	if err != nil {
		return false, err
	}

	dataSources, err := op.getDependentDataSources(generationQuery)
	if err != nil {
		return false, err
	}
	if uninitializedQueries, err := op.getUninitializedReportGenerationQueries(generationQueries); err != nil {
		return false, err
	} else if len(uninitializedQueries) > 0 {
		queriesStr := strings.Join(uninitializedQueries, ", ")
		logger.Warnf("the following ReportGenerationQueries for the query do not have their views created %s", queriesStr)
		if queueUninitialized {
			logger.Debugf("queueing uninitializedQueries: %s", queriesStr)
			for _, query := range uninitializedQueries {
				key, err := cache.MetaNamespaceKeyFunc(query)
				if err == nil {
					op.queues.reportGenerationQueryQueue.Add(key)
				}
			}
		}
		return false, nil
	}

	if uninitializedDataSources := op.getUnitilizedDataSources(dataSources); len(uninitializedDataSources) > 0 {
		dataSourcesStr := strings.Join(uninitializedDataSources, ", ")
		logger.Warnf("the following datasources for the query do not have their tables created %s", dataSourcesStr)
		if queueUninitialized {
			logger.Debugf("queueing uninitializedDataSources: %s", dataSourcesStr)
			for _, dataSource := range uninitializedDataSources {
				key, err := cache.MetaNamespaceKeyFunc(dataSource)
				if err == nil {
					op.queues.reportDataSourceQueue.Add(key)
				}
			}
		}
		return false, nil
	}
	return true, nil
}

func (op *Reporting) getUnitilizedDataSources(dataSources []*cbTypes.ReportDataSource) []string {
	var uninitializedDataSources []string
	for _, dataSource := range dataSources {
		if dataSource.TableName == "" {
			uninitializedDataSources = append(uninitializedDataSources, dataSource.Name)
		}
	}
	return uninitializedDataSources
}

func (op *Reporting) getUninitializedReportGenerationQueries(generationQueries []*cbTypes.ReportGenerationQuery) ([]string, error) {
	var uninitializedQueries, queriesWithDisabledView []string
	for _, query := range generationQueries {
		if query.ViewName == "" {
			if query.Spec.View.Disabled {
				queriesWithDisabledView = append(queriesWithDisabledView, query.Name)
			} else {
				uninitializedQueries = append(uninitializedQueries, query.Name)
			}
		}
	}

	if len(queriesWithDisabledView) > 0 {
		return nil, fmt.Errorf("invalid ReportGenerationQuery, references ReportGenerationQueries with spec.view.disabled=true: %s", strings.Join(queriesWithDisabledView, ", "))
	}
	return uninitializedQueries, nil
}

func (op *Reporting) getDependentGenerationQueries(generationQuery *cbTypes.ReportGenerationQuery, dynamicQueries bool) ([]*cbTypes.ReportGenerationQuery, error) {
	queriesAccumulator := make(map[string]*cbTypes.ReportGenerationQuery)
	const maxDepth = 100
	err := op.getDependentGenerationQueriesMemoized(generationQuery, 0, maxDepth, queriesAccumulator, dynamicQueries)
	if err != nil {
		return nil, err
	}
	queries := make([]*cbTypes.ReportGenerationQuery, 0, len(queriesAccumulator))
	for _, query := range queriesAccumulator {
		queries = append(queries, query)
	}
	return queries, nil
}

func (op *Reporting) getDependentGenerationQueriesMemoized(generationQuery *cbTypes.ReportGenerationQuery, depth, maxDepth int, queriesAccumulator map[string]*cbTypes.ReportGenerationQuery, dynamicQueries bool) error {
	if depth >= maxDepth {
		return fmt.Errorf("detected a cycle at depth %d for generationQuery %s", depth, generationQuery.Name)
	}
	var queries []string
	if dynamicQueries {
		queries = generationQuery.Spec.DynamicReportQueries
	} else {
		queries = generationQuery.Spec.ReportQueries
	}
	for _, queryName := range queries {
		if _, exists := queriesAccumulator[queryName]; exists {
			continue
		}
		genQuery, err := op.informers.Metering().V1alpha1().ReportGenerationQueries().Lister().ReportGenerationQueries(op.cfg.Namespace).Get(queryName)
		if err != nil {
			return err
		}
		err = op.getDependentGenerationQueriesMemoized(genQuery, depth+1, maxDepth, queriesAccumulator, dynamicQueries)
		if err != nil {
			return err
		}
		queriesAccumulator[genQuery.Name] = genQuery
	}
	return nil
}

func (op *Reporting) getDependentDataSources(generationQuery *cbTypes.ReportGenerationQuery) ([]*cbTypes.ReportDataSource, error) {
	dataSources := make([]*cbTypes.ReportDataSource, len(generationQuery.Spec.DataSources))
	for i, dataSourceName := range generationQuery.Spec.DataSources {
		dataSource, err := op.informers.Metering().V1alpha1().ReportDataSources().Lister().ReportDataSources(op.cfg.Namespace).Get(dataSourceName)
		if err != nil {
			return nil, err
		}
		dataSources[i] = dataSource
	}
	return dataSources, nil
}
