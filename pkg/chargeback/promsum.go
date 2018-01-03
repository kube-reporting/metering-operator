package chargeback

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/presto"
)

const (
	prestoQueryCap  = 1000000
	timestampFormat = "2006-01-02 15:04:05.000"
)

func (c *Chargeback) runPromsumWorker(stopCh <-chan struct{}) {
	logger := c.logger.WithField("component", "promsum")
	logger.Infof("Promsum collector worker started")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run collection immediately
	done := make(chan struct{})
	go func() {
		c.collectPromsumData(ctx, logger)
		close(done)
	}()
	// allow for cancellation
	select {
	case <-stopCh:
		return
	case <-done:
	}

	// From now on run collection every ticker interval
	ticker := time.NewTicker(c.promsumInterval)
	defer ticker.Stop()
	for {
		select {
		case <-stopCh:
			// if the stopCh is closed while we're waiting, cancel and return
			return
		case <-ticker.C:
			c.collectPromsumData(ctx, logger)
		}
	}
}

func (c *Chargeback) collectPromsumData(ctx context.Context, logger logrus.FieldLogger) {
	dataSources, err := c.informers.reportDataSourceLister.ReportDataSources(c.namespace).List(labels.Everything())
	if err != nil {
		logger.Errorf("couldn't list data stores: %v", err)
		return
	}

	g, ctx := errgroup.WithContext(ctx)
	now := time.Now().UTC()

	for _, dataSource := range dataSources {
		dataSource := dataSource
		g.Go(func() error {
			logger := logger.WithField("datasource", dataSource.Name)
			err := c.collectPromsumDataSourceData(logger, dataSource, now)
			if err != nil {
				logger.WithError(err).Errorf("error collecting promsum data for datasource")
				return err
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		logger.WithError(err).Errorf("some promsum datasources had errors when collecting data")
	}
}

func (c *Chargeback) collectPromsumDataSourceData(logger logrus.FieldLogger, dataSource *cbTypes.ReportDataSource, now time.Time) error {
	logger.Debugf("processing data store %q", dataSource.Name)
	if dataSource.Spec.Promsum == nil {
		logger.Debugf("not a promsum store, skipping %q", dataSource.Name)
		return nil
	}
	if dataSource.TableName == "" {
		// This data store doesn't have a table yet, let's skip it and
		// hope it'll have one next time.
		logger.Debugf("no table set, skipping collection for data store %q", dataSource.Name)
		key, err := cache.MetaNamespaceKeyFunc(dataSource)
		if err == nil {
			logger.Debugf("no table set, queueing %q", dataSource.Name)
			c.informers.reportDataSourceQueue.Add(key)
		}
		return nil
	}

	err := c.promsumCollectDataForQuery(logger, dataSource, now)
	if err != nil {
		return err
	}
	logger.Debugf("processing complete for data store %q", dataSource.Name)
	return nil
}

func (c *Chargeback) promsumCollectDataForQuery(logger logrus.FieldLogger, dataSource *cbTypes.ReportDataSource, now time.Time) error {
	timeRanges, err := c.promsumGetTimeRanges(logger, dataSource, now)
	if err != nil {
		return fmt.Errorf("couldn't get time ranges to query for dataSource %s: %v", dataSource.Name, err)
	}
	logger.Debugf("time ranges to query: %+v", timeRanges)

	if len(timeRanges) == 0 {
		logger.Info("no time ranges to query yet")
		return nil
	} ***REMOVED*** {
		begin := timeRanges[0].Start
		end := timeRanges[len(timeRanges)-1].End
		logger.Infof("querying for data between %s and %s", begin, end)
	}

	for _, queryRng := range timeRanges {
		query, err := c.informers.reportPrometheusQueryLister.ReportPrometheusQueries(c.namespace).Get(dataSource.Spec.Promsum.Query)
		if err != nil {
			return fmt.Errorf("could not get prometheus query: ", err)
		}

		records, err := c.promsumQuery(query, queryRng)
		if err != nil {
			return fmt.Errorf("failed to retrieve prometheus metrics for query '%s' in the range %v to %v: %v",
				query.Name, queryRng.Start, queryRng.End, err)
		}

		err = c.promsumStoreRecords(logger, dataSource, records)
		if err != nil {
			return fmt.Errorf("failed to store prometheus metrics for query '%s' in the range %v to %v: %v",
				query.Name, queryRng.Start, queryRng.End, err)
		}
	}
	return nil
}

func (c *Chargeback) promsumGetTimeRanges(logger logrus.FieldLogger, dataSource *cbTypes.ReportDataSource, now time.Time) ([]prom.Range, error) {
	// Get the most recent timestamp in the table for this query
	getLastTimestampQuery := fmt.Sprintf(`
				SELECT "timestamp"
				FROM %s
				ORDER BY "timestamp" DESC
				LIMIT 1`, dataSource.TableName)

	results, err := presto.ExecuteSelect(c.prestoConn, getLastTimestampQuery)
	if err != nil {
		return nil, fmt.Errorf("error getting last timestamp for table, maybe table doesn't exist yet? %v", err)
	}

	var lastTimestamp time.Time
	if len(results) == 0 {
		// Looks like we haven't populated any data in this table yet.
		// Let's back***REMOVED***ll our last 1 chunk.
		// we multiple by 2 because the most recent chunk will have a
		// chunkEnd == now, so it won't be queried, so this gets the chunk
		// before the latest
		lastTimestamp = now.Add(-2 * c.promsumChunkSize)
		logger.Debugf("no data in data store %s yet", dataSource.Name)
	} ***REMOVED*** {
		lastTimestamp = results[0]["timestamp"].(time.Time)
		logger.Debugf("last fetched data for data store %s at %s", dataSource.Name, lastTimestamp.String())
	}

	// We don't want to duplicate the lastTimestamp record so add
	// the step size so that we start at the next interval no longer in
	// our range.
	chunkStart := truncateToMinute(lastTimestamp.Add(c.promsumStepSize))
	chunkEnd := truncateToMinute(chunkStart.Add(c.promsumChunkSize))

	var timeRanges []prom.Range
	for {
		// Never collect data in the future, that may not have all data points
		// collected yet
		if !chunkEnd.Before(now) {
			break
		}

		if chunkStart.Equal(chunkEnd) {
			break
		}
		// Only get chunks that are a full chunk size
		if chunkEnd.Sub(chunkStart) < c.promsumChunkSize {
			break
		}

		timeRanges = append(timeRanges, prom.Range{
			Start: chunkStart,
			End:   chunkEnd,
			Step:  c.promsumStepSize,
		})

		// Add the metrics step size to the start time so that we don't
		// re-query the previous ranges end time in this range
		chunkStart = truncateToMinute(chunkEnd.Add(c.promsumStepSize))
		// Add chunkSize to the end time to get our full chunk. If the end is
		// past the current time, then this chunk is skipped.
		chunkEnd = truncateToMinute(chunkStart.Add(c.promsumChunkSize))
	}

	return timeRanges, nil
}

func (c *Chargeback) promsumStoreRecords(logger logrus.FieldLogger, dataSource *cbTypes.ReportDataSource, records []BillingRecord) error {
	var queryValues [][]string

	for _, record := range records {
		queryValues = append(queryValues, []string{generateRecordValues(record)})
	}
	// capacity prestoQueryCap, length 0
	queryBuf := bytes.NewBuffer(make([]byte, 0, prestoQueryCap))
	// queryBuf.Reset()
	queryBuf.WriteString("VALUES ")
	queryBufIsEmpty := true
	for _, values := range queryValues {
		if !queryBufIsEmpty {
			queryBuf.WriteString(",")
		}

		currValue := fmt.Sprintf("(%s)", strings.Join(values, ","))

		queryCap := prestoQueryCap - len(presto.FormatInsertQuery(dataSource.TableName, ""))

		// There's a character limit of prestoQueryCap on insert
		// queries, so let's chunk them at that limit.
		if len(currValue)+queryBuf.Len() > queryCap {
			err := presto.ExecuteInsertQuery(c.prestoConn, dataSource.TableName, queryBuf.String())
			if err != nil {
				return fmt.Errorf("failed to store metrics into presto: %v", err)
			}
			queryBuf.Reset()
			queryBuf.WriteString("VALUES ")
			queryBuf.WriteString(currValue)
			queryBufIsEmpty = false
		} ***REMOVED*** {
			queryBuf.WriteString(currValue)
			queryBufIsEmpty = false
		}
	}
	if !queryBufIsEmpty {
		err := presto.ExecuteInsertQuery(c.prestoConn, dataSource.TableName, queryBuf.String())
		if err != nil {
			return fmt.Errorf("failed to store metrics into presto: %v", err)
		}
	}
	return nil
}

func generateRecordValues(record BillingRecord) string {
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

// BillingRecord is a receipt of a usage determined by a query within a speci***REMOVED***c time range.
type BillingRecord struct {
	Labels    map[string]string `json:"labels"`
	Amount    float64           `json:"amount"`
	StepSize  time.Duration     `json:"stepSize"`
	Timestamp time.Time         `json:"timestamp"`
}

func (c *Chargeback) promsumQuery(query *cbTypes.ReportPrometheusQuery, queryRng prom.Range) ([]BillingRecord, error) {
	pVal, err := c.promConn.QueryRange(context.Background(), query.Spec.Query, queryRng)
	if err != nil {
		return nil, fmt.Errorf("failed to perform billing query: %v", err)
	}

	matrix, ok := pVal.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("expected a matrix in response to query, got a %v", pVal.Type())
	}

	records := []BillingRecord{}
	// iterate over segments of contiguous billing records
	for _, sampleStream := range matrix {
		for _, value := range sampleStream.Values {
			labels := make(map[string]string, len(sampleStream.Metric))
			for k, v := range sampleStream.Metric {
				labels[string(k)] = string(v)
			}

			record := BillingRecord{
				Labels:    labels,
				Amount:    float64(value.Value),
				StepSize:  c.promsumStepSize,
				Timestamp: value.Timestamp.Time().UTC(),
			}
			records = append(records, record)
		}
	}
	return records, nil
}
