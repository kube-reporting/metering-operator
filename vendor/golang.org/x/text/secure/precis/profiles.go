// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

package precis

import (
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	// Implements the Nickname pro***REMOVED***le speci***REMOVED***ed in RFC 7700.
	// The nickname pro***REMOVED***le is not idempotent and may need to be applied multiple
	// times before being used for comparisons.
	Nickname *Pro***REMOVED***le = nickname

	// Implements the UsernameCaseMapped pro***REMOVED***le speci***REMOVED***ed in RFC 7613.
	UsernameCaseMapped *Pro***REMOVED***le = usernameCaseMap

	// Implements the UsernameCasePreserved pro***REMOVED***le speci***REMOVED***ed in RFC 7613.
	UsernameCasePreserved *Pro***REMOVED***le = usernameNoCaseMap

	// Implements the OpaqueString pro***REMOVED***le de***REMOVED***ned in RFC 7613 for passwords and other secure labels.
	OpaqueString *Pro***REMOVED***le = opaquestring
)

var (
	nickname = &Pro***REMOVED***le{
		options: getOpts(
			AdditionalMapping(func() transform.Transformer {
				return &nickAdditionalMapping{}
			}),
			IgnoreCase,
			Norm(norm.NFKC),
			DisallowEmpty,
		),
		class: freeform,
	}
	usernameCaseMap = &Pro***REMOVED***le{
		options: getOpts(
			FoldWidth,
			LowerCase(),
			Norm(norm.NFC),
			BidiRule,
		),
		class: identi***REMOVED***er,
	}
	usernameNoCaseMap = &Pro***REMOVED***le{
		options: getOpts(
			FoldWidth,
			Norm(norm.NFC),
			BidiRule,
		),
		class: identi***REMOVED***er,
	}
	opaquestring = &Pro***REMOVED***le{
		options: getOpts(
			AdditionalMapping(func() transform.Transformer {
				return mapSpaces
			}),
			Norm(norm.NFC),
			DisallowEmpty,
		),
		class: freeform,
	}
)

// mapSpaces is a shared value of a runes.Map transformer.
var mapSpaces transform.Transformer = runes.Map(func(r rune) rune {
	if unicode.Is(unicode.Zs, r) {
		return ' '
	}
	return r
})
