package jsoniter

import (
	"fmt"
	"github.com/modern-go/reflect2"
	"reflect"
	"sort"
	"strings"
	"unicode"
	"unsafe"
)

var typeDecoders = map[string]ValDecoder{}
var ***REMOVED***eldDecoders = map[string]ValDecoder{}
var typeEncoders = map[string]ValEncoder{}
var ***REMOVED***eldEncoders = map[string]ValEncoder{}
var extensions = []Extension{}

// StructDescriptor describe how should we encode/decode the struct
type StructDescriptor struct {
	Type   reflect2.Type
	Fields []*Binding
}

// GetField get one ***REMOVED***eld from the descriptor by its name.
// Can not use map here to keep ***REMOVED***eld orders.
func (structDescriptor *StructDescriptor) GetField(***REMOVED***eldName string) *Binding {
	for _, binding := range structDescriptor.Fields {
		if binding.Field.Name() == ***REMOVED***eldName {
			return binding
		}
	}
	return nil
}

// Binding describe how should we encode/decode the struct ***REMOVED***eld
type Binding struct {
	levels    []int
	Field     reflect2.StructField
	FromNames []string
	ToNames   []string
	Encoder   ValEncoder
	Decoder   ValDecoder
}

// Extension the one for all SPI. Customize encoding/decoding by specifying alternate encoder/decoder.
// Can also rename ***REMOVED***elds by UpdateStructDescriptor.
type Extension interface {
	UpdateStructDescriptor(structDescriptor *StructDescriptor)
	CreateMapKeyDecoder(typ reflect2.Type) ValDecoder
	CreateMapKeyEncoder(typ reflect2.Type) ValEncoder
	CreateDecoder(typ reflect2.Type) ValDecoder
	CreateEncoder(typ reflect2.Type) ValEncoder
	DecorateDecoder(typ reflect2.Type, decoder ValDecoder) ValDecoder
	DecorateEncoder(typ reflect2.Type, encoder ValEncoder) ValEncoder
}

// DummyExtension embed this type get dummy implementation for all methods of Extension
type DummyExtension struct {
}

// UpdateStructDescriptor No-op
func (extension *DummyExtension) UpdateStructDescriptor(structDescriptor *StructDescriptor) {
}

// CreateMapKeyDecoder No-op
func (extension *DummyExtension) CreateMapKeyDecoder(typ reflect2.Type) ValDecoder {
	return nil
}

// CreateMapKeyEncoder No-op
func (extension *DummyExtension) CreateMapKeyEncoder(typ reflect2.Type) ValEncoder {
	return nil
}

// CreateDecoder No-op
func (extension *DummyExtension) CreateDecoder(typ reflect2.Type) ValDecoder {
	return nil
}

// CreateEncoder No-op
func (extension *DummyExtension) CreateEncoder(typ reflect2.Type) ValEncoder {
	return nil
}

// DecorateDecoder No-op
func (extension *DummyExtension) DecorateDecoder(typ reflect2.Type, decoder ValDecoder) ValDecoder {
	return decoder
}

// DecorateEncoder No-op
func (extension *DummyExtension) DecorateEncoder(typ reflect2.Type, encoder ValEncoder) ValEncoder {
	return encoder
}

type EncoderExtension map[reflect2.Type]ValEncoder

// UpdateStructDescriptor No-op
func (extension EncoderExtension) UpdateStructDescriptor(structDescriptor *StructDescriptor) {
}

// CreateDecoder No-op
func (extension EncoderExtension) CreateDecoder(typ reflect2.Type) ValDecoder {
	return nil
}

// CreateEncoder get encoder from map
func (extension EncoderExtension) CreateEncoder(typ reflect2.Type) ValEncoder {
	return extension[typ]
}

// CreateMapKeyDecoder No-op
func (extension EncoderExtension) CreateMapKeyDecoder(typ reflect2.Type) ValDecoder {
	return nil
}

// CreateMapKeyEncoder No-op
func (extension EncoderExtension) CreateMapKeyEncoder(typ reflect2.Type) ValEncoder {
	return nil
}

// DecorateDecoder No-op
func (extension EncoderExtension) DecorateDecoder(typ reflect2.Type, decoder ValDecoder) ValDecoder {
	return decoder
}

// DecorateEncoder No-op
func (extension EncoderExtension) DecorateEncoder(typ reflect2.Type, encoder ValEncoder) ValEncoder {
	return encoder
}

