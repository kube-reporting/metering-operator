package aws

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
)

// UseServiceDefaultRetries instructs the con***REMOVED***g to use the service's own
// default number of retries. This will be the default action if
// Con***REMOVED***g.MaxRetries is nil also.
const UseServiceDefaultRetries = -1

// RequestRetryer is an alias for a type that implements the request.Retryer
// interface.
type RequestRetryer interface{}

// A Con***REMOVED***g provides service con***REMOVED***guration for service clients. By default,
// all clients will use the defaults.DefaultCon***REMOVED***g tructure.
//
//     // Create Session with MaxRetry con***REMOVED***guration to be shared by multiple
//     // service clients.
//     sess := session.Must(session.NewSession(&aws.Con***REMOVED***g{
//         MaxRetries: aws.Int(3),
//     }))
//
//     // Create S3 service client with a speci***REMOVED***c Region.
//     svc := s3.New(sess, &aws.Con***REMOVED***g{
//         Region: aws.String("us-west-2"),
//     })
type Con***REMOVED***g struct {
	// Enables verbose error printing of all credential chain errors.
	// Should be used when wanting to see all errors while attempting to
	// retrieve credentials.
	CredentialsChainVerboseErrors *bool

	// The credentials object to use when signing requests. Defaults to a
	// chain of credential providers to search for credentials in environment
	// variables, shared credential ***REMOVED***le, and EC2 Instance Roles.
	Credentials *credentials.Credentials

	// An optional endpoint URL (hostname only or fully quali***REMOVED***ed URI)
	// that overrides the default generated endpoint for a client. Set this
	// to `""` to use the default generated endpoint.
	//
	// @note You must still provide a `Region` value when specifying an
	//   endpoint for a client.
	Endpoint *string

	// The resolver to use for looking up endpoints for AWS service clients
	// to use based on region.
	EndpointResolver endpoints.Resolver

	// EnforceShouldRetryCheck is used in the AfterRetryHandler to always call
	// ShouldRetry regardless of whether or not if request.Retryable is set.
	// This will utilize ShouldRetry method of custom retryers. If EnforceShouldRetryCheck
	// is not set, then ShouldRetry will only be called if request.Retryable is nil.
	// Proper handling of the request.Retryable ***REMOVED***eld is important when setting this ***REMOVED***eld.
	EnforceShouldRetryCheck *bool

	// The region to send requests to. This parameter is required and must
	// be con***REMOVED***gured globally or on a per-client basis unless otherwise
	// noted. A full list of regions is found in the "Regions and Endpoints"
	// document.
	//
	// @see http://docs.aws.amazon.com/general/latest/gr/rande.html
	//   AWS Regions and Endpoints
	Region *string

	// Set this to `true` to disable SSL when sending requests. Defaults
	// to `false`.
	DisableSSL *bool

	// The HTTP client to use when sending requests. Defaults to
	// `http.DefaultClient`.
	HTTPClient *http.Client

	// An integer value representing the logging level. The default log level
	// is zero (LogOff), which represents no logging. To enable logging set
	// to a LogLevel Value.
	LogLevel *LogLevelType

	// The logger writer interface to write logging messages to. Defaults to
	// standard out.
	Logger Logger

	// The maximum number of times that a request will be retried for failures.
	// Defaults to -1, which defers the max retry setting to the service
	// speci***REMOVED***c con***REMOVED***guration.
	MaxRetries *int

	// Retryer guides how HTTP requests should be retried in case of
	// recoverable failures.
	//
	// When nil or the value does not implement the request.Retryer interface,
	// the client.DefaultRetryer will be used.
	//
	// When both Retryer and MaxRetries are non-nil, the former is used and
	// the latter ignored.
	//
	// To set the Retryer ***REMOVED***eld in a type-safe manner and with chaining, use
	// the request.WithRetryer helper function:
	//
	//   cfg := request.WithRetryer(aws.NewCon***REMOVED***g(), myRetryer)
	//
	Retryer RequestRetryer

	// Disables semantic parameter validation, which validates input for
	// missing required ***REMOVED***elds and/or other semantic request input errors.
	DisableParamValidation *bool

	// Disables the computation of request and response checksums, e.g.,
	// CRC32 checksums in Amazon DynamoDB.
	DisableComputeChecksums *bool

	// Set this to `true` to force the request to use path-style addressing,
	// i.e., `http://s3.amazonaws.com/BUCKET/KEY`. By default, the S3 client
	// will use virtual hosted bucket addressing when possible
	// (`http://BUCKET.s3.amazonaws.com/KEY`).
	//
	// @note This con***REMOVED***guration option is speci***REMOVED***c to the Amazon S3 service.
	// @see http://docs.aws.amazon.com/AmazonS3/latest/dev/VirtualHosting.html
	//   Amazon S3: Virtual Hosting of Buckets
	S3ForcePathStyle *bool

	// Set this to `true` to disable the SDK adding the `Expect: 100-Continue`
	// header to PUT requests over 2MB of content. 100-Continue instructs the
	// HTTP client not to send the body until the service responds with a
	// `continue` status. This is useful to prevent sending the request body
	// until after the request is authenticated, and validated.
	//
	// http://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectPUT.html
	//
	// 100-Continue is only enabled for Go 1.6 and above. See `http.Transport`'s
	// `ExpectContinueTimeout` for information on adjusting the continue wait
	// timeout. https://golang.org/pkg/net/http/#Transport
	//
	// You should use this flag to disble 100-Continue if you experience issues
	// with proxies or third party S3 compatible services.
	S3Disable100Continue *bool

	// Set this to `true` to enable S3 Accelerate feature. For all operations
	// compatible with S3 Accelerate will use the accelerate endpoint for
	// requests. Requests not compatible will fall back to normal S3 requests.
	//
	// The bucket must be enable for accelerate to be used with S3 client with
	// accelerate enabled. If the bucket is not enabled for accelerate an error
	// will be returned. The bucket name must be DNS compatible to also work
	// with accelerate.
	S3UseAccelerate *bool

	// S3DisableContentMD5Validation con***REMOVED***g option is temporarily disabled,
	// For S3 GetObject API calls, #1837.
	//
	// Set this to `true` to disable the S3 service client from automatically
	// adding the ContentMD5 to S3 Object Put and Upload API calls. This option
	// will also disable the SDK from performing object ContentMD5 validation
	// on GetObject API calls.
	S3DisableContentMD5Validation *bool

	// Set this to `true` to disable the EC2Metadata client from overriding the
	// default http.Client's Timeout. This is helpful if you do not want the
	// EC2Metadata client to create a new http.Client. This options is only
	// meaningful if you're not already using a custom HTTP client with the
	// SDK. Enabled by default.
	//
	// Must be set and provided to the session.NewSession() in order to disable
	// the EC2Metadata overriding the timeout for default credentials chain.
	//
	// Example:
	//    sess := session.Must(session.NewSession(aws.NewCon***REMOVED***g()
	//       .WithEC2MetadataDiableTimeoutOverride(true)))
	//
	//    svc := s3.New(sess)
	//
	EC2MetadataDisableTimeoutOverride *bool

	// Instructs the endpoint to be generated for a service client to
	// be the dual stack endpoint. The dual stack endpoint will support
	// both IPv4 and IPv6 addressing.
	//
	// Setting this for a service which does not support dual stack will fail
	// to make requets. It is not recommended to set this value on the session
	// as it will apply to all service clients created with the session. Even
	// services which don't support dual stack endpoints.
	//
	// If the Endpoint con***REMOVED***g value is also provided the UseDualStack flag
	// will be ignored.
	//
	// Only supported with.
	//
	//     sess := session.Must(session.NewSession())
	//
	//     svc := s3.New(sess, &aws.Con***REMOVED***g{
	//         UseDualStack: aws.Bool(true),
	//     })
	UseDualStack *bool

	// SleepDelay is an override for the func the SDK will call when sleeping
	// during the lifecycle of a request. Speci***REMOVED***cally this will be used for
	// request delays. This value should only be used for testing. To adjust
	// the delay of a request see the aws/client.DefaultRetryer and
	// aws/request.Retryer.
	//
	// SleepDelay will prevent any Context from being used for canceling retry
	// delay of an API operation. It is recommended to not use SleepDelay at all
	// and specify a Retryer instead.
	SleepDelay func(time.Duration)

	// DisableRestProtocolURICleaning will not clean the URL path when making rest protocol requests.
	// Will default to false. This would only be used for empty directory names in s3 requests.
	//
	// Example:
	//    sess := session.Must(session.NewSession(&aws.Con***REMOVED***g{
	//         DisableRestProtocolURICleaning: aws.Bool(true),
	//    }))
	//
	//    svc := s3.New(sess)
	//    out, err := svc.GetObject(&s3.GetObjectInput {
	//    	Bucket: aws.String("bucketname"),
	//    	Key: aws.String("//foo//bar//moo"),
	//    })
	DisableRestProtocolURICleaning *bool
}

