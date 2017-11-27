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

	logger = logger.WithFields(newLogIdenti***REMOVED***er())
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
		logger.Infof("new generationQuery discovered")
		if generationQuery.Spec.View.Disabled {
			logger.Infof("generationQuery has spec.view.disabled=true, skipping processing")
			return nil
		}
		viewName = generationQueryViewName(generationQuery.Name)
	} ***REMOVED*** {
		logger.Infof("existing generationQuery discovered, viewName: %s", generationQuery.ViewName)
		viewName = generationQuery.ViewName
	}

	if valid, err := c.validateGenerationQuery(logger, generationQuery, true); err != nil {
		return err
	} ***REMOVED*** if !valid {
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
func (c *Chargeback) validateGenerationQuery(logger log.FieldLogger, generationQuery *cbTypes.ReportGenerationQuery, queueUninitialized bool) (bool, error) {
	generationQueries, err := c.getDependentGenerationQueries(generationQuery)
	if err != nil {
		return false, err
	}
	dataStores, err := c.getDependentDatastores(generationQuery)
	if err != nil {
		return false, err
	}
	if uninitializedQueries, err := c.getUninitializedReportGenerationQueries(generationQueries); err != nil {
		return false, err
	} ***REMOVED*** if len(uninitializedQueries) > 0 {
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

	if uninitializedDataStores := c.getUnitilizedDatastores(dataStores); len(uninitializedDataStores) > 0 {
		dataStoresStr := strings.Join(uninitializedDataStores, ", ")
		logger.Warnf("the following datastores for the query do not have their tables created %s", dataStoresStr)
		if queueUninitialized {
			logger.Debugf("queueing uninitializedDataStores: %s", dataStoresStr)
			for _, dataStore := range uninitializedDataStores {
				key, err := cache.MetaNamespaceKeyFunc(dataStore)
				if err == nil {
					c.informers.reportDataStoreQueue.Add(key)
				}
			}
		}
		return false, nil
	}
	return true, nil
}

func (c *Chargeback) getUnitilizedDatastores(dataStores []*cbTypes.ReportDataStore) []string {
	var uninitializedDataStores []string
	for _, dataStore := range dataStores {
		if dataStore.TableName == "" {
			uninitializedDataStores = append(uninitializedDataStores, dataStore.Name)
		}
	}
	return uninitializedDataStores
}

func (c *Chargeback) getUninitializedReportGenerationQueries(generationQueries []*cbTypes.ReportGenerationQuery) ([]string, error) {
	var uninitializedQueries, queriesWithDisabledView []string
	for _, query := range generationQueries {
		if query.ViewName == "" {
			if query.Spec.View.Disabled {
				queriesWithDisabledView = append(queriesWithDisabledView, query.Name)
			} ***REMOVED*** {
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

func (c *Chargeback) getDependentDatastores(generationQuery *cbTypes.ReportGenerationQuery) ([]*cbTypes.ReportDataStore, error) {
	dataStores := make([]*cbTypes.ReportDataStore, len(generationQuery.Spec.DataStores))
	for i, dataStoreName := range generationQuery.Spec.DataStores {
		dataStore, err := c.informers.reportDataStoreLister.ReportDataStores(generationQuery.Namespace).Get(dataStoreName)
		if err != nil {
			return nil, err
		}
		dataStores[i] = dataStore
	}
	return dataStores, nil
}
