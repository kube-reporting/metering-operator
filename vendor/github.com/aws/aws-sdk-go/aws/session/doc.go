/*
Package session provides con***REMOVED***guration for the SDK's service clients. Sessions
can be shared across service clients that share the same base con***REMOVED***guration.

Sessions are safe to use concurrently as long as the Session is not being
modi***REMOVED***ed. Sessions should be cached when possible, because creating a new
Session will load all con***REMOVED***guration values from the environment, and con***REMOVED***g
***REMOVED***les each time the Session is created. Sharing the Session value across all of
your service clients will ensure the con***REMOVED***guration is loaded the fewest number
of times possible.

Sessions options from Shared Con***REMOVED***g

By default NewSession will only load credentials from the shared credentials
***REMOVED***le (~/.aws/credentials). If the AWS_SDK_LOAD_CONFIG environment variable is
set to a truthy value the Session will be created from the con***REMOVED***guration
values from the shared con***REMOVED***g (~/.aws/con***REMOVED***g) and shared credentials
(~/.aws/credentials) ***REMOVED***les. Using the NewSessionWithOptions with
SharedCon***REMOVED***gState set to SharedCon***REMOVED***gEnable will create the session as if the
AWS_SDK_LOAD_CONFIG environment variable was set.

Credential and con***REMOVED***g loading order

The Session will attempt to load con***REMOVED***guration and credentials from the
environment, con***REMOVED***guration ***REMOVED***les, and other credential sources. The order
con***REMOVED***guration is loaded in is:

  * Environment Variables
  * Shared Credentials ***REMOVED***le
  * Shared Con***REMOVED***guration ***REMOVED***le (if SharedCon***REMOVED***g is enabled)
  * EC2 Instance Metadata (credentials only)

The Environment variables for credentials will have precedence over shared
con***REMOVED***g even if SharedCon***REMOVED***g is enabled. To override this behavior, and use
shared con***REMOVED***g credentials instead specify the session.Options.Pro***REMOVED***le, (e.g.
when using credential_source=Environment to assume a role).

  sess, err := session.NewSessionWithOptions(session.Options{
	  Pro***REMOVED***le: "myPro***REMOVED***le",
  })

Creating Sessions

Creating a Session without additional options will load credentials region, and
pro***REMOVED***le loaded from the environment and shared con***REMOVED***g automatically. See,
"Environment Variables" section for information on environment variables used
by Session.

	// Create Session
	sess, err := session.NewSession()


When creating Sessions optional aws.Con***REMOVED***g values can be passed in that will
override the default, or loaded, con***REMOVED***g values the Session is being created
with. This allows you to provide additional, or case based, con***REMOVED***guration
as needed.

	// Create a Session with a custom region
	sess, err := session.NewSession(&aws.Con***REMOVED***g{
		Region: aws.String("us-west-2"),
	})

Use NewSessionWithOptions to provide additional con***REMOVED***guration driving how the
Session's con***REMOVED***guration will be loaded. Such as, specifying shared con***REMOVED***g
pro***REMOVED***le, or override the shared con***REMOVED***g state,  (AWS_SDK_LOAD_CONFIG).

	// Equivalent to session.NewSession()
	sess, err := session.NewSessionWithOptions(session.Options{
		// Options
	})

	sess, err := session.NewSessionWithOptions(session.Options{
		// Specify pro***REMOVED***le to load for the session's con***REMOVED***g
		Pro***REMOVED***le: "pro***REMOVED***le_name",

		// Provide SDK Con***REMOVED***g options, such as Region.
		Con***REMOVED***g: aws.Con***REMOVED***g{
			Region: aws.String("us-west-2"),
		},

		// Force enable Shared Con***REMOVED***g support
		SharedCon***REMOVED***gState: session.SharedCon***REMOVED***gEnable,
	})

Adding Handlers

You can add handlers to a session to decorate API operation, (e.g. adding HTTP
headers). All clients that use the Session receive a copy of the Session's
handlers. For example, the following request handler added to the Session logs
every requests made.

	// Create a session, and add additional handlers for all service
	// clients created with the Session to inherit. Adds logging handler.
	sess := session.Must(session.NewSession())

	sess.Handlers.Send.PushFront(func(r *request.Request) {
		// Log every request made and its payload
		logger.Printf("Request: %s/%s, Params: %s",
			r.ClientInfo.ServiceName, r.Operation, r.Params)
	})

Shared Con***REMOVED***g Fields

By default the SDK will only load the shared credentials ***REMOVED***le's
(~/.aws/credentials) credentials values, and all other con***REMOVED***g is provided by
the environment variables, SDK defaults, and user provided aws.Con***REMOVED***g values.

If the AWS_SDK_LOAD_CONFIG environment variable is set, or SharedCon***REMOVED***gEnable
option is used to create the Session the full shared con***REMOVED***g values will be
loaded. This includes credentials, region, and support for assume role. In
addition the Session will load its con***REMOVED***guration from both the shared con***REMOVED***g
***REMOVED***le (~/.aws/con***REMOVED***g) and shared credentials ***REMOVED***le (~/.aws/credentials). Both
***REMOVED***les have the same format.

If both con***REMOVED***g ***REMOVED***les are present the con***REMOVED***guration from both ***REMOVED***les will be
read. The Session will be created from con***REMOVED***guration values from the shared
credentials ***REMOVED***le (~/.aws/credentials) over those in the shared con***REMOVED***g ***REMOVED***le
(~/.aws/con***REMOVED***g).

Credentials are the values the SDK uses to authenticating requests with AWS
Services. When speci***REMOVED***ed in a ***REMOVED***le, both aws_access_key_id and
aws_secret_access_key must be provided together in the same ***REMOVED***le to be
considered valid. They will be ignored if both are not present.
aws_session_token is an optional ***REMOVED***eld that can be provided in addition to the
other two ***REMOVED***elds.

	aws_access_key_id = AKID
	aws_secret_access_key = SECRET
	aws_session_token = TOKEN

	; region only supported if SharedCon***REMOVED***gEnabled.
	region = us-east-1

Assume Role con***REMOVED***guration

The role_arn ***REMOVED***eld allows you to con***REMOVED***gure the SDK to assume an IAM role using
a set of credentials from another source. Such as when paired with static
credentials, "pro***REMOVED***le_source", "credential_process", or "credential_source"
***REMOVED***elds. If "role_arn" is provided, a source of credentials must also be
speci***REMOVED***ed, such as "source_pro***REMOVED***le", "credential_source", or
"credential_process".

	role_arn = arn:aws:iam::<account_number>:role/<role_name>
	source_pro***REMOVED***le = pro***REMOVED***le_with_creds
	external_id = 1234
	mfa_serial = <serial or mfa arn>
	role_session_name = session_name


The SDK supports assuming a role with MFA token. If "mfa_serial" is set, you
must also set the Session Option.AssumeRoleTokenProvider. The Session will fail
to load if the AssumeRoleTokenProvider is not speci***REMOVED***ed.

    sess := session.Must(session.NewSessionWithOptions(session.Options{
        AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
    }))

To setup Assume Role outside of a session see the stscreds.AssumeRoleProvider
documentation.

Environment Variables

When a Session is created several environment variables can be set to adjust
how the SDK functions, and what con***REMOVED***guration data it loads when creating
Sessions. All environment values are optional, but some values like credentials
require multiple of the values to set or the partial values will be ignored.
All environment variable values are strings unless otherwise noted.

Environment con***REMOVED***guration values. If set both Access Key ID and Secret Access
Key must be provided. Session Token and optionally also be provided, but is
not required.

	# Access Key ID
	AWS_ACCESS_KEY_ID=AKID
	AWS_ACCESS_KEY=AKID # only read if AWS_ACCESS_KEY_ID is not set.

	# Secret Access Key
	AWS_SECRET_ACCESS_KEY=SECRET
	AWS_SECRET_KEY=SECRET=SECRET # only read if AWS_SECRET_ACCESS_KEY is not set.

	# Session Token
	AWS_SESSION_TOKEN=TOKEN

Region value will instruct the SDK where to make service API requests to. If is
not provided in the environment the region must be provided before a service
client request is made.

	AWS_REGION=us-east-1

	# AWS_DEFAULT_REGION is only read if AWS_SDK_LOAD_CONFIG is also set,
	# and AWS_REGION is not also set.
	AWS_DEFAULT_REGION=us-east-1

Pro***REMOVED***le name the SDK should load use when loading shared con***REMOVED***g from the
con***REMOVED***guration ***REMOVED***les. If not provided "default" will be used as the pro***REMOVED***le name.

	AWS_PROFILE=my_pro***REMOVED***le

	# AWS_DEFAULT_PROFILE is only read if AWS_SDK_LOAD_CONFIG is also set,
	# and AWS_PROFILE is not also set.
	AWS_DEFAULT_PROFILE=my_pro***REMOVED***le

SDK load con***REMOVED***g instructs the SDK to load the shared con***REMOVED***g in addition to
shared credentials. This also expands the con***REMOVED***guration loaded so the shared
credentials will have parity with the shared con***REMOVED***g ***REMOVED***le. This also enables
Region and Pro***REMOVED***le support for the AWS_DEFAULT_REGION and AWS_DEFAULT_PROFILE
env values as well.

	AWS_SDK_LOAD_CONFIG=1

Shared credentials ***REMOVED***le path can be set to instruct the SDK to use an alternative
***REMOVED***le for the shared credentials. If not set the ***REMOVED***le will be loaded from
$HOME/.aws/credentials on Linux/Unix based systems, and
%USERPROFILE%\.aws\credentials on Windows.

	AWS_SHARED_CREDENTIALS_FILE=$HOME/my_shared_credentials

Shared con***REMOVED***g ***REMOVED***le path can be set to instruct the SDK to use an alternative
***REMOVED***le for the shared con***REMOVED***g. If not set the ***REMOVED***le will be loaded from
$HOME/.aws/con***REMOVED***g on Linux/Unix based systems, and
%USERPROFILE%\.aws\con***REMOVED***g on Windows.

	AWS_CONFIG_FILE=$HOME/my_shared_con***REMOVED***g

Path to a custom Credentials Authority (CA) bundle PEM ***REMOVED***le that the SDK
will use instead of the default system's root CA bundle. Use this only
if you want to replace the CA bundle the SDK uses for TLS requests.

	AWS_CA_BUNDLE=$HOME/my_custom_ca_bundle

Enabling this option will attempt to merge the Transport into the SDK's HTTP
client. If the client's Transport is not a http.Transport an error will be
returned. If the Transport's TLS con***REMOVED***g is set this option will cause the SDK
to overwrite the Transport's TLS con***REMOVED***g's  RootCAs value. If the CA bundle ***REMOVED***le
contains multiple certi***REMOVED***cates all of them will be loaded.

The Session option CustomCABundle is also available when creating sessions
to also enable this feature. CustomCABundle session option ***REMOVED***eld has priority
over the AWS_CA_BUNDLE environment variable, and will be used if both are set.

Setting a custom HTTPClient in the aws.Con***REMOVED***g options will override this setting.
To use this option and custom HTTP client, the HTTP client needs to be provided
when creating the session. Not the service client.
*/
package session
