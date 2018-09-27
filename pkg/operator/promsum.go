package operator

import (
	"context"
	"time"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	prometheusMetricNamespace = "metering"
)

var (
	prometheusReportDatasourceLabels = []string{
		"reportdatasource",
		"reportprometheusquery",
		"table_name",
	}

	prometheusReportDatasourceMetricsScrapedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "prometheus_reportdatasource_metrics_scraped_total",
			Help:      "Number of Prometheus metrics returned by a PrometheusQuery for a ReportDataSource.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceMetricsImportedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "prometheus_reportdatasource_metrics_imported_total",
			Help:      "Number of Prometheus ReportDatasource metrics imported.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceTotalImportsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "prometheus_reportdatasource_imports_total",
			Help:      "Number of Prometheus ReportDatasource metrics imports.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceFailedImportsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "prometheus_reportdatasource_failed_imports_total",
			Help:      "Number of failed Prometheus ReportDatasource metrics imports.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceTotalPrometheusQueriesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "prometheus_reportdatasource_prometheus_queries_total",
			Help:      "Number of Prometheus ReportDatasource Prometheus queries made for the ReportDataSource since start up.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceFailedPrometheusQueriesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "prometheus_reportdatasource_failed_prometheus_queries_total",
			Help:      "Number of failed Prometheus ReportDatasource Prometheus queries made for the ReportDataSource since start up.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceTotalPrestoStoresCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "prometheus_reportdatasource_presto_stores_total",
			Help:      "Number of Prometheus ReportDatasource calls to store all metrics collected into Presto.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceFailedPrestoStoresCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "prometheus_reportdatasource_failed_presto_stores_total",
			Help:      "Number of failed Prometheus ReportDatasource calls to store all metrics collected into Presto.",
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceImportDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "prometheus_reportdatasource_import_duration_seconds",
			Help:      "Duration to import Prometheus metrics into Presto.",
			Buckets:   []float64{30.0, 60.0, 300.0},
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourcePrometheusQueryDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "prometheus_reportdatasource_prometheus_query_duration_seconds",
			Help:      "Duration for a Prometheus query to return metrics to reporting-operator.",
			Buckets:   []float64{2.0, 10.0, 30.0, 60.0},
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourcePrestoreStoreDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: prometheusMetricNamespace,
			Name:      "prometheus_reportdatasource_presto_store_duration_seconds",
			Help:      "Duration to store all metrics fetched into Presto.",
			Buckets:   []float64{2.0, 10.0, 30.0, 60.0, 300.0},
		},
		prometheusReportDatasourceLabels,
	)

	prometheusReportDatasourceRunningImportsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: prometheusMetricNamespace,
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

type prometheusImporterFunc func(ctx context.Context, start, end time.Time) ([]*prometheusImportResults, error)

type prometheusImportResults struct {
	ReportDataSource     string `json:"reportDataSource"`
	MetricsImportedCount int    `json:"metricsImportedCount"`
}

func (op *Reporting) importPrometheusForTimeRange(ctx context.Context, start, end time.Time) ([]*prometheusImportResults, error) {
	reportDataSources, err := op.meteringClient.MeteringV1alpha1().ReportDataSources(op.cfg.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	logger := op.logger.WithField("component", "importPrometheusForTimeRange")
	const concurrency = 4
	// create a channel to act as a semaphore to limit the number of
	// imports happening in parallel
	semaphore := make(chan struct{}, concurrency)

	resultsCh := make(chan *prometheusImportResults)
	g, ctx := errgroup.WithContext(ctx)

	for _, reportDataSource := range reportDataSources.Items {
		reportDataSource := reportDataSource
		if reportDataSource.Spec.Promsum == nil {
			continue
		}

		// collect each dataSource concurrently
		g.Go(func() error {
			reportPromQuery, err := op.meteringClient.MeteringV1alpha1().ReportPrometheusQueries(reportDataSource.Namespace).Get(reportDataSource.Spec.Promsum.Query, metav1.GetOptions{})
			if err != nil {
				return err
			}

			dataSourceLogger := logger.WithFields(logrus.Fields{
				"queryName":        reportDataSource.Spec.Promsum.Query,
				"reportDataSource": reportDataSource.Name,
				"tableName":        dataSourceTableName(reportDataSource.Name),
			})
			importCfg := op.newPromImporterCfg(reportDataSource, reportPromQuery)
			metricsCollectors := op.newPromImporterMetricsCollectors(reportDataSource, reportPromQuery)

			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				return ctx.Err()
			}
			// decrement the semaphore at the end
			defer func() {
				<-semaphore
			}()

			var promConn prom.API
			if (reportDataSource.Spec.Promsum.PrometheusConfig != nil) && (reportDataSource.Spec.Promsum.PrometheusConfig.URL != "") {
				promConn, err = op.newPrometheusConnFromURL(reportDataSource.Spec.Promsum.PrometheusConfig.URL)
				if err != nil {
					return err
				}
			} else {
				promConn = op.promConn
			}

			importResults, err := prestostore.ImportFromTimeRange(dataSourceLogger, op.clock, promConn, op.prestoQueryer, metricsCollectors, ctx, start, end, importCfg, true)
			if err != nil {
				return err
			}
			resultsCh <- &prometheusImportResults{
				ReportDataSource:     reportDataSource.Name,
				MetricsImportedCount: len(importResults.Metrics),
			}
			return nil
		})
	}

	go func() {
		g.Wait()
		close(resultsCh)
	}()

	var results []*prometheusImportResults
	for importResults := range resultsCh {
		results = append(results, importResults)
	}

	return results, g.Wait()
}

