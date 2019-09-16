// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

// +build !go1.9

package google

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// DefaultCredentials holds Google credentials, including "Application Default Credentials".
// For more details, see:
// https://developers.google.com/accounts/docs/application-default-credentials
type DefaultCredentials struct {
	ProjectID   string // may be empty
	TokenSource oauth2.TokenSource

	// JSON contains the raw bytes from a JSON credentials ***REMOVED***le.
	// This ***REMOVED***eld may be nil if authentication is provided by the
	// environment and not with a credentials ***REMOVED***le, e.g. when code is
	// running on Google Cloud Platform.
	JSON []byte
}

// FindDefaultCredentials searches for "Application Default Credentials".
//
// It looks for credentials in the following places,
// preferring the ***REMOVED***rst location found:
//
//   1. A JSON ***REMOVED***le whose path is speci***REMOVED***ed by the
//      GOOGLE_APPLICATION_CREDENTIALS environment variable.
//   2. A JSON ***REMOVED***le in a location known to the gcloud command-line tool.
//      On Windows, this is %APPDATA%/gcloud/application_default_credentials.json.
//      On other systems, $HOME/.con***REMOVED***g/gcloud/application_default_credentials.json.
//   3. On Google App Engine it uses the appengine.AccessToken function.
//   4. On Google Compute Engine and Google App Engine Managed VMs, it fetches
//      credentials from the metadata server.
//      (In this ***REMOVED***nal case any provided scopes are ignored.)
func FindDefaultCredentials(ctx context.Context, scopes ...string) (*DefaultCredentials, error) {
	return ***REMOVED***ndDefaultCredentials(ctx, scopes)
}

// CredentialsFromJSON obtains Google credentials from a JSON value. The JSON can
// represent either a Google Developers Console client_credentials.json ***REMOVED***le (as in
// Con***REMOVED***gFromJSON) or a Google Developers service account key ***REMOVED***le (as in
// JWTCon***REMOVED***gFromJSON).
//
// Note: despite the name, the returned credentials may not be Application Default Credentials.
func CredentialsFromJSON(ctx context.Context, jsonData []byte, scopes ...string) (*DefaultCredentials, error) {
	return credentialsFromJSON(ctx, jsonData, scopes)
}
