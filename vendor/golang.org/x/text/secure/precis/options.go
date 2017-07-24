// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

package precis

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/width"
)

// An Option is used to de***REMOVED***ne the behavior and rules of a Pro***REMOVED***le.
type Option func(*options)

type options struct {
	// Preparation options
	foldWidth bool

	// Enforcement options
	cases         transform.Transformer
	disallow      runes.Set
	norm          norm.Form
	additional    []func() transform.Transformer
	width         *width.Transformer
	disallowEmpty bool
	bidiRule      bool

	// Comparison options
	ignorecase bool
}

func getOpts(o ...Option) (res options) {
	for _, f := range o {
		f(&res)
	}
	return
}

var (
	// The IgnoreCase option causes the pro***REMOVED***le to perform a case insensitive
	// comparison during the PRECIS comparison step.
	IgnoreCase Option = ignoreCase

	// The FoldWidth option causes the pro***REMOVED***le to map non-canonical wide and
	// narrow variants to their decomposition mapping. This is useful for
	// pro***REMOVED***les that are based on the identi***REMOVED***er class which would otherwise
	// disallow such characters.
	FoldWidth Option = foldWidth

	// The DisallowEmpty option causes the enforcement step to return an error if
	// the resulting string would be empty.
	DisallowEmpty Option = disallowEmpty

	// The BidiRule option causes the Bidi Rule de***REMOVED***ned in RFC 5893 to be
	// applied.
	BidiRule Option = bidiRule
)

var (
	ignoreCase = func(o *options) {
		o.ignorecase = true
	}
	foldWidth = func(o *options) {
		o.foldWidth = true
	}
	disallowEmpty = func(o *options) {
		o.disallowEmpty = true
	}
	bidiRule = func(o *options) {
		o.bidiRule = true
	}
)

// The AdditionalMapping option de***REMOVED***nes the additional mapping rule for the
// Pro***REMOVED***le by applying Transformer's in sequence.
func AdditionalMapping(t ...func() transform.Transformer) Option {
	return func(o *options) {
		o.additional = t
	}
}

// The Norm option de***REMOVED***nes a Pro***REMOVED***le's normalization rule. Defaults to NFC.
func Norm(f norm.Form) Option {
	return func(o *options) {
		o.norm = f
	}
}

// The FoldCase option de***REMOVED***nes a Pro***REMOVED***le's case mapping rule. Options can be
// provided to determine the type of case folding used.
func FoldCase(opts ...cases.Option) Option {
	return func(o *options) {
		o.cases = cases.Fold(opts...)
	}
}

// The Disallow option further restricts a Pro***REMOVED***le's allowed characters beyond
// what is disallowed by the underlying string class.
func Disallow(set runes.Set) Option {
	return func(o *options) {
		o.disallow = set
	}
}
