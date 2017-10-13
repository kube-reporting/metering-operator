package chargeback

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/apimachinery/pkg/labels"

	cbTypes "github.com/coreos-inc/kube-chargeback/pkg/apis/chargeback/v1alpha1"
	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback/v1"
	"github.com/coreos-inc/kube-chargeback/pkg/presto"
)

const (
	prestoQueryCap  = 1000000
	timestampFormat = "2006-01-02 15:04:05.000"
)

func (c *Chargeback) runPromsumWorker(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			// if the stopCh is closed while we're waiting, cancel and return
			return
		case <-time.After(c.promsumInterval):
			// after the timeout run the promsum logic
		}

		dataStores, err := c.informers.reportDataStoreLister.ReportDataStores(c.namespace).List(labels.Everything())
		if err != nil {
			c.logger.Errorf("couldn't list data stores: %v", err)
			continue
		}

		for _, dataStore := range dataStores {
			c.logger.Debugf("processing data store %q", dataStore.Name)
			if dataStore.Spec.DataStoreSource.Promsum == nil {
				c.logger.Debugf("not a promsum store, skipping %q", dataStore.Name)
				continue
			}
			if dataStore.TableName == "" {
				// This data store doesn't have a table yet, let's skip it and
				// hope it'll have one next time.
				c.logger.Debugf("no table set, skipping data store %q", dataStore.Name)
				continue
			}

			// Get the most recent timestamp in the table
			getLastTimestampQuery := fmt.Sprintf(`
				SELECT timestamp
				FROM %s
				ORDER BY timestamp DESC
				LIMIT 1`, dataStore.TableName)

			var lastTimestamp time.Time
			results, err := presto.ExecuteSelect(c.prestoConn, getLastTimestampQuery)
			if err != nil {
				c.logger.Warnf("error getting last timestamp for table, maybe table doesn't exist yet? %v", err)
				continue
			}
			if len(results) == 0 {
				// Looks like we haven't populated any data in this table yet.
				// Let's back***REMOVED***ll 6 hours.
				lastTimestamp = time.Now().UTC().Add(time.Hour * -6)
				c.logger.Debugf("no data in data store %s yet, back***REMOVED***lling 6 hours", dataStore.Name)
			} ***REMOVED*** {
				lastTimestamp = results[0]["timestamp"].(time.Time)
				c.logger.Debugf("last fetched data for data store %s at %s", dataStore.Name, lastTimestamp.String())
			}

			if lastTimestamp.After(time.Now().UTC()) {
				c.logger.Errorf("the last timestamp for this data store is in the future! %v", lastTimestamp.String())
				continue
			}

			timeIndex := lastTimestamp

			var timeRanges []cb.Range

			// Chunk the prometheus queries at 5x the precision level
			chunkDurationSize := c.promsumPrecision * 5

			for time.Since(timeIndex) > chunkDurationSize {
				timeRanges = append(timeRanges, cb.Range{
					Start: timeIndex,
					End:   timeIndex.Add(chunkDurationSize),
				})
				timeIndex = timeIndex.Add(chunkDurationSize)
			}

			timeRanges = append(timeRanges, cb.Range{
				Start: timeIndex,
				End:   time.Now().UTC(),
			})

			// capacity prestoQueryCap, length 0
			queryBuf := bytes.NewBuffer(make([]byte, 0, prestoQueryCap))

			for _, queryRng := range timeRanges {
				for _, queryName := range dataStore.Spec.DataStoreSource.Promsum.Queries {
					query, err := c.informers.reportPrometheusQueryLister.ReportPrometheusQueries(c.namespace).Get(queryName)
					if err != nil {
						c.logger.Fatal("Could not get prometheus query: ", err)
					}

					records, err := c.promsumMeter(query, queryRng)
					if err != nil {
						c.logger.Errorf("Failed to generate billing report for query '%s' in the range %v to %v: %v",
							query.Name, queryRng.Start, queryRng.End, err)
						continue
					}

					var queryValues [][]string

					for _, record := range records {
						queryValues = append(queryValues, []string{generateRecordValues(record)})
					}
					queryBuf.Reset()
					queryBuf.WriteString("VALUES ")
					queryBufIsEmpty := true
					for _, values := range queryValues {
						if !queryBufIsEmpty {
							queryBuf.WriteString(",")
						}

						currValue := fmt.Sprintf("(%s)", strings.Join(values, ","))

						// There's a character limit of prestoQueryCap on insert
						// queries, so let's chunk them at that limit.
						if len(currValue)+queryBuf.Len() > prestoQueryCap {
							err = presto.ExecuteInsertQuery(c.prestoConn, dataStore.TableName, queryBuf.String())
							if err != nil {
								c.logger.Errorf("Failed to record: %v", err)
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
						err = presto.ExecuteInsertQuery(c.prestoConn, dataStore.TableName, queryBuf.String())
						if err != nil {
							c.logger.Errorf("Failed to record: %v", err)
						}
					}
				}

			}
			c.logger.Debugf("processing complete for data store %q", dataStore.Name)
		}
	}
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
		record.QueryName, record.Amount, record.Timestamp.Format(timestampFormat), record.TimePrecision.Seconds(), keyString, valString)
}

// BillingRecord is a receipt of a usage determined by a query within a speci***REMOVED***c time range.
type BillingRecord struct {
	Labels        map[string]string `json:"labels"`
	QueryName     string            `json:"query"`
	Amount        float64           `json:"amount"`
	TimePrecision time.Duration     `json:"timePrecision"`
	Timestamp     time.Time         `json:"timestamp"`
}

func (c *Chargeback) promsumMeter(query *cbTypes.ReportPrometheusQuery, queryRng cb.Range) ([]BillingRecord, error) {
	pRng := prom.Range{
		Start: queryRng.Start,
		End:   queryRng.End,
		Step:  c.promsumPrecision,
	}

	pVal, err := c.promConn.QueryRange(context.Background(), query.Spec.Query, pRng)
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
				Labels:        labels,
				QueryName:     query.Name,
				Amount:        float64(value.Value),
				TimePrecision: c.promsumPrecision,
				Timestamp:     value.Timestamp.Time().UTC(),
			}
			records = append(records, record)
		}
	}
	return records, nil
}
