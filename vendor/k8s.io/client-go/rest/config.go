/*
Copyright 2016 The Kubernetes Authors.

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

package rest

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/***REMOVED***lepath"
	gruntime "runtime"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/pkg/version"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/klog"
)

const (
	DefaultQPS   float32 = 5.0
	DefaultBurst int     = 10
)

var ErrNotInCluster = errors.New("unable to load in-cluster con***REMOVED***guration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be de***REMOVED***ned")

// Con***REMOVED***g holds the common attributes that can be passed to a Kubernetes client on
// initialization.
type Con***REMOVED***g struct {
	// Host must be a host string, a host:port pair, or a URL to the base of the apiserver.
	// If a URL is given then the (optional) Path of that URL represents a pre***REMOVED***x that must
	// be appended to all request URIs used to access the apiserver. This allows a frontend
	// proxy to easily relocate all of the apiserver endpoints.
	Host string
	// APIPath is a sub-path that points to an API root.
	APIPath string

	// ContentCon***REMOVED***g contains settings that affect how objects are transformed when
	// sent to the server.
	ContentCon***REMOVED***g

	// Server requires Basic authentication
	Username string
	Password string

	// Server requires Bearer authentication. This client will not attempt to use
	// refresh tokens for an OAuth2 flow.
	// TODO: demonstrate an OAuth2 compatible client.
	BearerToken string

	// Path to a ***REMOVED***le containing a BearerToken.
	// If set, the contents are periodically read.
	// The last successfully read value takes precedence over BearerToken.
	BearerTokenFile string

	// Impersonate is the con***REMOVED***guration that RESTClient will use for impersonation.
	Impersonate ImpersonationCon***REMOVED***g

	// Server requires plugin-speci***REMOVED***ed authentication.
	AuthProvider *clientcmdapi.AuthProviderCon***REMOVED***g

	// Callback to persist con***REMOVED***g for AuthProvider.
	AuthCon***REMOVED***gPersister AuthProviderCon***REMOVED***gPersister

	// Exec-based authentication provider.
	ExecProvider *clientcmdapi.ExecCon***REMOVED***g

	// TLSClientCon***REMOVED***g contains settings to enable transport layer security
	TLSClientCon***REMOVED***g

	// UserAgent is an optional ***REMOVED***eld that speci***REMOVED***es the caller of this request.
	UserAgent string

	// Transport may be used for custom HTTP behavior. This attribute may not
	// be speci***REMOVED***ed with the TLS client certi***REMOVED***cate options. Use WrapTransport
	// for most client level operations.
	Transport http.RoundTripper
	// WrapTransport will be invoked for custom HTTP behavior after the underlying
	// transport is initialized (either the transport created from TLSClientCon***REMOVED***g,
	// Transport, or http.DefaultTransport). The con***REMOVED***g may layer other RoundTrippers
	// on top of the returned RoundTripper.
	WrapTransport func(rt http.RoundTripper) http.RoundTripper

	// QPS indicates the maximum QPS to the master from this client.
	// If it's zero, the created RESTClient will use DefaultQPS: 5
	QPS float32

	// Maximum burst for throttle.
	// If it's zero, the created RESTClient will use DefaultBurst: 10.
	Burst int

	// Rate limiter for limiting connections to the master from this client. If present overwrites QPS/Burst
	RateLimiter flowcontrol.RateLimiter

	// The maximum length of time to wait before giving up on a server request. A value of zero means no timeout.
	Timeout time.Duration

	// Dial speci***REMOVED***es the dial function for creating unencrypted TCP connections.
	Dial func(ctx context.Context, network, address string) (net.Conn, error)

	// Version forces a speci***REMOVED***c version to be used (if registered)
	// Do we need this?
	// Version string
}

// ImpersonationCon***REMOVED***g has all the available impersonation options
type ImpersonationCon***REMOVED***g struct {
	// UserName is the username to impersonate on each request.
	UserName string
	// Groups are the groups to impersonate on each request.
	Groups []string
	// Extra is a free-form ***REMOVED***eld which can be used to link some authentication information
	// to authorization information.  This ***REMOVED***eld allows you to impersonate it.
	Extra map[string][]string
}

// +k8s:deepcopy-gen=true
// TLSClientCon***REMOVED***g contains settings to enable transport layer security
type TLSClientCon***REMOVED***g struct {
	// Server should be accessed without verifying the TLS certi***REMOVED***cate. For testing only.
	Insecure bool
	// ServerName is passed to the server for SNI and is used in the client to check server
	// ceriti***REMOVED***cates against. If ServerName is empty, the hostname used to contact the
	// server is used.
	ServerName string

	// Server requires TLS client certi***REMOVED***cate authentication
	CertFile string
	// Server requires TLS client certi***REMOVED***cate authentication
	KeyFile string
	// Trusted root certi***REMOVED***cates for server
	CAFile string

	// CertData holds PEM-encoded bytes (typically read from a client certi***REMOVED***cate ***REMOVED***le).
	// CertData takes precedence over CertFile
	CertData []byte
	// KeyData holds PEM-encoded bytes (typically read from a client certi***REMOVED***cate key ***REMOVED***le).
	// KeyData takes precedence over KeyFile
	KeyData []byte
	// CAData holds PEM-encoded bytes (typically read from a root certi***REMOVED***cates bundle).
	// CAData takes precedence over CAFile
	CAData []byte
}

type ContentCon***REMOVED***g struct {
	// AcceptContentTypes speci***REMOVED***es the types the client will accept and is optional.
	// If not set, ContentType will be used to de***REMOVED***ne the Accept header
	AcceptContentTypes string
	// ContentType speci***REMOVED***es the wire format used to communicate with the server.
	// This value will be set as the Accept header on requests made to the server, and
	// as the default content type on any object sent to the server. If not set,
	// "application/json" is used.
	ContentType string
	// GroupVersion is the API version to talk to. Must be provided when initializing
	// a RESTClient directly. When initializing a Client, will be set with the default
	// code version.
	GroupVersion *schema.GroupVersion
	// NegotiatedSerializer is used for obtaining encoders and decoders for multiple
	// supported media types.
	NegotiatedSerializer runtime.NegotiatedSerializer
}

// RESTClientFor returns a RESTClient that satis***REMOVED***es the requested attributes on a client Con***REMOVED***g
// object. Note that a RESTClient may require ***REMOVED***elds that are optional when initializing a Client.
// A RESTClient created by this method is generic - it expects to operate on an API that follows
// the Kubernetes conventions, but may not be the Kubernetes API.
func RESTClientFor(con***REMOVED***g *Con***REMOVED***g) (*RESTClient, error) {
	if con***REMOVED***g.GroupVersion == nil {
		return nil, fmt.Errorf("GroupVersion is required when initializing a RESTClient")
	}
	if con***REMOVED***g.NegotiatedSerializer == nil {
		return nil, fmt.Errorf("NegotiatedSerializer is required when initializing a RESTClient")
	}
	qps := con***REMOVED***g.QPS
	if con***REMOVED***g.QPS == 0.0 {
		qps = DefaultQPS
	}
	burst := con***REMOVED***g.Burst
	if con***REMOVED***g.Burst == 0 {
		burst = DefaultBurst
	}

	baseURL, versionedAPIPath, err := defaultServerUrlFor(con***REMOVED***g)
	if err != nil {
		return nil, err
	}

	transport, err := TransportFor(con***REMOVED***g)
	if err != nil {
		return nil, err
	}

	var httpClient *http.Client
	if transport != http.DefaultTransport {
		httpClient = &http.Client{Transport: transport}
		if con***REMOVED***g.Timeout > 0 {
			httpClient.Timeout = con***REMOVED***g.Timeout
		}
	}

	return NewRESTClient(baseURL, versionedAPIPath, con***REMOVED***g.ContentCon***REMOVED***g, qps, burst, con***REMOVED***g.RateLimiter, httpClient)
}

// UnversionedRESTClientFor is the same as RESTClientFor, except that it allows
// the con***REMOVED***g.Version to be empty.
func UnversionedRESTClientFor(con***REMOVED***g *Con***REMOVED***g) (*RESTClient, error) {
	if con***REMOVED***g.NegotiatedSerializer == nil {
		return nil, fmt.Errorf("NegotiatedSerializer is required when initializing a RESTClient")
	}

	baseURL, versionedAPIPath, err := defaultServerUrlFor(con***REMOVED***g)
	if err != nil {
		return nil, err
	}

	transport, err := TransportFor(con***REMOVED***g)
	if err != nil {
		return nil, err
	}

	var httpClient *http.Client
	if transport != http.DefaultTransport {
		httpClient = &http.Client{Transport: transport}
		if con***REMOVED***g.Timeout > 0 {
			httpClient.Timeout = con***REMOVED***g.Timeout
		}
	}

	versionCon***REMOVED***g := con***REMOVED***g.ContentCon***REMOVED***g
	if versionCon***REMOVED***g.GroupVersion == nil {
		v := metav1.SchemeGroupVersion
		versionCon***REMOVED***g.GroupVersion = &v
	}

	return NewRESTClient(baseURL, versionedAPIPath, versionCon***REMOVED***g, con***REMOVED***g.QPS, con***REMOVED***g.Burst, con***REMOVED***g.RateLimiter, httpClient)
}

// SetKubernetesDefaults sets default values on the provided client con***REMOVED***g for accessing the
// Kubernetes API or returns an error if any of the defaults are impossible or invalid.
func SetKubernetesDefaults(con***REMOVED***g *Con***REMOVED***g) error {
	if len(con***REMOVED***g.UserAgent) == 0 {
		con***REMOVED***g.UserAgent = DefaultKubernetesUserAgent()
	}
	return nil
}

// adjustCommit returns suf***REMOVED***cient signi***REMOVED***cant ***REMOVED***gures of the commit's git hash.
func adjustCommit(c string) string {
	if len(c) == 0 {
		return "unknown"
	}
	if len(c) > 7 {
		return c[:7]
	}
	return c
}

// adjustVersion strips "alpha", "beta", etc. from version in form
// major.minor.patch-[alpha|beta|etc].
func adjustVersion(v string) string {
	if len(v) == 0 {
		return "unknown"
	}
	seg := strings.SplitN(v, "-", 2)
	return seg[0]
}

// adjustCommand returns the last component of the
// OS-speci***REMOVED***c command path for use in User-Agent.
func adjustCommand(p string) string {
	// Unlikely, but better than returning "".
	if len(p) == 0 {
		return "unknown"
	}
	return ***REMOVED***lepath.Base(p)
}

// buildUserAgent builds a User-Agent string from given args.
func buildUserAgent(command, version, os, arch, commit string) string {
	return fmt.Sprintf(
		"%s/%s (%s/%s) kubernetes/%s", command, version, os, arch, commit)
}

// DefaultKubernetesUserAgent returns a User-Agent string built from static global vars.
func DefaultKubernetesUserAgent() string {
	return buildUserAgent(
		adjustCommand(os.Args[0]),
		adjustVersion(version.Get().GitVersion),
		gruntime.GOOS,
		gruntime.GOARCH,
		adjustCommit(version.Get().GitCommit))
}

// InClusterCon***REMOVED***g returns a con***REMOVED***g object which uses the service account
// kubernetes gives to pods. It's intended for clients that expect to be
// running inside a pod running on kubernetes. It will return ErrNotInCluster
// if called from a process not running in a kubernetes environment.
func InClusterCon***REMOVED***g() (*Con***REMOVED***g, error) {
	const (
		tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
		rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	)
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		return nil, ErrNotInCluster
	}

	token, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	tlsClientCon***REMOVED***g := TLSClientCon***REMOVED***g{}

	if _, err := certutil.NewPool(rootCAFile); err != nil {
		klog.Errorf("Expected to load root CA con***REMOVED***g from %s, but got err: %v", rootCAFile, err)
	} ***REMOVED*** {
		tlsClientCon***REMOVED***g.CAFile = rootCAFile
	}

	return &Con***REMOVED***g{
		// TODO: switch to using cluster DNS.
		Host:            "https://" + net.JoinHostPort(host, port),
		TLSClientCon***REMOVED***g: tlsClientCon***REMOVED***g,
		BearerToken:     string(token),
		BearerTokenFile: tokenFile,
	}, nil
}

// IsCon***REMOVED***gTransportTLS returns true if and only if the provided
// con***REMOVED***g will result in a protected connection to the server when it
// is passed to restclient.RESTClientFor().  Use to determine when to
// send credentials over the wire.
//
// Note: the Insecure flag is ignored when testing for this value, so MITM attacks are
// still possible.
func IsCon***REMOVED***gTransportTLS(con***REMOVED***g Con***REMOVED***g) bool {
	baseURL, _, err := defaultServerUrlFor(&con***REMOVED***g)
	if err != nil {
		return false
	}
	return baseURL.Scheme == "https"
}

// LoadTLSFiles copies the data from the CertFile, KeyFile, and CAFile ***REMOVED***elds into the CertData,
// KeyData, and CAFile ***REMOVED***elds, or returns an error. If no error is returned, all three ***REMOVED***elds are
// either populated or were empty to start.
func LoadTLSFiles(c *Con***REMOVED***g) error {
	var err error
	c.CAData, err = dataFromSliceOrFile(c.CAData, c.CAFile)
	if err != nil {
		return err
	}

	c.CertData, err = dataFromSliceOrFile(c.CertData, c.CertFile)
	if err != nil {
		return err
	}

	c.KeyData, err = dataFromSliceOrFile(c.KeyData, c.KeyFile)
	if err != nil {
		return err
	}
	return nil
}

// dataFromSliceOrFile returns data from the slice (if non-empty), or from the ***REMOVED***le,
// or an error if an error occurred reading the ***REMOVED***le
func dataFromSliceOrFile(data []byte, ***REMOVED***le string) ([]byte, error) {
	if len(data) > 0 {
		return data, nil
	}
	if len(***REMOVED***le) > 0 {
		***REMOVED***leData, err := ioutil.ReadFile(***REMOVED***le)
		if err != nil {
			return []byte{}, err
		}
		return ***REMOVED***leData, nil
	}
	return nil, nil
}

func AddUserAgent(con***REMOVED***g *Con***REMOVED***g, userAgent string) *Con***REMOVED***g {
	fullUserAgent := DefaultKubernetesUserAgent() + "/" + userAgent
	con***REMOVED***g.UserAgent = fullUserAgent
	return con***REMOVED***g
}

// AnonymousClientCon***REMOVED***g returns a copy of the given con***REMOVED***g with all user credentials (cert/key, bearer token, and username/password) removed
func AnonymousClientCon***REMOVED***g(con***REMOVED***g *Con***REMOVED***g) *Con***REMOVED***g {
	// copy only known safe ***REMOVED***elds
	return &Con***REMOVED***g{
		Host:          con***REMOVED***g.Host,
		APIPath:       con***REMOVED***g.APIPath,
		ContentCon***REMOVED***g: con***REMOVED***g.ContentCon***REMOVED***g,
		TLSClientCon***REMOVED***g: TLSClientCon***REMOVED***g{
			Insecure:   con***REMOVED***g.Insecure,
			ServerName: con***REMOVED***g.ServerName,
			CAFile:     con***REMOVED***g.TLSClientCon***REMOVED***g.CAFile,
			CAData:     con***REMOVED***g.TLSClientCon***REMOVED***g.CAData,
		},
		RateLimiter:   con***REMOVED***g.RateLimiter,
		UserAgent:     con***REMOVED***g.UserAgent,
		Transport:     con***REMOVED***g.Transport,
		WrapTransport: con***REMOVED***g.WrapTransport,
		QPS:           con***REMOVED***g.QPS,
		Burst:         con***REMOVED***g.Burst,
		Timeout:       con***REMOVED***g.Timeout,
		Dial:          con***REMOVED***g.Dial,
	}
}

// CopyCon***REMOVED***g returns a copy of the given con***REMOVED***g
func CopyCon***REMOVED***g(con***REMOVED***g *Con***REMOVED***g) *Con***REMOVED***g {
	return &Con***REMOVED***g{
		Host:            con***REMOVED***g.Host,
		APIPath:         con***REMOVED***g.APIPath,
		ContentCon***REMOVED***g:   con***REMOVED***g.ContentCon***REMOVED***g,
		Username:        con***REMOVED***g.Username,
		Password:        con***REMOVED***g.Password,
		BearerToken:     con***REMOVED***g.BearerToken,
		BearerTokenFile: con***REMOVED***g.BearerTokenFile,
		Impersonate: ImpersonationCon***REMOVED***g{
			Groups:   con***REMOVED***g.Impersonate.Groups,
			Extra:    con***REMOVED***g.Impersonate.Extra,
			UserName: con***REMOVED***g.Impersonate.UserName,
		},
		AuthProvider:        con***REMOVED***g.AuthProvider,
		AuthCon***REMOVED***gPersister: con***REMOVED***g.AuthCon***REMOVED***gPersister,
		ExecProvider:        con***REMOVED***g.ExecProvider,
		TLSClientCon***REMOVED***g: TLSClientCon***REMOVED***g{
			Insecure:   con***REMOVED***g.TLSClientCon***REMOVED***g.Insecure,
			ServerName: con***REMOVED***g.TLSClientCon***REMOVED***g.ServerName,
			CertFile:   con***REMOVED***g.TLSClientCon***REMOVED***g.CertFile,
			KeyFile:    con***REMOVED***g.TLSClientCon***REMOVED***g.KeyFile,
			CAFile:     con***REMOVED***g.TLSClientCon***REMOVED***g.CAFile,
			CertData:   con***REMOVED***g.TLSClientCon***REMOVED***g.CertData,
			KeyData:    con***REMOVED***g.TLSClientCon***REMOVED***g.KeyData,
			CAData:     con***REMOVED***g.TLSClientCon***REMOVED***g.CAData,
		},
		UserAgent:     con***REMOVED***g.UserAgent,
		Transport:     con***REMOVED***g.Transport,
		WrapTransport: con***REMOVED***g.WrapTransport,
		QPS:           con***REMOVED***g.QPS,
		Burst:         con***REMOVED***g.Burst,
		RateLimiter:   con***REMOVED***g.RateLimiter,
		Timeout:       con***REMOVED***g.Timeout,
		Dial:          con***REMOVED***g.Dial,
	}
}
