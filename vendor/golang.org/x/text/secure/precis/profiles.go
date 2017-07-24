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
	Nickname              *Pro***REMOVED***le = nickname          // Implements the Nickname pro***REMOVED***le speci***REMOVED***ed in RFC 7700.
	UsernameCaseMapped    *Pro***REMOVED***le = usernameCaseMap   // Implements the UsernameCaseMapped pro***REMOVED***le speci***REMOVED***ed in RFC 7613.
	UsernameCasePreserved *Pro***REMOVED***le = usernameNoCaseMap // Implements the UsernameCasePreserved pro***REMOVED***le speci***REMOVED***ed in RFC 7613.
	OpaqueString          *Pro***REMOVED***le = opaquestring      // Implements the OpaqueString pro***REMOVED***le de***REMOVED***ned in RFC 7613 for passwords and other secure labels.
)

// TODO: mvl: "Ultimately, I would manually de***REMOVED***ne the structs for the internal
// pro***REMOVED***les. This avoid pulling in unneeded tables when they are not used."
var (
	nickname = NewFreeform(
		AdditionalMapping(func() transform.Transformer {
			return &nickAdditionalMapping{}
		}),
		IgnoreCase,
		Norm(norm.NFKC),
		DisallowEmpty,
	)
	usernameCaseMap = NewIdenti***REMOVED***er(
		FoldWidth,
		FoldCase(),
		Norm(norm.NFC),
		BidiRule,
	)
	usernameNoCaseMap = NewIdenti***REMOVED***er(
		FoldWidth,
		Norm(norm.NFC),
		BidiRule,
	)
	opaquestring = NewFreeform(
		AdditionalMapping(func() transform.Transformer {
			return runes.Map(func(r rune) rune {
				if unicode.Is(unicode.Zs, r) {
					return ' '
				}
				return r
			})
		}),
		Norm(norm.NFC),
		DisallowEmpty,
	)
)