type DecoderExtension map[reflect2.Type]ValDecoder

// UpdateStructDescriptor No-op
func (extension DecoderExtension) UpdateStructDescriptor(structDescriptor *StructDescriptor) {
}

// CreateMapKeyDecoder No-op
func (extension DecoderExtension) CreateMapKeyDecoder(typ reflect2.Type) ValDecoder {
	return nil
}

// CreateMapKeyEncoder No-op
func (extension DecoderExtension) CreateMapKeyEncoder(typ reflect2.Type) ValEncoder {
	return nil
}

// CreateDecoder get decoder from map
func (extension DecoderExtension) CreateDecoder(typ reflect2.Type) ValDecoder {
	return extension[typ]
}

// CreateEncoder No-op
func (extension DecoderExtension) CreateEncoder(typ reflect2.Type) ValEncoder {
	return nil
}

// DecorateDecoder No-op
func (extension DecoderExtension) DecorateDecoder(typ reflect2.Type, decoder ValDecoder) ValDecoder {
	return decoder
}

// DecorateEncoder No-op
func (extension DecoderExtension) DecorateEncoder(typ reflect2.Type, encoder ValEncoder) ValEncoder {
	return encoder
}

type funcDecoder struct {
	fun DecoderFunc
}

func (decoder *funcDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	decoder.fun(ptr, iter)
}

type funcEncoder struct {
	fun         EncoderFunc
	isEmptyFunc func(ptr unsafe.Pointer) bool
}

func (encoder *funcEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	encoder.fun(ptr, stream)
}

func (encoder *funcEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	if encoder.isEmptyFunc == nil {
		return false
	}
	return encoder.isEmptyFunc(ptr)
}

// DecoderFunc the function form of TypeDecoder
type DecoderFunc func(ptr unsafe.Pointer, iter *Iterator)

// EncoderFunc the function form of TypeEncoder
type EncoderFunc func(ptr unsafe.Pointer, stream *Stream)

// RegisterTypeDecoderFunc register TypeDecoder for a type with function
func RegisterTypeDecoderFunc(typ string, fun DecoderFunc) {
	typeDecoders[typ] = &funcDecoder{fun}
}

// RegisterTypeDecoder register TypeDecoder for a typ
func RegisterTypeDecoder(typ string, decoder ValDecoder) {
	typeDecoders[typ] = decoder
}

// RegisterFieldDecoderFunc register TypeDecoder for a struct ***REMOVED***eld with function
func RegisterFieldDecoderFunc(typ string, ***REMOVED***eld string, fun DecoderFunc) {
	RegisterFieldDecoder(typ, ***REMOVED***eld, &funcDecoder{fun})
}

// RegisterFieldDecoder register TypeDecoder for a struct ***REMOVED***eld
func RegisterFieldDecoder(typ string, ***REMOVED***eld string, decoder ValDecoder) {
	***REMOVED***eldDecoders[fmt.Sprintf("%s/%s", typ, ***REMOVED***eld)] = decoder
}

// RegisterTypeEncoderFunc register TypeEncoder for a type with encode/isEmpty function
func RegisterTypeEncoderFunc(typ string, fun EncoderFunc, isEmptyFunc func(unsafe.Pointer) bool) {
	typeEncoders[typ] = &funcEncoder{fun, isEmptyFunc}
}

// RegisterTypeEncoder register TypeEncoder for a type
func RegisterTypeEncoder(typ string, encoder ValEncoder) {
	typeEncoders[typ] = encoder
}

// RegisterFieldEncoderFunc register TypeEncoder for a struct ***REMOVED***eld with encode/isEmpty function
func RegisterFieldEncoderFunc(typ string, ***REMOVED***eld string, fun EncoderFunc, isEmptyFunc func(unsafe.Pointer) bool) {
	RegisterFieldEncoder(typ, ***REMOVED***eld, &funcEncoder{fun, isEmptyFunc})
}

// RegisterFieldEncoder register TypeEncoder for a struct ***REMOVED***eld
func RegisterFieldEncoder(typ string, ***REMOVED***eld string, encoder ValEncoder) {
	***REMOVED***eldEncoders[fmt.Sprintf("%s/%s", typ, ***REMOVED***eld)] = encoder
}

// RegisterExtension register extension
func RegisterExtension(extension Extension) {
	extensions = append(extensions, extension)
}

