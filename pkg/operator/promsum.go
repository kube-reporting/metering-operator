package operator

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/metering/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/operator/prestostore"
)

const (
	// Keep a cap on the number of time ranges we query per reconciliation.
	// If we get to defaultMaxPromTimeRanges, it means we're very backlogged,
	// or we have a small chunkSize and making tons of small queries all one
	// after another will cause undesired resource spikes, or both.  This will
	// make it take longer to catch up, but should help prevent memory from
	// exploding when we end up with a ton of time ranges.

	// defaultMaxPromTimeRanges is the number of time ranges for 24 hours if we
	// query in 5 minute chunks (the default).
	defaultMaxPromTimeRanges = (24 * 60) / 5 // 24 hours, 60 minutes per hour, default chunkSize is 5 minutes

	defaultMaxTimeDuration = 24 * time.Hour
)

var (
	prometheusReportDatasourceLabels = []string{
		"reportdatasource",
		"reportprometheusquery",
		"table_name",
	}

	prometheusReportDatasourceMetricsScrapedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "metering",
			Name:      "prometheus_reportdatasource_metrics_scraped_total",
			Help:      "Number of Prometheus metrics returned by a PrometheusQuery for a ReportDataSource.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceMetricsImportedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "metering",
			Name:      "prometheus_reportdatasource_metrics_imported_total",
			Help:      "Number of Prometheus ReportDatasource metrics imported.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceTotalImportsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "metering",
			Name:      "prometheus_reportdatasource_imports_total",
			Help:      "Number of Prometheus ReportDatasource metrics imports.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceFailedImportsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "metering",
			Name:      "prometheus_reportdatasource_failed_imports_total",
			Help:      "Number of failed Prometheus ReportDatasource metrics imports.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceTotalPrometheusQueriesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "metering",
			Name:      "prometheus_reportdatasource_prometheus_queries_total",
			Help:      "Number of Prometheus ReportDatasource Prometheus queries made for the ReportDataSource since start up.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceFailedPrometheusQueriesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "metering",
			Name:      "prometheus_reportdatasource_failed_prometheus_queries_total",
			Help:      "Number of failed Prometheus ReportDatasource Prometheus queries made for the ReportDataSource since start up.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceTotalPrestoStoresCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "metering",
			Name:      "prometheus_reportdatasource_presto_stores_total",
			Help:      "Number of Prometheus ReportDatasource calls to store all metrics collected into Presto.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceFailedPrestoStoresCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "metering",
			Name:      "prometheus_reportdatasource_failed_presto_stores_total",
			Help:      "Number of failed Prometheus ReportDatasource calls to store all metrics collected into Presto.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceImportDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "metering",
			Name:      "prometheus_reportdatasource_import_duration_seconds",
			Help:      "Duration to import Prometheus metrics into Presto.",
			Buckets:   []float64{30.0, 60.0, 300.0},
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourcePrometheusQueryDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "metering",
			Name:      "prometheus_reportdatasource_prometheus_query_duration_seconds",
			Help:      "Duration for a Prometheus query to return metrics to reporting-operator.",
			Buckets:   []float64{2.0, 10.0, 30.0, 60.0},
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourcePrestoreStoreDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "metering",
			Name:      "prometheus_reportdatasource_presto_store_duration_seconds",
			Help:      "Duration to store all metrics fetched into Presto.",
			Buckets:   []float64{2.0, 10.0, 30.0, 60.0, 300.0},
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceRunningImportsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "metering",
			Name:      "prometheus_reportdatasource_running_imports",
			Help:      "Number of Prometheus ReportDatasource imports currently running.",
		},
	)
)

func init() {
	prometheus.MustRegister(prometheusReportDatasourceMetricsScrapedCounter)
	prometheus.MustRegister(prometheusReportDatasourceMetricsImportedCounter)
	prometheus.MustRegister(prometheusReportDatasourceTotalImportsCounter)
	prometheus.MustRegister(prometheusReportDatasourceFailedImportsCounter)
	prometheus.MustRegister(prometheusReportDatasourceTotalPrometheusQueriesCounter)
	prometheus.MustRegister(prometheusReportDatasourceFailedPrometheusQueriesCounter)
	prometheus.MustRegister(prometheusReportDatasourceTotalPrestoStoresCounter)
	prometheus.MustRegister(prometheusReportDatasourceFailedPrestoStoresCounter)
	prometheus.MustRegister(prometheusReportDatasourceImportDurationHistogram)
	prometheus.MustRegister(prometheusReportDatasourcePrometheusQueryDurationHistogram)
	prometheus.MustRegister(prometheusReportDatasourcePrestoreStoreDurationHistogram)
	prometheus.MustRegister(prometheusReportDatasourceRunningImportsGauge)
}

