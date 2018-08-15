package chargeback

import (
	"context"
	"time"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/operator-framework/operator-metering/pkg/chargeback/prestostore"
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

func (c *Metering) runPrometheusImporterWorker(stopCh <-chan struct{}) {
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
	c.startPrometheusImporter(ctx)
}

type prometheusImporterFunc func(ctx context.Context, start, end time.Time) error

type prometheusImporterTimeRangeTrigger struct {
	start, end time.Time
	errCh      chan error
}

func (c *Metering) triggerPrometheusImporterForTimeRange(ctx context.Context, start, end time.Time) error {
	errCh := make(chan error)
	select {
	case c.prometheusImporterTriggerForTimeRangeCh <- prometheusImporterTimeRangeTrigger{start, end, errCh}:
		return <-errCh
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Metering) startPrometheusImporter(ctx context.Context) {
	logger := c.logger.WithField("component", "PrometheusImporter")
	logger.Infof("PrometheusImporter worker started")
	workers := make(map[string]*prometheusImporterWorker)
	importers := make(map[string]*prestostore.PrometheusImporter)

	const concurrency = 4
	// create a channel to act as a semaphore to limit the number of
	// imports happening in parallel
	semaphore := make(chan struct{}, concurrency)

	defer logger.Infof("PrometheusImporterWorker shutdown")

	if c.cfg.DisablePromsum {
		logger.Infof("Periodic Prometheus ReportDataSource importing disabled")
	}

	for {
		select {
		case <-ctx.Done():
			logger.Infof("got shutdown signal, shutting down PrometheusImporters")
			return
		case trigger := <-c.prometheusImporterTriggerForTimeRangeCh:
			// manually triggered import for a speci***REMOVED***c time range, usually from HTTP API

			g, ctx := errgroup.WithContext(ctx)
			for dataSourceName, importer := range importers {
				importer := importer
				dataSourceName := dataSourceName
				// collect each dataSource concurrently
				g.Go(func() error {
					return importPrometheusDataSourceData(ctx, logger, semaphore, dataSourceName, importer, func(ctx context.Context, importer *prestostore.PrometheusImporter) ([]prom.Range, error) {
						return importer.ImportMetrics(ctx, trigger.start, trigger.end, true)
					})
				})
			}
			err := g.Wait()
			if err != nil {
				logger.WithError(err).Errorf("PrometheusImporter worker encountered errors while importing data")
			}
			trigger.errCh <- err

		case dataSourceName := <-c.prometheusImporterDeletedDataSourceQueue:
			// if we have a worker for this ReportDataSource then we need to
			// stop it and remove it from our map
			if worker, exists := workers[dataSourceName]; exists {
				worker.stop()
				delete(workers, dataSourceName)
			}
			if _, exists := importers[dataSourceName]; exists {
				delete(importers, dataSourceName)
			}
		case reportDataSource := <-c.prometheusImporterNewDataSourceQueue:
			if reportDataSource.Spec.Promsum == nil {
				logger.Error("expected only Promsum ReportDataSources")
				continue
			}

			dataSourceName := reportDataSource.Name
			queryName := reportDataSource.Spec.Promsum.Query
			tableName := dataSourceTableName(dataSourceName)

			dataSourceLogger := logger.WithFields(logrus.Fields{
				"queryName":        queryName,
				"reportDataSource": dataSourceName,
				"tableName":        tableName,
			})

			reportPromQuery, err := c.informers.Metering().V1alpha1().ReportPrometheusQueries().Lister().ReportPrometheusQueries(reportDataSource.Namespace).Get(queryName)
			if err != nil {
				c.logger.WithError(err).Errorf("unable to ReportPrometheusQuery %s for ReportDataSource %s", queryName, dataSourceName)
				continue
			}

			promQuery := reportPromQuery.Spec.Query

			chunkSize := c.cfg.PrometheusQueryCon***REMOVED***g.ChunkSize.Duration
			stepSize := c.cfg.PrometheusQueryCon***REMOVED***g.StepSize.Duration
			queryInterval := c.cfg.PrometheusQueryCon***REMOVED***g.QueryInterval.Duration

			queryConf := reportDataSource.Spec.Promsum.QueryCon***REMOVED***g
			if queryConf != nil {
				if queryConf.ChunkSize != nil {
					chunkSize = queryConf.ChunkSize.Duration
				}
				if queryConf.StepSize != nil {
					stepSize = queryConf.StepSize.Duration
				}
				if queryConf.QueryInterval != nil {
					queryInterval = queryConf.QueryInterval.Duration
				}
			}

			cfg := prestostore.Con***REMOVED***g{
				PrometheusQuery:       promQuery,
				PrestoTableName:       tableName,
				ChunkSize:             chunkSize,
				StepSize:              stepSize,
				MaxTimeRanges:         defaultMaxPromTimeRanges,
				MaxQueryRangeDuration: defaultMaxTimeDuration,
			}

			importer, exists := importers[dataSourceName]
			if exists {
				dataSourceLogger.Debugf("ReportDataSource %s already has an importer, updating con***REMOVED***guration", dataSourceName)
				importer.UpdateCon***REMOVED***g(cfg)
			} ***REMOVED*** {
				importer = prestostore.NewPrometheusImporter(dataSourceLogger, c.promConn, c.prestoQueryer, c.clock, cfg)
				importers[dataSourceName] = importer
			}

			if !c.cfg.DisablePromsum {
				worker, workerExists := workers[dataSourceName]
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
			err := importPrometheusDataSourceData(ctx, logger, semaphore, dataSourceName, importer, func(ctx context.Context, importer *prestostore.PrometheusImporter) ([]prom.Range, error) {
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

type importFunc func(context.Context, *prestostore.PrometheusImporter) ([]prom.Range, error)

func importPrometheusDataSourceData(ctx context.Context, logger logrus.FieldLogger, semaphore chan struct{}, dataSourceName string, prometheusImporter *prestostore.PrometheusImporter, runImport importFunc) error {
	// blocks trying to increment the semaphore (sending on the
	// channel) or until the context is cancelled
	select {
	case semaphore <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	dataSourceLogger := logger.WithField("reportDataSource", dataSourceName)
	// decrement the semaphore at the end
	defer func() {
		dataSourceLogger.Infof("***REMOVED***nished import for Prometheus ReportDataSource %s", dataSourceName)
		<-semaphore
	}()
	dataSourceLogger.Infof("starting import for Prometheus ReportDataSource %s", dataSourceName)

	_, err := runImport(ctx, prometheusImporter)
	return err
}
