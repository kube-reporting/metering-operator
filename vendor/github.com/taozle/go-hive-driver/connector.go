package hive

import (
	"context"
	"database/sql/driver"
	"fmt"
	"net"
	"net/url"
	"strings"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/emersion/go-sasl"
	"github.com/taozle/go-hive-driver/thriftlib"
)

var (
	_ driver.Connector = &Connector{}
)

// Connector represents a fixed configuration for the hive driver with a given
// name. Connector satisfies the database/sql/driver Connector interface and
// can be used to create any number of DB Conn's via the database/sql OpenDB
// function.
//
// See https://golang.org/pkg/database/sql/driver/#Connector.
// See https://golang.org/pkg/database/sql/#OpenDB.
type Connector struct {
	opts   ConnectOptions
	dialer Dialer
}

func (c *Connector) Connect(ctx context.Context) (driver.Conn, error) {
	return c.connect(ctx)
}

func (c *Connector) Driver() driver.Driver {
	return &Driver{}
}

// dsn format:
//	hive://user@host:port?batch=100
func NewConnector(dsn string) (*Connector, error) {
	return NewConnectorWithDialer(DialWrapper{}, dsn)
}

func NewConnectorWithDialer(d Dialer, dsn string) (*Connector, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(u.Scheme) != "hive" {
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}

	opts, err := connectOptionsFromURL(u)
	if err != nil {
		return nil, err
	}

	if opts.BatchSize == 0 {
		opts.BatchSize = defaultBatchSize
	}

	return &Connector{
		dialer: d,
		opts:   opts,
	}, nil
}

// Open opens a new connection to the database. dsn is a connection string.
// Most users should only use it through database/sql package from the standard
// library.
func Open(dsn string) (driver.Conn, error) {
	return DialOpen(DialWrapper{}, dsn)
}

func DialOpen(d Dialer, dsn string) (driver.Conn, error) {
	c, err := NewConnector(dsn)
	if err != nil {
		return nil, err
	}
	c.dialer = d
	return c.connect(context.Background())
}

func (c *Connector) connect(ctx context.Context) (*Connection, error) {
	var conn net.Conn
	var err error
	if timeoutDialer, ok := c.dialer.(TimeoutDialer); ok && c.opts.Timeout != 0 {
		conn, err = timeoutDialer.DialTimeout("tcp", c.opts.Host, c.opts.Timeout)
	} else {
		conn, err = c.dialer.Dial("tcp", c.opts.Host)
	}
	if err != nil {
		return nil, err
	}

	var transport thrift.TTransport
	transport = thrift.NewTSocketFromConnTimeout(conn, c.opts.Timeout)

	if c.opts.AuthMode == "sasl" {
		transport = NewTSaslClientTransport(transport, func() sasl.Client {
			return sasl.NewPlainClient("", c.opts.Username, c.opts.Password)
		})
	}

	protocol := thrift.NewTBinaryProtocolTransport(transport)
	tclient := thrift.NewTStandardClient(protocol, protocol)
	service := thriftlib.NewTCLIServiceClient(tclient)

	req := thriftlib.NewTOpenSessionReq()
	req.ClientProtocol = thriftlib.TProtocolVersion_HIVE_CLI_SERVICE_PROTOCOL_V6
	if c.opts.Username != "" {
		req.Username = &c.opts.Username
	}
	if c.opts.Password != "" {
		req.Password = &c.opts.Password
	}

	session, err := service.OpenSession(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return &Connection{client: service, session: session.SessionHandle, batchSize: c.opts.BatchSize}, nil
}