// NewCon***REMOVED***g returns a new Con***REMOVED***g pointer that can be chained with builder
// methods to set multiple con***REMOVED***guration values inline without using pointers.
//
//     // Create Session with MaxRetry con***REMOVED***guration to be shared by multiple
//     // service clients.
//     sess := session.Must(session.NewSession(aws.NewCon***REMOVED***g().
//         WithMaxRetries(3),
//     ))
//
//     // Create S3 service client with a speci***REMOVED***c Region.
//     svc := s3.New(sess, aws.NewCon***REMOVED***g().
//         WithRegion("us-west-2"),
//     )
func NewCon***REMOVED***g() *Con***REMOVED***g {
	return &Con***REMOVED***g{}
}

// WithCredentialsChainVerboseErrors sets a con***REMOVED***g verbose errors boolean and returning
// a Con***REMOVED***g pointer.
func (c *Con***REMOVED***g) WithCredentialsChainVerboseErrors(verboseErrs bool) *Con***REMOVED***g {
	c.CredentialsChainVerboseErrors = &verboseErrs
	return c
}

// WithCredentials sets a con***REMOVED***g Credentials value returning a Con***REMOVED***g pointer
// for chaining.
func (c *Con***REMOVED***g) WithCredentials(creds *credentials.Credentials) *Con***REMOVED***g {
	c.Credentials = creds
	return c
}

