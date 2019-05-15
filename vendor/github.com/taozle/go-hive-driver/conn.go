package hive

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/taozle/go-hive-driver/thriftlib"
)

var (
	ErrUnsupportedPreparedStmt = errors.New("hive: hive doesn't support prepared stmt")
)

type Connection struct {
	client  *thriftlib.TCLIServiceClient
	session *thriftlib.TSessionHandle
	config  *config
}

func (*Connection) Prepare(query string) (driver.Stmt, error) {
	panic("hive: doesn't support prepared statements")
}

func (*Connection) Begin() (driver.Tx, error) {
	panic("hive: doesn't support transaction")
}

func (c *Connection) Close() error {
	req := thriftlib.NewTCloseSessionReq()
	req.SessionHandle = c.session

	resp, err := c.client.CloseSession(context.Background(), req)
	if err != nil {
		return err
	}

	if err := formatResponseErr(resp.GetStatus()); err != nil {
		return err
	}

	return nil
}

func (c *Connection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if len(args) > 0 {
		return nil, ErrUnsupportedPreparedStmt
	}

	req := thriftlib.NewTExecuteStatementReq()
	req.SessionHandle = c.session
	req.Statement = query
	ret, err := c.client.ExecuteStatement(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := formatResponseErr(ret.GetStatus()); err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *Connection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if len(args) > 0 {
		return nil, ErrUnsupportedPreparedStmt
	}

	req := thriftlib.NewTExecuteStatementReq()
	req.SessionHandle = c.session
	req.Statement = query
	ret, err := c.client.ExecuteStatement(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := formatResponseErr(ret.GetStatus()); err != nil {
		return nil, err
	}

	rs := newRowSet(c.client, ret.GetOperationHandle(), c.config)
	if err := rs.Bootstrap(); err != nil {
		return nil, err
	}

	return rs, nil
}

func formatResponseErr(status *thriftlib.TStatus) error {
	code := status.GetStatusCode()
	if code == thriftlib.TStatusCode_SUCCESS_STATUS || code == thriftlib.TStatusCode_SUCCESS_WITH_INFO_STATUS {
		return nil
	}

	return fmt.Errorf("hive: query failed. errmsg=%s", status.GetErrorMessage())
}
