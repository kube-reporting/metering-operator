// Package defaults is a collection of helpers to retrieve the SDK's default
// con***REMOVED***guration and handlers.
//
// Generally this package shouldn't be used directly, but session.Session
// instead. This package is useful when you need to reset the defaults
// of a session or service client to the SDK defaults before setting
// additional parameters.
package defaults

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/corehandlers"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/credentials/endpointcreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
)

// A Defaults provides a collection of default values for SDK clients.
type Defaults struct {
	Con***REMOVED***g   *aws.Con***REMOVED***g
	Handlers request.Handlers
}

// Get returns the SDK's default values with Con***REMOVED***g and handlers pre-con***REMOVED***gured.
func Get() Defaults {
	cfg := Con***REMOVED***g()
	handlers := Handlers()
	cfg.Credentials = CredChain(cfg, handlers)

	return Defaults{
		Con***REMOVED***g:   cfg,
		Handlers: handlers,
	}
}

// Con***REMOVED***g returns the default con***REMOVED***guration without credentials.
// To retrieve a con***REMOVED***g with credentials also included use
// `defaults.Get().Con***REMOVED***g` instead.
//
// Generally you shouldn't need to use this method directly, but
// is available if you need to reset the con***REMOVED***guration of an
// existing service client or session.
func Con***REMOVED***g() *aws.Con***REMOVED***g {
	return aws.NewCon***REMOVED***g().
		WithCredentials(credentials.AnonymousCredentials).
		WithRegion(os.Getenv("AWS_REGION")).
		WithHTTPClient(http.DefaultClient).
		WithMaxRetries(aws.UseServiceDefaultRetries).
		WithLogger(aws.NewDefaultLogger()).
		WithLogLevel(aws.LogOff).
		WithEndpointResolver(endpoints.DefaultResolver())
}

// Handlers returns the default request handlers.
//
// Generally you shouldn't need to use this method directly, but
// is available if you need to reset the request handlers of an
// existing service client or session.
func Handlers() request.Handlers {
	var handlers request.Handlers

	handlers.Validate.PushBackNamed(corehandlers.ValidateEndpointHandler)
	handlers.Validate.AfterEachFn = request.HandlerListStopOnError
	handlers.Build.PushBackNamed(corehandlers.SDKVersionUserAgentHandler)
	handlers.Build.AfterEachFn = request.HandlerListStopOnError
	handlers.Sign.PushBackNamed(corehandlers.BuildContentLengthHandler)
	handlers.Send.PushBackNamed(corehandlers.ValidateReqSigHandler)
	handlers.Send.PushBackNamed(corehandlers.SendHandler)
	handlers.AfterRetry.PushBackNamed(corehandlers.AfterRetryHandler)
	handlers.ValidateResponse.PushBackNamed(corehandlers.ValidateResponseHandler)

	return handlers
}

// CredChain returns the default credential chain.
//
// Generally you shouldn't need to use this method directly, but
// is available if you need to reset the credentials of an
// existing service client or session's Con***REMOVED***g.
func CredChain(cfg *aws.Con***REMOVED***g, handlers request.Handlers) *credentials.Credentials {
	return credentials.NewCredentials(&credentials.ChainProvider{
		VerboseErrors: aws.BoolValue(cfg.CredentialsChainVerboseErrors),
		Providers: []credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{Filename: "", Pro***REMOVED***le: ""},
			RemoteCredProvider(*cfg, handlers),
		},
	})
}

const (
	httpProviderEnvVar     = "AWS_CONTAINER_CREDENTIALS_FULL_URI"
	ecsCredsProviderEnvVar = "AWS_CONTAINER_CREDENTIALS_RELATIVE_URI"
)

// RemoteCredProvider returns a credentials provider for the default remote
// endpoints such as EC2 or ECS Roles.
func RemoteCredProvider(cfg aws.Con***REMOVED***g, handlers request.Handlers) credentials.Provider {
	if u := os.Getenv(httpProviderEnvVar); len(u) > 0 {
		return localHTTPCredProvider(cfg, handlers, u)
	}

	if uri := os.Getenv(ecsCredsProviderEnvVar); len(uri) > 0 {
		u := fmt.Sprintf("http://169.254.170.2%s", uri)
		return httpCredProvider(cfg, handlers, u)
	}

	return ec2RoleProvider(cfg, handlers)
}

func localHTTPCredProvider(cfg aws.Con***REMOVED***g, handlers request.Handlers, u string) credentials.Provider {
	var errMsg string

	parsed, err := url.Parse(u)
	if err != nil {
		errMsg = fmt.Sprintf("invalid URL, %v", err)
	} ***REMOVED*** if host := aws.URLHostname(parsed); !(host == "localhost" || host == "127.0.0.1") {
		errMsg = fmt.Sprintf("invalid host address, %q, only localhost and 127.0.0.1 are valid.", host)
	}

	if len(errMsg) > 0 {
		if cfg.Logger != nil {
			cfg.Logger.Log("Ignoring, HTTP credential provider", errMsg, err)
		}
		return credentials.ErrorProvider{
			Err:          awserr.New("CredentialsEndpointError", errMsg, err),
			ProviderName: endpointcreds.ProviderName,
		}
	}

	return httpCredProvider(cfg, handlers, u)
}

func httpCredProvider(cfg aws.Con***REMOVED***g, handlers request.Handlers, u string) credentials.Provider {
	return endpointcreds.NewProviderClient(cfg, handlers, u,
		func(p *endpointcreds.Provider) {
			p.ExpiryWindow = 5 * time.Minute
		},
	)
}

func ec2RoleProvider(cfg aws.Con***REMOVED***g, handlers request.Handlers) credentials.Provider {
	resolver := cfg.EndpointResolver
	if resolver == nil {
		resolver = endpoints.DefaultResolver()
	}

	e, _ := resolver.EndpointFor(endpoints.Ec2metadataServiceID, "")
	return &ec2rolecreds.EC2RoleProvider{
		Client:       ec2metadata.NewClient(cfg, handlers, e.URL, e.SigningRegion),
		ExpiryWindow: 5 * time.Minute,
	}
}