func getTypeDecoderFromExtension(ctx *ctx, typ reflect2.Type) ValDecoder {
	decoder := _getTypeDecoderFromExtension(ctx, typ)
	if decoder != nil {
		for _, extension := range extensions {
			decoder = extension.DecorateDecoder(typ, decoder)
		}
		for _, extension := range ctx.extensions {
			decoder = extension.DecorateDecoder(typ, decoder)
		}
	}
	return decoder
}
func _getTypeDecoderFromExtension(ctx *ctx, typ reflect2.Type) ValDecoder {
	for _, extension := range extensions {
		decoder := extension.CreateDecoder(typ)
		if decoder != nil {
			return decoder
		}
	}
	for _, extension := range ctx.extensions {
		decoder := extension.CreateDecoder(typ)
		if decoder != nil {
			return decoder
		}
	}
	typeName := typ.String()
	decoder := typeDecoders[typeName]
	if decoder != nil {
		return decoder
	}
	if typ.Kind() == reflect.Ptr {
		ptrType := typ.(*reflect2.UnsafePtrType)
		decoder := typeDecoders[ptrType.Elem().String()]
		if decoder != nil {
			return &OptionalDecoder{ptrType.Elem(), decoder}
		}
	}
	return nil
}

func getTypeEncoderFromExtension(ctx *ctx, typ reflect2.Type) ValEncoder {
	encoder := _getTypeEncoderFromExtension(ctx, typ)
	if encoder != nil {
		for _, extension := range extensions {
			encoder = extension.DecorateEncoder(typ, encoder)
		}
		for _, extension := range ctx.extensions {
			encoder = extension.DecorateEncoder(typ, encoder)
		}
	}
	return encoder
}

func _getTypeEncoderFromExtension(ctx *ctx, typ reflect2.Type) ValEncoder {
	for _, extension := range extensions {
		encoder := extension.CreateEncoder(typ)
		if encoder != nil {
			return encoder
		}
	}
	for _, extension := range ctx.extensions {
		encoder := extension.CreateEncoder(typ)
		if encoder != nil {
			return encoder
		}
	}
	typeName := typ.String()
	encoder := typeEncoders[typeName]
	if encoder != nil {
		return encoder
	}
	if typ.Kind() == reflect.Ptr {
		typePtr := typ.(*reflect2.UnsafePtrType)
		encoder := typeEncoders[typePtr.Elem().String()]
		if encoder != nil {
			return &OptionalEncoder{encoder}
		}
	}
	return nil
}

