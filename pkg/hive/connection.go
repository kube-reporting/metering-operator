package hive

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	hive "github.com/operator-framework/operator-metering/pkg/hive/hive_thrift"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	// ThriftVersion is the version of the Thrift protocol used to connect to Hive.
	ThriftVersion = hive.TProtocolVersion_HIVE_CLI_SERVICE_PROTOCOL_V8
)

// Connection to a Hive server.
type Connection struct {
	client    *hive.TCLIServiceClient
	transport *thrift.TSocket
	session   *hive.TSessionHandle
	queryLock sync.Mutex
}

// Connect to a Hive cluster.
func Connect(host string) (*Connection, error) {
	transport, err := thrift.NewTSocket(host)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to '%s': %v", host, err)
	}

	if err = transport.Open(); err != nil {
		return nil, err
	} ***REMOVED*** if transport == nil {
		return nil, errors.New("nil thrift socket")
	}

	protocol := thrift.NewTBinaryProtocolFactoryDefault()
	client := hive.NewTCLIServiceClientFactory(transport, protocol)

	req := hive.NewTOpenSessionReq()
	req.ClientProtocol = ThriftVersion

	resp, err := client.OpenSession(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("attempt to open session failed: %v", err)
	} ***REMOVED*** if resp.SessionHandle == nil {
		return nil, errors.New("session handler was nil")
	}

	return &Connection{
		client:    client,
		transport: transport,
		session:   resp.SessionHandle,
	}, nil
}

// Query a Hive server.
func (c *Connection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	// Only perform one query at a time
	c.queryLock.Lock()
	defer c.queryLock.Unlock()

	req := hive.NewTExecuteStatementReq()
	req.SessionHandle = c.session
	req.Statement = query

	resp, err := c.client.ExecuteStatement(context.Background(), req)
	if err != nil {
		return nil, err
	}

	switch resp.Status.GetStatusCode() {
	case hive.TStatusCode_SUCCESS_STATUS:
	case hive.TStatusCode_SUCCESS_WITH_INFO_STATUS:
	default:
		return nil, fmt.Errorf("encountered error: code: %d, sqlState: %s, message: %s", resp.Status.GetErrorCode(), resp.Status.GetSqlState(), resp.Status.GetErrorMessage())
	}
	return nil, nil
}

// Close connection to Hive server.
func (c *Connection) Close() error {
	// Wait for any current queries to ***REMOVED***nish
	c.queryLock.Lock()
	defer c.transport.Close()
	defer c.queryLock.Unlock()
	if c.session != nil {
		req := hive.NewTCloseSessionReq()
		req.SessionHandle = c.session
		if resp, err := c.client.CloseSession(context.Background(), req); err != nil {
			return fmt.Errorf("couldn't close connection: %+v, %v", resp, err)
		}
		c.session = nil
	}
	return nil
}

// reconnectingQueryer implements db.Queryer and will attempt to transparent
// reconnect when a query fails due to a connection related error
type reconnectingQueryer struct {
	hiveHost    string
	mu          sync.Mutex
	conn        *Connection
	logger      log.FieldLogger
	maxRetries  int
	connBackoff time.Duration
	ctx         context.Context
}

// NewReconnectingQueryer returns a reconnectingQueryer that will not attempt
// to reconnect once the ctx is cancelled.
func NewReconnectingQueryer(ctx context.Context, logger log.FieldLogger, hiveHost string, connBackoff time.Duration, maxRetries int) *reconnectingQueryer {
	return &reconnectingQueryer{
		hiveHost:    hiveHost,
		logger:      logger,
		connBackoff: connBackoff,
		maxRetries:  maxRetries,
		ctx:         ctx,
	}
}

func (q *reconnectingQueryer) Query(query string, args ...interface{}) (*sql.Rows, error) {
	for retries := 0; retries < q.maxRetries; retries++ {
		conn, err := q.getConnection(q.ctx)
		if err != nil {
			if err == io.EOF || isErrBrokenPipe(err) {
				q.logger.WithError(err).Debugf("error occurred while getting connection, attempting to create new connection and retry")
				q.Close()
				continue
			}
			// We don't close the connection here because we got an error while
			// getting it
			return nil, err
		}
		rows, err := conn.Query(query)
		if err != nil {
			if err == io.EOF || isErrBrokenPipe(err) {
				q.logger.WithError(err).Debugf("error occurred while making query, attempting to create new connection and retry")
				q.Close()
				continue
			}
			// We don't close the connection here because we got a good
			// connection, and made the query, but the query itself had an
			// error.
			return nil, err
		}
		return rows, nil
	}

	// We've tries 3 times, so close any connection and return an error
	q.Close()
	return nil, fmt.Errorf("unable to create new hive connection after existing hive connection closed")
}

func (q *reconnectingQueryer) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	var err error
	if q.conn != nil {
		err = q.conn.Close()
		// Discard our connection so we create a new one in getConnection
		q.conn = nil
	}
	return err
}

// getConnection will return the existing connection if one exists, or will
// attempt to create a new one if one doesn't exist
func (q *reconnectingQueryer) getConnection(ctx context.Context) (*Connection, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	var err error
	if q.conn == nil {
		q.conn, err = q.newConnection(ctx)
	}
	return q.conn, err
}

func (q *reconnectingQueryer) newConnection(ctx context.Context) (*Connection, error) {
	var conn *Connection
	backoff := wait.Backoff{
		Duration: q.connBackoff,
		Factor:   1.25,
		Steps:    q.maxRetries,
	}
	cond := func() (bool, error) {
		// check for cancellation
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			var err error
			conn, err = Connect(q.hiveHost)
			if err == nil {
				return true, nil
			} ***REMOVED*** {
				q.logger.WithError(err).Debugf("error encountered when connecting to hive, backing off and trying again")
			}
			return false, nil
		}
	}

	return conn, wait.ExponentialBackoff(backoff, cond)
}

func isErrBrokenPipe(err error) bool {
	if netErr, ok := err.(*net.OpError); ok {
		return netErr.Err == syscall.EPIPE
	}
	return false
}