// WithEndpoint sets a con***REMOVED***g Endpoint value returning a Con***REMOVED***g pointer for
// chaining.
func (c *Con***REMOVED***g) WithEndpoint(endpoint string) *Con***REMOVED***g {
	c.Endpoint = &endpoint
	return c
}

// WithEndpointResolver sets a con***REMOVED***g EndpointResolver value returning a
// Con***REMOVED***g pointer for chaining.
func (c *Con***REMOVED***g) WithEndpointResolver(resolver endpoints.Resolver) *Con***REMOVED***g {
	c.EndpointResolver = resolver
	return c
}

// WithRegion sets a con***REMOVED***g Region value returning a Con***REMOVED***g pointer for
// chaining.
func (c *Con***REMOVED***g) WithRegion(region string) *Con***REMOVED***g {
	c.Region = &region
	return c
}

// WithDisableSSL sets a con***REMOVED***g DisableSSL value returning a Con***REMOVED***g pointer
// for chaining.
func (c *Con***REMOVED***g) WithDisableSSL(disable bool) *Con***REMOVED***g {
	c.DisableSSL = &disable
	return c
}

// WithHTTPClient sets a con***REMOVED***g HTTPClient value returning a Con***REMOVED***g pointer
// for chaining.
func (c *Con***REMOVED***g) WithHTTPClient(client *http.Client) *Con***REMOVED***g {
	c.HTTPClient = client
	return c
}

// WithMaxRetries sets a con***REMOVED***g MaxRetries value returning a Con***REMOVED***g pointer
// for chaining.
func (c *Con***REMOVED***g) WithMaxRetries(max int) *Con***REMOVED***g {
	c.MaxRetries = &max
	return c
}

// WithDisableParamValidation sets a con***REMOVED***g DisableParamValidation value
// returning a Con***REMOVED***g pointer for chaining.
func (c *Con***REMOVED***g) WithDisableParamValidation(disable bool) *Con***REMOVED***g {
	c.DisableParamValidation = &disable
	return c
}

// WithDisableComputeChecksums sets a con***REMOVED***g DisableComputeChecksums value
// returning a Con***REMOVED***g pointer for chaining.
func (c *Con***REMOVED***g) WithDisableComputeChecksums(disable bool) *Con***REMOVED***g {
	c.DisableComputeChecksums = &disable
	return c
}

// WithLogLevel sets a con***REMOVED***g LogLevel value returning a Con***REMOVED***g pointer for
// chaining.
func (c *Con***REMOVED***g) WithLogLevel(level LogLevelType) *Con***REMOVED***g {
	c.LogLevel = &level
	return c
}

// WithLogger sets a con***REMOVED***g Logger value returning a Con***REMOVED***g pointer for
// chaining.
func (c *Con***REMOVED***g) WithLogger(logger Logger) *Con***REMOVED***g {
	c.Logger = logger
	return c
}

// WithS3ForcePathStyle sets a con***REMOVED***g S3ForcePathStyle value returning a Con***REMOVED***g
// pointer for chaining.
func (c *Con***REMOVED***g) WithS3ForcePathStyle(force bool) *Con***REMOVED***g {
	c.S3ForcePathStyle = &force
	return c
}

// WithS3Disable100Continue sets a con***REMOVED***g S3Disable100Continue value returning
// a Con***REMOVED***g pointer for chaining.
func (c *Con***REMOVED***g) WithS3Disable100Continue(disable bool) *Con***REMOVED***g {
	c.S3Disable100Continue = &disable
	return c
}

// WithS3UseAccelerate sets a con***REMOVED***g S3UseAccelerate value returning a Con***REMOVED***g
// pointer for chaining.
func (c *Con***REMOVED***g) WithS3UseAccelerate(enable bool) *Con***REMOVED***g {
	c.S3UseAccelerate = &enable
	return c

}

