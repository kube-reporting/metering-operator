package shareddefaults

import (
	"os"
	"path/***REMOVED***lepath"
	"runtime"
)

// SharedCredentialsFilename returns the SDK's default ***REMOVED***le path
// for the shared credentials ***REMOVED***le.
//
// Builds the shared con***REMOVED***g ***REMOVED***le path based on the OS's platform.
//
//   - Linux/Unix: $HOME/.aws/credentials
//   - Windows: %USERPROFILE%\.aws\credentials
func SharedCredentialsFilename() string {
	return ***REMOVED***lepath.Join(UserHomeDir(), ".aws", "credentials")
}

// SharedCon***REMOVED***gFilename returns the SDK's default ***REMOVED***le path for
// the shared con***REMOVED***g ***REMOVED***le.
//
// Builds the shared con***REMOVED***g ***REMOVED***le path based on the OS's platform.
//
//   - Linux/Unix: $HOME/.aws/con***REMOVED***g
//   - Windows: %USERPROFILE%\.aws\con***REMOVED***g
func SharedCon***REMOVED***gFilename() string {
	return ***REMOVED***lepath.Join(UserHomeDir(), ".aws", "con***REMOVED***g")
}

// UserHomeDir returns the home directory for the user the process is
// running under.
func UserHomeDir() string {
	if runtime.GOOS == "windows" { // Windows
		return os.Getenv("USERPROFILE")
	}

	// *nix
	return os.Getenv("HOME")
}
