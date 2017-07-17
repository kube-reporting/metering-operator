package hive

import (
	"errors"
	"fmt"

	"git.apache.org/thrift.git/lib/go/thrift"

	hive "github.com/coreos-inc/kube-chargeback/pkg/hive/hive_thrift"
)

var (
	// ThriftVersion is the version of the Thrift protocol used to connect to Hive.
	ThriftVersion = hive.TProtocolVersion_HIVE_CLI_SERVICE_PROTOCOL_V8
)

// Connection to a Hive server.
type Connection struct {
	client  *hive.TCLIServiceClient
	session *hive.TSessionHandle
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

	resp, err := client.OpenSession(req)
	if err != nil {
		return nil, fmt.Errorf("attempt to open session failed: %v", err)
	} ***REMOVED*** if resp.SessionHandle == nil {
		return nil, errors.New("session handler was nil")
	}

	return &Connection{
		client:  client,
		session: resp.SessionHandle,
	}, nil
}

// Query a Hive server.
func (c *Connection) Query(query string) error {
	req := hive.NewTExecuteStatementReq()
	req.SessionHandle = c.session
	req.Statement = query

	resp, err := c.client.ExecuteStatement(req)
	if err != nil {
		return fmt.Errorf("Error executing query '%s':  %+v, %v", query, resp, err)
	}

	switch resp.Status.GetStatusCode() {
	case hive.TStatusCode_SUCCESS_STATUS:
	case hive.TStatusCode_SUCCESS_WITH_INFO_STATUS:
	default:
		return fmt.Errorf("encountered error: %s", resp.Status.String())
	}
	return nil
}

// Close connection to Hive server.
func (c *Connection) Close() error {
	if c.session != nil {
		req := hive.NewTCloseSessionReq()
		req.SessionHandle = c.session
		if resp, err := c.client.CloseSession(req); err != nil {
			return fmt.Errorf("couldn't close connection: %+v, %v", resp, err)
		}
	}
	return nil
}
