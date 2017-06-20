// Copyright (c) 2014, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE ***REMOVED***le.

package util

// Range is a key range.
type Range struct {
	// Start of the key range, include in the range.
	Start []byte

	// Limit of the key range, not include in the range.
	Limit []byte
}

// BytesPre***REMOVED***x returns key range that satisfy the given pre***REMOVED***x.
// This only applicable for the standard 'bytes comparer'.
func BytesPre***REMOVED***x(pre***REMOVED***x []byte) *Range {
	var limit []byte
	for i := len(pre***REMOVED***x) - 1; i >= 0; i-- {
		c := pre***REMOVED***x[i]
		if c < 0xff {
			limit = make([]byte, i+1)
			copy(limit, pre***REMOVED***x)
			limit[i] = c + 1
			break
		}
	}
	return &Range{pre***REMOVED***x, limit}
}
