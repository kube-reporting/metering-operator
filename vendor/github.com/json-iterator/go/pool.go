package jsoniter

import (
	"io"
)

// IteratorPool a thread safe pool of iterators with same con***REMOVED***guration
type IteratorPool interface {
	BorrowIterator(data []byte) *Iterator
	ReturnIterator(iter *Iterator)
}

// StreamPool a thread safe pool of streams with same con***REMOVED***guration
type StreamPool interface {
	BorrowStream(writer io.Writer) *Stream
	ReturnStream(stream *Stream)
}

func (cfg *frozenCon***REMOVED***g) BorrowStream(writer io.Writer) *Stream {
	stream := cfg.streamPool.Get().(*Stream)
	stream.Reset(writer)
	return stream
}

func (cfg *frozenCon***REMOVED***g) ReturnStream(stream *Stream) {
	stream.Error = nil
	stream.Attachment = nil
	cfg.streamPool.Put(stream)
}

func (cfg *frozenCon***REMOVED***g) BorrowIterator(data []byte) *Iterator {
	iter := cfg.iteratorPool.Get().(*Iterator)
	iter.ResetBytes(data)
	return iter
}

func (cfg *frozenCon***REMOVED***g) ReturnIterator(iter *Iterator) {
	iter.Error = nil
	iter.Attachment = nil
	cfg.iteratorPool.Put(iter)
}
