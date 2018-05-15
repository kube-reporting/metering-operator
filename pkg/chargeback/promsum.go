package chargeback

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/operator-framework/operator-metering/pkg/apis/chargeback/v1alpha1"
	"github.com/operator-framework/operator-metering/pkg/presto"
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

	c.collectPromsumData(ctx, logger, timeBoundsGetter, defaultMaxPromTimeRanges, false)
}

// promsumDataSourceTimeBoundsGetter takes a dataSource and returns the time
// which we should begin collecting data and end time we should collect data
// until.
type promsumDataSourceTimeBoundsGetter func(dataSource *cbTypes.ReportDataSource) (startTime, endTime time.Time, err error)

func (c *Chargeback) collectPromsumData(ctx context.Context, logger logrus.FieldLogger, timeBoundsGetter promsumDataSourceTimeBoundsGetter, maxPromTimeRanges int64, allowIncompleteChunks bool) {
	dataSources, err := c.informers.Chargeback().V1alpha1().ReportDataSources().Lister().ReportDataSources(c.cfg.Namespace).List(labels.Everything())
	if err != nil {
		logger.Errorf("couldn't list data stores: %v", err)
		return
	}

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

		g.Go(func() error {
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
				logger.WithError(err).Errorf("error collecting promsum data for datasource")
				return err
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		logger.WithError(err).Errorf("some promsum datasources had errors when collecting data")
	}
	logger.Debugf("all promsum datasource collectors have finished")
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
	timeRanges, err := promsumGetTimeRanges(startTime, endTime, c.cfg.PromsumChunkSize, c.cfg.PromsumStepSize, maxPromTimeRanges, allowIncompleteChunks)
	if err != nil {
		return fmt.Errorf("couldn't get time ranges to query for dataSource %s: %v", dataSource.Name, err)
	}

	if len(timeRanges) == 0 {
		logger.Info("no time ranges to query yet")
		return nil
	} else {
		begin := timeRanges[0].Start
		end := timeRanges[len(timeRanges)-1].End
		logger.Infof("querying for data between %s and %s (chunks: %d)", begin, end, len(timeRanges))
	}

	query, err := c.informers.Chargeback().V1alpha1().ReportPrometheusQueries().Lister().ReportPrometheusQueries(c.cfg.Namespace).Get(dataSource.Spec.Promsum.Query)
	if err != nil {
		return fmt.Errorf("could not get prometheus query: %s", err)
	}

	for _, queryRng := range timeRanges {
		// check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// continue processing if context isn't cancelled.
		}
		records, err := c.promsumQuery(query, queryRng)
		if err != nil {
			return fmt.Errorf("failed to retrieve prometheus metrics for query '%s' in the range %v to %v: %v",
				query.Name, queryRng.Start, queryRng.End, err)
		}

		// check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// continue processing if context isn't cancelled.
		}

		err = c.promsumStoreRecords(ctx, logger, dataSource.TableName, records)
		if err != nil {
			return fmt.Errorf("failed to store prometheus metrics for query '%s' in the range %v to %v: %v",
				query.Name, queryRng.Start, queryRng.End, err)
		}
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

func promsumGetTimeRanges(beginTime, endTime time.Time, chunkSize, stepSize time.Duration, maxPromTimeRanges int64, allowIncompleteChunks bool) ([]prom.Range, error) {
	chunkStart := truncateToMinute(beginTime)
	chunkEnd := truncateToMinute(chunkStart.Add(chunkSize))

	// don't set a limit if negative or zero
	disableMax := maxPromTimeRanges <= 0

	var timeRanges []prom.Range
	for i := int64(0); disableMax || (i < maxPromTimeRanges); i++ {
		if allowIncompleteChunks {
			if chunkEnd.After(endTime) {
				chunkEnd = truncateToMinute(endTime)
			}
			if chunkEnd.Equal(chunkStart) {
				break
			}
		} else {
			// Do not collect data after endTime
			if chunkEnd.After(endTime) {
				break
			}

			// Only get chunks that are a full chunk size
			if chunkEnd.Sub(chunkStart) < chunkSize {
				break
			}
		}
		timeRanges = append(timeRanges, prom.Range{
			Start: chunkStart.UTC(),
			End:   chunkEnd.UTC(),
			Step:  stepSize,
		})

		if allowIncompleteChunks && chunkEnd.Equal(truncateToMinute(endTime)) {
			break
		}

		// Add the metrics step size to the start time so that we don't
		// re-query the previous ranges end time in this range
		chunkStart = truncateToMinute(chunkEnd.Add(stepSize))
		// Add chunkSize to the end time to get our full chunk. If the chunkEnd
		// is past the endTime, then this chunk is skipped.
		chunkEnd = truncateToMinute(chunkStart.Add(chunkSize))
	}

	return timeRanges, nil
}

func (c *Chargeback) promsumQuery(query *cbTypes.ReportPrometheusQuery, queryRng prom.Range) ([]*PromsumRecord, error) {
	pVal, err := c.promConn.QueryRange(context.Background(), query.Spec.Query, queryRng)
	if err != nil {
		return nil, fmt.Errorf("failed to perform billing query: %v", err)
	}

	matrix, ok := pVal.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("expected a matrix in response to query, got a %v", pVal.Type())
	}

	var records []*PromsumRecord
	// iterate over segments of contiguous billing records
	for _, sampleStream := range matrix {
		for _, value := range sampleStream.Values {
			labels := make(map[string]string, len(sampleStream.Metric))
			for k, v := range sampleStream.Metric {
				labels[string(k)] = string(v)
			}

			record := &PromsumRecord{
				Labels:    labels,
				Amount:    float64(value.Value),
				StepSize:  c.cfg.PromsumStepSize,
				Timestamp: value.Timestamp.Time().UTC(),
			}
			records = append(records, record)
		}
	}
	return records, nil
}

func (c *Chargeback) promsumStoreRecords(ctx context.Context, logger logrus.FieldLogger, tableName string, records []*PromsumRecord) error {
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

func generateRecordSQLValues(record *PromsumRecord) string {
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

// PromsumRecord is a receipt of a usage determined by a query within a specific time range.
type PromsumRecord struct {
	Labels    map[string]string `json:"labels"`
	Amount    float64           `json:"amount"`
	StepSize  time.Duration     `json:"stepSize"`
	Timestamp time.Time         `json:"timestamp"`
}
