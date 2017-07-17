/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this ***REMOVED***le except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the speci***REMOVED***c language governing permissions and
limitations under the License.
*/

package resource

import (
	"strconv"
)

type suf***REMOVED***x string

// suf***REMOVED***xer can interpret and construct suf***REMOVED***xes.
type suf***REMOVED***xer interface {
	interpret(suf***REMOVED***x) (base, exponent int32, fmt Format, ok bool)
	construct(base, exponent int32, fmt Format) (s suf***REMOVED***x, ok bool)
	constructBytes(base, exponent int32, fmt Format) (s []byte, ok bool)
}

// quantitySuf***REMOVED***xer handles suf***REMOVED***xes for all three formats that quantity
// can handle.
var quantitySuf***REMOVED***xer = newSuf***REMOVED***xer()

type bePair struct {
	base, exponent int32
}

type listSuf***REMOVED***xer struct {
	suf***REMOVED***xToBE      map[suf***REMOVED***x]bePair
	beToSuf***REMOVED***x      map[bePair]suf***REMOVED***x
	beToSuf***REMOVED***xBytes map[bePair][]byte
}

func (ls *listSuf***REMOVED***xer) addSuf***REMOVED***x(s suf***REMOVED***x, pair bePair) {
	if ls.suf***REMOVED***xToBE == nil {
		ls.suf***REMOVED***xToBE = map[suf***REMOVED***x]bePair{}
	}
	if ls.beToSuf***REMOVED***x == nil {
		ls.beToSuf***REMOVED***x = map[bePair]suf***REMOVED***x{}
	}
	if ls.beToSuf***REMOVED***xBytes == nil {
		ls.beToSuf***REMOVED***xBytes = map[bePair][]byte{}
	}
	ls.suf***REMOVED***xToBE[s] = pair
	ls.beToSuf***REMOVED***x[pair] = s
	ls.beToSuf***REMOVED***xBytes[pair] = []byte(s)
}

func (ls *listSuf***REMOVED***xer) lookup(s suf***REMOVED***x) (base, exponent int32, ok bool) {
	pair, ok := ls.suf***REMOVED***xToBE[s]
	if !ok {
		return 0, 0, false
	}
	return pair.base, pair.exponent, true
}

func (ls *listSuf***REMOVED***xer) construct(base, exponent int32) (s suf***REMOVED***x, ok bool) {
	s, ok = ls.beToSuf***REMOVED***x[bePair{base, exponent}]
	return
}

func (ls *listSuf***REMOVED***xer) constructBytes(base, exponent int32) (s []byte, ok bool) {
	s, ok = ls.beToSuf***REMOVED***xBytes[bePair{base, exponent}]
	return
}

type suf***REMOVED***xHandler struct {
	decSuf***REMOVED***xes listSuf***REMOVED***xer
	binSuf***REMOVED***xes listSuf***REMOVED***xer
}

type fastLookup struct {
	*suf***REMOVED***xHandler
}

func (l fastLookup) interpret(s suf***REMOVED***x) (base, exponent int32, format Format, ok bool) {
	switch s {
	case "":
		return 10, 0, DecimalSI, true
	case "n":
		return 10, -9, DecimalSI, true
	case "u":
		return 10, -6, DecimalSI, true
	case "m":
		return 10, -3, DecimalSI, true
	case "k":
		return 10, 3, DecimalSI, true
	case "M":
		return 10, 6, DecimalSI, true
	case "G":
		return 10, 9, DecimalSI, true
	}
	return l.suf***REMOVED***xHandler.interpret(s)
}

