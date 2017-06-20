// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE ***REMOVED***le.

package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb/***REMOVED***lter"
)

type iFilter struct {
	***REMOVED***lter.Filter
}

func (f iFilter) Contains(***REMOVED***lter, key []byte) bool {
	return f.Filter.Contains(***REMOVED***lter, internalKey(key).ukey())
}

func (f iFilter) NewGenerator() ***REMOVED***lter.FilterGenerator {
	return iFilterGenerator{f.Filter.NewGenerator()}
}

type iFilterGenerator struct {
	***REMOVED***lter.FilterGenerator
}

func (g iFilterGenerator) Add(key []byte) {
	g.FilterGenerator.Add(internalKey(key).ukey())
}
