package jsoniter

import (
	"fmt"
	"github.com/modern-go/reflect2"
	"io"
	"reflect"
	"unsafe"
)

func encoderOfStruct(ctx *ctx, typ reflect2.Type) ValEncoder {
	type bindingTo struct {
		binding *Binding
		toName  string
		ignored bool
	}
	orderedBindings := []*bindingTo{}
	structDescriptor := describeStruct(ctx, typ)
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
				old.ignored, new.ignored = resolveConflictBinding(ctx.frozenCon***REMOVED***g, old.binding, new.binding)
			}
			orderedBindings = append(orderedBindings, new)
		}
	}
	if len(orderedBindings) == 0 {
		return &emptyStructEncoder{}
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
	return &structEncoder{typ, ***REMOVED***nalOrderedFields}
}

func createCheckIsEmpty(ctx *ctx, typ reflect2.Type) checkIsEmpty {
	encoder := createEncoderOfNative(ctx, typ)
	if encoder != nil {
		return encoder
	}
	kind := typ.Kind()
	switch kind {
	case reflect.Interface:
		return &dynamicEncoder{typ}
	case reflect.Struct:
		return &structEncoder{typ: typ}
	case reflect.Array:
		return &arrayEncoder{}
	case reflect.Slice:
		return &sliceEncoder{}
	case reflect.Map:
		return encoderOfMap(ctx, typ)
	case reflect.Ptr:
		return &OptionalEncoder{}
	default:
		return &lazyErrorEncoder{err: fmt.Errorf("unsupported type: %v", typ)}
	}
}

func resolveConflictBinding(cfg *frozenCon***REMOVED***g, old, new *Binding) (ignoreOld, ignoreNew bool) {
	newTagged := new.Field.Tag().Get(cfg.getTagKey()) != ""
	oldTagged := old.Field.Tag().Get(cfg.getTagKey()) != ""
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

type structFieldEncoder struct {
	***REMOVED***eld        reflect2.StructField
	***REMOVED***eldEncoder ValEncoder
	omitempty    bool
}

func (encoder *structFieldEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	***REMOVED***eldPtr := encoder.***REMOVED***eld.UnsafeGet(ptr)
	encoder.***REMOVED***eldEncoder.Encode(***REMOVED***eldPtr, stream)
	if stream.Error != nil && stream.Error != io.EOF {
		stream.Error = fmt.Errorf("%s: %s", encoder.***REMOVED***eld.Name(), stream.Error.Error())
	}
}

func (encoder *structFieldEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	***REMOVED***eldPtr := encoder.***REMOVED***eld.UnsafeGet(ptr)
	return encoder.***REMOVED***eldEncoder.IsEmpty(***REMOVED***eldPtr)
}

func (encoder *structFieldEncoder) IsEmbeddedPtrNil(ptr unsafe.Pointer) bool {
	isEmbeddedPtrNil, converted := encoder.***REMOVED***eldEncoder.(IsEmbeddedPtrNil)
	if !converted {
		return false
	}
	***REMOVED***eldPtr := encoder.***REMOVED***eld.UnsafeGet(ptr)
	return isEmbeddedPtrNil.IsEmbeddedPtrNil(***REMOVED***eldPtr)
}

type IsEmbeddedPtrNil interface {
	IsEmbeddedPtrNil(ptr unsafe.Pointer) bool
}

type structEncoder struct {
	typ    reflect2.Type
	***REMOVED***elds []structFieldTo
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
		if ***REMOVED***eld.encoder.IsEmbeddedPtrNil(ptr) {
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
	if stream.Error != nil && stream.Error != io.EOF {
		stream.Error = fmt.Errorf("%v.%s", encoder.typ, stream.Error.Error())
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

func (encoder *emptyStructEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return false
}

type stringModeNumberEncoder struct {
	elemEncoder ValEncoder
}

func (encoder *stringModeNumberEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	stream.writeByte('"')
	encoder.elemEncoder.Encode(ptr, stream)
	stream.writeByte('"')
}

func (encoder *stringModeNumberEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return encoder.elemEncoder.IsEmpty(ptr)
}

type stringModeStringEncoder struct {
	elemEncoder ValEncoder
	cfg         *frozenCon***REMOVED***g
}

func (encoder *stringModeStringEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	tempStream := encoder.cfg.BorrowStream(nil)
	defer encoder.cfg.ReturnStream(tempStream)
	encoder.elemEncoder.Encode(ptr, tempStream)
	stream.WriteString(string(tempStream.Buffer()))
}

func (encoder *stringModeStringEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return encoder.elemEncoder.IsEmpty(ptr)
}