func newSuf***REMOVED***xer() suf***REMOVED***xer {
	sh := &suf***REMOVED***xHandler{}

	// IMPORTANT: if you change this section you must change fastLookup

	sh.binSuf***REMOVED***xes.addSuf***REMOVED***x("Ki", bePair{2, 10})
	sh.binSuf***REMOVED***xes.addSuf***REMOVED***x("Mi", bePair{2, 20})
	sh.binSuf***REMOVED***xes.addSuf***REMOVED***x("Gi", bePair{2, 30})
	sh.binSuf***REMOVED***xes.addSuf***REMOVED***x("Ti", bePair{2, 40})
	sh.binSuf***REMOVED***xes.addSuf***REMOVED***x("Pi", bePair{2, 50})
	sh.binSuf***REMOVED***xes.addSuf***REMOVED***x("Ei", bePair{2, 60})
	// Don't emit an error when trying to produce
	// a suf***REMOVED***x for 2^0.
	sh.decSuf***REMOVED***xes.addSuf***REMOVED***x("", bePair{2, 0})

	sh.decSuf***REMOVED***xes.addSuf***REMOVED***x("n", bePair{10, -9})
	sh.decSuf***REMOVED***xes.addSuf***REMOVED***x("u", bePair{10, -6})
	sh.decSuf***REMOVED***xes.addSuf***REMOVED***x("m", bePair{10, -3})
	sh.decSuf***REMOVED***xes.addSuf***REMOVED***x("", bePair{10, 0})
	sh.decSuf***REMOVED***xes.addSuf***REMOVED***x("k", bePair{10, 3})
	sh.decSuf***REMOVED***xes.addSuf***REMOVED***x("M", bePair{10, 6})
	sh.decSuf***REMOVED***xes.addSuf***REMOVED***x("G", bePair{10, 9})
	sh.decSuf***REMOVED***xes.addSuf***REMOVED***x("T", bePair{10, 12})
	sh.decSuf***REMOVED***xes.addSuf***REMOVED***x("P", bePair{10, 15})
	sh.decSuf***REMOVED***xes.addSuf***REMOVED***x("E", bePair{10, 18})

	return fastLookup{sh}
}

func (sh *suf***REMOVED***xHandler) construct(base, exponent int32, fmt Format) (s suf***REMOVED***x, ok bool) {
	switch fmt {
	case DecimalSI:
		return sh.decSuf***REMOVED***xes.construct(base, exponent)
	case BinarySI:
		return sh.binSuf***REMOVED***xes.construct(base, exponent)
	case DecimalExponent:
		if base != 10 {
			return "", false
		}
		if exponent == 0 {
			return "", true
		}
		return suf***REMOVED***x("e" + strconv.FormatInt(int64(exponent), 10)), true
	}
	return "", false
}

func (sh *suf***REMOVED***xHandler) constructBytes(base, exponent int32, format Format) (s []byte, ok bool) {
	switch format {
	case DecimalSI:
		return sh.decSuf***REMOVED***xes.constructBytes(base, exponent)
	case BinarySI:
		return sh.binSuf***REMOVED***xes.constructBytes(base, exponent)
	case DecimalExponent:
		if base != 10 {
			return nil, false
		}
		if exponent == 0 {
			return nil, true
		}
		result := make([]byte, 8, 8)
		result[0] = 'e'
		number := strconv.AppendInt(result[1:1], int64(exponent), 10)
		if &result[1] == &number[0] {
			return result[:1+len(number)], true
		}
		result = append(result[:1], number...)
		return result, true
	}
	return nil, false
}

func (sh *suf***REMOVED***xHandler) interpret(suf***REMOVED***x suf***REMOVED***x) (base, exponent int32, fmt Format, ok bool) {
	// Try lookup tables ***REMOVED***rst
	if b, e, ok := sh.decSuf***REMOVED***xes.lookup(suf***REMOVED***x); ok {
		return b, e, DecimalSI, true
	}
	if b, e, ok := sh.binSuf***REMOVED***xes.lookup(suf***REMOVED***x); ok {
		return b, e, BinarySI, true
	}

	if len(suf***REMOVED***x) > 1 && (suf***REMOVED***x[0] == 'E' || suf***REMOVED***x[0] == 'e') {
		parsed, err := strconv.ParseInt(string(suf***REMOVED***x[1:]), 10, 64)
		if err != nil {
			return 0, 0, DecimalExponent, false
		}
		return 10, int32(parsed), DecimalExponent, true
	}

	return 0, 0, DecimalExponent, false
}
