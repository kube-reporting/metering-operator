/*
Copyright 2014 The Kubernetes Authors.

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
	"crypto/tls"
	"errors"
	"net/http"

	"k8s.io/client-go/plugin/pkg/client/auth/exec"
	"k8s.io/client-go/transport"
)

// TLSCon***REMOVED***gFor returns a tls.Con***REMOVED***g that will provide the transport level security de***REMOVED***ned
// by the provided Con***REMOVED***g. Will return nil if no transport level security is requested.
func TLSCon***REMOVED***gFor(con***REMOVED***g *Con***REMOVED***g) (*tls.Con***REMOVED***g, error) {
	cfg, err := con***REMOVED***g.TransportCon***REMOVED***g()
	if err != nil {
		return nil, err
	}
	return transport.TLSCon***REMOVED***gFor(cfg)
}

// TransportFor returns an http.RoundTripper that will provide the authentication
// or transport level security de***REMOVED***ned by the provided Con***REMOVED***g. Will return the
// default http.DefaultTransport if no special case behavior is needed.
func TransportFor(con***REMOVED***g *Con***REMOVED***g) (http.RoundTripper, error) {
	cfg, err := con***REMOVED***g.TransportCon***REMOVED***g()
	if err != nil {
		return nil, err
	}
	return transport.New(cfg)
}

// HTTPWrappersForCon***REMOVED***g wraps a round tripper with any relevant layered behavior from the
// con***REMOVED***g. Exposed to allow more clients that need HTTP-like behavior but then must hijack
// the underlying connection (like WebSocket or HTTP2 clients). Pure HTTP clients should use
// the higher level TransportFor or RESTClientFor methods.
func HTTPWrappersForCon***REMOVED***g(con***REMOVED***g *Con***REMOVED***g, rt http.RoundTripper) (http.RoundTripper, error) {
	cfg, err := con***REMOVED***g.TransportCon***REMOVED***g()
	if err != nil {
		return nil, err
	}
	return transport.HTTPWrappersForCon***REMOVED***g(cfg, rt)
}

// TransportCon***REMOVED***g converts a client con***REMOVED***g to an appropriate transport con***REMOVED***g.
func (c *Con***REMOVED***g) TransportCon***REMOVED***g() (*transport.Con***REMOVED***g, error) {
	conf := &transport.Con***REMOVED***g{
		UserAgent:     c.UserAgent,
		Transport:     c.Transport,
		WrapTransport: c.WrapTransport,
		TLS: transport.TLSCon***REMOVED***g{
			Insecure:   c.Insecure,
			ServerName: c.ServerName,
			CAFile:     c.CAFile,
			CAData:     c.CAData,
			CertFile:   c.CertFile,
			CertData:   c.CertData,
			KeyFile:    c.KeyFile,
			KeyData:    c.KeyData,
		},
		Username:    c.Username,
		Password:    c.Password,
		BearerToken: c.BearerToken,
		Impersonate: transport.ImpersonationCon***REMOVED***g{
			UserName: c.Impersonate.UserName,
			Groups:   c.Impersonate.Groups,
			Extra:    c.Impersonate.Extra,
		},
		Dial: c.Dial,
	}

	if c.ExecProvider != nil && c.AuthProvider != nil {
		return nil, errors.New("execProvider and authProvider cannot be used in combination")
	}

	if c.ExecProvider != nil {
		provider, err := exec.GetAuthenticator(c.ExecProvider)
		if err != nil {
			return nil, err
		}
		if err := provider.UpdateTransportCon***REMOVED***g(conf); err != nil {
			return nil, err
		}
	}
	if c.AuthProvider != nil {
		provider, err := GetAuthProvider(c.Host, c.AuthProvider, c.AuthCon***REMOVED***gPersister)
		if err != nil {
			return nil, err
		}
		wt := conf.WrapTransport
		if wt != nil {
			conf.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
				return provider.WrapTransport(wt(rt))
			}
		} ***REMOVED*** {
			conf.WrapTransport = provider.WrapTransport
		}
	}
	return conf, nil
}
