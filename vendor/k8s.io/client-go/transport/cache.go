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

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	utilnet "k8s.io/apimachinery/pkg/util/net"
)

// TlsTransportCache caches TLS http.RoundTrippers different con***REMOVED***gurations. The
// same RoundTripper will be returned for con***REMOVED***gs with identical TLS options If
// the con***REMOVED***g has no custom TLS options, http.DefaultTransport is returned.
type tlsTransportCache struct {
	mu         sync.Mutex
	transports map[string]*http.Transport
}

const idleConnsPerHost = 25

var tlsCache = &tlsTransportCache{transports: make(map[string]*http.Transport)}

func (c *tlsTransportCache) get(con***REMOVED***g *Con***REMOVED***g) (http.RoundTripper, error) {
	key, err := tlsCon***REMOVED***gKey(con***REMOVED***g)
	if err != nil {
		return nil, err
	}

	// Ensure we only create a single transport for the given TLS options
	c.mu.Lock()
	defer c.mu.Unlock()

	// See if we already have a custom transport for this con***REMOVED***g
	if t, ok := c.transports[key]; ok {
		return t, nil
	}

	// Get the TLS options for this client con***REMOVED***g
	tlsCon***REMOVED***g, err := TLSCon***REMOVED***gFor(con***REMOVED***g)
	if err != nil {
		return nil, err
	}
	// The options didn't require a custom TLS con***REMOVED***g
	if tlsCon***REMOVED***g == nil {
		return http.DefaultTransport, nil
	}

	// Cache a single transport for these options
	c.transports[key] = utilnet.SetTransportDefaults(&http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientCon***REMOVED***g:     tlsCon***REMOVED***g,
		MaxIdleConnsPerHost: idleConnsPerHost,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
	})
	return c.transports[key], nil
}

// tlsCon***REMOVED***gKey returns a unique key for tls.Con***REMOVED***g objects returned from TLSCon***REMOVED***gFor
func tlsCon***REMOVED***gKey(c *Con***REMOVED***g) (string, error) {
	// Make sure ca/key/cert content is loaded
	if err := loadTLSFiles(c); err != nil {
		return "", err
	}
	// Only include the things that actually affect the tls.Con***REMOVED***g
	return fmt.Sprintf("%v/%x/%x/%x", c.TLS.Insecure, c.TLS.CAData, c.TLS.CertData, c.TLS.KeyData), nil
}
