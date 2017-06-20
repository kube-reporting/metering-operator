// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE ***REMOVED***le.

package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb/***REMOVED***lter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func dupOptions(o *opt.Options) *opt.Options {
	newo := &opt.Options{}
	if o != nil {
		*newo = *o
	}
	if newo.Strict == 0 {
		newo.Strict = opt.DefaultStrict
	}
	return newo
}

func (s *session) setOptions(o *opt.Options) {
	no := dupOptions(o)
	// Alternative ***REMOVED***lters.
	if ***REMOVED***lters := o.GetAltFilters(); len(***REMOVED***lters) > 0 {
		no.AltFilters = make([]***REMOVED***lter.Filter, len(***REMOVED***lters))
		for i, ***REMOVED***lter := range ***REMOVED***lters {
			no.AltFilters[i] = &iFilter{***REMOVED***lter}
		}
	}
	// Comparer.
	s.icmp = &iComparer{o.GetComparer()}
	no.Comparer = s.icmp
	// Filter.
	if ***REMOVED***lter := o.GetFilter(); ***REMOVED***lter != nil {
		no.Filter = &iFilter{***REMOVED***lter}
	}

	s.o = &cachedOptions{Options: no}
	s.o.cache()
}

const optCachedLevel = 7

type cachedOptions struct {
	*opt.Options

	compactionExpandLimit []int
	compactionGPOverlaps  []int
	compactionSourceLimit []int
	compactionTableSize   []int
	compactionTotalSize   []int64
}

func (co *cachedOptions) cache() {
	co.compactionExpandLimit = make([]int, optCachedLevel)
	co.compactionGPOverlaps = make([]int, optCachedLevel)
	co.compactionSourceLimit = make([]int, optCachedLevel)
	co.compactionTableSize = make([]int, optCachedLevel)
	co.compactionTotalSize = make([]int64, optCachedLevel)

	for level := 0; level < optCachedLevel; level++ {
		co.compactionExpandLimit[level] = co.Options.GetCompactionExpandLimit(level)
		co.compactionGPOverlaps[level] = co.Options.GetCompactionGPOverlaps(level)
		co.compactionSourceLimit[level] = co.Options.GetCompactionSourceLimit(level)
		co.compactionTableSize[level] = co.Options.GetCompactionTableSize(level)
		co.compactionTotalSize[level] = co.Options.GetCompactionTotalSize(level)
	}
}

func (co *cachedOptions) GetCompactionExpandLimit(level int) int {
	if level < optCachedLevel {
		return co.compactionExpandLimit[level]
	}
	return co.Options.GetCompactionExpandLimit(level)
}

func (co *cachedOptions) GetCompactionGPOverlaps(level int) int {
	if level < optCachedLevel {
		return co.compactionGPOverlaps[level]
	}
	return co.Options.GetCompactionGPOverlaps(level)
}

func (co *cachedOptions) GetCompactionSourceLimit(level int) int {
	if level < optCachedLevel {
		return co.compactionSourceLimit[level]
	}
	return co.Options.GetCompactionSourceLimit(level)
}

func (co *cachedOptions) GetCompactionTableSize(level int) int {
	if level < optCachedLevel {
		return co.compactionTableSize[level]
	}
	return co.Options.GetCompactionTableSize(level)
}

func (co *cachedOptions) GetCompactionTotalSize(level int) int64 {
	if level < optCachedLevel {
		return co.compactionTotalSize[level]
	}
	return co.Options.GetCompactionTotalSize(level)
}
