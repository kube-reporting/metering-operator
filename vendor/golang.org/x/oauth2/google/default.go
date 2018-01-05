// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

package google

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/***REMOVED***lepath"
	"runtime"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// DefaultCredentials holds "Application Default Credentials".
// For more details, see:
// https://developers.google.com/accounts/docs/application-default-credentials
type DefaultCredentials struct {
	ProjectID   string // may be empty
	TokenSource oauth2.TokenSource
}

// DefaultClient returns an HTTP Client that uses the
// DefaultTokenSource to obtain authentication credentials.
func DefaultClient(ctx context.Context, scope ...string) (*http.Client, error) {
	ts, err := DefaultTokenSource(ctx, scope...)
	if err != nil {
		return nil, err
	}
	return oauth2.NewClient(ctx, ts), nil
}

// DefaultTokenSource returns the token source for
// "Application Default Credentials".
// It is a shortcut for FindDefaultCredentials(ctx, scope).TokenSource.
func DefaultTokenSource(ctx context.Context, scope ...string) (oauth2.TokenSource, error) {
	creds, err := FindDefaultCredentials(ctx, scope...)
	if err != nil {
		return nil, err
	}
	return creds.TokenSource, nil
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
func FindDefaultCredentials(ctx context.Context, scope ...string) (*DefaultCredentials, error) {
	// First, try the environment variable.
	const envVar = "GOOGLE_APPLICATION_CREDENTIALS"
	if ***REMOVED***lename := os.Getenv(envVar); ***REMOVED***lename != "" {
		creds, err := readCredentialsFile(ctx, ***REMOVED***lename, scope)
		if err != nil {
			return nil, fmt.Errorf("google: error getting credentials using %v environment variable: %v", envVar, err)
		}
		return creds, nil
	}

	// Second, try a well-known ***REMOVED***le.
	***REMOVED***lename := wellKnownFile()
	if creds, err := readCredentialsFile(ctx, ***REMOVED***lename, scope); err == nil {
		return creds, nil
	} ***REMOVED*** if !os.IsNotExist(err) {
		return nil, fmt.Errorf("google: error getting credentials using well-known ***REMOVED***le (%v): %v", ***REMOVED***lename, err)
	}

	// Third, if we're on Google App Engine use those credentials.
	if appengineTokenFunc != nil && !appengineFlex {
		return &DefaultCredentials{
			ProjectID:   appengineAppIDFunc(ctx),
			TokenSource: AppEngineTokenSource(ctx, scope...),
		}, nil
	}

	// Fourth, if we're on Google Compute Engine use the metadata server.
	if metadata.OnGCE() {
		id, _ := metadata.ProjectID()
		return &DefaultCredentials{
			ProjectID:   id,
			TokenSource: ComputeTokenSource(""),
		}, nil
	}

	// None are found; return helpful error.
	const url = "https://developers.google.com/accounts/docs/application-default-credentials"
	return nil, fmt.Errorf("google: could not ***REMOVED***nd default credentials. See %v for more information.", url)
}

func wellKnownFile() string {
	const f = "application_default_credentials.json"
	if runtime.GOOS == "windows" {
		return ***REMOVED***lepath.Join(os.Getenv("APPDATA"), "gcloud", f)
	}
	return ***REMOVED***lepath.Join(guessUnixHomeDir(), ".con***REMOVED***g", "gcloud", f)
}

func readCredentialsFile(ctx context.Context, ***REMOVED***lename string, scopes []string) (*DefaultCredentials, error) {
	b, err := ioutil.ReadFile(***REMOVED***lename)
	if err != nil {
		return nil, err
	}
	var f credentialsFile
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	ts, err := f.tokenSource(ctx, append([]string(nil), scopes...))
	if err != nil {
		return nil, err
	}
	return &DefaultCredentials{
		ProjectID:   f.ProjectID,
		TokenSource: ts,
	}, nil
}
