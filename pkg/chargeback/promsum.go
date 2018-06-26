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
	// If we get to 2000, it means we're very backlogged, or we have a small
	// chunkSize and making tons of small queries all one after another will
	// cause undesired resource spikes, or both.
	// This will make it take longer to catch up, but should help prevent
	// memory from exploding when we end up with a ton of time ranges.
	defaultMaxPromTimeRanges = 2000
)

func (c *Chargeback) runPrometheusImporterWorker(stopCh <-chan struct{}) {
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

func (c *Chargeback) triggerPrometheusImporterFromLastTimestamp(ctx context.Context) error {
	select {
	case c.prometheusImporterTriggerFromLastTimestampCh <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type prometheusImporterTimeRangeTrigger struct {
	start, end time.Time
	errCh      chan error
}

func (c *Chargeback) triggerPrometheusImporterForTimeRange(ctx context.Context, start, end time.Time) error {
	errCh := make(chan error)
	select {
	case c.prometheusImporterTriggerForTimeRangeCh <- prometheusImporterTimeRangeTrigger{start, end, errCh}:
		return <-errCh
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Chargeback) startPrometheusImporter(ctx context.Context) {
	logger := c.logger.WithField("component", "PrometheusImporter")
	logger.Infof("PrometheusImporter worker started")
	prometheusImporters := make(map[string]*prestostore.PrometheusImporter)

	defer logger.Infof("PrometheusImporterWorker shutdown")

	var timeCh <-chan time.Time
	if c.cfg.DisablePromsum {
		logger.Infof("Periodic Prometheus ReportDataSource importing disabled")
	} else {
		logger.Infof("Periodiccally importing Prometheus ReportDataSource every %s", c.cfg.PromsumInterval)
		ticker := time.NewTicker(c.cfg.PromsumInterval)
		timeCh = ticker.C

		defer ticker.Stop()
		// this go routine runs the trigger import function every PollInterval tick
		// causing the importer to collect and store data
		go func() {
			for {
				select {
				case <-timeCh:
					if err := c.triggerPrometheusImporterFromLastTimestamp(ctx); err != nil {
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			logger.Infof("got shutdown signal, shutting down PrometheusImporters")
			return
		case trigger := <-c.prometheusImporterTriggerForTimeRangeCh:
			// manually triggered import for a specific time range, usually from HTTP API
			err := c.importPrometheusDataSourceDataForTimeRange(ctx, logger, prometheusImporters, trigger.start, trigger.end)
			trigger.errCh <- err
		case <-c.prometheusImporterTriggerFromLastTimestampCh:
			// every tick on timeCh this import Prometheus data for multiple
			// ReportDataSources in parallel.
			// we ignore the error because it's already logged and handled, but
			// we may want to check it for cancellation in the future
			_ = c.importPrometheusDataSourceDataFromLastTimestamp(ctx, logger, prometheusImporters)
		case dataSourceName := <-c.prometheusImporterDeletedDataSourceQueue:
			// if we have an importer for this ReportDataSource then we need to
			// remove it from our map so that the next time we import  Metrics
			// it's not processed
			if _, exists := prometheusImporters[dataSourceName]; exists {
				delete(prometheusImporters, dataSourceName)
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

			reportPromQuery, err := c.informers.Chargeback().V1alpha1().ReportPrometheusQueries().Lister().ReportPrometheusQueries(reportDataSource.Namespace).Get(queryName)
			if err != nil {
				c.logger.WithError(err).Errorf("unable to ReportPrometheusQuery %s for ReportDataSource %s", queryName, dataSourceName)
				continue
			}
			promQuery := reportPromQuery.Spec.Query

			cfg := prestostore.Config{
				PrometheusQuery:       promQuery,
				PrestoTableName:       tableName,
				ChunkSize:             c.cfg.PromsumChunkSize,
				StepSize:              c.cfg.PromsumStepSize,
				MaxTimeRanges:         defaultMaxPromTimeRanges,
				AllowIncompleteChunks: true,
			}

			if importer, exists := prometheusImporters[dataSourceName]; exists {
				dataSourceLogger.Debugf("ReportDataSource %s already has an importer, updating configuration", dataSourceName)
				importer.UpdateConfig(cfg)
			} else {
				importer := prestostore.NewPrometheusImporter(dataSourceLogger, c.promConn, c.prestoConn, c.clock, cfg)
				prometheusImporters[dataSourceName] = importer
			}
		}
	}
}

type importFunc func(context.Context, *prestostore.PrometheusImporter) ([]prom.Range, error)

func (c *Chargeback) importPrometheusDataSourceDataFromLastTimestamp(ctx context.Context, logger logrus.FieldLogger, prometheusImporters map[string]*prestostore.PrometheusImporter) error {
	return c.importPrometheusDataSourceData(ctx, logger, prometheusImporters, func(ctx context.Context, importer *prestostore.PrometheusImporter) ([]prom.Range, error) {
		return importer.ImportFromLastTimestamp(ctx)
	})
}

func (c *Chargeback) importPrometheusDataSourceDataForTimeRange(ctx context.Context, logger logrus.FieldLogger, prometheusImporters map[string]*prestostore.PrometheusImporter, start, end time.Time) error {
	return c.importPrometheusDataSourceData(ctx, logger, prometheusImporters, func(ctx context.Context, importer *prestostore.PrometheusImporter) ([]prom.Range, error) {
		return importer.ImportMetrics(ctx, start, end)
	})
}

func (c *Chargeback) importPrometheusDataSourceData(ctx context.Context, logger logrus.FieldLogger, prometheusImporters map[string]*prestostore.PrometheusImporter, runImport importFunc) error {
	logger.Infof("Importing Prometheus metrics to Presto")

	const concurrency = 4
	// create a channel to act as a semaphore to limit the number of
	// importFuncs happening in parallel
	semaphore := make(chan struct{}, concurrency)
	g, ctx := errgroup.WithContext(ctx)

	// start a go routine for each worker, where each Go routine will
	// attempt to increment the semaphore, blocking if there are
	// already `concurrency` go routines doing work. When a go routine
	// is no longer importing, it decrements the semaphore allowing
	// other importer Go routines to run
	for dataSourceName, prometheusImporter := range prometheusImporters {
		prometheusImporter := prometheusImporter
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
				dataSourceLogger.Infof("finished import for Prometheus ReportDataSource %s", dataSourceName)
				<-semaphore
			}()
			dataSourceLogger.Infof("starting import for Prometheus ReportDataSource %s", dataSourceName)

			_, err := runImport(ctx, prometheusImporter)
			return err
		})
	}
	err := g.Wait()
	if err != nil {
		logger.WithError(err).Errorf("PrometheusImporter worker encountered errors while importing data")
		return err
	}
	return nil
}
