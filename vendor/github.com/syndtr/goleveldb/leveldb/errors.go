// Copyright (c) 2014, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE ***REMOVED***le.

package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb/errors"
)

// Common errors.
var (
	ErrNotFound         = errors.ErrNotFound
	ErrReadOnly         = errors.New("leveldb: read-only mode")
	ErrSnapshotReleased = errors.New("leveldb: snapshot released")
	ErrIterReleased     = errors.New("leveldb: iterator released")
	ErrClosed           = errors.New("leveldb: closed")
)