// WithS3DisableContentMD5Validation sets a con***REMOVED***g
// S3DisableContentMD5Validation value returning a Con***REMOVED***g pointer for chaining.
func (c *Con***REMOVED***g) WithS3DisableContentMD5Validation(enable bool) *Con***REMOVED***g {
	c.S3DisableContentMD5Validation = &enable
	return c

}

// WithUseDualStack sets a con***REMOVED***g UseDualStack value returning a Con***REMOVED***g
// pointer for chaining.
func (c *Con***REMOVED***g) WithUseDualStack(enable bool) *Con***REMOVED***g {
	c.UseDualStack = &enable
	return c
}

// WithEC2MetadataDisableTimeoutOverride sets a con***REMOVED***g EC2MetadataDisableTimeoutOverride value
// returning a Con***REMOVED***g pointer for chaining.
func (c *Con***REMOVED***g) WithEC2MetadataDisableTimeoutOverride(enable bool) *Con***REMOVED***g {
	c.EC2MetadataDisableTimeoutOverride = &enable
	return c
}

// WithSleepDelay overrides the function used to sleep while waiting for the
// next retry. Defaults to time.Sleep.
func (c *Con***REMOVED***g) WithSleepDelay(fn func(time.Duration)) *Con***REMOVED***g {
	c.SleepDelay = fn
	return c
}

// MergeIn merges the passed in con***REMOVED***gs into the existing con***REMOVED***g object.
func (c *Con***REMOVED***g) MergeIn(cfgs ...*Con***REMOVED***g) {
	for _, other := range cfgs {
		mergeInCon***REMOVED***g(c, other)
	}
}

func mergeInCon***REMOVED***g(dst *Con***REMOVED***g, other *Con***REMOVED***g) {
	if other == nil {
		return
	}

	if other.CredentialsChainVerboseErrors != nil {
		dst.CredentialsChainVerboseErrors = other.CredentialsChainVerboseErrors
	}

	if other.Credentials != nil {
		dst.Credentials = other.Credentials
	}

	if other.Endpoint != nil {
		dst.Endpoint = other.Endpoint
	}

	if other.EndpointResolver != nil {
		dst.EndpointResolver = other.EndpointResolver
	}

	if other.Region != nil {
		dst.Region = other.Region
	}

	if other.DisableSSL != nil {
		dst.DisableSSL = other.DisableSSL
	}

	if other.HTTPClient != nil {
		dst.HTTPClient = other.HTTPClient
	}

	if other.LogLevel != nil {
		dst.LogLevel = other.LogLevel
	}

	if other.Logger != nil {
		dst.Logger = other.Logger
	}

	if other.MaxRetries != nil {
		dst.MaxRetries = other.MaxRetries
	}

	if other.Retryer != nil {
		dst.Retryer = other.Retryer
	}

	if other.DisableParamValidation != nil {
		dst.DisableParamValidation = other.DisableParamValidation
	}

	if other.DisableComputeChecksums != nil {
		dst.DisableComputeChecksums = other.DisableComputeChecksums
	}

	if other.S3ForcePathStyle != nil {
		dst.S3ForcePathStyle = other.S3ForcePathStyle
	}

	if other.S3Disable100Continue != nil {
		dst.S3Disable100Continue = other.S3Disable100Continue
	}

	if other.S3UseAccelerate != nil {
		dst.S3UseAccelerate = other.S3UseAccelerate
	}

	if other.S3DisableContentMD5Validation != nil {
		dst.S3DisableContentMD5Validation = other.S3DisableContentMD5Validation
	}

	if other.UseDualStack != nil {
		dst.UseDualStack = other.UseDualStack
	}

	if other.EC2MetadataDisableTimeoutOverride != nil {
		dst.EC2MetadataDisableTimeoutOverride = other.EC2MetadataDisableTimeoutOverride
	}

	if other.SleepDelay != nil {
		dst.SleepDelay = other.SleepDelay
	}

	if other.DisableRestProtocolURICleaning != nil {
		dst.DisableRestProtocolURICleaning = other.DisableRestProtocolURICleaning
	}

	if other.EnforceShouldRetryCheck != nil {
		dst.EnforceShouldRetryCheck = other.EnforceShouldRetryCheck
	}
}

// Copy will return a shallow copy of the Con***REMOVED***g object. If any additional
// con***REMOVED***gurations are provided they will be merged into the new con***REMOVED***g returned.
func (c *Con***REMOVED***g) Copy(cfgs ...*Con***REMOVED***g) *Con***REMOVED***g {
	dst := &Con***REMOVED***g{}
	dst.MergeIn(c)

	for _, cfg := range cfgs {
		dst.MergeIn(cfg)
	}

	return dst
}
