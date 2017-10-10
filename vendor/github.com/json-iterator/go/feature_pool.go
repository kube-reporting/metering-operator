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
	select {
	case stream := <-cfg.streamPool:
		stream.Reset(writer)
		return stream
	default:
		return NewStream(cfg, writer, 512)
	}
}

func (cfg *frozenCon***REMOVED***g) ReturnStream(stream *Stream) {
	stream.Error = nil
	select {
	case cfg.streamPool <- stream:
		return
	default:
		return
	}
}

func (cfg *frozenCon***REMOVED***g) BorrowIterator(data []byte) *Iterator {
	select {
	case iter := <-cfg.iteratorPool:
		iter.ResetBytes(data)
		return iter
	default:
		return ParseBytes(cfg, data)
	}
}

func (cfg *frozenCon***REMOVED***g) ReturnIterator(iter *Iterator) {
	iter.Error = nil
	select {
	case cfg.iteratorPool <- iter:
		return
	default:
		return
	}
}
