package jsoniter

import (
	"fmt"
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
	onePtrEmbedded     bool
	onePtrOptimization bool
	Type               reflect.Type
	Fields             []*Binding
}

// GetField get one ***REMOVED***eld from the descriptor by its name.
// Can not use map here to keep ***REMOVED***eld orders.
func (structDescriptor *StructDescriptor) GetField(***REMOVED***eldName string) *Binding {
	for _, binding := range structDescriptor.Fields {
		if binding.Field.Name == ***REMOVED***eldName {
			return binding
		}
	}
	return nil
}

// Binding describe how should we encode/decode the struct ***REMOVED***eld
type Binding struct {
	levels    []int
	Field     *reflect.StructField
	FromNames []string
	ToNames   []string
	Encoder   ValEncoder
	Decoder   ValDecoder
}

// Extension the one for all SPI. Customize encoding/decoding by specifying alternate encoder/decoder.
// Can also rename ***REMOVED***elds by UpdateStructDescriptor.
type Extension interface {
	UpdateStructDescriptor(structDescriptor *StructDescriptor)
	CreateDecoder(typ reflect.Type) ValDecoder
	CreateEncoder(typ reflect.Type) ValEncoder
	DecorateDecoder(typ reflect.Type, decoder ValDecoder) ValDecoder
	DecorateEncoder(typ reflect.Type, encoder ValEncoder) ValEncoder
}

// DummyExtension embed this type get dummy implementation for all methods of Extension
type DummyExtension struct {
}

// UpdateStructDescriptor No-op
func (extension *DummyExtension) UpdateStructDescriptor(structDescriptor *StructDescriptor) {
}

// CreateDecoder No-op
func (extension *DummyExtension) CreateDecoder(typ reflect.Type) ValDecoder {
	return nil
}

// CreateEncoder No-op
func (extension *DummyExtension) CreateEncoder(typ reflect.Type) ValEncoder {
	return nil
}

// DecorateDecoder No-op
func (extension *DummyExtension) DecorateDecoder(typ reflect.Type, decoder ValDecoder) ValDecoder {
	return decoder
}

