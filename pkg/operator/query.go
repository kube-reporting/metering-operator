package operator

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/db"
	"github.com/operator-framework/operator-metering/pkg/operator/reporting"
	"github.com/operator-framework/operator-metering/pkg/operator/reportingutil"
	"github.com/operator-framework/operator-metering/pkg/presto"
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
	var viewName string
	createView := false
	if generationQuery.Spec.View.Disabled {
		logger.Infof("ReportGenerationQuery has spec.view.disabled=true, skipping view creation")
	} else if generationQuery.Status.ViewName == "" {
		logger.Infof("new ReportGenerationQuery discovered")
		viewName = reportingutil.GenerationQueryViewName(generationQuery.Namespace, generationQuery.Name)
		createView = true
	} else {
		logger.Infof("existing ReportGenerationQuery discovered, viewName: %s", generationQuery.Status.ViewName)
		viewName = generationQuery.Status.ViewName
	}

	queryDependencies, err := reporting.GetAndValidateGenerationQueryDependencies(
		reporting.NewReportGenerationQueryListerGetter(op.reportGenerationQueryLister),
		reporting.NewReportDataSourceListerGetter(op.reportDataSourceLister),
		reporting.NewReportListerGetter(op.reportLister),
		generationQuery,
		op.uninitialiedDependendenciesHandler(),
	)
	if err != nil {
		if reporting.IsUninitializedDependencyError(err) {
			logger.Warnf("unable to validate ReportGenerationQuery %s, has uninitialized dependencies: %v", generationQuery.Name, err)
			// We do not return an error because we do not need to requeue this
			// query. Instead we can wait until this queries uninitialized
			// dependencies become initialized. After they're initialized they
			// will queue anything that depends on them, including this query.
			return nil
		} else if reporting.IsInvalidDependencyError(err) {
			logger.WithError(err).Errorf("unable to validate ReportGenerationQuery %s, has invalid dependencies, dropping off queue", generationQuery.Name)
			// Invalid dependency means it will not resolve itself, so do not
			// return an error since we do not want to be requeued unless the
			// resource is modified, or it's dependencies are modified.
			return nil
		} else {
			// The error occurred when getting the dependencies or for an
			// unknown reason so we want to retry up to a limit. This most
			// commonly occurs when fetching a dependency from the API fails,
			// or if there is a cyclic dependency.
			return fmt.Errorf("unable to get or validate ReportGenerationQuery dependencies %s: %v", generationQuery.Name, err)
		}
	}

	if createView {
		tmplCtx := &reporting.ReportQueryTemplateContext{
			DynamicDependentQueries: queryDependencies.DynamicReportGenerationQueries,
			Report:                  nil,
		}
		renderedQuery, err := reporting.RenderQuery(generationQuery.Spec.Query, generationQuery.Namespace, tmplCtx)
		if err != nil {
			return err
		}

		logger.Infof("creating view %s", viewName)
		err = op.prestoViewCreator.CreateView(viewName, renderedQuery)
		if err != nil {
			return fmt.Errorf("error creating view %s for ReportGenerationQuery %s: %v", viewName, generationQuery.Name, err)
		}
		logger.Infof("created view %s", viewName)

		err = op.updateReportQueryViewName(logger, generationQuery, viewName)
		if err != nil {
			return err
		}
	}

	// enqueue any queries depending on this one
	if err := op.queueDependentReportGenerationQueriesForQuery(generationQuery); err != nil {
		logger.WithError(err).Errorf("error queuing ReportGenerationQuery dependents of ReportGenerationQuery %s", generationQuery.Name)
	}
	// enqueue any reports depending on this one
	if err := op.queueDependentReportsForQuery(generationQuery); err != nil {
		logger.WithError(err).Errorf("error queuing Report dependents of ReportGenerationQuery %s", generationQuery.Name)
	}

	return nil
}

func (op *Reporting) updateReportQueryViewName(logger log.FieldLogger, generationQuery *cbTypes.ReportGenerationQuery, viewName string) error {
	generationQuery.Status.ViewName = viewName
	_, err := op.meteringClient.MeteringV1alpha1().ReportGenerationQueries(generationQuery.Namespace).Update(generationQuery)
	if err != nil {
		logger.WithError(err).Errorf("failed to update ReportGenerationQuery view name for %q", generationQuery.Name)
		return err
	}
	return nil
}

func (op *Reporting) uninitialiedDependendenciesHandler() *reporting.UninitialiedDependendenciesHandler {
	return &reporting.UninitialiedDependendenciesHandler{
		HandleUninitializedReportGenerationQuery: op.enqueueReportGenerationQuery,
		HandleUninitializedReportDataSource:      op.enqueueReportDataSource,
	}
}

// queueDependentReportGenerationQueriesForQuery will queue all ReportGenerationQueries in the namespace which have a dependency on the generationQuery
func (op *Reporting) queueDependentReportGenerationQueriesForQuery(generationQuery *cbTypes.ReportGenerationQuery) error {
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
		// look at the list of ReportGenerationQuery dependencies
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

func (op *Reporting) queueDependentReportsForQuery(generationQuery *cbTypes.ReportGenerationQuery) error {
	reportLister := op.meteringClient.MeteringV1alpha1().Reports(generationQuery.Namespace)
	reports, err := reportLister.List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, report := range reports.Items {
		if report.Spec.GenerationQueryName == generationQuery.Name {
			op.enqueueReport(report)
		}
	}
	return nil
}

type PrestoViewCreator interface {
	CreateView(viewName, query string) error
}

type prestoViewCreator struct {
	queryer db.Queryer
}

func (c *prestoViewCreator) CreateView(viewName, query string) error {
	return presto.CreateView(c.queryer, viewName, query, true)
}
