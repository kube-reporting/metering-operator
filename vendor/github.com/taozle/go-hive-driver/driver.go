package hive

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/emersion/go-sasl"
	"github.com/taozle/go-hive-driver/thriftlib"
	"net/url"
)

var (
	ErrNoPassword = errors.New("hive: passwd is required")
)

func init() {
	sql.Register("hive", &Driver{})
}

type Driver struct {
}

//
// connString format:
//	hive://user@host:port?batch=100
func (*Driver) Open(connString string) (driver.Conn, error) {
	info, err := url.Parse(connString)
	if err != nil {
		return nil, err
	}

	var transport thrift.TTransport
	transport, err = thrift.NewTSocket(info.Host)
	if err != nil {
		return nil, err
	}

	config := parseConfigFromQuery(info.Query())
	switch config.auth {
	case "sasl":
		passwd, ok := info.User.Password()
		if !ok {
			return nil, ErrNoPassword
		}

		transport = NewTSaslClientTransport(transport, func() sasl.Client {
			return sasl.NewPlainClient("", info.User.Username(), passwd)
		})
	default:
	}

	if err := transport.Open(); err != nil {
		return nil, err
	}

	protocol := thrift.NewTBinaryProtocolTransport(transport)
	tclient := thrift.NewTStandardClient(protocol, protocol)
	service := thriftlib.NewTCLIServiceClient(tclient)

	req := thriftlib.NewTOpenSessionReq()
	req.ClientProtocol = thriftlib.TProtocolVersion_HIVE_CLI_SERVICE_PROTOCOL_V6
	if name := info.User.Username(); name != "" {
		req.Username = &name
	}
	if password, ok := info.User.Password(); ok {
		req.Password = &password
	}

	session, err := service.OpenSession(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return &Connection{client: service, session: session.SessionHandle, config: config}, nil
}
