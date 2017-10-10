/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this ***REMOVED***le except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the speci***REMOVED***c language governing permissions and
limitations under the License.
*/

package transport

import "net/http"

// Con***REMOVED***g holds various options for establishing a transport.
type Con***REMOVED***g struct {
	// UserAgent is an optional ***REMOVED***eld that speci***REMOVED***es the caller of this
	// request.
	UserAgent string

	// The base TLS con***REMOVED***guration for this transport.
	TLS TLSCon***REMOVED***g

	// Username and password for basic authentication
	Username string
	Password string

	// Bearer token for authentication
	BearerToken string

	// CacheDir is the directory where we'll store HTTP cached responses.
	// If set to empty string, no caching mechanism will be used.
	CacheDir string

	// Impersonate is the con***REMOVED***g that this Con***REMOVED***g will impersonate using
	Impersonate ImpersonationCon***REMOVED***g

	// Transport may be used for custom HTTP behavior. This attribute may
	// not be speci***REMOVED***ed with the TLS client certi***REMOVED***cate options. Use
	// WrapTransport for most client level operations.
	Transport http.RoundTripper

	// WrapTransport will be invoked for custom HTTP behavior after the
	// underlying transport is initialized (either the transport created
	// from TLSClientCon***REMOVED***g, Transport, or http.DefaultTransport). The
	// con***REMOVED***g may layer other RoundTrippers on top of the returned
	// RoundTripper.
	WrapTransport func(rt http.RoundTripper) http.RoundTripper
}

// ImpersonationCon***REMOVED***g has all the available impersonation options
type ImpersonationCon***REMOVED***g struct {
	// UserName matches user.Info.GetName()
	UserName string
	// Groups matches user.Info.GetGroups()
	Groups []string
	// Extra matches user.Info.GetExtra()
	Extra map[string][]string
}

// HasCA returns whether the con***REMOVED***guration has a certi***REMOVED***cate authority or not.
func (c *Con***REMOVED***g) HasCA() bool {
	return len(c.TLS.CAData) > 0 || len(c.TLS.CAFile) > 0
}

// HasBasicAuth returns whether the con***REMOVED***guration has basic authentication or not.
func (c *Con***REMOVED***g) HasBasicAuth() bool {
	return len(c.Username) != 0
}

// HasTokenAuth returns whether the con***REMOVED***guration has token authentication or not.
func (c *Con***REMOVED***g) HasTokenAuth() bool {
	return len(c.BearerToken) != 0
}

// HasCertAuth returns whether the con***REMOVED***guration has certi***REMOVED***cate authentication or not.
func (c *Con***REMOVED***g) HasCertAuth() bool {
	return len(c.TLS.CertData) != 0 || len(c.TLS.CertFile) != 0
}

// TLSCon***REMOVED***g holds the information needed to set up a TLS transport.
type TLSCon***REMOVED***g struct {
	CAFile   string // Path of the PEM-encoded server trusted root certi***REMOVED***cates.
	CertFile string // Path of the PEM-encoded client certi***REMOVED***cate.
	KeyFile  string // Path of the PEM-encoded client key.

	Insecure   bool   // Server should be accessed without verifying the certi***REMOVED***cate. For testing only.
	ServerName string // Override for the server name passed to the server for SNI and used to verify certi***REMOVED***cates.

	CAData   []byte // Bytes of the PEM-encoded server trusted root certi***REMOVED***cates. Supercedes CAFile.
	CertData []byte // Bytes of the PEM-encoded client certi***REMOVED***cate. Supercedes CertFile.
	KeyData  []byte // Bytes of the PEM-encoded client key. Supercedes KeyFile.
}
