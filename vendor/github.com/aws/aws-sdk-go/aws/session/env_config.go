package session

import (
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
)

// EnvProviderName provides a name of the provider when con***REMOVED***g is loaded from environment.
const EnvProviderName = "EnvCon***REMOVED***gCredentials"

// envCon***REMOVED***g is a collection of environment values the SDK will read
// setup con***REMOVED***g from. All environment values are optional. But some values
// such as credentials require multiple values to be complete or the values
// will be ignored.
type envCon***REMOVED***g struct {
	// Environment con***REMOVED***guration values. If set both Access Key ID and Secret Access
	// Key must be provided. Session Token and optionally also be provided, but is
	// not required.
	//
	//	# Access Key ID
	//	AWS_ACCESS_KEY_ID=AKID
	//	AWS_ACCESS_KEY=AKID # only read if AWS_ACCESS_KEY_ID is not set.
	//
	//	# Secret Access Key
	//	AWS_SECRET_ACCESS_KEY=SECRET
	//	AWS_SECRET_KEY=SECRET=SECRET # only read if AWS_SECRET_ACCESS_KEY is not set.
	//
	//	# Session Token
	//	AWS_SESSION_TOKEN=TOKEN
	Creds credentials.Value

	// Region value will instruct the SDK where to make service API requests to. If is
	// not provided in the environment the region must be provided before a service
	// client request is made.
	//
	//	AWS_REGION=us-east-1
	//
	//	# AWS_DEFAULT_REGION is only read if AWS_SDK_LOAD_CONFIG is also set,
	//	# and AWS_REGION is not also set.
	//	AWS_DEFAULT_REGION=us-east-1
	Region string

	// Pro***REMOVED***le name the SDK should load use when loading shared con***REMOVED***guration from the
	// shared con***REMOVED***guration ***REMOVED***les. If not provided "default" will be used as the
	// pro***REMOVED***le name.
	//
	//	AWS_PROFILE=my_pro***REMOVED***le
	//
	//	# AWS_DEFAULT_PROFILE is only read if AWS_SDK_LOAD_CONFIG is also set,
	//	# and AWS_PROFILE is not also set.
	//	AWS_DEFAULT_PROFILE=my_pro***REMOVED***le
	Pro***REMOVED***le string

	// SDK load con***REMOVED***g instructs the SDK to load the shared con***REMOVED***g in addition to
	// shared credentials. This also expands the con***REMOVED***guration loaded from the shared
	// credentials to have parity with the shared con***REMOVED***g ***REMOVED***le. This also enables
	// Region and Pro***REMOVED***le support for the AWS_DEFAULT_REGION and AWS_DEFAULT_PROFILE
	// env values as well.
	//
	//	AWS_SDK_LOAD_CONFIG=1
	EnableSharedCon***REMOVED***g bool

	// Shared credentials ***REMOVED***le path can be set to instruct the SDK to use an alternate
	// ***REMOVED***le for the shared credentials. If not set the ***REMOVED***le will be loaded from
	// $HOME/.aws/credentials on Linux/Unix based systems, and
	// %USERPROFILE%\.aws\credentials on Windows.
	//
	//	AWS_SHARED_CREDENTIALS_FILE=$HOME/my_shared_credentials
	SharedCredentialsFile string

	// Shared con***REMOVED***g ***REMOVED***le path can be set to instruct the SDK to use an alternate
	// ***REMOVED***le for the shared con***REMOVED***g. If not set the ***REMOVED***le will be loaded from
	// $HOME/.aws/con***REMOVED***g on Linux/Unix based systems, and
	// %USERPROFILE%\.aws\con***REMOVED***g on Windows.
	//
	//	AWS_CONFIG_FILE=$HOME/my_shared_con***REMOVED***g
	SharedCon***REMOVED***gFile string

	// Sets the path to a custom Credentials Authroity (CA) Bundle PEM ***REMOVED***le
	// that the SDK will use instead of the system's root CA bundle.
	// Only use this if you want to con***REMOVED***gure the SDK to use a custom set
	// of CAs.
	//
	// Enabling this option will attempt to merge the Transport
	// into the SDK's HTTP client. If the client's Transport is
	// not a http.Transport an error will be returned. If the
	// Transport's TLS con***REMOVED***g is set this option will cause the
	// SDK to overwrite the Transport's TLS con***REMOVED***g's  RootCAs value.
	//
	// Setting a custom HTTPClient in the aws.Con***REMOVED***g options will override this setting.
	// To use this option and custom HTTP client, the HTTP client needs to be provided
	// when creating the session. Not the service client.
	//
	//  AWS_CA_BUNDLE=$HOME/my_custom_ca_bundle
	CustomCABundle string
}

