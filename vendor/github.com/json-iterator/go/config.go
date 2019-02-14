package jsoniter

import (
	"encoding/json"
	"io"
	"reflect"
	"sync"
	"unsafe"

	"github.com/modern-go/concurrent"
	"github.com/modern-go/reflect2"
)

// Con***REMOVED***g customize how the API should behave.
// The API is created from Con***REMOVED***g by Froze.
type Con***REMOVED***g struct {
	IndentionStep                 int
	MarshalFloatWith6Digits       bool
	EscapeHTML                    bool
	SortMapKeys                   bool
	UseNumber                     bool
	DisallowUnknownFields         bool
	TagKey                        string
	OnlyTaggedField               bool
	ValidateJsonRawMessage        bool
	ObjectFieldMustBeSimpleString bool
	CaseSensitive                 bool
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
	Valid(data []byte) bool
	RegisterExtension(extension Extension)
	DecoderOf(typ reflect2.Type) ValDecoder
	EncoderOf(typ reflect2.Type) ValEncoder
}

// Con***REMOVED***gDefault the default API
var Con***REMOVED***gDefault = Con***REMOVED***g{
	EscapeHTML: true,
}.Froze()

// Con***REMOVED***gCompatibleWithStandardLibrary tries to be 100% compatible with standard library behavior
var Con***REMOVED***gCompatibleWithStandardLibrary = Con***REMOVED***g{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
}.Froze()

// Con***REMOVED***gFastest marshals float with only 6 digits precision
var Con***REMOVED***gFastest = Con***REMOVED***g{
	EscapeHTML:                    false,
	MarshalFloatWith6Digits:       true, // will lose precession
	ObjectFieldMustBeSimpleString: true, // do not unescape object ***REMOVED***eld
}.Froze()

type frozenCon***REMOVED***g struct {
	con***REMOVED***gBeforeFrozen            Con***REMOVED***g
	sortMapKeys                   bool
	indentionStep                 int
	objectFieldMustBeSimpleString bool
	onlyTaggedField               bool
	disallowUnknownFields         bool
	decoderCache                  *concurrent.Map
	encoderCache                  *concurrent.Map
	encoderExtension              Extension
	decoderExtension              Extension
	extraExtensions               []Extension
	streamPool                    *sync.Pool
	iteratorPool                  *sync.Pool
	caseSensitive                 bool
}

func (cfg *frozenCon***REMOVED***g) initCache() {
	cfg.decoderCache = concurrent.NewMap()
	cfg.encoderCache = concurrent.NewMap()
}

func (cfg *frozenCon***REMOVED***g) addDecoderToCache(cacheKey uintptr, decoder ValDecoder) {
	cfg.decoderCache.Store(cacheKey, decoder)
}

func (cfg *frozenCon***REMOVED***g) addEncoderToCache(cacheKey uintptr, encoder ValEncoder) {
	cfg.encoderCache.Store(cacheKey, encoder)
}

func (cfg *frozenCon***REMOVED***g) getDecoderFromCache(cacheKey uintptr) ValDecoder {
	decoder, found := cfg.decoderCache.Load(cacheKey)
	if found {
		return decoder.(ValDecoder)
	}
	return nil
}

func (cfg *frozenCon***REMOVED***g) getEncoderFromCache(cacheKey uintptr) ValEncoder {
	encoder, found := cfg.encoderCache.Load(cacheKey)
	if found {
		return encoder.(ValEncoder)
	}
	return nil
}

var cfgCache = concurrent.NewMap()

func getFrozenCon***REMOVED***gFromCache(cfg Con***REMOVED***g) *frozenCon***REMOVED***g {
	obj, found := cfgCache.Load(cfg)
	if found {
		return obj.(*frozenCon***REMOVED***g)
	}
	return nil
}

func addFrozenCon***REMOVED***gToCache(cfg Con***REMOVED***g, frozenCon***REMOVED***g *frozenCon***REMOVED***g) {
	cfgCache.Store(cfg, frozenCon***REMOVED***g)
}

// Froze forge API from con***REMOVED***g
func (cfg Con***REMOVED***g) Froze() API {
	api := &frozenCon***REMOVED***g{
		sortMapKeys:                   cfg.SortMapKeys,
		indentionStep:                 cfg.IndentionStep,
		objectFieldMustBeSimpleString: cfg.ObjectFieldMustBeSimpleString,
		onlyTaggedField:               cfg.OnlyTaggedField,
		disallowUnknownFields:         cfg.DisallowUnknownFields,
		caseSensitive:                 cfg.CaseSensitive,
	}
	api.streamPool = &sync.Pool{
		New: func() interface{} {
			return NewStream(api, nil, 512)
		},
	}
	api.iteratorPool = &sync.Pool{
		New: func() interface{} {
			return NewIterator(api)
		},
	}
	api.initCache()
	encoderExtension := EncoderExtension{}
	decoderExtension := DecoderExtension{}
	if cfg.MarshalFloatWith6Digits {
		api.marshalFloatWith6Digits(encoderExtension)
	}
	if cfg.EscapeHTML {
		api.escapeHTML(encoderExtension)
	}
	if cfg.UseNumber {
		api.useNumber(decoderExtension)
	}
	if cfg.ValidateJsonRawMessage {
		api.validateJsonRawMessage(encoderExtension)
	}
	api.encoderExtension = encoderExtension
	api.decoderExtension = decoderExtension
	api.con***REMOVED***gBeforeFrozen = cfg
	return api
}

