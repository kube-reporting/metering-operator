// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

package yaml

import (
	"bytes"
	"encoding"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// indirect walks down v allocating pointers as needed,
// until it gets to a non-pointer.
// if it encounters an Unmarshaler, indirect stops and returns that.
// if decodingNull is true, indirect stops at the last pointer so it can be set to nil.
func indirect(v reflect.Value, decodingNull bool) (json.Unmarshaler, encoding.TextUnmarshaler, reflect.Value) {
	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we ***REMOVED***nd them.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable.
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() && (!decodingNull || e.Elem().Kind() == reflect.Ptr) {
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		if v.Elem().Kind() != reflect.Ptr && decodingNull && v.CanSet() {
			break
		}
		if v.IsNil() {
			if v.CanSet() {
				v.Set(reflect.New(v.Type().Elem()))
			} ***REMOVED*** {
				v = reflect.New(v.Type().Elem())
			}
		}
		if v.Type().NumMethod() > 0 {
			if u, ok := v.Interface().(json.Unmarshaler); ok {
				return u, nil, reflect.Value{}
			}
			if u, ok := v.Interface().(encoding.TextUnmarshaler); ok {
				return nil, u, reflect.Value{}
			}
		}
		v = v.Elem()
	}
	return nil, nil, v
}

// A ***REMOVED***eld represents a single ***REMOVED***eld found in a struct.
type ***REMOVED***eld struct {
	name      string
	nameBytes []byte                 // []byte(name)
	equalFold func(s, t []byte) bool // bytes.EqualFold or equivalent

	tag       bool
	index     []int
	typ       reflect.Type
	omitEmpty bool
	quoted    bool
}

func ***REMOVED***llField(f ***REMOVED***eld) ***REMOVED***eld {
	f.nameBytes = []byte(f.name)
	f.equalFold = foldFunc(f.nameBytes)
	return f
}

// byName sorts ***REMOVED***eld by name, breaking ties with depth,
// then breaking ties with "name came from json tag", then
// breaking ties with index sequence.
type byName []***REMOVED***eld

func (x byName) Len() int { return len(x) }

func (x byName) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x byName) Less(i, j int) bool {
	if x[i].name != x[j].name {
		return x[i].name < x[j].name
	}
	if len(x[i].index) != len(x[j].index) {
		return len(x[i].index) < len(x[j].index)
	}
	if x[i].tag != x[j].tag {
		return x[i].tag
	}
	return byIndex(x).Less(i, j)
}

// byIndex sorts ***REMOVED***eld by index sequence.
type byIndex []***REMOVED***eld

func (x byIndex) Len() int { return len(x) }

func (x byIndex) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x byIndex) Less(i, j int) bool {
	for k, xik := range x[i].index {
		if k >= len(x[j].index) {
			return false
		}
		if xik != x[j].index[k] {
			return xik < x[j].index[k]
		}
	}
	return len(x[i].index) < len(x[j].index)
}