var (
	credAccessEnvKey = []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_ACCESS_KEY",
	}
	credSecretEnvKey = []string{
		"AWS_SECRET_ACCESS_KEY",
		"AWS_SECRET_KEY",
	}
	credSessionEnvKey = []string{
		"AWS_SESSION_TOKEN",
	}

	regionEnvKeys = []string{
		"AWS_REGION",
		"AWS_DEFAULT_REGION", // Only read if AWS_SDK_LOAD_CONFIG is also set
	}
	pro***REMOVED***leEnvKeys = []string{
		"AWS_PROFILE",
		"AWS_DEFAULT_PROFILE", // Only read if AWS_SDK_LOAD_CONFIG is also set
	}
	sharedCredsFileEnvKey = []string{
		"AWS_SHARED_CREDENTIALS_FILE",
	}
	sharedCon***REMOVED***gFileEnvKey = []string{
		"AWS_CONFIG_FILE",
	}
)

// loadEnvCon***REMOVED***g retrieves the SDK's environment con***REMOVED***guration.
// See `envCon***REMOVED***g` for the values that will be retrieved.
//
// If the environment variable `AWS_SDK_LOAD_CONFIG` is set to a truthy value
// the shared SDK con***REMOVED***g will be loaded in addition to the SDK's speci***REMOVED***c
// con***REMOVED***guration values.
func loadEnvCon***REMOVED***g() envCon***REMOVED***g {
	enableSharedCon***REMOVED***g, _ := strconv.ParseBool(os.Getenv("AWS_SDK_LOAD_CONFIG"))
	return envCon***REMOVED***gLoad(enableSharedCon***REMOVED***g)
}

// loadEnvSharedCon***REMOVED***g retrieves the SDK's environment con***REMOVED***guration, and the
// SDK shared con***REMOVED***g. See `envCon***REMOVED***g` for the values that will be retrieved.
//
// Loads the shared con***REMOVED***guration in addition to the SDK's speci***REMOVED***c con***REMOVED***guration.
// This will load the same values as `loadEnvCon***REMOVED***g` if the `AWS_SDK_LOAD_CONFIG`
// environment variable is set.
func loadSharedEnvCon***REMOVED***g() envCon***REMOVED***g {
	return envCon***REMOVED***gLoad(true)
}

func envCon***REMOVED***gLoad(enableSharedCon***REMOVED***g bool) envCon***REMOVED***g {
	cfg := envCon***REMOVED***g{}

	cfg.EnableSharedCon***REMOVED***g = enableSharedCon***REMOVED***g

	setFromEnvVal(&cfg.Creds.AccessKeyID, credAccessEnvKey)
	setFromEnvVal(&cfg.Creds.SecretAccessKey, credSecretEnvKey)
	setFromEnvVal(&cfg.Creds.SessionToken, credSessionEnvKey)

	// Require logical grouping of credentials
	if len(cfg.Creds.AccessKeyID) == 0 || len(cfg.Creds.SecretAccessKey) == 0 {
		cfg.Creds = credentials.Value{}
	} ***REMOVED*** {
		cfg.Creds.ProviderName = EnvProviderName
	}

	regionKeys := regionEnvKeys
	pro***REMOVED***leKeys := pro***REMOVED***leEnvKeys
	if !cfg.EnableSharedCon***REMOVED***g {
		regionKeys = regionKeys[:1]
		pro***REMOVED***leKeys = pro***REMOVED***leKeys[:1]
	}

	setFromEnvVal(&cfg.Region, regionKeys)
	setFromEnvVal(&cfg.Pro***REMOVED***le, pro***REMOVED***leKeys)

	setFromEnvVal(&cfg.SharedCredentialsFile, sharedCredsFileEnvKey)
	setFromEnvVal(&cfg.SharedCon***REMOVED***gFile, sharedCon***REMOVED***gFileEnvKey)

	if len(cfg.SharedCredentialsFile) == 0 {
		cfg.SharedCredentialsFile = defaults.SharedCredentialsFilename()
	}
	if len(cfg.SharedCon***REMOVED***gFile) == 0 {
		cfg.SharedCon***REMOVED***gFile = defaults.SharedCon***REMOVED***gFilename()
	}

	cfg.CustomCABundle = os.Getenv("AWS_CA_BUNDLE")

	return cfg
}

func setFromEnvVal(dst *string, keys []string) {
	for _, k := range keys {
		if v := os.Getenv(k); len(v) > 0 {
			*dst = v
			break
		}
	}
}
