package jsoniter

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"sync/atomic"
	"unsafe"
)

// Con***REMOVED***g customize how the API should behave.
// The API is created from Con***REMOVED***g by Froze.
type Con***REMOVED***g struct {
	IndentionStep           int
	MarshalFloatWith6Digits bool
	EscapeHTML              bool
	SortMapKeys             bool
	UseNumber               bool
	TagKey                  string
}

type frozenCon***REMOVED***g struct {
	con***REMOVED***gBeforeFrozen Con***REMOVED***g
	sortMapKeys        bool
	indentionStep      int
	decoderCache       unsafe.Pointer
	encoderCache       unsafe.Pointer
	extensions         []Extension
	streamPool         chan *Stream
	iteratorPool       chan *Iterator
}

// API the public interface of this package.
// Primary Marshal and Unmarshal.
type API interface {
	IteratorPool
	StreamPool
	MarshalToString(v interface{}) (string, error)
	Marshal(v interface{}) ([]byte, error)
	MarshalIndent(v interface{}, pre***REMOVED***x, indent string) ([]byte, error)
	UnmarshalFromString(str string, v interface{}) error
	Unmarshal(data []byte, v interface{}) error
	Get(data []byte, path ...interface{}) Any
	NewEncoder(writer io.Writer) *Encoder
	NewDecoder(reader io.Reader) *Decoder
}

// Con***REMOVED***gDefault the default API
var Con***REMOVED***gDefault = Con***REMOVED***g{
	EscapeHTML: true,
}.Froze()

// Con***REMOVED***gCompatibleWithStandardLibrary tries to be 100% compatible with standard library behavior
var Con***REMOVED***gCompatibleWithStandardLibrary = Con***REMOVED***g{
	EscapeHTML:  true,
	SortMapKeys: true,
}.Froze()

// Con***REMOVED***gFastest marshals float with only 6 digits precision
var Con***REMOVED***gFastest = Con***REMOVED***g{
	EscapeHTML:              false,
	MarshalFloatWith6Digits: true,
}.Froze()

// Froze forge API from con***REMOVED***g
func (cfg Con***REMOVED***g) Froze() API {
	// TODO: cache frozen con***REMOVED***g
	frozenCon***REMOVED***g := &frozenCon***REMOVED***g{
		sortMapKeys:   cfg.SortMapKeys,
		indentionStep: cfg.IndentionStep,
		streamPool:    make(chan *Stream, 16),
		iteratorPool:  make(chan *Iterator, 16),
	}
	atomic.StorePointer(&frozenCon***REMOVED***g.decoderCache, unsafe.Pointer(&map[string]ValDecoder{}))
	atomic.StorePointer(&frozenCon***REMOVED***g.encoderCache, unsafe.Pointer(&map[string]ValEncoder{}))
	if cfg.MarshalFloatWith6Digits {
		frozenCon***REMOVED***g.marshalFloatWith6Digits()
	}
	if cfg.EscapeHTML {
		frozenCon***REMOVED***g.escapeHTML()
	}
	if cfg.UseNumber {
		frozenCon***REMOVED***g.useNumber()
	}
	frozenCon***REMOVED***g.con***REMOVED***gBeforeFrozen = cfg
	return frozenCon***REMOVED***g
}

func (cfg *frozenCon***REMOVED***g) useNumber() {
	cfg.addDecoderToCache(reflect.TypeOf((*interface{})(nil)).Elem(), &funcDecoder{func(ptr unsafe.Pointer, iter *Iterator) {
		if iter.WhatIsNext() == NumberValue {
			*((*interface{})(ptr)) = json.Number(iter.readNumberAsString())
		} ***REMOVED*** {
			*((*interface{})(ptr)) = iter.Read()
		}
	}})
}
func (cfg *frozenCon***REMOVED***g) getTagKey() string {
	tagKey := cfg.con***REMOVED***gBeforeFrozen.TagKey
	if tagKey == "" {
		return "json"
	}
	return tagKey
}

