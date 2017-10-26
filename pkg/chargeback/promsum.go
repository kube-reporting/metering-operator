package chargeback

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	"github.com/coreos-inc/kube-chargeback/pkg/presto"
)

const (
	prestoQueryCap  = 1000000
	timestampFormat = "2006-01-02 15:04:05.000"
)

func (c *Chargeback) runPromsumWorker(stopCh <-chan struct{}) {
	ticker := time.NewTicker(c.promsumInterval)
	defer ticker.Stop()

	c.collectPromsumData()

	for {
		select {
		case <-stopCh:
			// if the stopCh is closed while we're waiting, cancel and return
			return
		case <-ticker.C:
			c.collectPromsumData()
		}
	}
}

func (c *Chargeback) collectPromsumData() {
	dataStores, err := c.informers.reportDataStoreLister.ReportDataStores(c.namespace).List(labels.Everything())
	if err != nil {
		c.logger.Errorf("couldn't list data stores: %v", err)
		return
	}

	now := time.Now().UTC()
	for _, dataStore := range dataStores {
		logger := c.logger.WithField("datastore", dataStore.Name)
		err := c.collectPromsumDatastoreData(logger, dataStore, now)
		if err != nil {
			logger.WithError(err).Errorf("error collecting promsum data for datastore")
		}
	}
}

func (c *Chargeback) collectPromsumDatastoreData(logger logrus.FieldLogger, dataStore *cbTypes.ReportDataStore, now time.Time) error {
	logger.Debugf("processing data store %q", dataStore.Name)
	if dataStore.Spec.DataStoreSource.Promsum == nil {
		logger.Debugf("not a promsum store, skipping %q", dataStore.Name)
		return nil
	}
	if dataStore.TableName == "" {
		// This data store doesn't have a table yet, let's skip it and
		// hope it'll have one next time.
		logger.Debugf("no table set, skipping data store %q", dataStore.Name)
		return nil
	}

	for _, queryName := range dataStore.Spec.DataStoreSource.Promsum.Queries {
		timeRanges, err := c.promsumGetTimeRanges(logger, dataStore, queryName, now)
		if err != nil {
			return fmt.Errorf("couldn't get time ranges to for query %s: %v", queryName, err)
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
			query, err := c.informers.reportPrometheusQueryLister.ReportPrometheusQueries(c.namespace).Get(queryName)
			if err != nil {
				return fmt.Errorf("could not get prometheus query: ", err)
			}

			records, err := c.promsumQuery(query, queryRng)
			if err != nil {
				return fmt.Errorf("failed to retrieve prometheus metrics for query '%s' in the range %v to %v: %v",
					query.Name, queryRng.Start, queryRng.End, err)
			}

			err = c.promsumStoreRecords(logger, dataStore, records)
			if err != nil {
				return fmt.Errorf("failed to store prometheus metrics for query '%s' in the range %v to %v: %v",
					query.Name, queryRng.Start, queryRng.End, err)
			}
		}
	}
	logger.Debugf("processing complete for data store %q", dataStore.Name)
	return nil
}

func (c *Chargeback) promsumGetTimeRanges(logger logrus.FieldLogger, dataStore *cbTypes.ReportDataStore, queryName string, now time.Time) ([]prom.Range, error) {
	// Get the most recent timestamp in the table for this query
	getLastTimestampQuery := fmt.Sprintf(`
				SELECT "timestamp"
				FROM %s
				WHERE query = '%s'
				ORDER BY "timestamp" DESC
				LIMIT 1`, dataStore.TableName, resourceNameReplacer.Replace(queryName))

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
		logger.Debugf("no data in data store %s yet", dataStore.Name)
	} ***REMOVED*** {
		lastTimestamp = results[0]["timestamp"].(time.Time)
		// We don't want duplicate the lastTimestamp record so add
		// the step size so that we start at the next interval no longer in
		// our range.
		lastTimestamp = lastTimestamp.Add(c.promsumStepSize)
		logger.Debugf("last fetched data for data store %s at %s", dataStore.Name, lastTimestamp.String())
	}

	if lastTimestamp.After(now) {
		return nil, fmt.Errorf("the last timestamp for this data store is in the future! %v", lastTimestamp.String())
	}

	chunkStart := lastTimestamp
	chunkEnd := chunkStart.Add(c.promsumChunkSize)

	if chunkStart.Equal(chunkEnd) {
		return nil, fmt.Errorf("error querying for data, start and end are the same: %v", chunkStart.String())
	}

	var timeRanges []prom.Range
	// Only get chunks that are a full chunk size
	for chunkEnd.Before(now) {
		if chunkStart.Equal(chunkEnd) {
			logger.Warnf("skipping querying for data, start and end are the same: %v", chunkStart.String())
			continue
		}
		timeRanges = append(timeRanges, prom.Range{
			Start: chunkStart,
			End:   chunkEnd,
			Step:  c.promsumStepSize,
		})

		// Add the metrics step size to the start time so that we don't
		// re-query the previous ranges end time in this range
		chunkStart = chunkEnd.Add(c.promsumStepSize)
		chunkEnd = chunkStart.Add(c.promsumChunkSize)
	}

	return timeRanges, nil
}

func (c *Chargeback) promsumStoreRecords(logger logrus.FieldLogger, dataStore *cbTypes.ReportDataStore, records []BillingRecord) error {
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

		queryCap := prestoQueryCap - len(presto.FormatInsertQuery(dataStore.TableName, ""))

		// There's a character limit of prestoQueryCap on insert
		// queries, so let's chunk them at that limit.
		if len(currValue)+queryBuf.Len() > queryCap {
			err := presto.ExecuteInsertQuery(c.prestoConn, dataStore.TableName, queryBuf.String())
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
		err := presto.ExecuteInsertQuery(c.prestoConn, dataStore.TableName, queryBuf.String())
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
	return fmt.Sprintf("('%s',%f,timestamp '%s',%f,map(%s,%s))",
		record.QueryName, record.Amount, record.Timestamp.Format(timestampFormat), record.StepSize.Seconds(), keyString, valString)
}

// BillingRecord is a receipt of a usage determined by a query within a speci***REMOVED***c time range.
type BillingRecord struct {
	Labels    map[string]string `json:"labels"`
	QueryName string            `json:"query"`
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
				QueryName: resourceNameReplacer.Replace(query.Name),
				Amount:    float64(value.Value),
				StepSize:  c.promsumStepSize,
				Timestamp: value.Timestamp.Time().UTC(),
			}
			records = append(records, record)
		}
	}
	return records, nil
}
