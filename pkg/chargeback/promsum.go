package chargeback

import (
	"context"
	"time"

	"github.com/operator-framework/operator-metering/pkg/chargeback/promexporter"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	// Keep a cap on the number of time ranges we query per reconciliation.
	// If we get to 2000, it means we're very backlogged, or we have a small
	// chunkSize and making tons of small queries all one after another will
	// cause undesired resource spikes, or both.
	// This will make it take longer to catch up, but should help prevent
	// memory from exploding when we end up with a ton of time ranges.
	defaultMaxPromTimeRanges = 2000
)

func (c *Chargeback) runPrometheusExporterWorker(stopCh <-chan struct{}) {
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
	c.startPrometheusExporter(ctx)
}

func (c *Chargeback) triggerPromExporterFromLastTimestamp(ctx context.Context) error {
	select {
	case c.promExporterTriggerFromLastTimestampCh <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type promExporterTimeRangeTrigger struct {
	start, end time.Time
	errCh      chan error
}

func (c *Chargeback) triggerPromExporterForTimeRange(ctx context.Context, start, end time.Time) error {
	errCh := make(chan error)
	select {
	case c.promExporterTriggerForTimeRangeCh <- promExporterTimeRangeTrigger{start, end, errCh}:
		return <-errCh
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Chargeback) startPrometheusExporter(ctx context.Context) {
	logger := c.logger.WithField("component", "PrometheusExporter")
	logger.Infof("PrometheusExporter worker started")
	ticker := time.NewTicker(c.cfg.PromsumInterval)
	promExporters := make(map[string]*promexporter.PrestoExporter)

	defer func() {
		logger.Infof("PrometheusExporterWorker shutdown")
		ticker.Stop()
	}()

	timeCh := ticker.C

	// this go routine runs the trigger export function every PollInterval tick
	// causing the exporter to collect and store data
	go func() {
		for {
			select {
			case <-timeCh:
				if err := c.triggerPromExporterFromLastTimestamp(ctx); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			logger.Infof("got shutdown signal, shutting down PrometheusExporters")
			return
		case trigger := <-c.promExporterTriggerForTimeRangeCh:
			// manually triggered export for a specific time range, usually from HTTP API
			err := c.exportPrometheusDataSourceDataForTimeRange(ctx, logger, promExporters, trigger.start, trigger.end)
			trigger.errCh <- err
		case <-c.promExporterTriggerFromLastTimestampCh:
			// every tick on timeCh this export Prometheus data for multiple
			// ReportDataSources in parallel.
			// we ignore the error because it's already logged and handled, but
			// we may want to check it for cancellation in the future
			_ = c.exportPrometheusDataSourceDataFromLastTimestamp(ctx, logger, promExporters)
		case dataSourceName := <-c.prometheusExporterDeletedDataSourceQueue:
			// if we have an exporter for this ReportDataSource then we need to
			// remove it from our map so that the next time we export  Metrics
			// it's not processed
			if _, exists := promExporters[dataSourceName]; exists {
				delete(promExporters, dataSourceName)
			}
		case reportDataSource := <-c.prometheusExporterNewDataSourceQueue:
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

			if _, exists := promExporters[dataSourceName]; exists {
				// We've already got an exporter for this ReportDataSource
				// so we just need to update it
				dataSourceLogger.Debugf("ReportDataSource %s already has an exporter, updating configuration")
			}

			reportPromQuery, err := c.informers.Chargeback().V1alpha1().ReportPrometheusQueries().Lister().ReportPrometheusQueries(reportDataSource.Namespace).Get(queryName)
			if err != nil {
				c.logger.WithError(err).Errorf("unable to ReportPrometheusQuery %s for ReportDataSource %s", queryName, dataSourceName)
				continue
			}
			promQuery := reportPromQuery.Spec.Query

			cfg := promexporter.Config{
				PrometheusQuery:       promQuery,
				PrestoTableName:       tableName,
				ChunkSize:             c.cfg.PromsumChunkSize,
				StepSize:              c.cfg.PromsumStepSize,
				MaxTimeRanges:         defaultMaxPromTimeRanges,
				AllowIncompleteChunks: true,
			}
			exporter := promexporter.NewPrestoExporter(dataSourceLogger, c.promConn, c.prestoConn, c.clock, cfg)
			promExporters[dataSourceName] = exporter
		}
	}
}

type exporterFunc func(context.Context, *promexporter.PrestoExporter) ([]*promexporter.Record, error)

func (c *Chargeback) exportPrometheusDataSourceDataFromLastTimestamp(ctx context.Context, logger logrus.FieldLogger, promExporters map[string]*promexporter.PrestoExporter) error {
	return c.exportPrometheusDataSourceData(ctx, logger, promExporters, func(ctx context.Context, exporter *promexporter.PrestoExporter) ([]*promexporter.Record, error) {
		return exporter.ExportFromLastTimestamp(ctx)
	})
}

func (c *Chargeback) exportPrometheusDataSourceDataForTimeRange(ctx context.Context, logger logrus.FieldLogger, promExporters map[string]*promexporter.PrestoExporter, start, end time.Time) error {
	return c.exportPrometheusDataSourceData(ctx, logger, promExporters, func(ctx context.Context, exporter *promexporter.PrestoExporter) ([]*promexporter.Record, error) {
		return exporter.Export(ctx, start, end)
	})
}

func (c *Chargeback) exportPrometheusDataSourceData(ctx context.Context, logger logrus.FieldLogger, promExporters map[string]*promexporter.PrestoExporter, export exporterFunc) error {
	logger.Infof("Exporting Prometheus metrics to Presto")

	const concurrency = 4
	// create a channel to act as a semaphore to limit the number of
	// exports happening in parallel
	semaphore := make(chan struct{}, concurrency)
	g, ctx := errgroup.WithContext(ctx)

	// start a go routine for each worker, where each Go routine will
	// attempt to increment the semaphore, blocking if there are
	// already `concurrency` go routines doing work. When a go routine
	// is no longer exporting, it decrements the semaphore allowing
	// other exporter Go routines to run
	for dataSourceName, exporter := range promExporters {
		exporter := exporter
		g.Go(func() error {
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
				dataSourceLogger.Infof("finished export for Prometheus ReportDataSource %s", dataSourceName)
				<-semaphore
			}()
			dataSourceLogger.Infof("starting export for Prometheus ReportDataSource %s", dataSourceName)

			_, err := export(ctx, exporter)
			return err
		})
	}
	err := g.Wait()
	if err != nil {
		logger.WithError(err).Errorf("PrometheusExporter worker encountered errors while exporting data")
		return err
	}
	return nil
}
