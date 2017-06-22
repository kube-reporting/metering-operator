package defaults

import (
	"github.com/aws/aws-sdk-go/internal/shareddefaults"
)

// SharedCredentialsFilename returns the SDK's default ***REMOVED***le path
// for the shared credentials ***REMOVED***le.
//
// Builds the shared con***REMOVED***g ***REMOVED***le path based on the OS's platform.
//
//   - Linux/Unix: $HOME/.aws/credentials
//   - Windows: %USERPROFILE%\.aws\credentials
func SharedCredentialsFilename() string {
	return shareddefaults.SharedCredentialsFilename()
}

// SharedCon***REMOVED***gFilename returns the SDK's default ***REMOVED***le path for
// the shared con***REMOVED***g ***REMOVED***le.
//
// Builds the shared con***REMOVED***g ***REMOVED***le path based on the OS's platform.
//
//   - Linux/Unix: $HOME/.aws/con***REMOVED***g
//   - Windows: %USERPROFILE%\.aws\con***REMOVED***g
func SharedCon***REMOVED***gFilename() string {
	return shareddefaults.SharedCon***REMOVED***gFilename()
}
