package reflect2

import (
	"reflect"
	"unsafe"
)

type safeField struct {
	reflect.StructField
}

func (***REMOVED***eld *safeField) Offset() uintptr {
	return ***REMOVED***eld.StructField.Offset
}

func (***REMOVED***eld *safeField) Name() string {
	return ***REMOVED***eld.StructField.Name
}

func (***REMOVED***eld *safeField) PkgPath() string {
	return ***REMOVED***eld.StructField.PkgPath
}

func (***REMOVED***eld *safeField) Type() Type {
	panic("not implemented")
}

func (***REMOVED***eld *safeField) Tag() reflect.StructTag {
	return ***REMOVED***eld.StructField.Tag
}

func (***REMOVED***eld *safeField) Index() []int {
	return ***REMOVED***eld.StructField.Index
}

func (***REMOVED***eld *safeField) Anonymous() bool {
	return ***REMOVED***eld.StructField.Anonymous
}

func (***REMOVED***eld *safeField) Set(obj interface{}, value interface{}) {
	val := reflect.ValueOf(obj).Elem()
	val.FieldByIndex(***REMOVED***eld.Index()).Set(reflect.ValueOf(value).Elem())
}

func (***REMOVED***eld *safeField) UnsafeSet(obj unsafe.Pointer, value unsafe.Pointer) {
	panic("unsafe operation is not supported")
}

func (***REMOVED***eld *safeField) Get(obj interface{}) interface{} {
	val := reflect.ValueOf(obj).Elem().FieldByIndex(***REMOVED***eld.Index())
	ptr := reflect.New(val.Type())
	ptr.Elem().Set(val)
	return ptr.Interface()
}

func (***REMOVED***eld *safeField) UnsafeGet(obj unsafe.Pointer) unsafe.Pointer {
	panic("does not support unsafe operation")
}
