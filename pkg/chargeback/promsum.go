package chargeback

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/presto"
	"github.com/operator-framework/operator-metering/pkg/promcollector"
)

const (
	prestoQueryCap  = 1000000
	timestampFormat = "2006-01-02 15:04:05.000"

	// Keep a cap on the number of time ranges we query per reconciliation.
	// If we get to 2000, it means we're very backlogged, or we have a small
	// chunkSize and making tons of small queries all one after another will
	// cause undesired resource spikes, or both.
	// This will make it take longer to catch up, but should help prevent
	// memory from exploding when we end up with a ton of time ranges.
	defaultMaxPromTimeRanges = 2000
)

func (c *Chargeback) runPromsumWorker(stopCh <-chan struct{}) {
	logger := c.logger.WithField("component", "promsum")
	logger.Infof("Promsum collector worker started")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// run a go routine that waits for the stopCh to be closed and propagates
	// the shutdown to the collectors by calling cancel()
	go func() {
		<-stopCh
		logger.Infof("got shutdown signal, shutting down promsum collectors")
		// if the stopCh is closed while we're waiting, cancel and wait for
		// everything to return
		cancel()
	}()

	var wg sync.WaitGroup
	ticker := time.NewTicker(c.cfg.PromsumInterval)
	tickerCh := ticker.C
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tickerCh:
				c.collectPromsumDataWithDefaultTimeBounds(ctx, logger)
			}
		}
	}()

	wg.Wait()
	logger.Infof("promsum collectors shutdown")
}

func (c *Chargeback) collectPromsumDataWithDefaultTimeBounds(ctx context.Context, logger logrus.FieldLogger) {
	timeBoundsGetter := promsumDataSourceTimeBoundsGetter(func(dataSource *cbTypes.ReportDataSource) (startTime, endTime time.Time, err error) {
		logger := logger.WithField("datasource", dataSource.Name)
		startTime, endTime, err = c.promsumGetTimeBounds(logger, dataSource)
		if err != nil {
			return startTime, endTime, fmt.Errorf("couldn't determine time bounds for dataSource %s: %v", dataSource.Name, err)
		}
		return startTime, endTime, nil
	})

	err := c.collectPromsumData(ctx, logger, timeBoundsGetter, defaultMaxPromTimeRanges, false)
	if err != nil {
		logger.WithError(err).Errorf("unable to collect prometheus data")
	}
}

// promsumDataSourceTimeBoundsGetter takes a dataSource and returns the time
// which we should begin collecting data and end time we should collect data
// until.
type promsumDataSourceTimeBoundsGetter func(dataSource *cbTypes.ReportDataSource) (startTime, endTime time.Time, err error)

