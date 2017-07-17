// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

package websocket

import (
	"crypto/tls"
	"net"
)

func dialWithDialer(dialer *net.Dialer, con***REMOVED***g *Con***REMOVED***g) (conn net.Conn, err error) {
	switch con***REMOVED***g.Location.Scheme {
	case "ws":
		conn, err = dialer.Dial("tcp", parseAuthority(con***REMOVED***g.Location))

	case "wss":
		conn, err = tls.DialWithDialer(dialer, "tcp", parseAuthority(con***REMOVED***g.Location), con***REMOVED***g.TlsCon***REMOVED***g)

	default:
		err = ErrBadScheme
	}
	return
}
