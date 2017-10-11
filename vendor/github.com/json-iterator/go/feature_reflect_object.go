package jsoniter

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"unsafe"
)

func encoderOfStruct(cfg *frozenCon***REMOVED***g, typ reflect.Type) (ValEncoder, error) {
	type bindingTo struct {
		binding *Binding
		toName  string
		ignored bool
	}
	orderedBindings := []*bindingTo{}
	structDescriptor, err := describeStruct(cfg, typ)
	if err != nil {
		return nil, err
	}
	for _, binding := range structDescriptor.Fields {
		for _, toName := range binding.ToNames {
			new := &bindingTo{
				binding: binding,
				toName:  toName,
			}
			for _, old := range orderedBindings {
				if old.toName != toName {
					continue
				}
				old.ignored, new.ignored = resolveConflictBinding(cfg, old.binding, new.binding)
			}
			orderedBindings = append(orderedBindings, new)
		}
	}
	if len(orderedBindings) == 0 {
		return &emptyStructEncoder{}, nil
	}
	***REMOVED***nalOrderedFields := []structFieldTo{}
	for _, bindingTo := range orderedBindings {
		if !bindingTo.ignored {
			***REMOVED***nalOrderedFields = append(***REMOVED***nalOrderedFields, structFieldTo{
				encoder: bindingTo.binding.Encoder.(*structFieldEncoder),
				toName:  bindingTo.toName,
			})
		}
	}
	return &structEncoder{structDescriptor.onePtrEmbedded, structDescriptor.onePtrOptimization, ***REMOVED***nalOrderedFields}, nil
}

func resolveConflictBinding(cfg *frozenCon***REMOVED***g, old, new *Binding) (ignoreOld, ignoreNew bool) {
	newTagged := new.Field.Tag.Get(cfg.getTagKey()) != ""
	oldTagged := old.Field.Tag.Get(cfg.getTagKey()) != ""
	if newTagged {
		if oldTagged {
			if len(old.levels) > len(new.levels) {
				return true, false
			} ***REMOVED*** if len(new.levels) > len(old.levels) {
				return false, true
			} ***REMOVED*** {
				return true, true
			}
		} ***REMOVED*** {
			return true, false
		}
	} ***REMOVED*** {
		if oldTagged {
			return true, false
		}
		if len(old.levels) > len(new.levels) {
			return true, false
		} ***REMOVED*** if len(new.levels) > len(old.levels) {
			return false, true
		} ***REMOVED*** {
			return true, true
		}
	}
}

func decoderOfStruct(cfg *frozenCon***REMOVED***g, typ reflect.Type) (ValDecoder, error) {
	bindings := map[string]*Binding{}
	structDescriptor, err := describeStruct(cfg, typ)
	if err != nil {
		return nil, err
	}
	for _, binding := range structDescriptor.Fields {
		for _, fromName := range binding.FromNames {
			old := bindings[fromName]
			if old == nil {
				bindings[fromName] = binding
				continue
			}
			ignoreOld, ignoreNew := resolveConflictBinding(cfg, old, binding)
			if ignoreOld {
				delete(bindings, fromName)
			}
			if !ignoreNew {
				bindings[fromName] = binding
			}
		}
	}
	***REMOVED***elds := map[string]*structFieldDecoder{}
	for k, binding := range bindings {
		***REMOVED***elds[strings.ToLower(k)] = binding.Decoder.(*structFieldDecoder)
	}
	return createStructDecoder(typ, ***REMOVED***elds)
}

type structFieldEncoder struct {
	***REMOVED***eld        *reflect.StructField
	***REMOVED***eldEncoder ValEncoder
	omitempty    bool
}

func (encoder *structFieldEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	***REMOVED***eldPtr := unsafe.Pointer(uintptr(ptr) + encoder.***REMOVED***eld.Offset)
	encoder.***REMOVED***eldEncoder.Encode(***REMOVED***eldPtr, stream)
	if stream.Error != nil && stream.Error != io.EOF {
		stream.Error = fmt.Errorf("%s: %s", encoder.***REMOVED***eld.Name, stream.Error.Error())
	}
}

func (encoder *structFieldEncoder) EncodeInterface(val interface{}, stream *Stream) {
	WriteToStream(val, stream, encoder)
}

func (encoder *structFieldEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	***REMOVED***eldPtr := unsafe.Pointer(uintptr(ptr) + encoder.***REMOVED***eld.Offset)
	return encoder.***REMOVED***eldEncoder.IsEmpty(***REMOVED***eldPtr)
}

type structEncoder struct {
	onePtrEmbedded     bool
	onePtrOptimization bool
	***REMOVED***elds             []structFieldTo
}

type structFieldTo struct {
	encoder *structFieldEncoder
	toName  string
}

func (encoder *structEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	stream.WriteObjectStart()
	isNotFirst := false
	for _, ***REMOVED***eld := range encoder.***REMOVED***elds {
		if ***REMOVED***eld.encoder.omitempty && ***REMOVED***eld.encoder.IsEmpty(ptr) {
			continue
		}
		if isNotFirst {
			stream.WriteMore()
		}
		stream.WriteObjectField(***REMOVED***eld.toName)
		***REMOVED***eld.encoder.Encode(ptr, stream)
		isNotFirst = true
	}
	stream.WriteObjectEnd()
}

func (encoder *structEncoder) EncodeInterface(val interface{}, stream *Stream) {
	e := (*emptyInterface)(unsafe.Pointer(&val))
	if encoder.onePtrOptimization {
		if e.word == nil && encoder.onePtrEmbedded {
			stream.WriteObjectStart()
			stream.WriteObjectEnd()
			return
		}
		ptr := uintptr(e.word)
		e.word = unsafe.Pointer(&ptr)
	}
	if reflect.TypeOf(val).Kind() == reflect.Ptr {
		encoder.Encode(unsafe.Pointer(&e.word), stream)
	} ***REMOVED*** {
		encoder.Encode(e.word, stream)
	}
}

func (encoder *structEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return false
}

type emptyStructEncoder struct {
}

func (encoder *emptyStructEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	stream.WriteEmptyObject()
}

func (encoder *emptyStructEncoder) EncodeInterface(val interface{}, stream *Stream) {
	WriteToStream(val, stream, encoder)
}

func (encoder *emptyStructEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return false
}