func (cfg Con***REMOVED***g) frozeWithCacheReuse(extraExtensions []Extension) *frozenCon***REMOVED***g {
	api := getFrozenCon***REMOVED***gFromCache(cfg)
	if api != nil {
		return api
	}
	api = cfg.Froze().(*frozenCon***REMOVED***g)
	for _, extension := range extraExtensions {
		api.RegisterExtension(extension)
	}
	addFrozenCon***REMOVED***gToCache(cfg, api)
	return api
}

func (cfg *frozenCon***REMOVED***g) validateJsonRawMessage(extension EncoderExtension) {
	encoder := &funcEncoder{func(ptr unsafe.Pointer, stream *Stream) {
		rawMessage := *(*json.RawMessage)(ptr)
		iter := cfg.BorrowIterator([]byte(rawMessage))
		iter.Read()
		if iter.Error != nil {
			stream.WriteRaw("null")
		} ***REMOVED*** {
			cfg.ReturnIterator(iter)
			stream.WriteRaw(string(rawMessage))
		}
	}, func(ptr unsafe.Pointer) bool {
		return len(*((*json.RawMessage)(ptr))) == 0
	}}
	extension[reflect2.TypeOfPtr((*json.RawMessage)(nil)).Elem()] = encoder
	extension[reflect2.TypeOfPtr((*RawMessage)(nil)).Elem()] = encoder
}

func (cfg *frozenCon***REMOVED***g) useNumber(extension DecoderExtension) {
	extension[reflect2.TypeOfPtr((*interface{})(nil)).Elem()] = &funcDecoder{func(ptr unsafe.Pointer, iter *Iterator) {
		exitingValue := *((*interface{})(ptr))
		if exitingValue != nil && reflect.TypeOf(exitingValue).Kind() == reflect.Ptr {
			iter.ReadVal(exitingValue)
			return
		}
		if iter.WhatIsNext() == NumberValue {
			*((*interface{})(ptr)) = json.Number(iter.readNumberAsString())
		} ***REMOVED*** {
			*((*interface{})(ptr)) = iter.Read()
		}
	}}
}
func (cfg *frozenCon***REMOVED***g) getTagKey() string {
	tagKey := cfg.con***REMOVED***gBeforeFrozen.TagKey
	if tagKey == "" {
		return "json"
	}
	return tagKey
}

func (cfg *frozenCon***REMOVED***g) RegisterExtension(extension Extension) {
	cfg.extraExtensions = append(cfg.extraExtensions, extension)
	copied := cfg.con***REMOVED***gBeforeFrozen
	cfg.con***REMOVED***gBeforeFrozen = copied
}

type lossyFloat32Encoder struct {
}

func (encoder *lossyFloat32Encoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	stream.WriteFloat32Lossy(*((*float32)(ptr)))
}

func (encoder *lossyFloat32Encoder) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*float32)(ptr)) == 0
}

type lossyFloat64Encoder struct {
}

func (encoder *lossyFloat64Encoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	stream.WriteFloat64Lossy(*((*float64)(ptr)))
}

func (encoder *lossyFloat64Encoder) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*float64)(ptr)) == 0
}

// EnableLossyFloatMarshalling keeps 10**(-6) precision
// for float variables for better performance.
func (cfg *frozenCon***REMOVED***g) marshalFloatWith6Digits(extension EncoderExtension) {
	// for better performance
	extension[reflect2.TypeOfPtr((*float32)(nil)).Elem()] = &lossyFloat32Encoder{}
	extension[reflect2.TypeOfPtr((*float64)(nil)).Elem()] = &lossyFloat64Encoder{}
}

type htmlEscapedStringEncoder struct {
}

func (encoder *htmlEscapedStringEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	str := *((*string)(ptr))
	stream.WriteStringWithHTMLEscaped(str)
}

func (encoder *htmlEscapedStringEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*string)(ptr)) == ""
}

func (cfg *frozenCon***REMOVED***g) escapeHTML(encoderExtension EncoderExtension) {
	encoderExtension[reflect2.TypeOfPtr((*string)(nil)).Elem()] = &htmlEscapedStringEncoder{}
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
	return newCfg.frozeWithCacheReuse(cfg.extraExtensions).Marshal(v)
}

func (cfg *frozenCon***REMOVED***g) UnmarshalFromString(str string, v interface{}) error {
	data := []byte(str)
	iter := cfg.BorrowIterator(data)
	defer cfg.ReturnIterator(iter)
	iter.ReadVal(v)
	c := iter.nextToken()
	if c == 0 {
		if iter.Error == io.EOF {
			return nil
		}
		return iter.Error
	}
	iter.ReportError("Unmarshal", "there are bytes left after unmarshal")
	return iter.Error
}

func (cfg *frozenCon***REMOVED***g) Get(data []byte, path ...interface{}) Any {
	iter := cfg.BorrowIterator(data)
	defer cfg.ReturnIterator(iter)
	return locatePath(iter, path)
}

func (cfg *frozenCon***REMOVED***g) Unmarshal(data []byte, v interface{}) error {
	iter := cfg.BorrowIterator(data)
	defer cfg.ReturnIterator(iter)
	iter.ReadVal(v)
	c := iter.nextToken()
	if c == 0 {
		if iter.Error == io.EOF {
			return nil
		}
		return iter.Error
	}
	iter.ReportError("Unmarshal", "there are bytes left after unmarshal")
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

func (cfg *frozenCon***REMOVED***g) Valid(data []byte) bool {
	iter := cfg.BorrowIterator(data)
	defer cfg.ReturnIterator(iter)
	iter.Skip()
	return iter.Error == nil
}