// typeFields returns a list of ***REMOVED***elds that JSON should recognize for the given type.
// The algorithm is breadth-***REMOVED***rst search over the set of structs to include - the top struct
// and then any reachable anonymous structs.
func typeFields(t reflect.Type) []***REMOVED***eld {
	// Anonymous ***REMOVED***elds to explore at the current level and the next.
	current := []***REMOVED***eld{}
	next := []***REMOVED***eld{{typ: t}}

	// Count of queued names for current level and the next.
	count := map[reflect.Type]int{}
	nextCount := map[reflect.Type]int{}

	// Types already visited at an earlier level.
	visited := map[reflect.Type]bool{}

	// Fields found.
	var ***REMOVED***elds []***REMOVED***eld

	for len(next) > 0 {
		current, next = next, current[:0]
		count, nextCount = nextCount, map[reflect.Type]int{}

		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true

			// Scan f.typ for ***REMOVED***elds to include.
			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)
				if sf.PkgPath != "" { // unexported
					continue
				}
				tag := sf.Tag.Get("json")
				if tag == "-" {
					continue
				}
				name, opts := parseTag(tag)
				if !isValidTag(name) {
					name = ""
				}
				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i

				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
					// Follow pointer.
					ft = ft.Elem()
				}

				// Record found ***REMOVED***eld and index sequence.
				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = sf.Name
					}
					***REMOVED***elds = append(***REMOVED***elds, ***REMOVED***llField(***REMOVED***eld{
						name:      name,
						tag:       tagged,
						index:     index,
						typ:       ft,
						omitEmpty: opts.Contains("omitempty"),
						quoted:    opts.Contains("string"),
					}))
					if count[f.typ] > 1 {
						// If there were multiple instances, add a second,
						// so that the annihilation code will see a duplicate.
						// It only cares about the distinction between 1 or 2,
						// so don't bother generating any more copies.
						***REMOVED***elds = append(***REMOVED***elds, ***REMOVED***elds[len(***REMOVED***elds)-1])
					}
					continue
				}

				// Record new anonymous struct to explore in next round.
				nextCount[ft]++
				if nextCount[ft] == 1 {
					next = append(next, ***REMOVED***llField(***REMOVED***eld{name: ft.Name(), index: index, typ: ft}))
				}
			}
		}
	}

	sort.Sort(byName(***REMOVED***elds))

	// Delete all ***REMOVED***elds that are hidden by the Go rules for embedded ***REMOVED***elds,
	// except that ***REMOVED***elds with JSON tags are promoted.

	// The ***REMOVED***elds are sorted in primary order of name, secondary order
	// of ***REMOVED***eld index length. Loop over names; for each name, delete
	// hidden ***REMOVED***elds by choosing the one dominant ***REMOVED***eld that survives.
	out := ***REMOVED***elds[:0]
	for advance, i := 0, 0; i < len(***REMOVED***elds); i += advance {
		// One iteration per name.
		// Find the sequence of ***REMOVED***elds with the name of this ***REMOVED***rst ***REMOVED***eld.
		***REMOVED*** := ***REMOVED***elds[i]
		name := ***REMOVED***.name
		for advance = 1; i+advance < len(***REMOVED***elds); advance++ {
			fj := ***REMOVED***elds[i+advance]
			if fj.name != name {
				break
			}
		}
		if advance == 1 { // Only one ***REMOVED***eld with this name
			out = append(out, ***REMOVED***)
			continue
		}
		dominant, ok := dominantField(***REMOVED***elds[i : i+advance])
		if ok {
			out = append(out, dominant)
		}
	}

	***REMOVED***elds = out
	sort.Sort(byIndex(***REMOVED***elds))

	return ***REMOVED***elds
}

// dominantField looks through the ***REMOVED***elds, all of which are known to
// have the same name, to ***REMOVED***nd the single ***REMOVED***eld that dominates the
// others using Go's embedding rules, modi***REMOVED***ed by the presence of
// JSON tags. If there are multiple top-level ***REMOVED***elds, the boolean
// will be false: This condition is an error in Go and we skip all
// the ***REMOVED***elds.
func dominantField(***REMOVED***elds []***REMOVED***eld) (***REMOVED***eld, bool) {
	// The ***REMOVED***elds are sorted in increasing index-length order. The winner
	// must therefore be one with the shortest index length. Drop all
	// longer entries, which is easy: just truncate the slice.
	length := len(***REMOVED***elds[0].index)
	tagged := -1 // Index of ***REMOVED***rst tagged ***REMOVED***eld.
	for i, f := range ***REMOVED***elds {
		if len(f.index) > length {
			***REMOVED***elds = ***REMOVED***elds[:i]
			break
		}
		if f.tag {
			if tagged >= 0 {
				// Multiple tagged ***REMOVED***elds at the same level: conflict.
				// Return no ***REMOVED***eld.
				return ***REMOVED***eld{}, false
			}
			tagged = i
		}
	}
	if tagged >= 0 {
		return ***REMOVED***elds[tagged], true
	}
	// All remaining ***REMOVED***elds have the same length. If there's more than one,
	// we have a conflict (two ***REMOVED***elds named "X" at the same level) and we
	// return no ***REMOVED***eld.
	if len(***REMOVED***elds) > 1 {
		return ***REMOVED***eld{}, false
	}
	return ***REMOVED***elds[0], true
}

var ***REMOVED***eldCache struct {
	sync.RWMutex
	m map[reflect.Type][]***REMOVED***eld
}

// cachedTypeFields is like typeFields but uses a cache to avoid repeated work.
func cachedTypeFields(t reflect.Type) []***REMOVED***eld {
	***REMOVED***eldCache.RLock()
	f := ***REMOVED***eldCache.m[t]
	***REMOVED***eldCache.RUnlock()
	if f != nil {
		return f
	}

	// Compute ***REMOVED***elds without lock.
	// Might duplicate effort but won't hold other computations back.
	f = typeFields(t)
	if f == nil {
		f = []***REMOVED***eld{}
	}

	***REMOVED***eldCache.Lock()
	if ***REMOVED***eldCache.m == nil {
		***REMOVED***eldCache.m = map[reflect.Type][]***REMOVED***eld{}
	}
	***REMOVED***eldCache.m[t] = f
	***REMOVED***eldCache.Unlock()
	return f
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}

