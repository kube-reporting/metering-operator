package chargeback

import (
	"fmt"
	"strings"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
)

func (c *Chargeback) runReportGenerationQueryWorker() {
	logger := c.logger.WithField("component", "reportGenerationQueryWorker")
	logger.Infof("ReportGenerationQuery worker started")
	for c.processReportGenerationQuery(logger) {

	}
}

func (c *Chargeback) processReportGenerationQuery(logger log.FieldLogger) bool {
	key, quit := c.informers.reportGenerationQueryQueue.Get()
	if quit {
		return false
	}
	defer c.informers.reportGenerationQueryQueue.Done(key)

	logger = logger.WithFields(c.newLogIdentifier())
	err := c.syncReportGenerationQuery(logger, key.(string))
	c.handleErr(logger, err, "ReportGenerationQuery", key, c.informers.reportGenerationQueryQueue)
	return true
}

func (c *Chargeback) syncReportGenerationQuery(logger log.FieldLogger, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.WithError(err).Errorf("invalid resource key :%s", key)
		return nil
	}

	logger = logger.WithField("generationQuery", name)
	reportGenerationQuery, err := c.informers.reportGenerationQueryLister.ReportGenerationQueries(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("ReportGenerationQuery %s does not exist anymore", key)
			return nil
		}
		return err
	}

	logger.Infof("syncing reportGenerationQuery %s", reportGenerationQuery.GetName())
	err = c.handleReportGenerationQuery(logger, reportGenerationQuery)
	if err != nil {
		logger.WithError(err).Errorf("error syncing reportGenerationQuery %s", reportGenerationQuery.GetName())
		return err
	}
	logger.Infof("successfully synced reportGenerationQuery %s", reportGenerationQuery.GetName())
	return nil
}

func (c *Chargeback) handleReportGenerationQuery(logger log.FieldLogger, generationQuery *cbTypes.ReportGenerationQuery) error {
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

	if valid, err := c.validateGenerationQuery(logger, generationQuery, true); err != nil {
		return err
	} else if !valid {
		logger.Warnf("cannot create view for reportGenerationQuery, it has uninitialized dependencies")
		return nil
	}

	renderedQuery, err := renderGenerationQuery(generationQuery)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("CREATE OR REPLACE VIEW %s AS %s", viewName, renderedQuery)
	_, err = c.prestoConn.Query(query)
	if err != nil {
		return err
	}

	return c.updateReportQueryViewName(logger, generationQuery, viewName)
}

func (c *Chargeback) updateReportQueryViewName(logger log.FieldLogger, generationQuery *cbTypes.ReportGenerationQuery, viewName string) error {
	generationQuery.ViewName = viewName
	_, err := c.chargebackClient.ChargebackV1alpha1().ReportGenerationQueries(generationQuery.Namespace).Update(generationQuery)
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
func (c *Chargeback) validateGenerationQuery(logger log.FieldLogger, generationQuery *cbTypes.ReportGenerationQuery, queueUninitialized bool) (bool, error) {
	generationQueries, err := c.getDependentGenerationQueries(generationQuery)
	if err != nil {
		return false, err
	}
	dataSources, err := c.getDependentDataSources(generationQuery)
	if err != nil {
		return false, err
	}
	if uninitializedQueries, err := c.getUninitializedReportGenerationQueries(generationQueries); err != nil {
		return false, err
	} else if len(uninitializedQueries) > 0 {
		queriesStr := strings.Join(uninitializedQueries, ", ")
		logger.Warnf("the following ReportGenerationQueries for the query do not have their views created %s", queriesStr)
		if queueUninitialized {
			logger.Debugf("queueing uninitializedQueries: %s", queriesStr)
			for _, query := range uninitializedQueries {
				key, err := cache.MetaNamespaceKeyFunc(query)
				if err == nil {
					c.informers.reportGenerationQueryQueue.Add(key)
				}
			}
		}
		return false, nil
	}

	if uninitializedDataSources := c.getUnitilizedDataSources(dataSources); len(uninitializedDataSources) > 0 {
		dataSourcesStr := strings.Join(uninitializedDataSources, ", ")
		logger.Warnf("the following datasources for the query do not have their tables created %s", dataSourcesStr)
		if queueUninitialized {
			logger.Debugf("queueing uninitializedDataSources: %s", dataSourcesStr)
			for _, dataSource := range uninitializedDataSources {
				key, err := cache.MetaNamespaceKeyFunc(dataSource)
				if err == nil {
					c.informers.reportDataSourceQueue.Add(key)
				}
			}
		}
		return false, nil
	}
	return true, nil
}

func (c *Chargeback) getUnitilizedDataSources(dataSources []*cbTypes.ReportDataSource) []string {
	var uninitializedDataSources []string
	for _, dataSource := range dataSources {
		if dataSource.TableName == "" {
			uninitializedDataSources = append(uninitializedDataSources, dataSource.Name)
		}
	}
	return uninitializedDataSources
}

func (c *Chargeback) getUninitializedReportGenerationQueries(generationQueries []*cbTypes.ReportGenerationQuery) ([]string, error) {
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

func (c *Chargeback) getDependentGenerationQueries(generationQuery *cbTypes.ReportGenerationQuery) ([]*cbTypes.ReportGenerationQuery, error) {
	queriesAccumulator := make(map[string]*cbTypes.ReportGenerationQuery)
	err := c.getDependentGenerationQueriesMemoized(generationQuery, queriesAccumulator)
	if err != nil {
		return nil, err
	}
	queries := make([]*cbTypes.ReportGenerationQuery, 0, len(queriesAccumulator))
	for _, query := range queries {
		queries = append(queries, query)
	}
	return queries, nil
}

func (c *Chargeback) getDependentGenerationQueriesMemoized(generationQuery *cbTypes.ReportGenerationQuery, queriesAccumulator map[string]*cbTypes.ReportGenerationQuery) error {
	for _, queryName := range generationQuery.Spec.ReportQueries {
		if _, exists := queriesAccumulator[queryName]; exists {
			continue
		}
		genQuery, err := c.informers.reportGenerationQueryLister.ReportGenerationQueries(generationQuery.Namespace).Get(queryName)
		if err != nil {
			return err
		}
		err = c.getDependentGenerationQueriesMemoized(genQuery, queriesAccumulator)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Chargeback) getDependentDataSources(generationQuery *cbTypes.ReportGenerationQuery) ([]*cbTypes.ReportDataSource, error) {
	dataSources := make([]*cbTypes.ReportDataSource, len(generationQuery.Spec.DataSources))
	for i, dataSourceName := range generationQuery.Spec.DataSources {
		dataSource, err := c.informers.reportDataSourceLister.ReportDataSources(generationQuery.Namespace).Get(dataSourceName)
		if err != nil {
			return nil, err
		}
		dataSources[i] = dataSource
	}
	return dataSources, nil
}