// DecorateEncoder No-op
func (extension *DummyExtension) DecorateEncoder(typ reflect.Type, encoder ValEncoder) ValEncoder {
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

func (encoder *funcEncoder) EncodeInterface(val interface{}, stream *Stream) {
	WriteToStream(val, stream, encoder)
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

func getTypeDecoderFromExtension(typ reflect.Type) ValDecoder {
	decoder := _getTypeDecoderFromExtension(typ)
	if decoder != nil {
		for _, extension := range extensions {
			decoder = extension.DecorateDecoder(typ, decoder)
		}
	}
	return decoder
}
func _getTypeDecoderFromExtension(typ reflect.Type) ValDecoder {
	for _, extension := range extensions {
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
		decoder := typeDecoders[typ.Elem().String()]
		if decoder != nil {
			return &optionalDecoder{typ.Elem(), decoder}
		}
	}
	return nil
}

func getTypeEncoderFromExtension(typ reflect.Type) ValEncoder {
	encoder := _getTypeEncoderFromExtension(typ)
	if encoder != nil {
		for _, extension := range extensions {
			encoder = extension.DecorateEncoder(typ, encoder)
		}
	}
	return encoder
}

func _getTypeEncoderFromExtension(typ reflect.Type) ValEncoder {
	for _, extension := range extensions {
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
		encoder := typeEncoders[typ.Elem().String()]
		if encoder != nil {
			return &optionalEncoder{encoder}
		}
	}
	return nil
}

func describeStruct(cfg *frozenCon***REMOVED***g, typ reflect.Type) (*StructDescriptor, error) {
	embeddedBindings := []*Binding{}
	bindings := []*Binding{}
	for i := 0; i < typ.NumField(); i++ {
		***REMOVED***eld := typ.Field(i)
		tag := ***REMOVED***eld.Tag.Get(cfg.getTagKey())
		tagParts := strings.Split(tag, ",")
		if tag == "-" {
			continue
		}
		if ***REMOVED***eld.Anonymous && (tag == "" || tagParts[0] == "") {
			if ***REMOVED***eld.Type.Kind() == reflect.Struct {
				structDescriptor, err := describeStruct(cfg, ***REMOVED***eld.Type)
				if err != nil {
					return nil, err
				}
				for _, binding := range structDescriptor.Fields {
					binding.levels = append([]int{i}, binding.levels...)
					omitempty := binding.Encoder.(*structFieldEncoder).omitempty
					binding.Encoder = &structFieldEncoder{&***REMOVED***eld, binding.Encoder, omitempty}
					binding.Decoder = &structFieldDecoder{&***REMOVED***eld, binding.Decoder}
					embeddedBindings = append(embeddedBindings, binding)
				}
				continue
			} ***REMOVED*** if ***REMOVED***eld.Type.Kind() == reflect.Ptr && ***REMOVED***eld.Type.Elem().Kind() == reflect.Struct {
				structDescriptor, err := describeStruct(cfg, ***REMOVED***eld.Type.Elem())
				if err != nil {
					return nil, err
				}
				for _, binding := range structDescriptor.Fields {
					binding.levels = append([]int{i}, binding.levels...)
					omitempty := binding.Encoder.(*structFieldEncoder).omitempty
					binding.Encoder = &optionalEncoder{binding.Encoder}
					binding.Encoder = &structFieldEncoder{&***REMOVED***eld, binding.Encoder, omitempty}
					binding.Decoder = &deferenceDecoder{***REMOVED***eld.Type.Elem(), binding.Decoder}
					binding.Decoder = &structFieldDecoder{&***REMOVED***eld, binding.Decoder}
					embeddedBindings = append(embeddedBindings, binding)
				}
				continue
			}
		}
		***REMOVED***eldNames := calcFieldNames(***REMOVED***eld.Name, tagParts[0], tag)
		***REMOVED***eldCacheKey := fmt.Sprintf("%s/%s", typ.String(), ***REMOVED***eld.Name)
		decoder := ***REMOVED***eldDecoders[***REMOVED***eldCacheKey]
		if decoder == nil {
			var err error
			decoder, err = decoderOfType(cfg, ***REMOVED***eld.Type)
			if err != nil {
				return nil, err
			}
		}
		encoder := ***REMOVED***eldEncoders[***REMOVED***eldCacheKey]
		if encoder == nil {
			var err error
			encoder, err = encoderOfType(cfg, ***REMOVED***eld.Type)
			if err != nil {
				return nil, err
			}
			// map is stored as pointer in the struct
			if ***REMOVED***eld.Type.Kind() == reflect.Map {
				encoder = &optionalEncoder{encoder}
			}
		}
		binding := &Binding{
			Field:     &***REMOVED***eld,
			FromNames: ***REMOVED***eldNames,
			ToNames:   ***REMOVED***eldNames,
			Decoder:   decoder,
			Encoder:   encoder,
		}
		binding.levels = []int{i}
		bindings = append(bindings, binding)
	}
	return createStructDescriptor(cfg, typ, bindings, embeddedBindings), nil
}
func createStructDescriptor(cfg *frozenCon***REMOVED***g, typ reflect.Type, bindings []*Binding, embeddedBindings []*Binding) *StructDescriptor {
	onePtrEmbedded := false
	onePtrOptimization := false
	if typ.NumField() == 1 {
		***REMOVED***rstField := typ.Field(0)
		switch ***REMOVED***rstField.Type.Kind() {
		case reflect.Ptr:
			if ***REMOVED***rstField.Anonymous && ***REMOVED***rstField.Type.Elem().Kind() == reflect.Struct {
				onePtrEmbedded = true
			}
			fallthrough
		case reflect.Map:
			onePtrOptimization = true
		case reflect.Struct:
			onePtrOptimization = isStructOnePtr(***REMOVED***rstField.Type)
		}
	}
	structDescriptor := &StructDescriptor{
		onePtrEmbedded:     onePtrEmbedded,
		onePtrOptimization: onePtrOptimization,
		Type:               typ,
		Fields:             bindings,
	}
	for _, extension := range extensions {
		extension.UpdateStructDescriptor(structDescriptor)
	}
	processTags(structDescriptor, cfg)
	// merge normal & embedded bindings & sort with original order
	allBindings := sortableBindings(append(embeddedBindings, structDescriptor.Fields...))
	sort.Sort(allBindings)
	structDescriptor.Fields = allBindings
	return structDescriptor
}

func isStructOnePtr(typ reflect.Type) bool {
	if typ.NumField() == 1 {
		***REMOVED***rstField := typ.Field(0)
		switch ***REMOVED***rstField.Type.Kind() {
		case reflect.Ptr:
			return true
		case reflect.Map:
			return true
		case reflect.Struct:
			return isStructOnePtr(***REMOVED***rstField.Type)
		}
	}
	return false
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
		tagParts := strings.Split(binding.Field.Tag.Get(cfg.getTagKey()), ",")
		for _, tagPart := range tagParts[1:] {
			if tagPart == "omitempty" {
				shouldOmitEmpty = true
			} ***REMOVED*** if tagPart == "string" {
				if binding.Field.Type.Kind() == reflect.String {
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