func describeStruct(ctx *ctx, typ reflect2.Type) *StructDescriptor {
	structType := typ.(*reflect2.UnsafeStructType)
	embeddedBindings := []*Binding{}
	bindings := []*Binding{}
	for i := 0; i < structType.NumField(); i++ {
		***REMOVED***eld := structType.Field(i)
		tag, hastag := ***REMOVED***eld.Tag().Lookup(ctx.getTagKey())
		if ctx.onlyTaggedField && !hastag {
			continue
		}
		tagParts := strings.Split(tag, ",")
		if tag == "-" {
			continue
		}
		if ***REMOVED***eld.Anonymous() && (tag == "" || tagParts[0] == "") {
			if ***REMOVED***eld.Type().Kind() == reflect.Struct {
				structDescriptor := describeStruct(ctx, ***REMOVED***eld.Type())
				for _, binding := range structDescriptor.Fields {
					binding.levels = append([]int{i}, binding.levels...)
					omitempty := binding.Encoder.(*structFieldEncoder).omitempty
					binding.Encoder = &structFieldEncoder{***REMOVED***eld, binding.Encoder, omitempty}
					binding.Decoder = &structFieldDecoder{***REMOVED***eld, binding.Decoder}
					embeddedBindings = append(embeddedBindings, binding)
				}
				continue
			} ***REMOVED*** if ***REMOVED***eld.Type().Kind() == reflect.Ptr {
				ptrType := ***REMOVED***eld.Type().(*reflect2.UnsafePtrType)
				if ptrType.Elem().Kind() == reflect.Struct {
					structDescriptor := describeStruct(ctx, ptrType.Elem())
					for _, binding := range structDescriptor.Fields {
						binding.levels = append([]int{i}, binding.levels...)
						omitempty := binding.Encoder.(*structFieldEncoder).omitempty
						binding.Encoder = &dereferenceEncoder{binding.Encoder}
						binding.Encoder = &structFieldEncoder{***REMOVED***eld, binding.Encoder, omitempty}
						binding.Decoder = &dereferenceDecoder{ptrType.Elem(), binding.Decoder}
						binding.Decoder = &structFieldDecoder{***REMOVED***eld, binding.Decoder}
						embeddedBindings = append(embeddedBindings, binding)
					}
					continue
				}
			}
		}
		***REMOVED***eldNames := calcFieldNames(***REMOVED***eld.Name(), tagParts[0], tag)
		***REMOVED***eldCacheKey := fmt.Sprintf("%s/%s", typ.String(), ***REMOVED***eld.Name())
		decoder := ***REMOVED***eldDecoders[***REMOVED***eldCacheKey]
		if decoder == nil {
			decoder = decoderOfType(ctx.append(***REMOVED***eld.Name()), ***REMOVED***eld.Type())
		}
		encoder := ***REMOVED***eldEncoders[***REMOVED***eldCacheKey]
		if encoder == nil {
			encoder = encoderOfType(ctx.append(***REMOVED***eld.Name()), ***REMOVED***eld.Type())
		}
		binding := &Binding{
			Field:     ***REMOVED***eld,
			FromNames: ***REMOVED***eldNames,
			ToNames:   ***REMOVED***eldNames,
			Decoder:   decoder,
			Encoder:   encoder,
		}
		binding.levels = []int{i}
		bindings = append(bindings, binding)
	}
	return createStructDescriptor(ctx, typ, bindings, embeddedBindings)
}
func createStructDescriptor(ctx *ctx, typ reflect2.Type, bindings []*Binding, embeddedBindings []*Binding) *StructDescriptor {
	structDescriptor := &StructDescriptor{
		Type:   typ,
		Fields: bindings,
	}
	for _, extension := range extensions {
		extension.UpdateStructDescriptor(structDescriptor)
	}
	for _, extension := range ctx.extensions {
		extension.UpdateStructDescriptor(structDescriptor)
	}
	processTags(structDescriptor, ctx.frozenCon***REMOVED***g)
	// merge normal & embedded bindings & sort with original order
	allBindings := sortableBindings(append(embeddedBindings, structDescriptor.Fields...))
	sort.Sort(allBindings)
	structDescriptor.Fields = allBindings
	return structDescriptor
}

type sortableBindings []*Binding

func (bindings sortableBindings) Len() int {
	return len(bindings)
}

func (bindings sortableBindings) Less(i, j int) bool {
	left := bindings[i].levels
	right := bindings[j].levels
	k := 0
	for {
		if left[k] < right[k] {
			return true
		} ***REMOVED*** if left[k] > right[k] {
			return false
		}
		k++
	}
}

func (bindings sortableBindings) Swap(i, j int) {
	bindings[i], bindings[j] = bindings[j], bindings[i]
}

func processTags(structDescriptor *StructDescriptor, cfg *frozenCon***REMOVED***g) {
	for _, binding := range structDescriptor.Fields {
		shouldOmitEmpty := false
		tagParts := strings.Split(binding.Field.Tag().Get(cfg.getTagKey()), ",")
		for _, tagPart := range tagParts[1:] {
			if tagPart == "omitempty" {
				shouldOmitEmpty = true
			} ***REMOVED*** if tagPart == "string" {
				if binding.Field.Type().Kind() == reflect.String {
					binding.Decoder = &stringModeStringDecoder{binding.Decoder, cfg}
					binding.Encoder = &stringModeStringEncoder{binding.Encoder, cfg}
				} ***REMOVED*** {
					binding.Decoder = &stringModeNumberDecoder{binding.Decoder}
					binding.Encoder = &stringModeNumberEncoder{binding.Encoder}
				}
			}
		}
		binding.Decoder = &structFieldDecoder{binding.Field, binding.Decoder}
		binding.Encoder = &structFieldEncoder{binding.Field, binding.Encoder, shouldOmitEmpty}
	}
}

func calcFieldNames(originalFieldName string, tagProvidedFieldName string, wholeTag string) []string {
	// ignore?
	if wholeTag == "-" {
		return []string{}
	}
	// rename?
	var ***REMOVED***eldNames []string
	if tagProvidedFieldName == "" {
		***REMOVED***eldNames = []string{originalFieldName}
	} ***REMOVED*** {
		***REMOVED***eldNames = []string{tagProvidedFieldName}
	}
	// private?
	isNotExported := unicode.IsLower(rune(originalFieldName[0]))
	if isNotExported {
		***REMOVED***eldNames = []string{}
	}
	return ***REMOVED***eldNames
}
