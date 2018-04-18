package jsoniter

import (
	"fmt"
	"github.com/modern-go/reflect2"
	"io"
	"strings"
	"unsafe"
)

func decoderOfStruct(ctx *ctx, typ reflect2.Type) ValDecoder {
	bindings := map[string]*Binding{}
	structDescriptor := describeStruct(ctx, typ)
	for _, binding := range structDescriptor.Fields {
		for _, fromName := range binding.FromNames {
			old := bindings[fromName]
			if old == nil {
				bindings[fromName] = binding
				continue
			}
			ignoreOld, ignoreNew := resolveConflictBinding(ctx.frozenCon***REMOVED***g, old, binding)
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
		***REMOVED***elds[k] = binding.Decoder.(*structFieldDecoder)
	}
	return createStructDecoder(ctx, typ, ***REMOVED***elds)
}

func createStructDecoder(ctx *ctx, typ reflect2.Type, ***REMOVED***elds map[string]*structFieldDecoder) ValDecoder {
	if ctx.disallowUnknownFields {
		return &generalStructDecoder{typ: typ, ***REMOVED***elds: ***REMOVED***elds, disallowUnknownFields: true}
	}
	knownHash := map[int64]struct{}{
		0: {},
	}
	switch len(***REMOVED***elds) {
	case 0:
		return &skipObjectDecoder{typ}
	case 1:
		for ***REMOVED***eldName, ***REMOVED***eldDecoder := range ***REMOVED***elds {
			***REMOVED***eldHash := calcHash(***REMOVED***eldName)
			_, known := knownHash[***REMOVED***eldHash]
			if known {
				return &generalStructDecoder{typ, ***REMOVED***elds, false}
			}
			knownHash[***REMOVED***eldHash] = struct{}{}
			return &oneFieldStructDecoder{typ, ***REMOVED***eldHash, ***REMOVED***eldDecoder}
		}
	case 2:
		var ***REMOVED***eldHash1 int64
		var ***REMOVED***eldHash2 int64
		var ***REMOVED***eldDecoder1 *structFieldDecoder
		var ***REMOVED***eldDecoder2 *structFieldDecoder
		for ***REMOVED***eldName, ***REMOVED***eldDecoder := range ***REMOVED***elds {
			***REMOVED***eldHash := calcHash(***REMOVED***eldName)
			_, known := knownHash[***REMOVED***eldHash]
			if known {
				return &generalStructDecoder{typ, ***REMOVED***elds, false}
			}
			knownHash[***REMOVED***eldHash] = struct{}{}
			if ***REMOVED***eldHash1 == 0 {
				***REMOVED***eldHash1 = ***REMOVED***eldHash
				***REMOVED***eldDecoder1 = ***REMOVED***eldDecoder
			} ***REMOVED*** {
				***REMOVED***eldHash2 = ***REMOVED***eldHash
				***REMOVED***eldDecoder2 = ***REMOVED***eldDecoder
			}
		}
		return &twoFieldsStructDecoder{typ, ***REMOVED***eldHash1, ***REMOVED***eldDecoder1, ***REMOVED***eldHash2, ***REMOVED***eldDecoder2}
	case 3:
		var ***REMOVED***eldName1 int64
		var ***REMOVED***eldName2 int64
		var ***REMOVED***eldName3 int64
		var ***REMOVED***eldDecoder1 *structFieldDecoder
		var ***REMOVED***eldDecoder2 *structFieldDecoder
		var ***REMOVED***eldDecoder3 *structFieldDecoder
		for ***REMOVED***eldName, ***REMOVED***eldDecoder := range ***REMOVED***elds {
			***REMOVED***eldHash := calcHash(***REMOVED***eldName)
			_, known := knownHash[***REMOVED***eldHash]
			if known {
				return &generalStructDecoder{typ, ***REMOVED***elds, false}
			}
			knownHash[***REMOVED***eldHash] = struct{}{}
			if ***REMOVED***eldName1 == 0 {
				***REMOVED***eldName1 = ***REMOVED***eldHash
				***REMOVED***eldDecoder1 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName2 == 0 {
				***REMOVED***eldName2 = ***REMOVED***eldHash
				***REMOVED***eldDecoder2 = ***REMOVED***eldDecoder
			} ***REMOVED*** {
				***REMOVED***eldName3 = ***REMOVED***eldHash
				***REMOVED***eldDecoder3 = ***REMOVED***eldDecoder
			}
		}
		return &threeFieldsStructDecoder{typ,
			***REMOVED***eldName1, ***REMOVED***eldDecoder1,
			***REMOVED***eldName2, ***REMOVED***eldDecoder2,
			***REMOVED***eldName3, ***REMOVED***eldDecoder3}
	case 4:
		var ***REMOVED***eldName1 int64
		var ***REMOVED***eldName2 int64
		var ***REMOVED***eldName3 int64
		var ***REMOVED***eldName4 int64
		var ***REMOVED***eldDecoder1 *structFieldDecoder
		var ***REMOVED***eldDecoder2 *structFieldDecoder
		var ***REMOVED***eldDecoder3 *structFieldDecoder
		var ***REMOVED***eldDecoder4 *structFieldDecoder
		for ***REMOVED***eldName, ***REMOVED***eldDecoder := range ***REMOVED***elds {
			***REMOVED***eldHash := calcHash(***REMOVED***eldName)
			_, known := knownHash[***REMOVED***eldHash]
			if known {
				return &generalStructDecoder{typ, ***REMOVED***elds, false}
			}
			knownHash[***REMOVED***eldHash] = struct{}{}
			if ***REMOVED***eldName1 == 0 {
				***REMOVED***eldName1 = ***REMOVED***eldHash
				***REMOVED***eldDecoder1 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName2 == 0 {
				***REMOVED***eldName2 = ***REMOVED***eldHash
				***REMOVED***eldDecoder2 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName3 == 0 {
				***REMOVED***eldName3 = ***REMOVED***eldHash
				***REMOVED***eldDecoder3 = ***REMOVED***eldDecoder
			} ***REMOVED*** {
				***REMOVED***eldName4 = ***REMOVED***eldHash
				***REMOVED***eldDecoder4 = ***REMOVED***eldDecoder
			}
		}
		return &fourFieldsStructDecoder{typ,
			***REMOVED***eldName1, ***REMOVED***eldDecoder1,
			***REMOVED***eldName2, ***REMOVED***eldDecoder2,
			***REMOVED***eldName3, ***REMOVED***eldDecoder3,
			***REMOVED***eldName4, ***REMOVED***eldDecoder4}
	case 5:
		var ***REMOVED***eldName1 int64
		var ***REMOVED***eldName2 int64
		var ***REMOVED***eldName3 int64
		var ***REMOVED***eldName4 int64
		var ***REMOVED***eldName5 int64
		var ***REMOVED***eldDecoder1 *structFieldDecoder
		var ***REMOVED***eldDecoder2 *structFieldDecoder
		var ***REMOVED***eldDecoder3 *structFieldDecoder
		var ***REMOVED***eldDecoder4 *structFieldDecoder
		var ***REMOVED***eldDecoder5 *structFieldDecoder
		for ***REMOVED***eldName, ***REMOVED***eldDecoder := range ***REMOVED***elds {
			***REMOVED***eldHash := calcHash(***REMOVED***eldName)
			_, known := knownHash[***REMOVED***eldHash]
			if known {
				return &generalStructDecoder{typ, ***REMOVED***elds, false}
			}
			knownHash[***REMOVED***eldHash] = struct{}{}
			if ***REMOVED***eldName1 == 0 {
				***REMOVED***eldName1 = ***REMOVED***eldHash
				***REMOVED***eldDecoder1 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName2 == 0 {
				***REMOVED***eldName2 = ***REMOVED***eldHash
				***REMOVED***eldDecoder2 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName3 == 0 {
				***REMOVED***eldName3 = ***REMOVED***eldHash
				***REMOVED***eldDecoder3 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName4 == 0 {
				***REMOVED***eldName4 = ***REMOVED***eldHash
				***REMOVED***eldDecoder4 = ***REMOVED***eldDecoder
			} ***REMOVED*** {
				***REMOVED***eldName5 = ***REMOVED***eldHash
				***REMOVED***eldDecoder5 = ***REMOVED***eldDecoder
			}
		}
		return &***REMOVED***veFieldsStructDecoder{typ,
			***REMOVED***eldName1, ***REMOVED***eldDecoder1,
			***REMOVED***eldName2, ***REMOVED***eldDecoder2,
			***REMOVED***eldName3, ***REMOVED***eldDecoder3,
			***REMOVED***eldName4, ***REMOVED***eldDecoder4,
			***REMOVED***eldName5, ***REMOVED***eldDecoder5}
	case 6:
		var ***REMOVED***eldName1 int64
		var ***REMOVED***eldName2 int64
		var ***REMOVED***eldName3 int64
		var ***REMOVED***eldName4 int64
		var ***REMOVED***eldName5 int64
		var ***REMOVED***eldName6 int64
		var ***REMOVED***eldDecoder1 *structFieldDecoder
		var ***REMOVED***eldDecoder2 *structFieldDecoder
		var ***REMOVED***eldDecoder3 *structFieldDecoder
		var ***REMOVED***eldDecoder4 *structFieldDecoder
		var ***REMOVED***eldDecoder5 *structFieldDecoder
		var ***REMOVED***eldDecoder6 *structFieldDecoder
		for ***REMOVED***eldName, ***REMOVED***eldDecoder := range ***REMOVED***elds {
			***REMOVED***eldHash := calcHash(***REMOVED***eldName)
			_, known := knownHash[***REMOVED***eldHash]
			if known {
				return &generalStructDecoder{typ, ***REMOVED***elds, false}
			}
			knownHash[***REMOVED***eldHash] = struct{}{}
			if ***REMOVED***eldName1 == 0 {
				***REMOVED***eldName1 = ***REMOVED***eldHash
				***REMOVED***eldDecoder1 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName2 == 0 {
				***REMOVED***eldName2 = ***REMOVED***eldHash
				***REMOVED***eldDecoder2 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName3 == 0 {
				***REMOVED***eldName3 = ***REMOVED***eldHash
				***REMOVED***eldDecoder3 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName4 == 0 {
				***REMOVED***eldName4 = ***REMOVED***eldHash
				***REMOVED***eldDecoder4 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName5 == 0 {
				***REMOVED***eldName5 = ***REMOVED***eldHash
				***REMOVED***eldDecoder5 = ***REMOVED***eldDecoder
			} ***REMOVED*** {
				***REMOVED***eldName6 = ***REMOVED***eldHash
				***REMOVED***eldDecoder6 = ***REMOVED***eldDecoder
			}
		}
		return &sixFieldsStructDecoder{typ,
			***REMOVED***eldName1, ***REMOVED***eldDecoder1,
			***REMOVED***eldName2, ***REMOVED***eldDecoder2,
			***REMOVED***eldName3, ***REMOVED***eldDecoder3,
			***REMOVED***eldName4, ***REMOVED***eldDecoder4,
			***REMOVED***eldName5, ***REMOVED***eldDecoder5,
			***REMOVED***eldName6, ***REMOVED***eldDecoder6}
	case 7:
		var ***REMOVED***eldName1 int64
		var ***REMOVED***eldName2 int64
		var ***REMOVED***eldName3 int64
		var ***REMOVED***eldName4 int64
		var ***REMOVED***eldName5 int64
		var ***REMOVED***eldName6 int64
		var ***REMOVED***eldName7 int64
		var ***REMOVED***eldDecoder1 *structFieldDecoder
		var ***REMOVED***eldDecoder2 *structFieldDecoder
		var ***REMOVED***eldDecoder3 *structFieldDecoder
		var ***REMOVED***eldDecoder4 *structFieldDecoder
		var ***REMOVED***eldDecoder5 *structFieldDecoder
		var ***REMOVED***eldDecoder6 *structFieldDecoder
		var ***REMOVED***eldDecoder7 *structFieldDecoder
		for ***REMOVED***eldName, ***REMOVED***eldDecoder := range ***REMOVED***elds {
			***REMOVED***eldHash := calcHash(***REMOVED***eldName)
			_, known := knownHash[***REMOVED***eldHash]
			if known {
				return &generalStructDecoder{typ, ***REMOVED***elds, false}
			}
			knownHash[***REMOVED***eldHash] = struct{}{}
			if ***REMOVED***eldName1 == 0 {
				***REMOVED***eldName1 = ***REMOVED***eldHash
				***REMOVED***eldDecoder1 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName2 == 0 {
				***REMOVED***eldName2 = ***REMOVED***eldHash
				***REMOVED***eldDecoder2 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName3 == 0 {
				***REMOVED***eldName3 = ***REMOVED***eldHash
				***REMOVED***eldDecoder3 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName4 == 0 {
				***REMOVED***eldName4 = ***REMOVED***eldHash
				***REMOVED***eldDecoder4 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName5 == 0 {
				***REMOVED***eldName5 = ***REMOVED***eldHash
				***REMOVED***eldDecoder5 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName6 == 0 {
				***REMOVED***eldName6 = ***REMOVED***eldHash
				***REMOVED***eldDecoder6 = ***REMOVED***eldDecoder
			} ***REMOVED*** {
				***REMOVED***eldName7 = ***REMOVED***eldHash
				***REMOVED***eldDecoder7 = ***REMOVED***eldDecoder
			}
		}
		return &sevenFieldsStructDecoder{typ,
			***REMOVED***eldName1, ***REMOVED***eldDecoder1,
			***REMOVED***eldName2, ***REMOVED***eldDecoder2,
			***REMOVED***eldName3, ***REMOVED***eldDecoder3,
			***REMOVED***eldName4, ***REMOVED***eldDecoder4,
			***REMOVED***eldName5, ***REMOVED***eldDecoder5,
			***REMOVED***eldName6, ***REMOVED***eldDecoder6,
			***REMOVED***eldName7, ***REMOVED***eldDecoder7}
	case 8:
		var ***REMOVED***eldName1 int64
		var ***REMOVED***eldName2 int64
		var ***REMOVED***eldName3 int64
		var ***REMOVED***eldName4 int64
		var ***REMOVED***eldName5 int64
		var ***REMOVED***eldName6 int64
		var ***REMOVED***eldName7 int64
		var ***REMOVED***eldName8 int64
		var ***REMOVED***eldDecoder1 *structFieldDecoder
		var ***REMOVED***eldDecoder2 *structFieldDecoder
		var ***REMOVED***eldDecoder3 *structFieldDecoder
		var ***REMOVED***eldDecoder4 *structFieldDecoder
		var ***REMOVED***eldDecoder5 *structFieldDecoder
		var ***REMOVED***eldDecoder6 *structFieldDecoder
		var ***REMOVED***eldDecoder7 *structFieldDecoder
		var ***REMOVED***eldDecoder8 *structFieldDecoder
		for ***REMOVED***eldName, ***REMOVED***eldDecoder := range ***REMOVED***elds {
			***REMOVED***eldHash := calcHash(***REMOVED***eldName)
			_, known := knownHash[***REMOVED***eldHash]
			if known {
				return &generalStructDecoder{typ, ***REMOVED***elds, false}
			}
			knownHash[***REMOVED***eldHash] = struct{}{}
			if ***REMOVED***eldName1 == 0 {
				***REMOVED***eldName1 = ***REMOVED***eldHash
				***REMOVED***eldDecoder1 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName2 == 0 {
				***REMOVED***eldName2 = ***REMOVED***eldHash
				***REMOVED***eldDecoder2 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName3 == 0 {
				***REMOVED***eldName3 = ***REMOVED***eldHash
				***REMOVED***eldDecoder3 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName4 == 0 {
				***REMOVED***eldName4 = ***REMOVED***eldHash
				***REMOVED***eldDecoder4 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName5 == 0 {
				***REMOVED***eldName5 = ***REMOVED***eldHash
				***REMOVED***eldDecoder5 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName6 == 0 {
				***REMOVED***eldName6 = ***REMOVED***eldHash
				***REMOVED***eldDecoder6 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName7 == 0 {
				***REMOVED***eldName7 = ***REMOVED***eldHash
				***REMOVED***eldDecoder7 = ***REMOVED***eldDecoder
			} ***REMOVED*** {
				***REMOVED***eldName8 = ***REMOVED***eldHash
				***REMOVED***eldDecoder8 = ***REMOVED***eldDecoder
			}
		}
		return &eightFieldsStructDecoder{typ,
			***REMOVED***eldName1, ***REMOVED***eldDecoder1,
			***REMOVED***eldName2, ***REMOVED***eldDecoder2,
			***REMOVED***eldName3, ***REMOVED***eldDecoder3,
			***REMOVED***eldName4, ***REMOVED***eldDecoder4,
			***REMOVED***eldName5, ***REMOVED***eldDecoder5,
			***REMOVED***eldName6, ***REMOVED***eldDecoder6,
			***REMOVED***eldName7, ***REMOVED***eldDecoder7,
			***REMOVED***eldName8, ***REMOVED***eldDecoder8}
	case 9:
		var ***REMOVED***eldName1 int64
		var ***REMOVED***eldName2 int64
		var ***REMOVED***eldName3 int64
		var ***REMOVED***eldName4 int64
		var ***REMOVED***eldName5 int64
		var ***REMOVED***eldName6 int64
		var ***REMOVED***eldName7 int64
		var ***REMOVED***eldName8 int64
		var ***REMOVED***eldName9 int64
		var ***REMOVED***eldDecoder1 *structFieldDecoder
		var ***REMOVED***eldDecoder2 *structFieldDecoder
		var ***REMOVED***eldDecoder3 *structFieldDecoder
		var ***REMOVED***eldDecoder4 *structFieldDecoder
		var ***REMOVED***eldDecoder5 *structFieldDecoder
		var ***REMOVED***eldDecoder6 *structFieldDecoder
		var ***REMOVED***eldDecoder7 *structFieldDecoder
		var ***REMOVED***eldDecoder8 *structFieldDecoder
		var ***REMOVED***eldDecoder9 *structFieldDecoder
		for ***REMOVED***eldName, ***REMOVED***eldDecoder := range ***REMOVED***elds {
			***REMOVED***eldHash := calcHash(***REMOVED***eldName)
			_, known := knownHash[***REMOVED***eldHash]
			if known {
				return &generalStructDecoder{typ, ***REMOVED***elds, false}
			}
			knownHash[***REMOVED***eldHash] = struct{}{}
			if ***REMOVED***eldName1 == 0 {
				***REMOVED***eldName1 = ***REMOVED***eldHash
				***REMOVED***eldDecoder1 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName2 == 0 {
				***REMOVED***eldName2 = ***REMOVED***eldHash
				***REMOVED***eldDecoder2 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName3 == 0 {
				***REMOVED***eldName3 = ***REMOVED***eldHash
				***REMOVED***eldDecoder3 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName4 == 0 {
				***REMOVED***eldName4 = ***REMOVED***eldHash
				***REMOVED***eldDecoder4 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName5 == 0 {
				***REMOVED***eldName5 = ***REMOVED***eldHash
				***REMOVED***eldDecoder5 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName6 == 0 {
				***REMOVED***eldName6 = ***REMOVED***eldHash
				***REMOVED***eldDecoder6 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName7 == 0 {
				***REMOVED***eldName7 = ***REMOVED***eldHash
				***REMOVED***eldDecoder7 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName8 == 0 {
				***REMOVED***eldName8 = ***REMOVED***eldHash
				***REMOVED***eldDecoder8 = ***REMOVED***eldDecoder
			} ***REMOVED*** {
				***REMOVED***eldName9 = ***REMOVED***eldHash
				***REMOVED***eldDecoder9 = ***REMOVED***eldDecoder
			}
		}
		return &nineFieldsStructDecoder{typ,
			***REMOVED***eldName1, ***REMOVED***eldDecoder1,
			***REMOVED***eldName2, ***REMOVED***eldDecoder2,
			***REMOVED***eldName3, ***REMOVED***eldDecoder3,
			***REMOVED***eldName4, ***REMOVED***eldDecoder4,
			***REMOVED***eldName5, ***REMOVED***eldDecoder5,
			***REMOVED***eldName6, ***REMOVED***eldDecoder6,
			***REMOVED***eldName7, ***REMOVED***eldDecoder7,
			***REMOVED***eldName8, ***REMOVED***eldDecoder8,
			***REMOVED***eldName9, ***REMOVED***eldDecoder9}
	case 10:
		var ***REMOVED***eldName1 int64
		var ***REMOVED***eldName2 int64
		var ***REMOVED***eldName3 int64
		var ***REMOVED***eldName4 int64
		var ***REMOVED***eldName5 int64
		var ***REMOVED***eldName6 int64
		var ***REMOVED***eldName7 int64
		var ***REMOVED***eldName8 int64
		var ***REMOVED***eldName9 int64
		var ***REMOVED***eldName10 int64
		var ***REMOVED***eldDecoder1 *structFieldDecoder
		var ***REMOVED***eldDecoder2 *structFieldDecoder
		var ***REMOVED***eldDecoder3 *structFieldDecoder
		var ***REMOVED***eldDecoder4 *structFieldDecoder
		var ***REMOVED***eldDecoder5 *structFieldDecoder
		var ***REMOVED***eldDecoder6 *structFieldDecoder
		var ***REMOVED***eldDecoder7 *structFieldDecoder
		var ***REMOVED***eldDecoder8 *structFieldDecoder
		var ***REMOVED***eldDecoder9 *structFieldDecoder
		var ***REMOVED***eldDecoder10 *structFieldDecoder
		for ***REMOVED***eldName, ***REMOVED***eldDecoder := range ***REMOVED***elds {
			***REMOVED***eldHash := calcHash(***REMOVED***eldName)
			_, known := knownHash[***REMOVED***eldHash]
			if known {
				return &generalStructDecoder{typ, ***REMOVED***elds, false}
			}
			knownHash[***REMOVED***eldHash] = struct{}{}
			if ***REMOVED***eldName1 == 0 {
				***REMOVED***eldName1 = ***REMOVED***eldHash
				***REMOVED***eldDecoder1 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName2 == 0 {
				***REMOVED***eldName2 = ***REMOVED***eldHash
				***REMOVED***eldDecoder2 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName3 == 0 {
				***REMOVED***eldName3 = ***REMOVED***eldHash
				***REMOVED***eldDecoder3 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName4 == 0 {
				***REMOVED***eldName4 = ***REMOVED***eldHash
				***REMOVED***eldDecoder4 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName5 == 0 {
				***REMOVED***eldName5 = ***REMOVED***eldHash
				***REMOVED***eldDecoder5 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName6 == 0 {
				***REMOVED***eldName6 = ***REMOVED***eldHash
				***REMOVED***eldDecoder6 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName7 == 0 {
				***REMOVED***eldName7 = ***REMOVED***eldHash
				***REMOVED***eldDecoder7 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName8 == 0 {
				***REMOVED***eldName8 = ***REMOVED***eldHash
				***REMOVED***eldDecoder8 = ***REMOVED***eldDecoder
			} ***REMOVED*** if ***REMOVED***eldName9 == 0 {
				***REMOVED***eldName9 = ***REMOVED***eldHash
				***REMOVED***eldDecoder9 = ***REMOVED***eldDecoder
			} ***REMOVED*** {
				***REMOVED***eldName10 = ***REMOVED***eldHash
				***REMOVED***eldDecoder10 = ***REMOVED***eldDecoder
			}
		}
		return &tenFieldsStructDecoder{typ,
			***REMOVED***eldName1, ***REMOVED***eldDecoder1,
			***REMOVED***eldName2, ***REMOVED***eldDecoder2,
			***REMOVED***eldName3, ***REMOVED***eldDecoder3,
			***REMOVED***eldName4, ***REMOVED***eldDecoder4,
			***REMOVED***eldName5, ***REMOVED***eldDecoder5,
			***REMOVED***eldName6, ***REMOVED***eldDecoder6,
			***REMOVED***eldName7, ***REMOVED***eldDecoder7,
			***REMOVED***eldName8, ***REMOVED***eldDecoder8,
			***REMOVED***eldName9, ***REMOVED***eldDecoder9,
			***REMOVED***eldName10, ***REMOVED***eldDecoder10}
	}
	return &generalStructDecoder{typ, ***REMOVED***elds, false}
}

type generalStructDecoder struct {
	typ                   reflect2.Type
	***REMOVED***elds                map[string]*structFieldDecoder
	disallowUnknownFields bool
}

func (decoder *generalStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	decoder.decodeOneField(ptr, iter)
	for iter.nextToken() == ',' {
		decoder.decodeOneField(ptr, iter)
	}
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%v.%s", decoder.typ, iter.Error.Error())
	}
}

func (decoder *generalStructDecoder) decodeOneField(ptr unsafe.Pointer, iter *Iterator) {
	var ***REMOVED***eld string
	var ***REMOVED***eldDecoder *structFieldDecoder
	if iter.cfg.objectFieldMustBeSimpleString {
		***REMOVED***eldBytes := iter.ReadStringAsSlice()
		***REMOVED***eld = *(*string)(unsafe.Pointer(&***REMOVED***eldBytes))
		***REMOVED***eldDecoder = decoder.***REMOVED***elds[***REMOVED***eld]
		if ***REMOVED***eldDecoder == nil {
			***REMOVED***eldDecoder = decoder.***REMOVED***elds[strings.ToLower(***REMOVED***eld)]
		}
	} ***REMOVED*** {
		***REMOVED***eld = iter.ReadString()
		***REMOVED***eldDecoder = decoder.***REMOVED***elds[***REMOVED***eld]
		if ***REMOVED***eldDecoder == nil {
			***REMOVED***eldDecoder = decoder.***REMOVED***elds[strings.ToLower(***REMOVED***eld)]
		}
	}
	if ***REMOVED***eldDecoder == nil {
		msg := "found unknown ***REMOVED***eld: " + ***REMOVED***eld
		if decoder.disallowUnknownFields {
			iter.ReportError("ReadObject", msg)
		}
		c := iter.nextToken()
		if c != ':' {
			iter.ReportError("ReadObject", "expect : after object ***REMOVED***eld, but found "+string([]byte{c}))
		}
		iter.Skip()
		return
	}
	c := iter.nextToken()
	if c != ':' {
		iter.ReportError("ReadObject", "expect : after object ***REMOVED***eld, but found "+string([]byte{c}))
	}
	***REMOVED***eldDecoder.Decode(ptr, iter)
}

type skipObjectDecoder struct {
	typ reflect2.Type
}

func (decoder *skipObjectDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	valueType := iter.WhatIsNext()
	if valueType != ObjectValue && valueType != NilValue {
		iter.ReportError("skipObjectDecoder", "expect object or null")
		return
	}
	iter.Skip()
}

type oneFieldStructDecoder struct {
	typ          reflect2.Type
	***REMOVED***eldHash    int64
	***REMOVED***eldDecoder *structFieldDecoder
}

func (decoder *oneFieldStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	for {
		if iter.readFieldHash() == decoder.***REMOVED***eldHash {
			decoder.***REMOVED***eldDecoder.Decode(ptr, iter)
		} ***REMOVED*** {
			iter.Skip()
		}
		if iter.isObjectEnd() {
			break
		}
	}
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%v.%s", decoder.typ, iter.Error.Error())
	}
}

type twoFieldsStructDecoder struct {
	typ           reflect2.Type
	***REMOVED***eldHash1    int64
	***REMOVED***eldDecoder1 *structFieldDecoder
	***REMOVED***eldHash2    int64
	***REMOVED***eldDecoder2 *structFieldDecoder
}

func (decoder *twoFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case decoder.***REMOVED***eldHash1:
			decoder.***REMOVED***eldDecoder1.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash2:
			decoder.***REMOVED***eldDecoder2.Decode(ptr, iter)
		default:
			iter.Skip()
		}
		if iter.isObjectEnd() {
			break
		}
	}
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%v.%s", decoder.typ, iter.Error.Error())
	}
}

type threeFieldsStructDecoder struct {
	typ           reflect2.Type
	***REMOVED***eldHash1    int64
	***REMOVED***eldDecoder1 *structFieldDecoder
	***REMOVED***eldHash2    int64
	***REMOVED***eldDecoder2 *structFieldDecoder
	***REMOVED***eldHash3    int64
	***REMOVED***eldDecoder3 *structFieldDecoder
}

func (decoder *threeFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case decoder.***REMOVED***eldHash1:
			decoder.***REMOVED***eldDecoder1.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash2:
			decoder.***REMOVED***eldDecoder2.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash3:
			decoder.***REMOVED***eldDecoder3.Decode(ptr, iter)
		default:
			iter.Skip()
		}
		if iter.isObjectEnd() {
			break
		}
	}
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%v.%s", decoder.typ, iter.Error.Error())
	}
}

type fourFieldsStructDecoder struct {
	typ           reflect2.Type
	***REMOVED***eldHash1    int64
	***REMOVED***eldDecoder1 *structFieldDecoder
	***REMOVED***eldHash2    int64
	***REMOVED***eldDecoder2 *structFieldDecoder
	***REMOVED***eldHash3    int64
	***REMOVED***eldDecoder3 *structFieldDecoder
	***REMOVED***eldHash4    int64
	***REMOVED***eldDecoder4 *structFieldDecoder
}

func (decoder *fourFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case decoder.***REMOVED***eldHash1:
			decoder.***REMOVED***eldDecoder1.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash2:
			decoder.***REMOVED***eldDecoder2.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash3:
			decoder.***REMOVED***eldDecoder3.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash4:
			decoder.***REMOVED***eldDecoder4.Decode(ptr, iter)
		default:
			iter.Skip()
		}
		if iter.isObjectEnd() {
			break
		}
	}
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%v.%s", decoder.typ, iter.Error.Error())
	}
}

type ***REMOVED***veFieldsStructDecoder struct {
	typ           reflect2.Type
	***REMOVED***eldHash1    int64
	***REMOVED***eldDecoder1 *structFieldDecoder
	***REMOVED***eldHash2    int64
	***REMOVED***eldDecoder2 *structFieldDecoder
	***REMOVED***eldHash3    int64
	***REMOVED***eldDecoder3 *structFieldDecoder
	***REMOVED***eldHash4    int64
	***REMOVED***eldDecoder4 *structFieldDecoder
	***REMOVED***eldHash5    int64
	***REMOVED***eldDecoder5 *structFieldDecoder
}

func (decoder ****REMOVED***veFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case decoder.***REMOVED***eldHash1:
			decoder.***REMOVED***eldDecoder1.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash2:
			decoder.***REMOVED***eldDecoder2.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash3:
			decoder.***REMOVED***eldDecoder3.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash4:
			decoder.***REMOVED***eldDecoder4.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash5:
			decoder.***REMOVED***eldDecoder5.Decode(ptr, iter)
		default:
			iter.Skip()
		}
		if iter.isObjectEnd() {
			break
		}
	}
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%v.%s", decoder.typ, iter.Error.Error())
	}
}

type sixFieldsStructDecoder struct {
	typ           reflect2.Type
	***REMOVED***eldHash1    int64
	***REMOVED***eldDecoder1 *structFieldDecoder
	***REMOVED***eldHash2    int64
	***REMOVED***eldDecoder2 *structFieldDecoder
	***REMOVED***eldHash3    int64
	***REMOVED***eldDecoder3 *structFieldDecoder
	***REMOVED***eldHash4    int64
	***REMOVED***eldDecoder4 *structFieldDecoder
	***REMOVED***eldHash5    int64
	***REMOVED***eldDecoder5 *structFieldDecoder
	***REMOVED***eldHash6    int64
	***REMOVED***eldDecoder6 *structFieldDecoder
}

func (decoder *sixFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case decoder.***REMOVED***eldHash1:
			decoder.***REMOVED***eldDecoder1.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash2:
			decoder.***REMOVED***eldDecoder2.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash3:
			decoder.***REMOVED***eldDecoder3.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash4:
			decoder.***REMOVED***eldDecoder4.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash5:
			decoder.***REMOVED***eldDecoder5.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash6:
			decoder.***REMOVED***eldDecoder6.Decode(ptr, iter)
		default:
			iter.Skip()
		}
		if iter.isObjectEnd() {
			break
		}
	}
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%v.%s", decoder.typ, iter.Error.Error())
	}
}

type sevenFieldsStructDecoder struct {
	typ           reflect2.Type
	***REMOVED***eldHash1    int64
	***REMOVED***eldDecoder1 *structFieldDecoder
	***REMOVED***eldHash2    int64
	***REMOVED***eldDecoder2 *structFieldDecoder
	***REMOVED***eldHash3    int64
	***REMOVED***eldDecoder3 *structFieldDecoder
	***REMOVED***eldHash4    int64
	***REMOVED***eldDecoder4 *structFieldDecoder
	***REMOVED***eldHash5    int64
	***REMOVED***eldDecoder5 *structFieldDecoder
	***REMOVED***eldHash6    int64
	***REMOVED***eldDecoder6 *structFieldDecoder
	***REMOVED***eldHash7    int64
	***REMOVED***eldDecoder7 *structFieldDecoder
}

func (decoder *sevenFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case decoder.***REMOVED***eldHash1:
			decoder.***REMOVED***eldDecoder1.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash2:
			decoder.***REMOVED***eldDecoder2.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash3:
			decoder.***REMOVED***eldDecoder3.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash4:
			decoder.***REMOVED***eldDecoder4.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash5:
			decoder.***REMOVED***eldDecoder5.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash6:
			decoder.***REMOVED***eldDecoder6.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash7:
			decoder.***REMOVED***eldDecoder7.Decode(ptr, iter)
		default:
			iter.Skip()
		}
		if iter.isObjectEnd() {
			break
		}
	}
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%v.%s", decoder.typ, iter.Error.Error())
	}
}

type eightFieldsStructDecoder struct {
	typ           reflect2.Type
	***REMOVED***eldHash1    int64
	***REMOVED***eldDecoder1 *structFieldDecoder
	***REMOVED***eldHash2    int64
	***REMOVED***eldDecoder2 *structFieldDecoder
	***REMOVED***eldHash3    int64
	***REMOVED***eldDecoder3 *structFieldDecoder
	***REMOVED***eldHash4    int64
	***REMOVED***eldDecoder4 *structFieldDecoder
	***REMOVED***eldHash5    int64
	***REMOVED***eldDecoder5 *structFieldDecoder
	***REMOVED***eldHash6    int64
	***REMOVED***eldDecoder6 *structFieldDecoder
	***REMOVED***eldHash7    int64
	***REMOVED***eldDecoder7 *structFieldDecoder
	***REMOVED***eldHash8    int64
	***REMOVED***eldDecoder8 *structFieldDecoder
}

func (decoder *eightFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case decoder.***REMOVED***eldHash1:
			decoder.***REMOVED***eldDecoder1.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash2:
			decoder.***REMOVED***eldDecoder2.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash3:
			decoder.***REMOVED***eldDecoder3.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash4:
			decoder.***REMOVED***eldDecoder4.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash5:
			decoder.***REMOVED***eldDecoder5.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash6:
			decoder.***REMOVED***eldDecoder6.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash7:
			decoder.***REMOVED***eldDecoder7.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash8:
			decoder.***REMOVED***eldDecoder8.Decode(ptr, iter)
		default:
			iter.Skip()
		}
		if iter.isObjectEnd() {
			break
		}
	}
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%v.%s", decoder.typ, iter.Error.Error())
	}
}

type nineFieldsStructDecoder struct {
	typ           reflect2.Type
	***REMOVED***eldHash1    int64
	***REMOVED***eldDecoder1 *structFieldDecoder
	***REMOVED***eldHash2    int64
	***REMOVED***eldDecoder2 *structFieldDecoder
	***REMOVED***eldHash3    int64
	***REMOVED***eldDecoder3 *structFieldDecoder
	***REMOVED***eldHash4    int64
	***REMOVED***eldDecoder4 *structFieldDecoder
	***REMOVED***eldHash5    int64
	***REMOVED***eldDecoder5 *structFieldDecoder
	***REMOVED***eldHash6    int64
	***REMOVED***eldDecoder6 *structFieldDecoder
	***REMOVED***eldHash7    int64
	***REMOVED***eldDecoder7 *structFieldDecoder
	***REMOVED***eldHash8    int64
	***REMOVED***eldDecoder8 *structFieldDecoder
	***REMOVED***eldHash9    int64
	***REMOVED***eldDecoder9 *structFieldDecoder
}

func (decoder *nineFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case decoder.***REMOVED***eldHash1:
			decoder.***REMOVED***eldDecoder1.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash2:
			decoder.***REMOVED***eldDecoder2.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash3:
			decoder.***REMOVED***eldDecoder3.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash4:
			decoder.***REMOVED***eldDecoder4.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash5:
			decoder.***REMOVED***eldDecoder5.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash6:
			decoder.***REMOVED***eldDecoder6.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash7:
			decoder.***REMOVED***eldDecoder7.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash8:
			decoder.***REMOVED***eldDecoder8.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash9:
			decoder.***REMOVED***eldDecoder9.Decode(ptr, iter)
		default:
			iter.Skip()
		}
		if iter.isObjectEnd() {
			break
		}
	}
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%v.%s", decoder.typ, iter.Error.Error())
	}
}

type tenFieldsStructDecoder struct {
	typ            reflect2.Type
	***REMOVED***eldHash1     int64
	***REMOVED***eldDecoder1  *structFieldDecoder
	***REMOVED***eldHash2     int64
	***REMOVED***eldDecoder2  *structFieldDecoder
	***REMOVED***eldHash3     int64
	***REMOVED***eldDecoder3  *structFieldDecoder
	***REMOVED***eldHash4     int64
	***REMOVED***eldDecoder4  *structFieldDecoder
	***REMOVED***eldHash5     int64
	***REMOVED***eldDecoder5  *structFieldDecoder
	***REMOVED***eldHash6     int64
	***REMOVED***eldDecoder6  *structFieldDecoder
	***REMOVED***eldHash7     int64
	***REMOVED***eldDecoder7  *structFieldDecoder
	***REMOVED***eldHash8     int64
	***REMOVED***eldDecoder8  *structFieldDecoder
	***REMOVED***eldHash9     int64
	***REMOVED***eldDecoder9  *structFieldDecoder
	***REMOVED***eldHash10    int64
	***REMOVED***eldDecoder10 *structFieldDecoder
}

func (decoder *tenFieldsStructDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	if !iter.readObjectStart() {
		return
	}
	for {
		switch iter.readFieldHash() {
		case decoder.***REMOVED***eldHash1:
			decoder.***REMOVED***eldDecoder1.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash2:
			decoder.***REMOVED***eldDecoder2.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash3:
			decoder.***REMOVED***eldDecoder3.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash4:
			decoder.***REMOVED***eldDecoder4.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash5:
			decoder.***REMOVED***eldDecoder5.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash6:
			decoder.***REMOVED***eldDecoder6.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash7:
			decoder.***REMOVED***eldDecoder7.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash8:
			decoder.***REMOVED***eldDecoder8.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash9:
			decoder.***REMOVED***eldDecoder9.Decode(ptr, iter)
		case decoder.***REMOVED***eldHash10:
			decoder.***REMOVED***eldDecoder10.Decode(ptr, iter)
		default:
			iter.Skip()
		}
		if iter.isObjectEnd() {
			break
		}
	}
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%v.%s", decoder.typ, iter.Error.Error())
	}
}

type structFieldDecoder struct {
	***REMOVED***eld        reflect2.StructField
	***REMOVED***eldDecoder ValDecoder
}

func (decoder *structFieldDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	***REMOVED***eldPtr := decoder.***REMOVED***eld.UnsafeGet(ptr)
	decoder.***REMOVED***eldDecoder.Decode(***REMOVED***eldPtr, iter)
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%s: %s", decoder.***REMOVED***eld.Name(), iter.Error.Error())
	}
}

type stringModeStringDecoder struct {
	elemDecoder ValDecoder
	cfg         *frozenCon***REMOVED***g
}

func (decoder *stringModeStringDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	decoder.elemDecoder.Decode(ptr, iter)
	str := *((*string)(ptr))
	tempIter := decoder.cfg.BorrowIterator([]byte(str))
	defer decoder.cfg.ReturnIterator(tempIter)
	*((*string)(ptr)) = tempIter.ReadString()
}

type stringModeNumberDecoder struct {
	elemDecoder ValDecoder
}

func (decoder *stringModeNumberDecoder) Decode(ptr unsafe.Pointer, iter *Iterator) {
	c := iter.nextToken()
	if c != '"' {
		iter.ReportError("stringModeNumberDecoder", `expect ", but found `+string([]byte{c}))
		return
	}
	decoder.elemDecoder.Decode(ptr, iter)
	if iter.Error != nil {
		return
	}
	c = iter.readByte()
	if c != '"' {
		iter.ReportError("stringModeNumberDecoder", `expect ", but found `+string([]byte{c}))
		return
	}
}