func (cfg *frozenCon***REMOVED***g) registerExtension(extension Extension) {
	cfg.extensions = append(cfg.extensions, extension)
}

type lossyFloat32Encoder struct {
}

func (encoder *lossyFloat32Encoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	stream.WriteFloat32Lossy(*((*float32)(ptr)))
}

func (encoder *lossyFloat32Encoder) EncodeInterface(val interface{}, stream *Stream) {
	WriteToStream(val, stream, encoder)
}

func (encoder *lossyFloat32Encoder) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*float32)(ptr)) == 0
}

type lossyFloat64Encoder struct {
}

func (encoder *lossyFloat64Encoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	stream.WriteFloat64Lossy(*((*float64)(ptr)))
}

func (encoder *lossyFloat64Encoder) EncodeInterface(val interface{}, stream *Stream) {
	WriteToStream(val, stream, encoder)
}

func (encoder *lossyFloat64Encoder) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*float64)(ptr)) == 0
}

// EnableLossyFloatMarshalling keeps 10**(-6) precision
// for float variables for better performance.
func (cfg *frozenCon***REMOVED***g) marshalFloatWith6Digits() {
	// for better performance
	cfg.addEncoderToCache(reflect.TypeOf((*float32)(nil)).Elem(), &lossyFloat32Encoder{})
	cfg.addEncoderToCache(reflect.TypeOf((*float64)(nil)).Elem(), &lossyFloat64Encoder{})
}

type htmlEscapedStringEncoder struct {
}

func (encoder *htmlEscapedStringEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	str := *((*string)(ptr))
	stream.WriteStringWithHTMLEscaped(str)
}

func (encoder *htmlEscapedStringEncoder) EncodeInterface(val interface{}, stream *Stream) {
	WriteToStream(val, stream, encoder)
}

func (encoder *htmlEscapedStringEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*string)(ptr)) == ""
}

func (cfg *frozenCon***REMOVED***g) escapeHTML() {
	cfg.addEncoderToCache(reflect.TypeOf((*string)(nil)).Elem(), &htmlEscapedStringEncoder{})
}

func (cfg *frozenCon***REMOVED***g) addDecoderToCache(cacheKey reflect.Type, decoder ValDecoder) {
	done := false
	for !done {
		ptr := atomic.LoadPointer(&cfg.decoderCache)
		cache := *(*map[reflect.Type]ValDecoder)(ptr)
		copied := map[reflect.Type]ValDecoder{}
		for k, v := range cache {
			copied[k] = v
		}
		copied[cacheKey] = decoder
		done = atomic.CompareAndSwapPointer(&cfg.decoderCache, ptr, unsafe.Pointer(&copied))
	}
}

func (cfg *frozenCon***REMOVED***g) addEncoderToCache(cacheKey reflect.Type, encoder ValEncoder) {
	done := false
	for !done {
		ptr := atomic.LoadPointer(&cfg.encoderCache)
		cache := *(*map[reflect.Type]ValEncoder)(ptr)
		copied := map[reflect.Type]ValEncoder{}
		for k, v := range cache {
			copied[k] = v
		}
		copied[cacheKey] = encoder
		done = atomic.CompareAndSwapPointer(&cfg.encoderCache, ptr, unsafe.Pointer(&copied))
	}
}

func (cfg *frozenCon***REMOVED***g) getDecoderFromCache(cacheKey reflect.Type) ValDecoder {
	ptr := atomic.LoadPointer(&cfg.decoderCache)
	cache := *(*map[reflect.Type]ValDecoder)(ptr)
	return cache[cacheKey]
}

func (cfg *frozenCon***REMOVED***g) getEncoderFromCache(cacheKey reflect.Type) ValEncoder {
	ptr := atomic.LoadPointer(&cfg.encoderCache)
	cache := *(*map[reflect.Type]ValEncoder)(ptr)
	return cache[cacheKey]
}

