// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

// +build go1.7,!go1.8

package http2

import "crypto/tls"

// temporary copy of Go 1.7's private tls.Con***REMOVED***g.clone:
func cloneTLSCon***REMOVED***g(c *tls.Con***REMOVED***g) *tls.Con***REMOVED***g {
	return &tls.Con***REMOVED***g{
		Rand:                        c.Rand,
		Time:                        c.Time,
		Certi***REMOVED***cates:                c.Certi***REMOVED***cates,
		NameToCerti***REMOVED***cate:           c.NameToCerti***REMOVED***cate,
		GetCerti***REMOVED***cate:              c.GetCerti***REMOVED***cate,
		RootCAs:                     c.RootCAs,
		NextProtos:                  c.NextProtos,
		ServerName:                  c.ServerName,
		ClientAuth:                  c.ClientAuth,
		ClientCAs:                   c.ClientCAs,
		InsecureSkipVerify:          c.InsecureSkipVerify,
		CipherSuites:                c.CipherSuites,
		PreferServerCipherSuites:    c.PreferServerCipherSuites,
		SessionTicketsDisabled:      c.SessionTicketsDisabled,
		SessionTicketKey:            c.SessionTicketKey,
		ClientSessionCache:          c.ClientSessionCache,
		MinVersion:                  c.MinVersion,
		MaxVersion:                  c.MaxVersion,
		CurvePreferences:            c.CurvePreferences,
		DynamicRecordSizingDisabled: c.DynamicRecordSizingDisabled,
		Renegotiation:               c.Renegotiation,
	}
}
