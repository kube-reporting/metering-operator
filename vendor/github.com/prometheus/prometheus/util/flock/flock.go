// Copyright 2016 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this ***REMOVED***le except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the speci***REMOVED***c language governing permissions and
// limitations under the License.

// Package flock provides portable ***REMOVED***le locking. It is essentially ripped out
// from the code of github.com/syndtr/goleveldb. Strange enough that the
// standard library does not provide this functionality. Once this package has
// proven to work as expected, we should probably turn it into a separate
// general purpose package for humanity.
package flock

import (
	"os"
	"path/***REMOVED***lepath"
)

// Releaser provides the Release method to release a ***REMOVED***le lock.
type Releaser interface {
	Release() error
}

// New locks the ***REMOVED***le with the provided name. If the ***REMOVED***le does not exist, it is
// created. The returned Releaser is used to release the lock. existed is true
// if the ***REMOVED***le to lock already existed. A non-nil error is returned if the
// locking has failed. Neither this function nor the returned Releaser is
// goroutine-safe.
func New(***REMOVED***leName string) (r Releaser, existed bool, err error) {
	if err = os.MkdirAll(***REMOVED***lepath.Dir(***REMOVED***leName), 0755); err != nil {
		return
	}

	_, err = os.Stat(***REMOVED***leName)
	existed = err == nil

	r, err = newLock(***REMOVED***leName)
	return
}
