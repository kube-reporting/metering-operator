// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE ***REMOVED***le.

// Package storage provides storage abstraction for LevelDB.
package storage

import (
	"errors"
	"fmt"
	"io"
)

// FileType represent a ***REMOVED***le type.
type FileType int

// File types.
const (
	TypeManifest FileType = 1 << iota
	TypeJournal
	TypeTable
	TypeTemp

	TypeAll = TypeManifest | TypeJournal | TypeTable | TypeTemp
)

func (t FileType) String() string {
	switch t {
	case TypeManifest:
		return "manifest"
	case TypeJournal:
		return "journal"
	case TypeTable:
		return "table"
	case TypeTemp:
		return "temp"
	}
	return fmt.Sprintf("<unknown:%d>", t)
}

// Common error.
var (
	ErrInvalidFile = errors.New("leveldb/storage: invalid ***REMOVED***le for argument")
	ErrLocked      = errors.New("leveldb/storage: already locked")
	ErrClosed      = errors.New("leveldb/storage: closed")
)

// ErrCorrupted is the type that wraps errors that indicate corruption of
// a ***REMOVED***le. Package storage has its own type instead of using
// errors.ErrCorrupted to prevent circular import.
type ErrCorrupted struct {
	Fd  FileDesc
	Err error
}

func (e *ErrCorrupted) Error() string {
	if !e.Fd.Zero() {
		return fmt.Sprintf("%v [***REMOVED***le=%v]", e.Err, e.Fd)
	}
	return e.Err.Error()
}

// Syncer is the interface that wraps basic Sync method.
type Syncer interface {
	// Sync commits the current contents of the ***REMOVED***le to stable storage.
	Sync() error
}

// Reader is the interface that groups the basic Read, Seek, ReadAt and Close
// methods.
type Reader interface {
	io.ReadSeeker
	io.ReaderAt
	io.Closer
}

// Writer is the interface that groups the basic Write, Sync and Close
// methods.
type Writer interface {
	io.WriteCloser
	Syncer
}

// Locker is the interface that wraps Unlock method.
type Locker interface {
	Unlock()
}

// FileDesc is a '***REMOVED***le descriptor'.
type FileDesc struct {
	Type FileType
	Num  int64
}

func (fd FileDesc) String() string {
	switch fd.Type {
	case TypeManifest:
		return fmt.Sprintf("MANIFEST-%06d", fd.Num)
	case TypeJournal:
		return fmt.Sprintf("%06d.log", fd.Num)
	case TypeTable:
		return fmt.Sprintf("%06d.ldb", fd.Num)
	case TypeTemp:
		return fmt.Sprintf("%06d.tmp", fd.Num)
	default:
		return fmt.Sprintf("%#x-%d", fd.Type, fd.Num)
	}
}

// Zero returns true if fd == (FileDesc{}).
func (fd FileDesc) Zero() bool {
	return fd == (FileDesc{})
}

// FileDescOk returns true if fd is a valid '***REMOVED***le descriptor'.
func FileDescOk(fd FileDesc) bool {
	switch fd.Type {
	case TypeManifest:
	case TypeJournal:
	case TypeTable:
	case TypeTemp:
	default:
		return false
	}
	return fd.Num >= 0
}

// Storage is the storage. A storage instance must be safe for concurrent use.
type Storage interface {
	// Lock locks the storage. Any subsequent attempt to call Lock will fail
	// until the last lock released.
	// Caller should call Unlock method after use.
	Lock() (Locker, error)

	// Log logs a string. This is used for logging.
	// An implementation may write to a ***REMOVED***le, stdout or simply do nothing.
	Log(str string)

	// SetMeta store '***REMOVED***le descriptor' that can later be acquired using GetMeta
	// method. The '***REMOVED***le descriptor' should point to a valid ***REMOVED***le.
	// SetMeta should be implemented in such way that changes should happen
	// atomically.
	SetMeta(fd FileDesc) error

	// GetMeta returns '***REMOVED***le descriptor' stored in meta. The '***REMOVED***le descriptor'
	// can be updated using SetMeta method.
	// Returns os.ErrNotExist if meta doesn't store any '***REMOVED***le descriptor', or
	// '***REMOVED***le descriptor' point to nonexistent ***REMOVED***le.
	GetMeta() (FileDesc, error)

	// List returns ***REMOVED***le descriptors that match the given ***REMOVED***le types.
	// The ***REMOVED***le types may be OR'ed together.
	List(ft FileType) ([]FileDesc, error)

	// Open opens ***REMOVED***le with the given '***REMOVED***le descriptor' read-only.
	// Returns os.ErrNotExist error if the ***REMOVED***le does not exist.
	// Returns ErrClosed if the underlying storage is closed.
	Open(fd FileDesc) (Reader, error)

	// Create creates ***REMOVED***le with the given '***REMOVED***le descriptor', truncate if already
	// exist and opens write-only.
	// Returns ErrClosed if the underlying storage is closed.
	Create(fd FileDesc) (Writer, error)

	// Remove removes ***REMOVED***le with the given '***REMOVED***le descriptor'.
	// Returns ErrClosed if the underlying storage is closed.
	Remove(fd FileDesc) error

	// Rename renames ***REMOVED***le from oldfd to newfd.
	// Returns ErrClosed if the underlying storage is closed.
	Rename(oldfd, newfd FileDesc) error

	// Close closes the storage.
	// It is valid to call Close multiple times. Other methods should not be
	// called after the storage has been closed.
	Close() error
}
