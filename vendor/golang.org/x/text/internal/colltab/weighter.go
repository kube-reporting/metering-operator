// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

package colltab // import "golang.org/x/text/internal/colltab"

// A Weighter can be used as a source for Collator and Searcher.
type Weighter interface {
	// Start ***REMOVED***nds the start of the segment that includes position p.
	Start(p int, b []byte) int

	// StartString ***REMOVED***nds the start of the segment that includes position p.
	StartString(p int, s string) int

	// AppendNext appends Elems to buf corresponding to the longest match
	// of a single character or contraction from the start of s.
	// It returns the new buf and the number of bytes consumed.
	AppendNext(buf []Elem, s []byte) (ce []Elem, n int)

	// AppendNextString appends Elems to buf corresponding to the longest match
	// of a single character or contraction from the start of s.
	// It returns the new buf and the number of bytes consumed.
	AppendNextString(buf []Elem, s string) (ce []Elem, n int)

	// Domain returns a slice of all single characters and contractions for which
	// collation elements are de***REMOVED***ned in this table.
	Domain() []string

	// Top returns the highest variable primary value.
	Top() uint32
}
