// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

// +build !go1.9

// Package google provides support for making OAuth2 authorized and authenticated
// HTTP requests to Google APIs. It supports the Web server flow, client-side
// credentials, service accounts, Google Compute Engine service accounts, and Google
// App Engine service accounts.
//
// A brief overview of the package follows. For more information, please read
// https://developers.google.com/accounts/docs/OAuth2
// and
// https://developers.google.com/accounts/docs/application-default-credentials.
//
// OAuth2 Con***REMOVED***gs
//
// Two functions in this package return golang.org/x/oauth2.Con***REMOVED***g values from Google credential
// data. Google supports two JSON formats for OAuth2 credentials: one is handled by Con***REMOVED***gFromJSON,
// the other by JWTCon***REMOVED***gFromJSON. The returned Con***REMOVED***g can be used to obtain a TokenSource or
// create an http.Client.
//
//
// Credentials
//
// The DefaultCredentials type represents Google Application Default Credentials, as
// well as other forms of credential.
//
// Use FindDefaultCredentials to obtain Application Default Credentials.
// FindDefaultCredentials looks in some well-known places for a credentials ***REMOVED***le, and
// will call AppEngineTokenSource or ComputeTokenSource as needed.
//
// DefaultClient and DefaultTokenSource are convenience methods. They ***REMOVED***rst call FindDefaultCredentials,
// then use the credentials to construct an http.Client or an oauth2.TokenSource.
//
// Use CredentialsFromJSON to obtain credentials from either of the two JSON
// formats described in OAuth2 Con***REMOVED***gs, above. (The DefaultCredentials returned may
// not be "Application Default Credentials".) The TokenSource in the returned value
// is the same as the one obtained from the oauth2.Con***REMOVED***g returned from
// Con***REMOVED***gFromJSON or JWTCon***REMOVED***gFromJSON, but the DefaultCredentials may contain
// additional information that is useful is some circumstances.
package google // import "golang.org/x/oauth2/google"
