package session

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/corehandlers"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
)

// A Session provides a central location to create service clients from and
// store con***REMOVED***gurations and request handlers for those services.
//
// Sessions are safe to create service clients concurrently, but it is not safe
// to mutate the Session concurrently.
//
// The Session satis***REMOVED***es the service client's client.ClientCon***REMOVED***gProvider.
type Session struct {
	Con***REMOVED***g   *aws.Con***REMOVED***g
	Handlers request.Handlers
}

// New creates a new instance of the handlers merging in the provided con***REMOVED***gs
// on top of the SDK's default con***REMOVED***gurations. Once the Session is created it
// can be mutated to modify the Con***REMOVED***g or Handlers. The Session is safe to be
// read concurrently, but it should not be written to concurrently.
//
// If the AWS_SDK_LOAD_CONFIG environment is set to a truthy value, the New
// method could now encounter an error when loading the con***REMOVED***guration. When
// The environment variable is set, and an error occurs, New will return a
// session that will fail all requests reporting the error that occurred while
// loading the session. Use NewSession to get the error when creating the
// session.
//
// If the AWS_SDK_LOAD_CONFIG environment variable is set to a truthy value
// the shared con***REMOVED***g ***REMOVED***le (~/.aws/con***REMOVED***g) will also be loaded, in addition to
// the shared credentials ***REMOVED***le (~/.aws/credentials). Values set in both the
// shared con***REMOVED***g, and shared credentials will be taken from the shared
// credentials ***REMOVED***le.
//
// Deprecated: Use NewSession functions to create sessions instead. NewSession
// has the same functionality as New except an error can be returned when the
// func is called instead of waiting to receive an error until a request is made.
func New(cfgs ...*aws.Con***REMOVED***g) *Session {
	// load initial con***REMOVED***g from environment
	envCfg := loadEnvCon***REMOVED***g()

	if envCfg.EnableSharedCon***REMOVED***g {
		s, err := newSession(Options{}, envCfg, cfgs...)
		if err != nil {
			// Old session.New expected all errors to be discovered when
			// a request is made, and would report the errors then. This
			// needs to be replicated if an error occurs while creating
			// the session.
			msg := "failed to create session with AWS_SDK_LOAD_CONFIG enabled. " +
				"Use session.NewSession to handle errors occurring during session creation."

			// Session creation failed, need to report the error and prevent
			// any requests from succeeding.
			s = &Session{Con***REMOVED***g: defaults.Con***REMOVED***g()}
			s.Con***REMOVED***g.MergeIn(cfgs...)
			s.Con***REMOVED***g.Logger.Log("ERROR:", msg, "Error:", err)
			s.Handlers.Validate.PushBack(func(r *request.Request) {
				r.Error = err
			})
		}
		return s
	}

	return deprecatedNewSession(cfgs...)
}

// NewSession returns a new Session created from SDK defaults, con***REMOVED***g ***REMOVED***les,
// environment, and user provided con***REMOVED***g ***REMOVED***les. Once the Session is created
// it can be mutated to modify the Con***REMOVED***g or Handlers. The Session is safe to
// be read concurrently, but it should not be written to concurrently.
//
// If the AWS_SDK_LOAD_CONFIG environment variable is set to a truthy value
// the shared con***REMOVED***g ***REMOVED***le (~/.aws/con***REMOVED***g) will also be loaded in addition to
// the shared credentials ***REMOVED***le (~/.aws/credentials). Values set in both the
// shared con***REMOVED***g, and shared credentials will be taken from the shared
// credentials ***REMOVED***le. Enabling the Shared Con***REMOVED***g will also allow the Session
// to be built with retrieving credentials with AssumeRole set in the con***REMOVED***g.
//
// See the NewSessionWithOptions func for information on how to override or
// control through code how the Session will be created. Such as specifying the
// con***REMOVED***g pro***REMOVED***le, and controlling if shared con***REMOVED***g is enabled or not.
func NewSession(cfgs ...*aws.Con***REMOVED***g) (*Session, error) {
	opts := Options{}
	opts.Con***REMOVED***g.MergeIn(cfgs...)

	return NewSessionWithOptions(opts)
}

// SharedCon***REMOVED***gState provides the ability to optionally override the state
// of the session's creation based on the shared con***REMOVED***g being enabled or
// disabled.
type SharedCon***REMOVED***gState int