const (
	caseMask     = ^byte(0x20) // Mask to ignore case in ASCII.
	kelvin       = '\u212a'
	smallLongEss = '\u017f'
)

// foldFunc returns one of four different case folding equivalence
// functions, from most general (and slow) to fastest:
//
// 1) bytes.EqualFold, if the key s contains any non-ASCII UTF-8
// 2) equalFoldRight, if s contains special folding ASCII ('k', 'K', 's', 'S')
// 3) asciiEqualFold, no special, but includes non-letters (including _)
// 4) simpleLetterEqualFold, no specials, no non-letters.
//
// The letters S and K are special because they map to 3 runes, not just 2:
//  * S maps to s and to U+017F 'ſ' Latin small letter long s
//  * k maps to K and to U+212A 'K' Kelvin sign
// See http://play.golang.org/p/tTxjOc0OGo
//
// The returned function is specialized for matching against s and
// should only be given s. It's not curried for performance reasons.
func foldFunc(s []byte) func(s, t []byte) bool {
	nonLetter := false
	special := false // special letter
	for _, b := range s {
		if b >= utf8.RuneSelf {
			return bytes.EqualFold
		}
		upper := b & caseMask
		if upper < 'A' || upper > 'Z' {
			nonLetter = true
		} ***REMOVED*** if upper == 'K' || upper == 'S' {
			// See above for why these letters are special.
			special = true
		}
	}
	if special {
		return equalFoldRight
	}
	if nonLetter {
		return asciiEqualFold
	}
	return simpleLetterEqualFold
}

// equalFoldRight is a specialization of bytes.EqualFold when s is
// known to be all ASCII (including punctuation), but contains an 's',
// 'S', 'k', or 'K', requiring a Unicode fold on the bytes in t.
// See comments on foldFunc.
func equalFoldRight(s, t []byte) bool {
	for _, sb := range s {
		if len(t) == 0 {
			return false
		}
		tb := t[0]
		if tb < utf8.RuneSelf {
			if sb != tb {
				sbUpper := sb & caseMask
				if 'A' <= sbUpper && sbUpper <= 'Z' {
					if sbUpper != tb&caseMask {
						return false
					}
				} ***REMOVED*** {
					return false
				}
			}
			t = t[1:]
			continue
		}
		// sb is ASCII and t is not. t must be either kelvin
		// sign or long s; sb must be s, S, k, or K.
		tr, size := utf8.DecodeRune(t)
		switch sb {
		case 's', 'S':
			if tr != smallLongEss {
				return false
			}
		case 'k', 'K':
			if tr != kelvin {
				return false
			}
		default:
			return false
		}
		t = t[size:]

	}
	if len(t) > 0 {
		return false
	}
	return true
}

// asciiEqualFold is a specialization of bytes.EqualFold for use when
// s is all ASCII (but may contain non-letters) and contains no
// special-folding letters.
// See comments on foldFunc.
func asciiEqualFold(s, t []byte) bool {
	if len(s) != len(t) {
		return false
	}
	for i, sb := range s {
		tb := t[i]
		if sb == tb {
			continue
		}
		if ('a' <= sb && sb <= 'z') || ('A' <= sb && sb <= 'Z') {
			if sb&caseMask != tb&caseMask {
				return false
			}
		} ***REMOVED*** {
			return false
		}
	}
	return true
}

// simpleLetterEqualFold is a specialization of bytes.EqualFold for
// use when s is all ASCII letters (no underscores, etc) and also
// doesn't contain 'k', 'K', 's', or 'S'.
// See comments on foldFunc.
func simpleLetterEqualFold(s, t []byte) bool {
	if len(s) != len(t) {
		return false
	}
	for i, b := range s {
		if b&caseMask != t[i]&caseMask {
			return false
		}
	}
	return true
}

// tagOptions is the string following a comma in a struct ***REMOVED***eld's "json"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

// parseTag splits a struct ***REMOVED***eld's json tag into its name and
// comma-separated options.
func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, tagOptions("")
}

// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}