func (op *Reporting) runPrometheusImporterWorker(stopCh <-chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// run a go routine that waits for the stopCh to be closed and propagates
	// the shutdown to the collectors by calling cancel()
	go func() {
		<-stopCh
		// if the stopCh is closed while we're waiting, cancel and wait for
		// everything to return
		cancel()
	}()
	op.startPrometheusImporter(ctx)
}

type prometheusImporterFunc func(ctx context.Context, start, end time.Time) ([]*prometheusImportResults, error)

type prometheusImportResults struct {
	ReportDataSource     string `json:"reportDataSource"`
	MetricsImportedCount int    `json:"metricsImportedCount"`
}

type prometheusImporterTimeRangeTriggerResult struct {
	err           error
	importResults []*prometheusImportResults
}

type prometheusImporterTimeRangeTrigger struct {
	start, end time.Time
	result     chan prometheusImporterTimeRangeTriggerResult
}

func (op *Reporting) triggerPrometheusImporterForTimeRange(ctx context.Context, start, end time.Time) ([]*prometheusImportResults, error) {
	resultCh := make(chan prometheusImporterTimeRangeTriggerResult)
	select {
	case op.prometheusImporterTriggerForTimeRangeCh <- prometheusImporterTimeRangeTrigger{start, end, resultCh}:
		result := <-resultCh
		return result.importResults, result.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (op *Reporting) startPrometheusImporter(ctx context.Context) {
	logger := op.logger.WithField("component", "PrometheusImporter")
	logger.Infof("PrometheusImporter worker started")
	workers := make(map[string]*prometheusImporterWorker)
	importers := make(map[string]*prestostore.PrometheusImporter)

	const concurrency = 4
	// create a channel to act as a semaphore to limit the number of
	// imports happening in parallel
	semaphore := make(chan struct{}, concurrency)

	defer logger.Infof("PrometheusImporterWorker shutdown")

	if op.cfg.DisablePromsum {
		logger.Infof("Periodic Prometheus ReportDataSource importing disabled")
	}

	for {
		select {
		case <-ctx.Done():
			logger.Infof("got shutdown signal, shutting down PrometheusImporters")
			return
		case trigger := <-op.prometheusImporterTriggerForTimeRangeCh:
			// manually triggered import for a speci***REMOVED***c time range, usually from HTTP API

			var results []*prometheusImportResults
			resultsCh := make(chan *prometheusImportResults)

			go func() {
				for importResults := range resultsCh {
					results = append(results, importResults)
				}
			}()

			g, ctx := errgroup.WithContext(ctx)
			for dataSourceName, importer := range importers {
				importer := importer
				dataSourceName := dataSourceName
				// collect each dataSource concurrently
				g.Go(func() error {
					importResults, err := importPrometheusDataSourceData(ctx, logger, semaphore, dataSourceName, importer, func(ctx context.Context, importer *prestostore.PrometheusImporter) (*prestostore.PrometheusImportResults, error) {
						return importer.ImportMetrics(ctx, trigger.start, trigger.end, true)
					})
					resultsCh <- &prometheusImportResults{
						ReportDataSource:     dataSourceName,
						MetricsImportedCount: len(importResults.Metrics),
					}
					return err
				})

			}
			err := g.Wait()
			if err != nil {
				logger.WithError(err).Errorf("PrometheusImporter worker encountered errors while importing data")
			}
			close(resultsCh)
			trigger.result <- prometheusImporterTimeRangeTriggerResult{
				err:           err,
				importResults: results,
			}
		case dataSourceName := <-op.prometheusImporterDeletedDataSourceQueue:
			// if we have a worker for this ReportDataSource then we need to
			// stop it and remove it from our map
			if worker, exists := workers[dataSourceName]; exists {
				worker.stop()
				delete(workers, dataSourceName)
			}
			if _, exists := importers[dataSourceName]; exists {
				delete(importers, dataSourceName)
			}
		case reportDataSource := <-op.prometheusImporterNewDataSourceQueue:
			if reportDataSource.Spec.Promsum == nil {
				logger.Error("expected only Promsum ReportDataSources")
				continue
			}

			dataSourceName := reportDataSource.Name
			queryName := reportDataSource.Spec.Promsum.Query
			tableName := dataSourceTableName(dataSourceName)

			reportPromQuery, err := op.informers.Metering().V1alpha1().ReportPrometheusQueries().Lister().ReportPrometheusQueries(reportDataSource.Namespace).Get(queryName)
			if err != nil {
				op.logger.WithError(err).Errorf("unable to ReportPrometheusQuery %s for ReportDataSource %s", queryName, dataSourceName)
				continue
			}

			dataSourceLogger := logger.WithFields(logrus.Fields{
				"queryName":        queryName,
				"reportDataSource": dataSourceName,
				"tableName":        tableName,
			})

			importer, exists := importers[dataSourceName]
			if exists {
				dataSourceLogger.Debugf("ReportDataSource %s already has an importer, updating con***REMOVED***guration", dataSourceName)
				cfg := op.newPromImporterCfg(reportDataSource, reportPromQuery)
				importer.UpdateCon***REMOVED***g(cfg)
			} ***REMOVED*** {
				importer = op.newPromImporter(dataSourceLogger, reportDataSource, reportPromQuery)
				importers[dataSourceName] = importer
			}

			if !op.cfg.DisablePromsum {
				worker, workerExists := workers[dataSourceName]
				queryInterval := op.getQueryIntervalForReportDataSource(reportDataSource)
				if workerExists && worker.queryInterval != queryInterval {
					// queryInterval changed stop the existing worker from
					// collecting data, and create it with updated con***REMOVED***g
					worker.stop()
				} ***REMOVED*** if workerExists {
					// con***REMOVED***g hasn't changed skip the update
					continue
				}

				worker = newPromImportWorker(queryInterval)
				workers[dataSourceName] = worker

				// launch a go routine that periodically triggers a collection
				go worker.start(ctx, dataSourceLogger, semaphore, dataSourceName, importer)
			}
		}
	}
}

func (op *Reporting) getQueryIntervalForReportDataSource(reportDataSource *cbTypes.ReportDataSource) time.Duration {
	queryConf := reportDataSource.Spec.Promsum.QueryCon***REMOVED***g
	queryInterval := op.cfg.PrometheusQueryCon***REMOVED***g.QueryInterval.Duration
	if queryConf != nil {
		if queryConf.QueryInterval != nil {
			queryInterval = queryConf.QueryInterval.Duration
		}
	}
	return queryInterval
}

func (op *Reporting) newPromImporterCfg(reportDataSource *cbTypes.ReportDataSource, reportPromQuery *cbTypes.ReportPrometheusQuery) prestostore.Con***REMOVED***g {
	dataSourceName := reportDataSource.Name
	tableName := dataSourceTableName(dataSourceName)

	chunkSize := op.cfg.PrometheusQueryCon***REMOVED***g.ChunkSize.Duration
	stepSize := op.cfg.PrometheusQueryCon***REMOVED***g.StepSize.Duration

	queryConf := reportDataSource.Spec.Promsum.QueryCon***REMOVED***g
	if queryConf != nil {
		if queryConf.ChunkSize != nil {
			chunkSize = queryConf.ChunkSize.Duration
		}
		if queryConf.StepSize != nil {
			stepSize = queryConf.StepSize.Duration
		}
	}

	// round to the nearest second for chunk/step sizes
	chunkSize = chunkSize.Truncate(time.Second)
	stepSize = stepSize.Truncate(time.Second)

	return prestostore.Con***REMOVED***g{
		PrometheusQuery:       reportPromQuery.Spec.Query,
		PrestoTableName:       tableName,
		ChunkSize:             chunkSize,
		StepSize:              stepSize,
		MaxTimeRanges:         defaultMaxPromTimeRanges,
		MaxQueryRangeDuration: defaultMaxTimeDuration,
	}
}

func (op *Reporting) newPromImporter(logger logrus.FieldLogger, reportDataSource *cbTypes.ReportDataSource, reportPromQuery *cbTypes.ReportPrometheusQuery) *prestostore.PrometheusImporter {
	cfg := op.newPromImporterCfg(reportDataSource, reportPromQuery)

	promLabels := prometheus.Labels{
		"reportdatasource":      reportDataSource.Name,
		"reportprometheusquery": reportPromQuery.Name,
		"table_name":            cfg.PrestoTableName,
	}

	totalImportsCounter := prometheusReportDatasourceTotalImportsCounter.With(promLabels)
	failedImportsCounter := prometheusReportDatasourceFailedImportsCounter.With(promLabels)

	totalPrometheusQueriesCounter := prometheusReportDatasourceTotalPrometheusQueriesCounter.With(promLabels)
	failedPrometheusQueriesCounter := prometheusReportDatasourceFailedPrometheusQueriesCounter.With(promLabels)

	totalPrestoStoresCounter := prometheusReportDatasourceTotalPrestoStoresCounter.With(promLabels)
	failedPrestoStoresCounter := prometheusReportDatasourceFailedPrestoStoresCounter.With(promLabels)

	promQueryMetricsScrapedCounter := prometheusReportDatasourceMetricsScrapedCounter.With(promLabels)
	promQueryDurationHistogram := prometheusReportDatasourcePrometheusQueryDurationHistogram.With(promLabels)

	metricsImportedCounter := prometheusReportDatasourceMetricsImportedCounter.With(promLabels)
	importDurationHistogram := prometheusReportDatasourceImportDurationHistogram.With(promLabels)

	prestoStoreDurationHistogram := prometheusReportDatasourcePrestoreStoreDurationHistogram.With(promLabels)

	metricsCollectors := prestostore.ImporterMetricsCollectors{
		TotalImportsCounter:     totalImportsCounter,
		FailedImportsCounter:    failedImportsCounter,
		ImportDurationHistogram: importDurationHistogram,

		TotalPrometheusQueriesCounter:    totalPrometheusQueriesCounter,
		FailedPrometheusQueriesCounter:   failedPrometheusQueriesCounter,
		PrometheusQueryDurationHistogram: promQueryDurationHistogram,

		TotalPrestoStoresCounter:     totalPrestoStoresCounter,
		FailedPrestoStoresCounter:    failedPrestoStoresCounter,
		PrestoStoreDurationHistogram: prestoStoreDurationHistogram,

		MetricsScrapedCounter:  promQueryMetricsScrapedCounter,
		MetricsImportedCounter: metricsImportedCounter,
	}

	return prestostore.NewPrometheusImporter(logger, op.promConn, op.prestoQueryer, op.clock, cfg, metricsCollectors)
}

type prometheusImporterWorker struct {
	stopCh        chan struct{}
	doneCh        chan struct{}
	queryInterval time.Duration
}

func newPromImportWorker(queryInterval time.Duration) *prometheusImporterWorker {
	return &prometheusImporterWorker{
		queryInterval: queryInterval,
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
	}
}

// start begins periodic importing with the con***REMOVED***gured importer.
func (w *prometheusImporterWorker) start(ctx context.Context, logger logrus.FieldLogger, semaphore chan struct{}, dataSourceName string, importer *prestostore.PrometheusImporter) {
	ticker := time.NewTicker(w.queryInterval)
	defer close(w.doneCh)
	defer ticker.Stop()

	logger.Infof("Importing data for ReportDataSource %s every %s", dataSourceName, w.queryInterval)
	for {
		select {
		case <-w.stopCh:
			return
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			_, err := importPrometheusDataSourceData(ctx, logger, semaphore, dataSourceName, importer, func(ctx context.Context, importer *prestostore.PrometheusImporter) (*prestostore.PrometheusImportResults, error) {
				return importer.ImportFromLastTimestamp(ctx, false)
			})
			if err != nil {
				logger.WithError(err).Errorf("error collecting Prometheus DataSource data")
			}
		case <-ctx.Done():
			return
		}
	}
}

func (w *prometheusImporterWorker) stop() {
	close(w.stopCh)
	<-w.doneCh
}

type importFunc func(context.Context, *prestostore.PrometheusImporter) (*prestostore.PrometheusImportResults, error)

func importPrometheusDataSourceData(ctx context.Context, logger logrus.FieldLogger, semaphore chan struct{}, dataSourceName string, prometheusImporter *prestostore.PrometheusImporter, runImport importFunc) (*prestostore.PrometheusImportResults, error) {
	// blocks trying to increment the semaphore (sending on the
	// channel) or until the context is cancelled
	select {
	case semaphore <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	dataSourceLogger := logger.WithField("reportDataSource", dataSourceName)
	// decrement the semaphore at the end
	defer func() {
		dataSourceLogger.Infof("***REMOVED***nished import for Prometheus ReportDataSource %s", dataSourceName)
		prometheusReportDatasourceRunningImportsGauge.Dec()
		<-semaphore
	}()
	dataSourceLogger.Infof("starting import for Prometheus ReportDataSource %s", dataSourceName)
	prometheusReportDatasourceRunningImportsGauge.Inc()
	return runImport(ctx, prometheusImporter)
}
