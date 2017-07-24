// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

package precis

import (
	"unicode/utf8"
)

// TODO: Add contextual character rules from Appendix A of RFC5892.

// A class is a set of characters that match certain derived properties. The
// PRECIS framework de***REMOVED***nes two classes: The Freeform class and the Identi***REMOVED***er
// class. The freeform class should be used for pro***REMOVED***les where expressiveness is
// prioritized over safety such as nicknames or passwords. The identi***REMOVED***er class
// should be used for pro***REMOVED***les where safety is the ***REMOVED***rst priority such as
// addressable network labels and usernames.
type class struct {
	validFrom property
}

// Contains satis***REMOVED***es the runes.Set interface and returns whether the given rune
// is a member of the class.
func (c class) Contains(r rune) bool {
	b := make([]byte, 4)
	n := utf8.EncodeRune(b, r)

	trieval, _ := dpTrie.lookup(b[:n])
	return c.validFrom <= property(trieval)
}

var (
	identi***REMOVED***er = &class{validFrom: pValid}
	freeform   = &class{validFrom: idDisOrFreePVal}
)
