/*
Package session provides con***REMOVED***guration for the SDK's service clients.

Sessions can be shared across all service clients that share the same base
con***REMOVED***guration.  The Session is built from the SDK's default con***REMOVED***guration and
request handlers.

Sessions should be cached when possible, because creating a new Session will
load all con***REMOVED***guration values from the environment, and con***REMOVED***g ***REMOVED***les each time
the Session is created. Sharing the Session value across all of your service
clients will ensure the con***REMOVED***guration is loaded the fewest number of times possible.

Concurrency

Sessions are safe to use concurrently as long as the Session is not being
modi***REMOVED***ed. The SDK will not modify the Session once the Session has been created.
Creating service clients concurrently from a shared Session is safe.

Sessions from Shared Con***REMOVED***g

Sessions can be created using the method above that will only load the
additional con***REMOVED***g if the AWS_SDK_LOAD_CONFIG environment variable is set.
Alternatively you can explicitly create a Session with shared con***REMOVED***g enabled.
To do this you can use NewSessionWithOptions to con***REMOVED***gure how the Session will
be created. Using the NewSessionWithOptions with SharedCon***REMOVED***gState set to
SharedCon***REMOVED***gEnable will create the session as if the AWS_SDK_LOAD_CONFIG
environment variable was set.

Creating Sessions

When creating Sessions optional aws.Con***REMOVED***g values can be passed in that will
override the default, or loaded con***REMOVED***g values the Session is being created
with. This allows you to provide additional, or case based, con***REMOVED***guration
as needed.

By default NewSession will only load credentials from the shared credentials
***REMOVED***le (~/.aws/credentials). If the AWS_SDK_LOAD_CONFIG environment variable is
set to a truthy value the Session will be created from the con***REMOVED***guration
values from the shared con***REMOVED***g (~/.aws/con***REMOVED***g) and shared credentials
(~/.aws/credentials) ***REMOVED***les. See the section Sessions from Shared Con***REMOVED***g for
more information.

Create a Session with the default con***REMOVED***g and request handlers. With credentials
region, and pro***REMOVED***le loaded from the environment and shared con***REMOVED***g automatically.
Requires the AWS_PROFILE to be set, or "default" is used.

	// Create Session
	sess := session.Must(session.NewSession())

	// Create a Session with a custom region
	sess := session.Must(session.NewSession(&aws.Con***REMOVED***g{
		Region: aws.String("us-east-1"),
	}))

	// Create a S3 client instance from a session
	sess := session.Must(session.NewSession())

	svc := s3.New(sess)

Create Session With Option Overrides

In addition to NewSession, Sessions can be created using NewSessionWithOptions.
This func allows you to control and override how the Session will be created
through code instead of being driven by environment variables only.

Use NewSessionWithOptions when you want to provide the con***REMOVED***g pro***REMOVED***le, or
override the shared con***REMOVED***g state (AWS_SDK_LOAD_CONFIG).

	// Equivalent to session.NewSession()
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		// Options
	}))

	// Specify pro***REMOVED***le to load for the session's con***REMOVED***g
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		 Pro***REMOVED***le: "pro***REMOVED***le_name",
	}))

	// Specify pro***REMOVED***le for con***REMOVED***g and region for requests
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		 Con***REMOVED***g: aws.Con***REMOVED***g{Region: aws.String("us-east-1")},
		 Pro***REMOVED***le: "pro***REMOVED***le_name",
	}))

	// Force enable Shared Con***REMOVED***g support
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedCon***REMOVED***gState: session.SharedCon***REMOVED***gEnable,
	}))

Adding Handlers

You can add handlers to a session for processing HTTP requests. All service
clients that use the session inherit the handlers. For example, the following
handler logs every request and its payload made by a service client:

	// Create a session, and add additional handlers for all service
	// clients created with the Session to inherit. Adds logging handler.
	sess := session.Must(session.NewSession())

	sess.Handlers.Send.PushFront(func(r *request.Request) {
		// Log every request made and its payload
		logger.Println("Request: %s/%s, Payload: %s",
			r.ClientInfo.ServiceName, r.Operation, r.Params)
	})

Deprecated "New" function

The New session function has been deprecated because it does not provide good
way to return errors that occur when loading the con***REMOVED***guration ***REMOVED***les and values.
Because of this, NewSession was created so errors can be retrieved when
creating a session fails.

Shared Con***REMOVED***g Fields

By default the SDK will only load the shared credentials ***REMOVED***le's (~/.aws/credentials)
credentials values, and all other con***REMOVED***g is provided by the environment variables,
SDK defaults, and user provided aws.Con***REMOVED***g values.

If the AWS_SDK_LOAD_CONFIG environment variable is set, or SharedCon***REMOVED***gEnable
option is used to create the Session the full shared con***REMOVED***g values will be
loaded. This includes credentials, region, and support for assume role. In
addition the Session will load its con***REMOVED***guration from both the shared con***REMOVED***g
***REMOVED***le (~/.aws/con***REMOVED***g) and shared credentials ***REMOVED***le (~/.aws/credentials). Both
***REMOVED***les have the same format.

If both con***REMOVED***g ***REMOVED***les are present the con***REMOVED***guration from both ***REMOVED***les will be
read. The Session will be created from con***REMOVED***guration values from the shared
credentials ***REMOVED***le (~/.aws/credentials) over those in the shared con***REMOVED***g ***REMOVED***le (~/.aws/con***REMOVED***g).

Credentials are the values the SDK should use for authenticating requests with
AWS Services. They arfrom a con***REMOVED***guration ***REMOVED***le will need to include both
aws_access_key_id and aws_secret_access_key must be provided together in the
same ***REMOVED***le to be considered valid. The values will be ignored if not a complete
group. aws_session_token is an optional ***REMOVED***eld that can be provided if both of
the other two ***REMOVED***elds are also provided.

	aws_access_key_id = AKID
	aws_secret_access_key = SECRET
	aws_session_token = TOKEN

Assume Role values allow you to con***REMOVED***gure the SDK to assume an IAM role using
a set of credentials provided in a con***REMOVED***g ***REMOVED***le via the source_pro***REMOVED***le ***REMOVED***eld.
Both "role_arn" and "source_pro***REMOVED***le" are required. The SDK supports assuming
a role with MFA token if the session option AssumeRoleTokenProvider
is set.

	role_arn = arn:aws:iam::<account_number>:role/<role_name>
	source_pro***REMOVED***le = pro***REMOVED***le_with_creds
	external_id = 1234
	mfa_serial = <serial or mfa arn>
	role_session_name = session_name

Region is the region the SDK should use for looking up AWS service endpoints
and signing requests.

	region = us-east-1

Assume Role with MFA token

To create a session with support for assuming an IAM role with MFA set the
session option AssumeRoleTokenProvider to a function that will prompt for the
MFA token code when the SDK assumes the role and refreshes the role's credentials.
This allows you to con***REMOVED***gure the SDK via the shared con***REMOVED***g to assumea role
with MFA tokens.

In order for the SDK to assume a role with MFA the SharedCon***REMOVED***gState
session option must be set to SharedCon***REMOVED***gEnable, or AWS_SDK_LOAD_CONFIG
environment variable set.

The shared con***REMOVED***guration instructs the SDK to assume an IAM role with MFA
when the mfa_serial con***REMOVED***guration ***REMOVED***eld is set in the shared con***REMOVED***g
(~/.aws/con***REMOVED***g) or shared credentials (~/.aws/credentials) ***REMOVED***le.

If mfa_serial is set in the con***REMOVED***guration, the SDK will assume the role, and
the AssumeRoleTokenProvider session option is not set an an error will
be returned when creating the session.

    sess := session.Must(session.NewSessionWithOptions(session.Options{
        AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
    }))

    // Create service client value con***REMOVED***gured for credentials
    // from assumed role.
    svc := s3.New(sess)

To setup assume role outside of a session see the stscrds.AssumeRoleProvider
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
