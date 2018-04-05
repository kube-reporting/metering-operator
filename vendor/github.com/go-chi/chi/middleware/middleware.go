package middleware

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it ***REMOVED***ts in an interface{} without allocation. This technique
// for de***REMOVED***ning context keys was copied from Go 1.7's new use of context in net/http.
type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "chi/middleware context value " + k.name
}