func (cfg *frozenCon***REMOVED***g) cleanDecoders() {
	typeDecoders = map[string]ValDecoder{}
	***REMOVED***eldDecoders = map[string]ValDecoder{}
	*cfg = *(cfg.con***REMOVED***gBeforeFrozen.Froze().(*frozenCon***REMOVED***g))
}

func (cfg *frozenCon***REMOVED***g) cleanEncoders() {
	typeEncoders = map[string]ValEncoder{}
	***REMOVED***eldEncoders = map[string]ValEncoder{}
	*cfg = *(cfg.con***REMOVED***gBeforeFrozen.Froze().(*frozenCon***REMOVED***g))
}

func (cfg *frozenCon***REMOVED***g) MarshalToString(v interface{}) (string, error) {
	stream := cfg.BorrowStream(nil)
	defer cfg.ReturnStream(stream)
	stream.WriteVal(v)
	if stream.Error != nil {
		return "", stream.Error
	}
	return string(stream.Buffer()), nil
}

func (cfg *frozenCon***REMOVED***g) Marshal(v interface{}) ([]byte, error) {
	stream := cfg.BorrowStream(nil)
	defer cfg.ReturnStream(stream)
	stream.WriteVal(v)
	if stream.Error != nil {
		return nil, stream.Error
	}
	result := stream.Buffer()
	copied := make([]byte, len(result))
	copy(copied, result)
	return copied, nil
}

func (cfg *frozenCon***REMOVED***g) MarshalIndent(v interface{}, pre***REMOVED***x, indent string) ([]byte, error) {
	if pre***REMOVED***x != "" {
		panic("pre***REMOVED***x is not supported")
	}
	for _, r := range indent {
		if r != ' ' {
			panic("indent can only be space")
		}
	}
	newCfg := cfg.con***REMOVED***gBeforeFrozen
	newCfg.IndentionStep = len(indent)
	return newCfg.Froze().Marshal(v)
}

func (cfg *frozenCon***REMOVED***g) UnmarshalFromString(str string, v interface{}) error {
	data := []byte(str)
	data = data[:lastNotSpacePos(data)]
	iter := cfg.BorrowIterator(data)
	defer cfg.ReturnIterator(iter)
	iter.ReadVal(v)
	if iter.head == iter.tail {
		iter.loadMore()
	}
	if iter.Error == io.EOF {
		return nil
	}
	if iter.Error == nil {
		iter.ReportError("UnmarshalFromString", "there are bytes left after unmarshal")
	}
	return iter.Error
}

func (cfg *frozenCon***REMOVED***g) Get(data []byte, path ...interface{}) Any {
	iter := cfg.BorrowIterator(data)
	defer cfg.ReturnIterator(iter)
	return locatePath(iter, path)
}

func (cfg *frozenCon***REMOVED***g) Unmarshal(data []byte, v interface{}) error {
	data = data[:lastNotSpacePos(data)]
	iter := cfg.BorrowIterator(data)
	defer cfg.ReturnIterator(iter)
	typ := reflect.TypeOf(v)
	if typ.Kind() != reflect.Ptr {
		// return non-pointer error
		return errors.New("the second param must be ptr type")
	}
	iter.ReadVal(v)
	if iter.head == iter.tail {
		iter.loadMore()
	}
	if iter.Error == io.EOF {
		return nil
	}
	if iter.Error == nil {
		iter.ReportError("Unmarshal", "there are bytes left after unmarshal")
	}
	return iter.Error
}

func (cfg *frozenCon***REMOVED***g) NewEncoder(writer io.Writer) *Encoder {
	stream := NewStream(cfg, writer, 512)
	return &Encoder{stream}
}

func (cfg *frozenCon***REMOVED***g) NewDecoder(reader io.Reader) *Decoder {
	iter := Parse(cfg, reader, 512)
	return &Decoder{iter}
}
