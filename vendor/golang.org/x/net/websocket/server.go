// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

package websocket

import (
	"bu***REMOVED***o"
	"fmt"
	"io"
	"net/http"
)

func newServerConn(rwc io.ReadWriteCloser, buf *bu***REMOVED***o.ReadWriter, req *http.Request, con***REMOVED***g *Con***REMOVED***g, handshake func(*Con***REMOVED***g, *http.Request) error) (conn *Conn, err error) {
	var hs serverHandshaker = &hybiServerHandshaker{Con***REMOVED***g: con***REMOVED***g}
	code, err := hs.ReadHandshake(buf.Reader, req)
	if err == ErrBadWebSocketVersion {
		fmt.Fprintf(buf, "HTTP/1.1 %03d %s\r\n", code, http.StatusText(code))
		fmt.Fprintf(buf, "Sec-WebSocket-Version: %s\r\n", SupportedProtocolVersion)
		buf.WriteString("\r\n")
		buf.WriteString(err.Error())
		buf.Flush()
		return
	}
	if err != nil {
		fmt.Fprintf(buf, "HTTP/1.1 %03d %s\r\n", code, http.StatusText(code))
		buf.WriteString("\r\n")
		buf.WriteString(err.Error())
		buf.Flush()
		return
	}
	if handshake != nil {
		err = handshake(con***REMOVED***g, req)
		if err != nil {
			code = http.StatusForbidden
			fmt.Fprintf(buf, "HTTP/1.1 %03d %s\r\n", code, http.StatusText(code))
			buf.WriteString("\r\n")
			buf.Flush()
			return
		}
	}
	err = hs.AcceptHandshake(buf.Writer)
	if err != nil {
		code = http.StatusBadRequest
		fmt.Fprintf(buf, "HTTP/1.1 %03d %s\r\n", code, http.StatusText(code))
		buf.WriteString("\r\n")
		buf.Flush()
		return
	}
	conn = hs.NewServerConn(buf, rwc, req)
	return
}

// Server represents a server of a WebSocket.
type Server struct {
	// Con***REMOVED***g is a WebSocket con***REMOVED***guration for new WebSocket connection.
	Con***REMOVED***g

	// Handshake is an optional function in WebSocket handshake.
	// For example, you can check, or don't check Origin header.
	// Another example, you can select con***REMOVED***g.Protocol.
	Handshake func(*Con***REMOVED***g, *http.Request) error

	// Handler handles a WebSocket connection.
	Handler
}

// ServeHTTP implements the http.Handler interface for a WebSocket
func (s Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.serveWebSocket(w, req)
}

func (s Server) serveWebSocket(w http.ResponseWriter, req *http.Request) {
	rwc, buf, err := w.(http.Hijacker).Hijack()
	if err != nil {
		panic("Hijack failed: " + err.Error())
	}
	// The server should abort the WebSocket connection if it ***REMOVED***nds
	// the client did not send a handshake that matches with protocol
	// speci***REMOVED***cation.
	defer rwc.Close()
	conn, err := newServerConn(rwc, buf, req, &s.Con***REMOVED***g, s.Handshake)
	if err != nil {
		return
	}
	if conn == nil {
		panic("unexpected nil conn")
	}
	s.Handler(conn)
}

// Handler is a simple interface to a WebSocket browser client.
// It checks if Origin header is valid URL by default.
// You might want to verify websocket.Conn.Con***REMOVED***g().Origin in the func.
// If you use Server instead of Handler, you could call websocket.Origin and
// check the origin in your Handshake func. So, if you want to accept
// non-browser clients, which do not send an Origin header, set a
// Server.Handshake that does not check the origin.
type Handler func(*Conn)

func checkOrigin(con***REMOVED***g *Con***REMOVED***g, req *http.Request) (err error) {
	con***REMOVED***g.Origin, err = Origin(con***REMOVED***g, req)
	if err == nil && con***REMOVED***g.Origin == nil {
		return fmt.Errorf("null origin")
	}
	return err
}

// ServeHTTP implements the http.Handler interface for a WebSocket
func (h Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s := Server{Handler: h, Handshake: checkOrigin}
	s.serveWebSocket(w, req)
}
