package hive

import (
	"context"
	"database/sql/driver"
	"errors"
	"github.com/taozle/go-hive-driver/thriftlib"
	"io"
	"reflect"
	"sync"
)

type rowSet struct {
	client   *thriftlib.TCLIServiceClient
	opHandle *thriftlib.TOperationHandle
	config   *config

	columns    []*thriftlib.TColumnDesc
	columnOnce sync.Once
	values     [][]interface{}
}

func newRowSet(client *thriftlib.TCLIServiceClient, opHandle *thriftlib.TOperationHandle, config *config) *rowSet {
	return &rowSet{
		client:   client,
		opHandle: opHandle,
		config:   config,
	}
}

func (r *rowSet) Bootstrap() error {
	return r.fetchNext()
}

// getColumns should be called after fetching result once.
func (r *rowSet) getColumns() []*thriftlib.TColumnDesc {
	r.columnOnce.Do(func() {
		ctx := context.Background()
		req := thriftlib.NewTGetResultSetMetadataReq()
		req.OperationHandle = r.opHandle

		resp, err := r.client.GetResultSetMetadata(ctx, req)
		if err != nil {
			return
		}

		if err := formatResponseErr(resp.GetStatus()); err != nil {
			return
		}

		r.columns = resp.GetSchema().GetColumns()
	})

	return r.columns
}

func (r *rowSet) Columns() []string {
	var columns []string
	for _, column := range r.getColumns() {
		columns = append(columns, column.GetColumnName())
	}
	return columns
}

func (r *rowSet) Close() error {
	req := thriftlib.NewTCloseOperationReq()
	req.OperationHandle = r.opHandle

	resp, err := r.client.CloseOperation(context.Background(), req)
	if err != nil {
		return err
	}

	if err := formatResponseErr(resp.GetStatus()); err != nil {
		return err
	}

	return nil
}

func (r *rowSet) fetchNext() error {
	req := thriftlib.NewTFetchResultsReq()
	req.OperationHandle = r.opHandle
	req.Orientation = thriftlib.TFetchOrientation_FETCH_NEXT
	req.MaxRows = r.config.batchSize

	resp, err := r.client.FetchResults(context.Background(), req)
	if err != nil {
		return err
	}

	if err := formatResponseErr(resp.GetStatus()); err != nil {
		return err
	}

	columns := r.getColumns()
	if len(columns) == 0 {
		return errors.New("hive: fetch column meta info failed")
	}

	valLen := 0
	columnLen := len(columns)
	columnValues := make([][]interface{}, columnLen)
	for i, column := range columns {
		values, err := convertColumnValues(column, resp.GetResults().GetColumns()[i])
		if err != nil {
			return err
		}

		v := reflect.ValueOf(values)
		columnValues[i] = make([]interface{}, v.Len())
		for j := 0; j < v.Len(); j++ {
			columnValues[i][j] = v.Index(j).Interface()
		}
		valLen = v.Len()
	}

	for i := 0; i < valLen; i++ {
		value := make([]interface{}, columnLen)
		for j := 0; j < columnLen; j++ {
			value[j] = columnValues[j][i]
		}

		r.values = append(r.values, value)
	}

	return nil
}

func (r *rowSet) Next(dest []driver.Value) error {
	// fetch more values
	if len(r.values) == 0 {
		if err := r.fetchNext(); err != nil {
			return err
		}
	}
	if len(r.values) == 0 {
		return io.EOF
	}

	for i, v := range r.values[0] {
		dest[i] = v
	}
	r.values = r.values[1:]
	return nil
}
