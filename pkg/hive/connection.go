package hive

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/sirupsen/logrus"

	hive "github.com/operator-framework/operator-metering/pkg/hive/hive_thrift"
)

var (
	// ThriftVersion is the version of the Thrift protocol used to connect to Hive.
	ThriftVersion = hive.TProtocolVersion_HIVE_CLI_SERVICE_PROTOCOL_V8
)

// Connection to a Hive server.
type Connection struct {
	client     *hive.TCLIServiceClient
	transport  *thrift.TSocket
	session    *hive.TSessionHandle
	logger     log.FieldLogger
	logQueries bool
	queryLock  sync.Mutex
}

type Queryer interface {
	Query(query string) error
}

// Connect to a Hive cluster.
func Connect(host string) (*Connection, error) {
	logger := log.WithField("package", "hive")
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
		logger:    logger,
	}, nil
}

// Query a Hive server.
func (c *Connection) Query(query string) error {
	// Only perform one query at a time
	c.queryLock.Lock()
	defer c.queryLock.Unlock()

	req := hive.NewTExecuteStatementReq()
	req.SessionHandle = c.session
	req.Statement = query

	if c.logQueries {
		c.logger.Debugf("QUERY: \n%s\n", query)
	}
	resp, err := c.client.ExecuteStatement(context.Background(), req)
	if err != nil {
		return err
	}

	switch resp.Status.GetStatusCode() {
	case hive.TStatusCode_SUCCESS_STATUS:
	case hive.TStatusCode_SUCCESS_WITH_INFO_STATUS:
	default:
		return fmt.Errorf("encountered error: code: %d, sqlState: %s, message: %s", resp.Status.GetErrorCode(), resp.Status.GetSqlState(), resp.Status.GetErrorMessage())
	}
	return nil
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

func (c *Connection) SetLogQueries(logQueries bool) {
	c.logQueries = logQueries
}
