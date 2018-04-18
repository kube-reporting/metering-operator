package reflect2

import (
	"reflect"
	"unsafe"
)

type UnsafeStructField struct {
	reflect.StructField
	structType *UnsafeStructType
	rtype      unsafe.Pointer
	ptrRType   unsafe.Pointer
}

func newUnsafeStructField(structType *UnsafeStructType, structField reflect.StructField) *UnsafeStructField {
	return &UnsafeStructField{
		StructField: structField,
		rtype:       unpackEFace(structField.Type).data,
		ptrRType:    unpackEFace(reflect.PtrTo(structField.Type)).data,
		structType:  structType,
	}
}

func (***REMOVED***eld *UnsafeStructField) Offset() uintptr {
	return ***REMOVED***eld.StructField.Offset
}

func (***REMOVED***eld *UnsafeStructField) Name() string {
	return ***REMOVED***eld.StructField.Name
}

func (***REMOVED***eld *UnsafeStructField) PkgPath() string {
	return ***REMOVED***eld.StructField.PkgPath
}

func (***REMOVED***eld *UnsafeStructField) Type() Type {
	return ***REMOVED***eld.structType.cfg.Type2(***REMOVED***eld.StructField.Type)
}

func (***REMOVED***eld *UnsafeStructField) Tag() reflect.StructTag {
	return ***REMOVED***eld.StructField.Tag
}

func (***REMOVED***eld *UnsafeStructField) Index() []int {
	return ***REMOVED***eld.StructField.Index
}

func (***REMOVED***eld *UnsafeStructField) Anonymous() bool {
	return ***REMOVED***eld.StructField.Anonymous
}

func (***REMOVED***eld *UnsafeStructField) Set(obj interface{}, value interface{}) {
	objEFace := unpackEFace(obj)
	assertType("StructField.SetIndex argument 1", ***REMOVED***eld.structType.ptrRType, objEFace.rtype)
	valueEFace := unpackEFace(value)
	assertType("StructField.SetIndex argument 2", ***REMOVED***eld.ptrRType, valueEFace.rtype)
	***REMOVED***eld.UnsafeSet(objEFace.data, valueEFace.data)
}

func (***REMOVED***eld *UnsafeStructField) UnsafeSet(obj unsafe.Pointer, value unsafe.Pointer) {
	***REMOVED***eldPtr := add(obj, ***REMOVED***eld.StructField.Offset, "same as non-reflect &v.***REMOVED***eld")
	typedmemmove(***REMOVED***eld.rtype, ***REMOVED***eldPtr, value)
}

func (***REMOVED***eld *UnsafeStructField) Get(obj interface{}) interface{} {
	objEFace := unpackEFace(obj)
	assertType("StructField.GetIndex argument 1", ***REMOVED***eld.structType.ptrRType, objEFace.rtype)
	value := ***REMOVED***eld.UnsafeGet(objEFace.data)
	return packEFace(***REMOVED***eld.ptrRType, value)
}

func (***REMOVED***eld *UnsafeStructField) UnsafeGet(obj unsafe.Pointer) unsafe.Pointer {
	return add(obj, ***REMOVED***eld.StructField.Offset, "same as non-reflect &v.***REMOVED***eld")
}
