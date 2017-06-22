package credentials

import (
	"fmt"
	"os"

	"github.com/go-ini/ini"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/internal/shareddefaults"
)

// SharedCredsProviderName provides a name of SharedCreds provider
const SharedCredsProviderName = "SharedCredentialsProvider"

var (
	// ErrSharedCredentialsHomeNotFound is emitted when the user directory cannot be found.
	ErrSharedCredentialsHomeNotFound = awserr.New("UserHomeNotFound", "user home directory not found.", nil)
)

// A SharedCredentialsProvider retrieves credentials from the current user's home
// directory, and keeps track if those credentials are expired.
//
// Pro***REMOVED***le ini ***REMOVED***le example: $HOME/.aws/credentials
type SharedCredentialsProvider struct {
	// Path to the shared credentials ***REMOVED***le.
	//
	// If empty will look for "AWS_SHARED_CREDENTIALS_FILE" env variable. If the
	// env value is empty will default to current user's home directory.
	// Linux/OSX: "$HOME/.aws/credentials"
	// Windows:   "%USERPROFILE%\.aws\credentials"
	Filename string

	// AWS Pro***REMOVED***le to extract credentials from the shared credentials ***REMOVED***le. If empty
	// will default to environment variable "AWS_PROFILE" or "default" if
	// environment variable is also not set.
	Pro***REMOVED***le string

	// retrieved states if the credentials have been successfully retrieved.
	retrieved bool
}

// NewSharedCredentials returns a pointer to a new Credentials object
// wrapping the Pro***REMOVED***le ***REMOVED***le provider.
func NewSharedCredentials(***REMOVED***lename, pro***REMOVED***le string) *Credentials {
	return NewCredentials(&SharedCredentialsProvider{
		Filename: ***REMOVED***lename,
		Pro***REMOVED***le:  pro***REMOVED***le,
	})
}

// Retrieve reads and extracts the shared credentials from the current
// users home directory.
func (p *SharedCredentialsProvider) Retrieve() (Value, error) {
	p.retrieved = false

	***REMOVED***lename, err := p.***REMOVED***lename()
	if err != nil {
		return Value{ProviderName: SharedCredsProviderName}, err
	}

	creds, err := loadPro***REMOVED***le(***REMOVED***lename, p.pro***REMOVED***le())
	if err != nil {
		return Value{ProviderName: SharedCredsProviderName}, err
	}

	p.retrieved = true
	return creds, nil
}

// IsExpired returns if the shared credentials have expired.
func (p *SharedCredentialsProvider) IsExpired() bool {
	return !p.retrieved
}

// loadPro***REMOVED***les loads from the ***REMOVED***le pointed to by shared credentials ***REMOVED***lename for pro***REMOVED***le.
// The credentials retrieved from the pro***REMOVED***le will be returned or error. Error will be
// returned if it fails to read from the ***REMOVED***le, or the data is invalid.
func loadPro***REMOVED***le(***REMOVED***lename, pro***REMOVED***le string) (Value, error) {
	con***REMOVED***g, err := ini.Load(***REMOVED***lename)
	if err != nil {
		return Value{ProviderName: SharedCredsProviderName}, awserr.New("SharedCredsLoad", "failed to load shared credentials ***REMOVED***le", err)
	}
	iniPro***REMOVED***le, err := con***REMOVED***g.GetSection(pro***REMOVED***le)
	if err != nil {
		return Value{ProviderName: SharedCredsProviderName}, awserr.New("SharedCredsLoad", "failed to get pro***REMOVED***le", err)
	}

	id, err := iniPro***REMOVED***le.GetKey("aws_access_key_id")
	if err != nil {
		return Value{ProviderName: SharedCredsProviderName}, awserr.New("SharedCredsAccessKey",
			fmt.Sprintf("shared credentials %s in %s did not contain aws_access_key_id", pro***REMOVED***le, ***REMOVED***lename),
			err)
	}

	secret, err := iniPro***REMOVED***le.GetKey("aws_secret_access_key")
	if err != nil {
		return Value{ProviderName: SharedCredsProviderName}, awserr.New("SharedCredsSecret",
			fmt.Sprintf("shared credentials %s in %s did not contain aws_secret_access_key", pro***REMOVED***le, ***REMOVED***lename),
			nil)
	}

	// Default to empty string if not found
	token := iniPro***REMOVED***le.Key("aws_session_token")

	return Value{
		AccessKeyID:     id.String(),
		SecretAccessKey: secret.String(),
		SessionToken:    token.String(),
		ProviderName:    SharedCredsProviderName,
	}, nil
}

// ***REMOVED***lename returns the ***REMOVED***lename to use to read AWS shared credentials.
//
// Will return an error if the user's home directory path cannot be found.
func (p *SharedCredentialsProvider) ***REMOVED***lename() (string, error) {
	if len(p.Filename) != 0 {
		return p.Filename, nil
	}

	if p.Filename = os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); len(p.Filename) != 0 {
		return p.Filename, nil
	}

	if home := shareddefaults.UserHomeDir(); len(home) == 0 {
		// Backwards compatibility of home directly not found error being returned.
		// This error is too verbose, failure when opening the ***REMOVED***le would of been
		// a better error to return.
		return "", ErrSharedCredentialsHomeNotFound
	}

	p.Filename = shareddefaults.SharedCredentialsFilename()

	return p.Filename, nil
}

// pro***REMOVED***le returns the AWS shared credentials pro***REMOVED***le.  If empty will read
// environment variable "AWS_PROFILE". If that is not set pro***REMOVED***le will
// return "default".
func (p *SharedCredentialsProvider) pro***REMOVED***le() string {
	if p.Pro***REMOVED***le == "" {
		p.Pro***REMOVED***le = os.Getenv("AWS_PROFILE")
	}
	if p.Pro***REMOVED***le == "" {
		p.Pro***REMOVED***le = "default"
	}

	return p.Pro***REMOVED***le
}