func (c *Chargeback) collectPromsumData(ctx context.Context, logger logrus.FieldLogger, timeBoundsGetter promsumDataSourceTimeBoundsGetter, maxPromTimeRanges int64, allowIncompleteChunks bool) error {
	dataSources, err := c.informers.Chargeback().V1alpha1().ReportDataSources().Lister().ReportDataSources(c.cfg.Namespace).List(labels.Everything())
	if err != nil {
		return fmt.Errorf("couldn't list data stores: %v", err)
	}

	// sem acts as a semaphore limiting the number of active running
	// collections at once
	concurrency := 4
	sem := make(chan struct{}, concurrency)

	var g errgroup.Group
	for _, dataSource := range dataSources {
		dataSource := dataSource
		logger := logger.WithField("datasource", dataSource.Name)

		if dataSource.Spec.Promsum == nil {
			continue
		}
		if dataSource.TableName == "" {
			// This data store doesn't have a table yet, let's skip it and
			// hope it'll have one next time.
			logger.Debugf("no table set, skipping collection for data store %q", dataSource.Name)
			key, err := cache.MetaNamespaceKeyFunc(dataSource)
			if err == nil {
				logger.Debugf("no table set, queueing %q", dataSource.Name)
				c.queues.reportDataSourceQueue.Add(key)
			}
			continue
		}

		// this blocks if we're at the concurrency limit, and will return if we
		// get a context cancellation signal
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			if err := g.Wait(); err != nil {
				return err
			}
			return ctx.Err()
		}
		g.Go(func() error {
			// release the semaphore at the end
			defer func() {
				<-sem
			}()
			startTime, endTime, err := timeBoundsGetter(dataSource)
			if err != nil {
				logger.WithError(err).Errorf("error getting collection time boundries for datasource")
				return err
			}

			logger := logger.WithFields(logrus.Fields{
				"startTime": startTime,
				"endTime":   endTime,
			})
			err = c.collectPromsumDataSourceData(ctx, logger, dataSource, startTime, endTime, maxPromTimeRanges, allowIncompleteChunks)
			if err != nil {
				// if the error is from cancellation, then it's handled
				if err == context.Canceled {
					logger.Infof("promsum datasource collector shutdown")
					return nil
				}
				logger.WithError(err).Errorf("error collecting Prometheus data for datasource")
				return err
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("some Prometheus datasources had errors when collecting data, err: %v", err)
	}
	logger.Debugf("all promsum datasource collectors have finished")
	return nil
}

func (c *Chargeback) collectPromsumDataSourceData(ctx context.Context, logger logrus.FieldLogger, dataSource *cbTypes.ReportDataSource, startTime, endTime time.Time, maxPromTimeRanges int64, allowIncompleteChunks bool) error {
	logger.Debugf("processing data store %q", dataSource.Name)
	if dataSource.Spec.Promsum == nil {
		logger.Debugf("not a promsum store, skipping %q", dataSource.Name)
		return nil
	}
	err := c.promsumCollectDataForQuery(ctx, logger, dataSource, startTime, endTime, maxPromTimeRanges, allowIncompleteChunks)
	if err != nil {
		return err
	}
	logger.Debugf("processing complete for data store %q", dataSource.Name)
	return nil
}

func (c *Chargeback) promsumCollectDataForQuery(ctx context.Context, logger logrus.FieldLogger, dataSource *cbTypes.ReportDataSource, startTime, endTime time.Time, maxPromTimeRanges int64, allowIncompleteChunks bool) error {
	queryName := dataSource.Spec.Promsum.Query
	promQuery, err := c.informers.Chargeback().V1alpha1().ReportPrometheusQueries().Lister().ReportPrometheusQueries(dataSource.Namespace).Get(queryName)
	if err != nil {
		return fmt.Errorf("could not get Prometheus query '%s': %s", queryName, err)
	}

	preProcessingHandler := func(_ context.Context, timeRanges []prom.Range) error {
		if len(timeRanges) == 0 {
			logger.Info("no time ranges to query yet for ReportDataSource %s", dataSource.Name)
		} else {
			begin := timeRanges[0].Start
			end := timeRanges[len(timeRanges)-1].End
			logger.Infof("using query %s querying for data between %s and %s (chunks: %d)", queryName, begin, end, len(timeRanges))
		}
		return nil
	}

	postQueryHandler := func(ctx context.Context, timeRange prom.Range, records []*promcollector.Record) error {
		err = c.promsumStoreRecords(ctx, logger, dataSource.TableName, records)
		if err != nil {
			return fmt.Errorf("failed to store Prometheus metrics for ReportDataSource %s using query '%s' in the range %v to %v: %v",
				dataSource.Name, queryName, timeRange.Start, timeRange.End, err)
		}
		return nil
	}

	collector := promcollector.New(c.promConn, promQuery.Spec.Query, preProcessingHandler, nil, postQueryHandler)
	timeRangesCollected, err := collector.Collect(ctx, startTime, endTime, c.cfg.PromsumStepSize, c.cfg.PromsumChunkSize, maxPromTimeRanges, allowIncompleteChunks)
	if err != nil {
		return err
	}
	if len(timeRangesCollected) == 0 {
		logger.Infof("no data collected for ReportDataSource %s", dataSource.Name)
	}

	return nil
}

func (c *Chargeback) promsumGetLastTimestamp(logger logrus.FieldLogger, dataSource *cbTypes.ReportDataSource) (time.Time, error) {
	if dataSource.TableName == "" {
		return time.Time{}, fmt.Errorf("unable to get last timestamp for, dataSource %s no tableName is set", dataSource.Name)
	}
	// Get the most recent timestamp in the table for this query
	getLastTimestampQuery := fmt.Sprintf(`
				SELECT "timestamp"
				FROM %s
				ORDER BY "timestamp" DESC
				LIMIT 1`, dataSource.TableName)

	results, err := presto.ExecuteSelect(c.prestoConn, getLastTimestampQuery)
	if err != nil {
		return time.Time{}, fmt.Errorf("error getting last timestamp for dataSource %s, maybe table doesn't exist yet? %v", dataSource.Name, err)
	}

	var lastTimestamp time.Time
	if len(results) != 0 {
		lastTimestamp = results[0]["timestamp"].(time.Time)
	}
	return lastTimestamp, nil
}

func (c *Chargeback) promsumGetTimeBounds(logger logrus.FieldLogger, dataSource *cbTypes.ReportDataSource) (startTime, endTime time.Time, err error) {
	lastTimestamp, err := c.promsumGetLastTimestamp(logger, dataSource)
	if err != nil {
		return startTime, endTime, err
	}

	endTime = c.clock.Now()

	if !lastTimestamp.IsZero() {
		logger.Debugf("last fetched data for data store %s at %s", dataSource.Name, lastTimestamp.String())
	} else {
		// Looks like we haven't populated any data in this table yet.
		// Let's backfill our last 1 chunk.
		// we multiple by 2 because the most recent chunk will have a
		// chunkEnd == endTime, so it won't be queried, so this gets the chunk
		// before the latest
		lastTimestamp = endTime.Add(-2 * c.cfg.PromsumChunkSize)
		logger.Debugf("no data in data store %s yet", dataSource.Name)
	}
	// We don't want to duplicate the lastTimestamp record so add
	// the step size so that we start at the next interval no longer in
	// our range.
	startTime = lastTimestamp.Add(c.cfg.PromsumStepSize)

	const maxChunkDuration = 24 * time.Hour
	// If the lastTimestamp is too far back, we should limit this run to
	// maxChunkDuration so that if we're stopped for an extended amount of time,
	// this function won't return a slice with too many time ranges.
	totalChunkDuration := lastTimestamp.Sub(endTime)
	if totalChunkDuration >= maxChunkDuration {
		endTime = lastTimestamp.Add(maxChunkDuration)
	}
	return startTime, endTime, nil
}

func (c *Chargeback) promsumStoreRecords(ctx context.Context, logger logrus.FieldLogger, tableName string, records []*promcollector.Record) error {
	var queryValues []string

	for _, record := range records {
		recordValue := generateRecordSQLValues(record)
		queryValues = append(queryValues, recordValue)
	}
	// capacity prestoQueryCap, length 0
	queryBuf := bytes.NewBuffer(make([]byte, 0, prestoQueryCap))

	insertStatementLength := len(presto.FormatInsertQuery(tableName, ""))
	// calculate the queryCap with the "INSERT INTO $table_name" portion
	// accounted for
	queryCap := prestoQueryCap - insertStatementLength

	for _, value := range queryValues {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// continue processing if context isn't cancelled.
		}

		// If the buffer is empty, we add VALUES to it, and everything the
		// follows will be a single row to insert
		if queryBuf.Len() == 0 {
			queryBuf.WriteString("VALUES ")
		} else {
			// if the buffer isn't empty, then before we add more rows to the
			// insert query, add a comma to separate them.
			queryBuf.WriteString(",")
		}

		// There's a character limit of prestoQueryCap on insert
		// queries, so let's chunk them at that limit.
		bytesToWrite := len(value)
		newBufferSize := (bytesToWrite + queryBuf.Len())

		// if writing the current value to the buffer would exceed the
		// prestoQueryCap, preform the insert query, and reset the buffer
		if newBufferSize > queryCap {
			err := presto.ExecuteInsertQuery(c.prestoConn, tableName, queryBuf.String())
			if err != nil {
				return fmt.Errorf("failed to store metrics into presto: %v", err)
			}
			queryBuf.Reset()
		} else {
			queryBuf.WriteString(value)
		}
	}
	// if the buffer has unwritten values, perform the final insert
	if queryBuf.Len() != 0 {
		err := presto.ExecuteInsertQuery(c.prestoConn, tableName, queryBuf.String())
		if err != nil {
			return fmt.Errorf("failed to store metrics into presto: %v", err)
		}
	}
	return nil
}

func generateRecordSQLValues(record *promcollector.Record) string {
	var keys []string
	var vals []string
	for k, v := range record.Labels {
		keys = append(keys, "'"+k+"'")
		vals = append(vals, "'"+v+"'")
	}
	keyString := "ARRAY[" + strings.Join(keys, ",") + "]"
	valString := "ARRAY[" + strings.Join(vals, ",") + "]"
	return fmt.Sprintf("(%f,timestamp '%s',%f,map(%s,%s))",
		record.Amount, record.Timestamp.Format(timestampFormat), record.StepSize.Seconds(), keyString, valString)
}