const (
	// SharedCon***REMOVED***gStateFromEnv does not override any state of the
	// AWS_SDK_LOAD_CONFIG env var. It is the default value of the
	// SharedCon***REMOVED***gState type.
	SharedCon***REMOVED***gStateFromEnv SharedCon***REMOVED***gState = iota

	// SharedCon***REMOVED***gDisable overrides the AWS_SDK_LOAD_CONFIG env var value
	// and disables the shared con***REMOVED***g functionality.
	SharedCon***REMOVED***gDisable

	// SharedCon***REMOVED***gEnable overrides the AWS_SDK_LOAD_CONFIG env var value
	// and enables the shared con***REMOVED***g functionality.
	SharedCon***REMOVED***gEnable
)

// Options provides the means to control how a Session is created and what
// con***REMOVED***guration values will be loaded.
//
type Options struct {
	// Provides con***REMOVED***g values for the SDK to use when creating service clients
	// and making API requests to services. Any value set in with this ***REMOVED***eld
	// will override the associated value provided by the SDK defaults,
	// environment or con***REMOVED***g ***REMOVED***les where relevant.
	//
	// If not set, con***REMOVED***guration values from from SDK defaults, environment,
	// con***REMOVED***g will be used.
	Con***REMOVED***g aws.Con***REMOVED***g

	// Overrides the con***REMOVED***g pro***REMOVED***le the Session should be created from. If not
	// set the value of the environment variable will be loaded (AWS_PROFILE,
	// or AWS_DEFAULT_PROFILE if the Shared Con***REMOVED***g is enabled).
	//
	// If not set and environment variables are not set the "default"
	// (DefaultSharedCon***REMOVED***gPro***REMOVED***le) will be used as the pro***REMOVED***le to load the
	// session con***REMOVED***g from.
	Pro***REMOVED***le string

	// Instructs how the Session will be created based on the AWS_SDK_LOAD_CONFIG
	// environment variable. By default a Session will be created using the
	// value provided by the AWS_SDK_LOAD_CONFIG environment variable.
	//
	// Setting this value to SharedCon***REMOVED***gEnable or SharedCon***REMOVED***gDisable
	// will allow you to override the AWS_SDK_LOAD_CONFIG environment variable
	// and enable or disable the shared con***REMOVED***g functionality.
	SharedCon***REMOVED***gState SharedCon***REMOVED***gState

	// Ordered list of ***REMOVED***les the session will load con***REMOVED***guration from.
	// It will override environment variable AWS_SHARED_CREDENTIALS_FILE, AWS_CONFIG_FILE.
	SharedCon***REMOVED***gFiles []string

	// When the SDK's shared con***REMOVED***g is con***REMOVED***gured to assume a role with MFA
	// this option is required in order to provide the mechanism that will
	// retrieve the MFA token. There is no default value for this ***REMOVED***eld. If
	// it is not set an error will be returned when creating the session.
	//
	// This token provider will be called when ever the assumed role's
	// credentials need to be refreshed. Within the context of service clients
	// all sharing the same session the SDK will ensure calls to the token
	// provider are atomic. When sharing a token provider across multiple
	// sessions additional synchronization logic is needed to ensure the
	// token providers do not introduce race conditions. It is recommend to
	// share the session where possible.
	//
	// stscreds.StdinTokenProvider is a basic implementation that will prompt
	// from stdin for the MFA token code.
	//
	// This ***REMOVED***eld is only used if the shared con***REMOVED***guration is enabled, and
	// the con***REMOVED***g enables assume role wit MFA via the mfa_serial ***REMOVED***eld.
	AssumeRoleTokenProvider func() (string, error)

	// Reader for a custom Credentials Authority (CA) bundle in PEM format that
	// the SDK will use instead of the default system's root CA bundle. Use this
	// only if you want to replace the CA bundle the SDK uses for TLS requests.
	//
	// Enabling this option will attempt to merge the Transport into the SDK's HTTP
	// client. If the client's Transport is not a http.Transport an error will be
	// returned. If the Transport's TLS con***REMOVED***g is set this option will cause the SDK
	// to overwrite the Transport's TLS con***REMOVED***g's  RootCAs value. If the CA
	// bundle reader contains multiple certi***REMOVED***cates all of them will be loaded.
	//
	// The Session option CustomCABundle is also available when creating sessions
	// to also enable this feature. CustomCABundle session option ***REMOVED***eld has priority
	// over the AWS_CA_BUNDLE environment variable, and will be used if both are set.
	CustomCABundle io.Reader
}

