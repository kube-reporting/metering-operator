// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

package websocket

import (
	"bu***REMOVED***o"
	"io"
	"net"
	"net/http"
	"net/url"
)

// DialError is an error that occurs while dialling a websocket server.
type DialError struct {
	*Con***REMOVED***g
	Err error
}

func (e *DialError) Error() string {
	return "websocket.Dial " + e.Con***REMOVED***g.Location.String() + ": " + e.Err.Error()
}

// NewCon***REMOVED***g creates a new WebSocket con***REMOVED***g for client connection.
func NewCon***REMOVED***g(server, origin string) (con***REMOVED***g *Con***REMOVED***g, err error) {
	con***REMOVED***g = new(Con***REMOVED***g)
	con***REMOVED***g.Version = ProtocolVersionHybi13
	con***REMOVED***g.Location, err = url.ParseRequestURI(server)
	if err != nil {
		return
	}
	con***REMOVED***g.Origin, err = url.ParseRequestURI(origin)
	if err != nil {
		return
	}
	con***REMOVED***g.Header = http.Header(make(map[string][]string))
	return
}

// NewClient creates a new WebSocket client connection over rwc.
func NewClient(con***REMOVED***g *Con***REMOVED***g, rwc io.ReadWriteCloser) (ws *Conn, err error) {
	br := bu***REMOVED***o.NewReader(rwc)
	bw := bu***REMOVED***o.NewWriter(rwc)
	err = hybiClientHandshake(con***REMOVED***g, br, bw)
	if err != nil {
		return
	}
	buf := bu***REMOVED***o.NewReadWriter(br, bw)
	ws = newHybiClientConn(con***REMOVED***g, buf, rwc)
	return
}

// Dial opens a new client connection to a WebSocket.
func Dial(url_, protocol, origin string) (ws *Conn, err error) {
	con***REMOVED***g, err := NewCon***REMOVED***g(url_, origin)
	if err != nil {
		return nil, err
	}
	if protocol != "" {
		con***REMOVED***g.Protocol = []string{protocol}
	}
	return DialCon***REMOVED***g(con***REMOVED***g)
}

var portMap = map[string]string{
	"ws":  "80",
	"wss": "443",
}

func parseAuthority(location *url.URL) string {
	if _, ok := portMap[location.Scheme]; ok {
		if _, _, err := net.SplitHostPort(location.Host); err != nil {
			return net.JoinHostPort(location.Host, portMap[location.Scheme])
		}
	}
	return location.Host
}

// DialCon***REMOVED***g opens a new client connection to a WebSocket with a con***REMOVED***g.
func DialCon***REMOVED***g(con***REMOVED***g *Con***REMOVED***g) (ws *Conn, err error) {
	var client net.Conn
	if con***REMOVED***g.Location == nil {
		return nil, &DialError{con***REMOVED***g, ErrBadWebSocketLocation}
	}
	if con***REMOVED***g.Origin == nil {
		return nil, &DialError{con***REMOVED***g, ErrBadWebSocketOrigin}
	}
	dialer := con***REMOVED***g.Dialer
	if dialer == nil {
		dialer = &net.Dialer{}
	}
	client, err = dialWithDialer(dialer, con***REMOVED***g)
	if err != nil {
		goto Error
	}
	ws, err = NewClient(con***REMOVED***g, client)
	if err != nil {
		client.Close()
		goto Error
	}
	return

Error:
	return nil, &DialError{con***REMOVED***g, err}
}
