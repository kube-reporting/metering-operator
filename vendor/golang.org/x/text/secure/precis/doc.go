// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

// Package precis contains types and functions for the preparation,
// enforcement, and comparison of internationalized strings ("PRECIS") as
// de***REMOVED***ned in RFC 7564. It also contains several pre-de***REMOVED***ned pro***REMOVED***les for
// passwords, nicknames, and usernames as de***REMOVED***ned in RFC 7613 and RFC 7700.
//
// BE ADVISED: This package is under construction and the API may change in
// backwards incompatible ways and without notice.
package precis // import "golang.org/x/text/secure/precis"

//go:generate go run gen.go gen_trieval.go