// NewSessionWithOptions returns a new Session created from SDK defaults, con***REMOVED***g ***REMOVED***les,
// environment, and user provided con***REMOVED***g ***REMOVED***les. This func uses the Options
// values to con***REMOVED***gure how the Session is created.
//
// If the AWS_SDK_LOAD_CONFIG environment variable is set to a truthy value
// the shared con***REMOVED***g ***REMOVED***le (~/.aws/con***REMOVED***g) will also be loaded in addition to
// the shared credentials ***REMOVED***le (~/.aws/credentials). Values set in both the
// shared con***REMOVED***g, and shared credentials will be taken from the shared
// credentials ***REMOVED***le. Enabling the Shared Con***REMOVED***g will also allow the Session
// to be built with retrieving credentials with AssumeRole set in the con***REMOVED***g.
//
//     // Equivalent to session.New
//     sess := session.Must(session.NewSessionWithOptions(session.Options{}))
//
//     // Specify pro***REMOVED***le to load for the session's con***REMOVED***g
//     sess := session.Must(session.NewSessionWithOptions(session.Options{
//          Pro***REMOVED***le: "pro***REMOVED***le_name",
//     }))
//
//     // Specify pro***REMOVED***le for con***REMOVED***g and region for requests
//     sess := session.Must(session.NewSessionWithOptions(session.Options{
//          Con***REMOVED***g: aws.Con***REMOVED***g{Region: aws.String("us-east-1")},
//          Pro***REMOVED***le: "pro***REMOVED***le_name",
//     }))
//
//     // Force enable Shared Con***REMOVED***g support
//     sess := session.Must(session.NewSessionWithOptions(session.Options{
//         SharedCon***REMOVED***gState: session.SharedCon***REMOVED***gEnable,
//     }))
func NewSessionWithOptions(opts Options) (*Session, error) {
	var envCfg envCon***REMOVED***g
	if opts.SharedCon***REMOVED***gState == SharedCon***REMOVED***gEnable {
		envCfg = loadSharedEnvCon***REMOVED***g()
	} ***REMOVED*** {
		envCfg = loadEnvCon***REMOVED***g()
	}

	if len(opts.Pro***REMOVED***le) > 0 {
		envCfg.Pro***REMOVED***le = opts.Pro***REMOVED***le
	}

	switch opts.SharedCon***REMOVED***gState {
	case SharedCon***REMOVED***gDisable:
		envCfg.EnableSharedCon***REMOVED***g = false
	case SharedCon***REMOVED***gEnable:
		envCfg.EnableSharedCon***REMOVED***g = true
	}

	if len(envCfg.SharedCredentialsFile) == 0 {
		envCfg.SharedCredentialsFile = defaults.SharedCredentialsFilename()
	}
	if len(envCfg.SharedCon***REMOVED***gFile) == 0 {
		envCfg.SharedCon***REMOVED***gFile = defaults.SharedCon***REMOVED***gFilename()
	}

	// Only use AWS_CA_BUNDLE if session option is not provided.
	if len(envCfg.CustomCABundle) != 0 && opts.CustomCABundle == nil {
		f, err := os.Open(envCfg.CustomCABundle)
		if err != nil {
			return nil, awserr.New("LoadCustomCABundleError",
				"failed to open custom CA bundle PEM ***REMOVED***le", err)
		}
		defer f.Close()
		opts.CustomCABundle = f
	}

	return newSession(opts, envCfg, &opts.Con***REMOVED***g)
}

// Must is a helper function to ensure the Session is valid and there was no
// error when calling a NewSession function.
//
// This helper is intended to be used in variable initialization to load the
// Session and con***REMOVED***guration at startup. Such as:
//
//     var sess = session.Must(session.NewSession())
func Must(sess *Session, err error) *Session {
	if err != nil {
		panic(err)
	}

	return sess
}

func deprecatedNewSession(cfgs ...*aws.Con***REMOVED***g) *Session {
	cfg := defaults.Con***REMOVED***g()
	handlers := defaults.Handlers()

	// Apply the passed in con***REMOVED***gs so the con***REMOVED***guration can be applied to the
	// default credential chain
	cfg.MergeIn(cfgs...)
	if cfg.EndpointResolver == nil {
		// An endpoint resolver is required for a session to be able to provide
		// endpoints for service client con***REMOVED***gurations.
		cfg.EndpointResolver = endpoints.DefaultResolver()
	}
	cfg.Credentials = defaults.CredChain(cfg, handlers)

	// Reapply any passed in con***REMOVED***gs to override credentials if set
	cfg.MergeIn(cfgs...)

	s := &Session{
		Con***REMOVED***g:   cfg,
		Handlers: handlers,
	}

	initHandlers(s)

	return s
}

