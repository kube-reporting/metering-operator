// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

// +build go1.6

package http2

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

func con***REMOVED***gureTransport(t1 *http.Transport) (*Transport, error) {
	connPool := new(clientConnPool)
	t2 := &Transport{
		ConnPool: noDialClientConnPool{connPool},
		t1:       t1,
	}
	connPool.t = t2
	if err := registerHTTPSProtocol(t1, noDialH2RoundTripper{t2}); err != nil {
		return nil, err
	}
	if t1.TLSClientCon***REMOVED***g == nil {
		t1.TLSClientCon***REMOVED***g = new(tls.Con***REMOVED***g)
	}
	if !strSliceContains(t1.TLSClientCon***REMOVED***g.NextProtos, "h2") {
		t1.TLSClientCon***REMOVED***g.NextProtos = append([]string{"h2"}, t1.TLSClientCon***REMOVED***g.NextProtos...)
	}
	if !strSliceContains(t1.TLSClientCon***REMOVED***g.NextProtos, "http/1.1") {
		t1.TLSClientCon***REMOVED***g.NextProtos = append(t1.TLSClientCon***REMOVED***g.NextProtos, "http/1.1")
	}
	upgradeFn := func(authority string, c *tls.Conn) http.RoundTripper {
		addr := authorityAddr("https", authority)
		if used, err := connPool.addConnIfNeeded(addr, t2, c); err != nil {
			go c.Close()
			return erringRoundTripper{err}
		} ***REMOVED*** if !used {
			// Turns out we don't need this c.
			// For example, two goroutines made requests to the same host
			// at the same time, both kicking off TCP dials. (since protocol
			// was unknown)
			go c.Close()
		}
		return t2
	}
	if m := t1.TLSNextProto; len(m) == 0 {
		t1.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{
			"h2": upgradeFn,
		}
	} ***REMOVED*** {
		m["h2"] = upgradeFn
	}
	return t2, nil
}

// registerHTTPSProtocol calls Transport.RegisterProtocol but
// convering panics into errors.
func registerHTTPSProtocol(t *http.Transport, rt http.RoundTripper) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	t.RegisterProtocol("https", rt)
	return nil
}

// noDialH2RoundTripper is a RoundTripper which only tries to complete the request
// if there's already has a cached connection to the host.
type noDialH2RoundTripper struct{ t *Transport }

func (rt noDialH2RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := rt.t.RoundTrip(req)
	if err == ErrNoCachedConn {
		return nil, http.ErrSkipAltProtocol
	}
	return res, err
}
