// +build appengine

package logrus

import "io"

// IsTerminal returns true if stderr's ***REMOVED***le descriptor is a terminal.
func IsTerminal(f io.Writer) bool {
	return true
}
