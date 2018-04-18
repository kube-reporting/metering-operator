// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

package colltab

import (
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// Table holds all collation data for a given collation ordering.
type Table struct {
	Index Trie // main trie

	// expansion info
	ExpandElem []uint32

	// contraction info
	ContractTries  ContractTrieSet
	ContractElem   []uint32
	MaxContractLen int
	VariableTop    uint32
}

func (t *Table) AppendNext(w []Elem, b []byte) (res []Elem, n int) {
	return t.appendNext(w, source{bytes: b})
}

func (t *Table) AppendNextString(w []Elem, s string) (res []Elem, n int) {
	return t.appendNext(w, source{str: s})
}

func (t *Table) Start(p int, b []byte) int {
	// TODO: implement
	panic("not implemented")
}

func (t *Table) StartString(p int, s string) int {
	// TODO: implement
	panic("not implemented")
}

func (t *Table) Domain() []string {
	// TODO: implement
	panic("not implemented")
}

func (t *Table) Top() uint32 {
	return t.VariableTop
}

type source struct {
	str   string
	bytes []byte
}

func (src *source) lookup(t *Table) (ce Elem, sz int) {
	if src.bytes == nil {
		return t.Index.lookupString(src.str)
	}
	return t.Index.lookup(src.bytes)
}

func (src *source) tail(sz int) {
	if src.bytes == nil {
		src.str = src.str[sz:]
	} ***REMOVED*** {
		src.bytes = src.bytes[sz:]
	}
}

func (src *source) nfd(buf []byte, end int) []byte {
	if src.bytes == nil {
		return norm.NFD.AppendString(buf[:0], src.str[:end])
	}
	return norm.NFD.Append(buf[:0], src.bytes[:end]...)
}

func (src *source) rune() (r rune, sz int) {
	if src.bytes == nil {
		return utf8.DecodeRuneInString(src.str)
	}
	return utf8.DecodeRune(src.bytes)
}

func (src *source) properties(f norm.Form) norm.Properties {
	if src.bytes == nil {
		return f.PropertiesString(src.str)
	}
	return f.Properties(src.bytes)
}

// appendNext appends the weights corresponding to the next rune or
// contraction in s.  If a contraction is matched to a discontinuous
// sequence of runes, the weights for the interstitial runes are
// appended as well.  It returns a new slice that includes the appended
// weights and the number of bytes consumed from s.
func (t *Table) appendNext(w []Elem, src source) (res []Elem, n int) {
	ce, sz := src.lookup(t)
	tp := ce.ctype()
	if tp == ceNormal {
		if ce == 0 {
			r, _ := src.rune()
			const (
				hangulSize  = 3
				***REMOVED***rstHangul = 0xAC00
				lastHangul  = 0xD7A3
			)
			if r >= ***REMOVED***rstHangul && r <= lastHangul {
				// TODO: performance can be considerably improved here.
				n = sz
				var buf [16]byte // Used for decomposing Hangul.
				for b := src.nfd(buf[:0], hangulSize); len(b) > 0; b = b[sz:] {
					ce, sz = t.Index.lookup(b)
					w = append(w, ce)
				}
				return w, n
			}
			ce = makeImplicitCE(implicitPrimary(r))
		}
		w = append(w, ce)
	} ***REMOVED*** if tp == ceExpansionIndex {
		w = t.appendExpansion(w, ce)
	} ***REMOVED*** if tp == ceContractionIndex {
		n := 0
		src.tail(sz)
		if src.bytes == nil {
			w, n = t.matchContractionString(w, ce, src.str)
		} ***REMOVED*** {
			w, n = t.matchContraction(w, ce, src.bytes)
		}
		sz += n
	} ***REMOVED*** if tp == ceDecompose {
		// Decompose using NFKD and replace tertiary weights.
		t1, t2 := splitDecompose(ce)
		i := len(w)
		nfkd := src.properties(norm.NFKD).Decomposition()
		for p := 0; len(nfkd) > 0; nfkd = nfkd[p:] {
			w, p = t.appendNext(w, source{bytes: nfkd})
		}
		w[i] = w[i].updateTertiary(t1)
		if i++; i < len(w) {
			w[i] = w[i].updateTertiary(t2)
			for i++; i < len(w); i++ {
				w[i] = w[i].updateTertiary(maxTertiary)
			}
		}
	}
	return w, sz
}

func (t *Table) appendExpansion(w []Elem, ce Elem) []Elem {
	i := splitExpandIndex(ce)
	n := int(t.ExpandElem[i])
	i++
	for _, ce := range t.ExpandElem[i : i+n] {
		w = append(w, Elem(ce))
	}
	return w
}

func (t *Table) matchContraction(w []Elem, ce Elem, suf***REMOVED***x []byte) ([]Elem, int) {
	index, n, offset := splitContractIndex(ce)

	scan := t.ContractTries.scanner(index, n, suf***REMOVED***x)
	buf := [norm.MaxSegmentSize]byte{}
	bufp := 0
	p := scan.scan(0)

	if !scan.done && p < len(suf***REMOVED***x) && suf***REMOVED***x[p] >= utf8.RuneSelf {
		// By now we should have ***REMOVED***ltered most cases.
		p0 := p
		bufn := 0
		rune := norm.NFD.Properties(suf***REMOVED***x[p:])
		p += rune.Size()
		if rune.LeadCCC() != 0 {
			prevCC := rune.TrailCCC()
			// A gap may only occur in the last normalization segment.
			// This also ensures that len(scan.s) < norm.MaxSegmentSize.
			if end := norm.NFD.FirstBoundary(suf***REMOVED***x[p:]); end != -1 {
				scan.s = suf***REMOVED***x[:p+end]
			}
			for p < len(suf***REMOVED***x) && !scan.done && suf***REMOVED***x[p] >= utf8.RuneSelf {
				rune = norm.NFD.Properties(suf***REMOVED***x[p:])
				if ccc := rune.LeadCCC(); ccc == 0 || prevCC >= ccc {
					break
				}
				prevCC = rune.TrailCCC()
				if pp := scan.scan(p); pp != p {
					// Copy the interstitial runes for later processing.
					bufn += copy(buf[bufn:], suf***REMOVED***x[p0:p])
					if scan.pindex == pp {
						bufp = bufn
					}
					p, p0 = pp, pp
				} ***REMOVED*** {
					p += rune.Size()
				}
			}
		}
	}
	// Append weights for the matched contraction, which may be an expansion.
	i, n := scan.result()
	ce = Elem(t.ContractElem[i+offset])
	if ce.ctype() == ceNormal {
		w = append(w, ce)
	} ***REMOVED*** {
		w = t.appendExpansion(w, ce)
	}
	// Append weights for the runes in the segment not part of the contraction.
	for b, p := buf[:bufp], 0; len(b) > 0; b = b[p:] {
		w, p = t.appendNext(w, source{bytes: b})
	}
	return w, n
}

// TODO: unify the two implementations. This is best done after ***REMOVED***rst simplifying
// the algorithm taking into account the inclusion of both NFC and NFD forms
// in the table.
func (t *Table) matchContractionString(w []Elem, ce Elem, suf***REMOVED***x string) ([]Elem, int) {
	index, n, offset := splitContractIndex(ce)

	scan := t.ContractTries.scannerString(index, n, suf***REMOVED***x)
	buf := [norm.MaxSegmentSize]byte{}
	bufp := 0
	p := scan.scan(0)

	if !scan.done && p < len(suf***REMOVED***x) && suf***REMOVED***x[p] >= utf8.RuneSelf {
		// By now we should have ***REMOVED***ltered most cases.
		p0 := p
		bufn := 0
		rune := norm.NFD.PropertiesString(suf***REMOVED***x[p:])
		p += rune.Size()
		if rune.LeadCCC() != 0 {
			prevCC := rune.TrailCCC()
			// A gap may only occur in the last normalization segment.
			// This also ensures that len(scan.s) < norm.MaxSegmentSize.
			if end := norm.NFD.FirstBoundaryInString(suf***REMOVED***x[p:]); end != -1 {
				scan.s = suf***REMOVED***x[:p+end]
			}
			for p < len(suf***REMOVED***x) && !scan.done && suf***REMOVED***x[p] >= utf8.RuneSelf {
				rune = norm.NFD.PropertiesString(suf***REMOVED***x[p:])
				if ccc := rune.LeadCCC(); ccc == 0 || prevCC >= ccc {
					break
				}
				prevCC = rune.TrailCCC()
				if pp := scan.scan(p); pp != p {
					// Copy the interstitial runes for later processing.
					bufn += copy(buf[bufn:], suf***REMOVED***x[p0:p])
					if scan.pindex == pp {
						bufp = bufn
					}
					p, p0 = pp, pp
				} ***REMOVED*** {
					p += rune.Size()
				}
			}
		}
	}
	// Append weights for the matched contraction, which may be an expansion.
	i, n := scan.result()
	ce = Elem(t.ContractElem[i+offset])
	if ce.ctype() == ceNormal {
		w = append(w, ce)
	} ***REMOVED*** {
		w = t.appendExpansion(w, ce)
	}
	// Append weights for the runes in the segment not part of the contraction.
	for b, p := buf[:bufp], 0; len(b) > 0; b = b[p:] {
		w, p = t.appendNext(w, source{bytes: b})
	}
	return w, n
}