func (op *Reporting) getQueryIntervalForReportDataSource(reportDataSource *cbTypes.ReportDataSource) time.Duration {
	queryConf := reportDataSource.Spec.Promsum.QueryConfig
	queryInterval := op.cfg.PrometheusQueryConfig.QueryInterval.Duration
	if queryConf != nil {
		if queryConf.QueryInterval != nil {
			queryInterval = queryConf.QueryInterval.Duration
		}
	}
	return queryInterval
}

func (op *Reporting) newPromImporterCfg(reportDataSource *cbTypes.ReportDataSource, reportPromQuery *cbTypes.ReportPrometheusQuery) prestostore.Config {
	dataSourceName := reportDataSource.Name
	tableName := dataSourceTableName(dataSourceName)

	chunkSize := op.cfg.PrometheusQueryConfig.ChunkSize.Duration
	stepSize := op.cfg.PrometheusQueryConfig.StepSize.Duration

	queryConf := reportDataSource.Spec.Promsum.QueryConfig
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

	return prestostore.Config{
		PrometheusQuery:       reportPromQuery.Spec.Query,
		PrestoTableName:       tableName,
		ChunkSize:             chunkSize,
		StepSize:              stepSize,
		MaxTimeRanges:         defaultMaxPromTimeRanges,
		MaxQueryRangeDuration: defaultMaxTimeDuration,
	}
}

func (op *Reporting) newPromImporter(logger logrus.FieldLogger, reportDataSource *cbTypes.ReportDataSource, reportPromQuery *cbTypes.ReportPrometheusQuery) (*prestostore.PrometheusImporter, error) {
	cfg := op.newPromImporterCfg(reportDataSource, reportPromQuery)
	metricsCollectors := op.newPromImporterMetricsCollectors(reportDataSource, reportPromQuery)
	var promConn prom.API
	var err error
	if (reportDataSource.Spec.Promsum.PrometheusConfig != nil) && (reportDataSource.Spec.Promsum.PrometheusConfig.URL != "") {
		promConn, err = op.newPrometheusConnFromURL(reportDataSource.Spec.Promsum.PrometheusConfig.URL)
		if err != nil {
			return nil, err
		}
	} else {
		promConn = op.promConn
	}
	return prestostore.NewPrometheusImporter(logger, promConn, op.prestoQueryer, op.clock, cfg, metricsCollectors), nil
}

func (op *Reporting) newPromImporterMetricsCollectors(reportDataSource *cbTypes.ReportDataSource, reportPromQuery *cbTypes.ReportPrometheusQuery) prestostore.ImporterMetricsCollectors {
	promLabels := prometheus.Labels{
		"reportdatasource":      reportDataSource.Name,
		"reportprometheusquery": reportPromQuery.Name,
		"table_name":            dataSourceTableName(reportDataSource.Name),
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

	return prestostore.ImporterMetricsCollectors{
		TotalImportsCounter:     totalImportsCounter,
		FailedImportsCounter:    failedImportsCounter,
		ImportDurationHistogram: importDurationHistogram,
		ImportsRunningGauge:     prometheusReportDatasourceRunningImportsGauge,

		TotalPrometheusQueriesCounter:    totalPrometheusQueriesCounter,
		FailedPrometheusQueriesCounter:   failedPrometheusQueriesCounter,
		PrometheusQueryDurationHistogram: promQueryDurationHistogram,

		TotalPrestoStoresCounter:     totalPrestoStoresCounter,
		FailedPrestoStoresCounter:    failedPrestoStoresCounter,
		PrestoStoreDurationHistogram: prestoStoreDurationHistogram,

		MetricsScrapedCounter:  promQueryMetricsScrapedCounter,
		MetricsImportedCounter: metricsImportedCounter,
	}
}