func newSession(opts Options, envCfg envCon***REMOVED***g, cfgs ...*aws.Con***REMOVED***g) (*Session, error) {
	cfg := defaults.Con***REMOVED***g()
	handlers := defaults.Handlers()

	// Get a merged version of the user provided con***REMOVED***g to determine if
	// credentials were.
	userCfg := &aws.Con***REMOVED***g{}
	userCfg.MergeIn(cfgs...)

	// Ordered con***REMOVED***g ***REMOVED***les will be loaded in with later ***REMOVED***les overwriting
	// previous con***REMOVED***g ***REMOVED***le values.
	var cfgFiles []string
	if opts.SharedCon***REMOVED***gFiles != nil {
		cfgFiles = opts.SharedCon***REMOVED***gFiles
	} ***REMOVED*** {
		cfgFiles = []string{envCfg.SharedCon***REMOVED***gFile, envCfg.SharedCredentialsFile}
		if !envCfg.EnableSharedCon***REMOVED***g {
			// The shared con***REMOVED***g ***REMOVED***le (~/.aws/con***REMOVED***g) is only loaded if instructed
			// to load via the envCon***REMOVED***g.EnableSharedCon***REMOVED***g (AWS_SDK_LOAD_CONFIG).
			cfgFiles = cfgFiles[1:]
		}
	}

	// Load additional con***REMOVED***g from ***REMOVED***le(s)
	sharedCfg, err := loadSharedCon***REMOVED***g(envCfg.Pro***REMOVED***le, cfgFiles)
	if err != nil {
		return nil, err
	}

	if err := mergeCon***REMOVED***gSrcs(cfg, userCfg, envCfg, sharedCfg, handlers, opts); err != nil {
		return nil, err
	}

	s := &Session{
		Con***REMOVED***g:   cfg,
		Handlers: handlers,
	}

	initHandlers(s)

	// Setup HTTP client with custom cert bundle if enabled
	if opts.CustomCABundle != nil {
		if err := loadCustomCABundle(s, opts.CustomCABundle); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func loadCustomCABundle(s *Session, bundle io.Reader) error {
	var t *http.Transport
	switch v := s.Con***REMOVED***g.HTTPClient.Transport.(type) {
	case *http.Transport:
		t = v
	default:
		if s.Con***REMOVED***g.HTTPClient.Transport != nil {
			return awserr.New("LoadCustomCABundleError",
				"unable to load custom CA bundle, HTTPClient's transport unsupported type", nil)
		}
	}
	if t == nil {
		t = &http.Transport{}
	}

	p, err := loadCertPool(bundle)
	if err != nil {
		return err
	}
	if t.TLSClientCon***REMOVED***g == nil {
		t.TLSClientCon***REMOVED***g = &tls.Con***REMOVED***g{}
	}
	t.TLSClientCon***REMOVED***g.RootCAs = p

	s.Con***REMOVED***g.HTTPClient.Transport = t

	return nil
}

func loadCertPool(r io.Reader) (*x509.CertPool, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, awserr.New("LoadCustomCABundleError",
			"failed to read custom CA bundle PEM ***REMOVED***le", err)
	}

	p := x509.NewCertPool()
	if !p.AppendCertsFromPEM(b) {
		return nil, awserr.New("LoadCustomCABundleError",
			"failed to load custom CA bundle PEM ***REMOVED***le", err)
	}

	return p, nil
}

func mergeCon***REMOVED***gSrcs(cfg, userCfg *aws.Con***REMOVED***g, envCfg envCon***REMOVED***g, sharedCfg sharedCon***REMOVED***g, handlers request.Handlers, sessOpts Options) error {
	// Merge in user provided con***REMOVED***guration
	cfg.MergeIn(userCfg)

	// Region if not already set by user
	if len(aws.StringValue(cfg.Region)) == 0 {
		if len(envCfg.Region) > 0 {
			cfg.WithRegion(envCfg.Region)
		} ***REMOVED*** if envCfg.EnableSharedCon***REMOVED***g && len(sharedCfg.Region) > 0 {
			cfg.WithRegion(sharedCfg.Region)
		}
	}

	// Con***REMOVED***gure credentials if not already set
	if cfg.Credentials == credentials.AnonymousCredentials && userCfg.Credentials == nil {
		if len(envCfg.Creds.AccessKeyID) > 0 {
			cfg.Credentials = credentials.NewStaticCredentialsFromCreds(
				envCfg.Creds,
			)
		} ***REMOVED*** if envCfg.EnableSharedCon***REMOVED***g && len(sharedCfg.AssumeRole.RoleARN) > 0 && sharedCfg.AssumeRoleSource != nil {
			cfgCp := *cfg
			cfgCp.Credentials = credentials.NewStaticCredentialsFromCreds(
				sharedCfg.AssumeRoleSource.Creds,
			)
			if len(sharedCfg.AssumeRole.MFASerial) > 0 && sessOpts.AssumeRoleTokenProvider == nil {
				// AssumeRole Token provider is required if doing Assume Role
				// with MFA.
				return AssumeRoleTokenProviderNotSetError{}
			}
			cfg.Credentials = stscreds.NewCredentials(
				&Session{
					Con***REMOVED***g:   &cfgCp,
					Handlers: handlers.Copy(),
				},
				sharedCfg.AssumeRole.RoleARN,
				func(opt *stscreds.AssumeRoleProvider) {
					opt.RoleSessionName = sharedCfg.AssumeRole.RoleSessionName

					// Assume role with external ID
					if len(sharedCfg.AssumeRole.ExternalID) > 0 {
						opt.ExternalID = aws.String(sharedCfg.AssumeRole.ExternalID)
					}

					// Assume role with MFA
					if len(sharedCfg.AssumeRole.MFASerial) > 0 {
						opt.SerialNumber = aws.String(sharedCfg.AssumeRole.MFASerial)
						opt.TokenProvider = sessOpts.AssumeRoleTokenProvider
					}
				},
			)
		} ***REMOVED*** if len(sharedCfg.Creds.AccessKeyID) > 0 {
			cfg.Credentials = credentials.NewStaticCredentialsFromCreds(
				sharedCfg.Creds,
			)
		} ***REMOVED*** {
			// Fallback to default credentials provider, include mock errors
			// for the credential chain so user can identify why credentials
			// failed to be retrieved.
			cfg.Credentials = credentials.NewCredentials(&credentials.ChainProvider{
				VerboseErrors: aws.BoolValue(cfg.CredentialsChainVerboseErrors),
				Providers: []credentials.Provider{
					&credProviderError{Err: awserr.New("EnvAccessKeyNotFound", "failed to ***REMOVED***nd credentials in the environment.", nil)},
					&credProviderError{Err: awserr.New("SharedCredsLoad", fmt.Sprintf("failed to load pro***REMOVED***le, %s.", envCfg.Pro***REMOVED***le), nil)},
					defaults.RemoteCredProvider(*cfg, handlers),
				},
			})
		}
	}

	return nil
}

// AssumeRoleTokenProviderNotSetError is an error returned when creating a session when the
// MFAToken option is not set when shared con***REMOVED***g is con***REMOVED***gured load assume a
// role with an MFA token.
type AssumeRoleTokenProviderNotSetError struct{}

// Code is the short id of the error.
func (e AssumeRoleTokenProviderNotSetError) Code() string {
	return "AssumeRoleTokenProviderNotSetError"
}

// Message is the description of the error
func (e AssumeRoleTokenProviderNotSetError) Message() string {
	return fmt.Sprintf("assume role with MFA enabled, but AssumeRoleTokenProvider session option not set.")
}

// OrigErr is the underlying error that caused the failure.
func (e AssumeRoleTokenProviderNotSetError) OrigErr() error {
	return nil
}

// Error satis***REMOVED***es the error interface.
func (e AssumeRoleTokenProviderNotSetError) Error() string {
	return awserr.SprintError(e.Code(), e.Message(), "", nil)
}

type credProviderError struct {
	Err error
}

var emptyCreds = credentials.Value{}

func (c credProviderError) Retrieve() (credentials.Value, error) {
	return credentials.Value{}, c.Err
}
func (c credProviderError) IsExpired() bool {
	return true
}

func initHandlers(s *Session) {
	// Add the Validate parameter handler if it is not disabled.
	s.Handlers.Validate.Remove(corehandlers.ValidateParametersHandler)
	if !aws.BoolValue(s.Con***REMOVED***g.DisableParamValidation) {
		s.Handlers.Validate.PushBackNamed(corehandlers.ValidateParametersHandler)
	}
}

// Copy creates and returns a copy of the current Session, coping the con***REMOVED***g
// and handlers. If any additional con***REMOVED***gs are provided they will be merged
// on top of the Session's copied con***REMOVED***g.
//
//     // Create a copy of the current Session, con***REMOVED***gured for the us-west-2 region.
//     sess.Copy(&aws.Con***REMOVED***g{Region: aws.String("us-west-2")})
func (s *Session) Copy(cfgs ...*aws.Con***REMOVED***g) *Session {
	newSession := &Session{
		Con***REMOVED***g:   s.Con***REMOVED***g.Copy(cfgs...),
		Handlers: s.Handlers.Copy(),
	}

	initHandlers(newSession)

	return newSession
}

// ClientCon***REMOVED***g satis***REMOVED***es the client.Con***REMOVED***gProvider interface and is used to
// con***REMOVED***gure the service client instances. Passing the Session to the service
// client's constructor (New) will use this method to con***REMOVED***gure the client.
func (s *Session) ClientCon***REMOVED***g(serviceName string, cfgs ...*aws.Con***REMOVED***g) client.Con***REMOVED***g {
	// Backwards compatibility, the error will be eaten if user calls ClientCon***REMOVED***g
	// directly. All SDK services will use Clientcon***REMOVED***gWithError.
	cfg, _ := s.clientCon***REMOVED***gWithErr(serviceName, cfgs...)

	return cfg
}

func (s *Session) clientCon***REMOVED***gWithErr(serviceName string, cfgs ...*aws.Con***REMOVED***g) (client.Con***REMOVED***g, error) {
	s = s.Copy(cfgs...)

	var resolved endpoints.ResolvedEndpoint
	var err error

	region := aws.StringValue(s.Con***REMOVED***g.Region)

	if endpoint := aws.StringValue(s.Con***REMOVED***g.Endpoint); len(endpoint) != 0 {
		resolved.URL = endpoints.AddScheme(endpoint, aws.BoolValue(s.Con***REMOVED***g.DisableSSL))
		resolved.SigningRegion = region
	} ***REMOVED*** {
		resolved, err = s.Con***REMOVED***g.EndpointResolver.EndpointFor(
			serviceName, region,
			func(opt *endpoints.Options) {
				opt.DisableSSL = aws.BoolValue(s.Con***REMOVED***g.DisableSSL)
				opt.UseDualStack = aws.BoolValue(s.Con***REMOVED***g.UseDualStack)

				// Support the condition where the service is modeled but its
				// endpoint metadata is not available.
				opt.ResolveUnknownService = true
			},
		)
	}

	return client.Con***REMOVED***g{
		Con***REMOVED***g:        s.Con***REMOVED***g,
		Handlers:      s.Handlers,
		Endpoint:      resolved.URL,
		SigningRegion: resolved.SigningRegion,
		SigningName:   resolved.SigningName,
	}, err
}

// ClientCon***REMOVED***gNoResolveEndpoint is the same as ClientCon***REMOVED***g with the exception
// that the EndpointResolver will not be used to resolve the endpoint. The only
// endpoint set must come from the aws.Con***REMOVED***g.Endpoint ***REMOVED***eld.
func (s *Session) ClientCon***REMOVED***gNoResolveEndpoint(cfgs ...*aws.Con***REMOVED***g) client.Con***REMOVED***g {
	s = s.Copy(cfgs...)

	var resolved endpoints.ResolvedEndpoint

	region := aws.StringValue(s.Con***REMOVED***g.Region)

	if ep := aws.StringValue(s.Con***REMOVED***g.Endpoint); len(ep) > 0 {
		resolved.URL = endpoints.AddScheme(ep, aws.BoolValue(s.Con***REMOVED***g.DisableSSL))
		resolved.SigningRegion = region
	}

	return client.Con***REMOVED***g{
		Con***REMOVED***g:        s.Con***REMOVED***g,
		Handlers:      s.Handlers,
		Endpoint:      resolved.URL,
		SigningRegion: resolved.SigningRegion,
		SigningName:   resolved.SigningName,
	}
}
