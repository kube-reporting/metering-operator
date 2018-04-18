package client

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/request"
)

// A Con***REMOVED***g provides con***REMOVED***guration to a service client instance.
type Con***REMOVED***g struct {
	Con***REMOVED***g        *aws.Con***REMOVED***g
	Handlers      request.Handlers
	Endpoint      string
	SigningRegion string
	SigningName   string

	// States that the signing name did not come from a modeled source but
	// was derived based on other data. Used by service client constructors
	// to determine if the signin name can be overriden based on metadata the
	// service has.
	SigningNameDerived bool
}

// Con***REMOVED***gProvider provides a generic way for a service client to receive
// the ClientCon***REMOVED***g without circular dependencies.
type Con***REMOVED***gProvider interface {
	ClientCon***REMOVED***g(serviceName string, cfgs ...*aws.Con***REMOVED***g) Con***REMOVED***g
}

// Con***REMOVED***gNoResolveEndpointProvider same as Con***REMOVED***gProvider except it will not
// resolve the endpoint automatically. The service client's endpoint must be
// provided via the aws.Con***REMOVED***g.Endpoint ***REMOVED***eld.
type Con***REMOVED***gNoResolveEndpointProvider interface {
	ClientCon***REMOVED***gNoResolveEndpoint(cfgs ...*aws.Con***REMOVED***g) Con***REMOVED***g
}

// A Client implements the base client request and response handling
// used by all service clients.
type Client struct {
	request.Retryer
	metadata.ClientInfo

	Con***REMOVED***g   aws.Con***REMOVED***g
	Handlers request.Handlers
}

// New will return a pointer to a new initialized service client.
func New(cfg aws.Con***REMOVED***g, info metadata.ClientInfo, handlers request.Handlers, options ...func(*Client)) *Client {
	svc := &Client{
		Con***REMOVED***g:     cfg,
		ClientInfo: info,
		Handlers:   handlers.Copy(),
	}

	switch retryer, ok := cfg.Retryer.(request.Retryer); {
	case ok:
		svc.Retryer = retryer
	case cfg.Retryer != nil && cfg.Logger != nil:
		s := fmt.Sprintf("WARNING: %T does not implement request.Retryer; using DefaultRetryer instead", cfg.Retryer)
		cfg.Logger.Log(s)
		fallthrough
	default:
		maxRetries := aws.IntValue(cfg.MaxRetries)
		if cfg.MaxRetries == nil || maxRetries == aws.UseServiceDefaultRetries {
			maxRetries = 3
		}
		svc.Retryer = DefaultRetryer{NumMaxRetries: maxRetries}
	}

	svc.AddDebugHandlers()

	for _, option := range options {
		option(svc)
	}

	return svc
}

// NewRequest returns a new Request pointer for the service API
// operation and parameters.
func (c *Client) NewRequest(operation *request.Operation, params interface{}, data interface{}) *request.Request {
	return request.New(c.Con***REMOVED***g, c.ClientInfo, c.Handlers, c.Retryer, operation, params, data)
}

// AddDebugHandlers injects debug logging handlers into the service to log request
// debug information.
func (c *Client) AddDebugHandlers() {
	if !c.Con***REMOVED***g.LogLevel.AtLeast(aws.LogDebug) {
		return
	}

	c.Handlers.Send.PushFrontNamed(request.NamedHandler{Name: "awssdk.client.LogRequest", Fn: logRequest})
	c.Handlers.Send.PushBackNamed(request.NamedHandler{Name: "awssdk.client.LogResponse", Fn: logResponse})
}
