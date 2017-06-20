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

package local

import (
	"sync"
	"unsafe"

	"github.com/prometheus/common/model"
)

const (
	cacheLineSize = 64
)

// Avoid false sharing when using array of mutexes.
type paddedMutex struct {
	sync.Mutex
	pad [cacheLineSize - unsafe.Sizeof(sync.Mutex{})]byte
}

// ***REMOVED***ngerprintLocker allows locking individual ***REMOVED***ngerprints. To limit the number
// of mutexes needed for that, only a ***REMOVED***xed number of mutexes are
// allocated. Fingerprints to be locked are assigned to those pre-allocated
// mutexes by their value. Collisions are not detected. If two ***REMOVED***ngerprints get
// assigned to the same mutex, only one of them can be locked at the same
// time. As long as the number of pre-allocated mutexes is much larger than the
// number of goroutines requiring a ***REMOVED***ngerprint lock concurrently, the loss in
// ef***REMOVED***ciency is small. However, a goroutine must never lock more than one
// ***REMOVED***ngerprint at the same time. (In that case a collision would try to acquire
// the same mutex twice).
type ***REMOVED***ngerprintLocker struct {
	fpMtxs    []paddedMutex
	numFpMtxs uint
}

// newFingerprintLocker returns a new ***REMOVED***ngerprintLocker ready for use.  At least
// 1024 preallocated mutexes are used, even if preallocatedMutexes is lower.
func newFingerprintLocker(preallocatedMutexes int) ****REMOVED***ngerprintLocker {
	if preallocatedMutexes < 1024 {
		preallocatedMutexes = 1024
	}
	return &***REMOVED***ngerprintLocker{
		make([]paddedMutex, preallocatedMutexes),
		uint(preallocatedMutexes),
	}
}

// Lock locks the given ***REMOVED***ngerprint.
func (l ****REMOVED***ngerprintLocker) Lock(fp model.Fingerprint) {
	l.fpMtxs[hashFP(fp)%l.numFpMtxs].Lock()
}

// Unlock unlocks the given ***REMOVED***ngerprint.
func (l ****REMOVED***ngerprintLocker) Unlock(fp model.Fingerprint) {
	l.fpMtxs[hashFP(fp)%l.numFpMtxs].Unlock()
}

// hashFP simply moves entropy from the most signi***REMOVED***cant 48 bits of the
// ***REMOVED***ngerprint into the least signi***REMOVED***cant 16 bits (by XORing) so that a simple
// MOD on the result can be used to pick a mutex while still making use of
// changes in more signi***REMOVED***cant bits of the ***REMOVED***ngerprint. (The fast ***REMOVED***ngerprinting
// function we use is prone to only change a few bits for similar metrics. We
// really want to make use of every change in the ***REMOVED***ngerprint to vary mutex
// selection.)
func hashFP(fp model.Fingerprint) uint {
	return uint(fp ^ (fp >> 32) ^ (fp >> 16))
}
