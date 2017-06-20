// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE ***REMOVED***le.

// Package leveldb provides implementation of LevelDB key/value database.
//
// Create or open a database:
//
//	// The returned DB instance is safe for concurrent use. Which mean that all
//	// DB's methods may be called concurrently from multiple goroutine.
//	db, err := leveldb.OpenFile("path/to/db", nil)
//	...
//	defer db.Close()
//	...
//
// Read or modify the database content:
//
//	// Remember that the contents of the returned slice should not be modi***REMOVED***ed.
//	data, err := db.Get([]byte("key"), nil)
//	...
//	err = db.Put([]byte("key"), []byte("value"), nil)
//	...
//	err = db.Delete([]byte("key"), nil)
//	...
//
// Iterate over database content:
//
//	iter := db.NewIterator(nil, nil)
//	for iter.Next() {
//		// Remember that the contents of the returned slice should not be modi***REMOVED***ed, and
//		// only valid until the next call to Next.
//		key := iter.Key()
//		value := iter.Value()
//		...
//	}
//	iter.Release()
//	err = iter.Error()
//	...
//
// Iterate over subset of database content with a particular pre***REMOVED***x:
//	iter := db.NewIterator(util.BytesPre***REMOVED***x([]byte("foo-")), nil)
//	for iter.Next() {
//		// Use key/value.
//		...
//	}
//	iter.Release()
//	err = iter.Error()
//	...
//
// Seek-then-Iterate:
//
// 	iter := db.NewIterator(nil, nil)
// 	for ok := iter.Seek(key); ok; ok = iter.Next() {
// 		// Use key/value.
// 		...
// 	}
// 	iter.Release()
// 	err = iter.Error()
// 	...
//
// Iterate over subset of database content:
//
// 	iter := db.NewIterator(&util.Range{Start: []byte("foo"), Limit: []byte("xoo")}, nil)
// 	for iter.Next() {
// 		// Use key/value.
// 		...
// 	}
// 	iter.Release()
// 	err = iter.Error()
// 	...
//
// Batch writes:
//
//	batch := new(leveldb.Batch)
//	batch.Put([]byte("foo"), []byte("value"))
//	batch.Put([]byte("bar"), []byte("another value"))
//	batch.Delete([]byte("baz"))
//	err = db.Write(batch, nil)
//	...
//
// Use bloom ***REMOVED***lter:
//
//	o := &opt.Options{
//		Filter: ***REMOVED***lter.NewBloomFilter(10),
//	}
//	db, err := leveldb.OpenFile("path/to/db", o)
//	...
//	defer db.Close()
//	...
package leveldb
