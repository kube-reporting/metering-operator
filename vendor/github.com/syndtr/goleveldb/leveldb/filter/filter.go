// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE ***REMOVED***le.

// Package ***REMOVED***lter provides interface and implementation of probabilistic
// data structure.
//
// The ***REMOVED***lter is resposible for creating small ***REMOVED***lter from a set of keys.
// These ***REMOVED***lter will then used to test whether a key is a member of the set.
// In many cases, a ***REMOVED***lter can cut down the number of disk seeks from a
// handful to a single disk seek per DB.Get call.
package ***REMOVED***lter

// Buffer is the interface that wraps basic Alloc, Write and WriteByte methods.
type Buffer interface {
	// Alloc allocs n bytes of slice from the buffer. This also advancing
	// write offset.
	Alloc(n int) []byte

	// Write appends the contents of p to the buffer.
	Write(p []byte) (n int, err error)

	// WriteByte appends the byte c to the buffer.
	WriteByte(c byte) error
}

// Filter is the ***REMOVED***lter.
type Filter interface {
	// Name returns the name of this policy.
	//
	// Note that if the ***REMOVED***lter encoding changes in an incompatible way,
	// the name returned by this method must be changed. Otherwise, old
	// incompatible ***REMOVED***lters may be passed to methods of this type.
	Name() string

	// NewGenerator creates a new ***REMOVED***lter generator.
	NewGenerator() FilterGenerator

	// Contains returns true if the ***REMOVED***lter contains the given key.
	//
	// The ***REMOVED***lter are ***REMOVED***lters generated by the ***REMOVED***lter generator.
	Contains(***REMOVED***lter, key []byte) bool
}

// FilterGenerator is the ***REMOVED***lter generator.
type FilterGenerator interface {
	// Add adds a key to the ***REMOVED***lter generator.
	//
	// The key may become invalid after call to this method end, therefor
	// key must be copied if implementation require keeping key for later
	// use. The key should not modi***REMOVED***ed directly, doing so may cause
	// unde***REMOVED***ned results.
	Add(key []byte)

	// Generate generates ***REMOVED***lters based on keys passed so far. After call
	// to Generate the ***REMOVED***lter generator maybe resetted, depends on implementation.
	Generate(b Buffer)
}
