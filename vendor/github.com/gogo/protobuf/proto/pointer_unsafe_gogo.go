// Protocol Buffers for Go with Gadgets
//
// Copyright (c) 2013, The GoGo Authors. All rights reserved.
// http://github.com/gogo/protobuf
//
// Redistribution and use in source and binary forms, with or without
// modi***REMOVED***cation, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// +build !appengine,!js

// This ***REMOVED***le contains the implementation of the proto ***REMOVED***eld accesses using package unsafe.

package proto

import (
	"reflect"
	"unsafe"
)

func structPointer_InterfaceAt(p structPointer, f ***REMOVED***eld, t reflect.Type) interface{} {
	point := unsafe.Pointer(uintptr(p) + uintptr(f))
	r := reflect.NewAt(t, point)
	return r.Interface()
}

func structPointer_InterfaceRef(p structPointer, f ***REMOVED***eld, t reflect.Type) interface{} {
	point := unsafe.Pointer(uintptr(p) + uintptr(f))
	r := reflect.NewAt(t, point)
	if r.Elem().IsNil() {
		return nil
	}
	return r.Elem().Interface()
}

func copyUintPtr(oldptr, newptr uintptr, size int) {
	oldbytes := make([]byte, 0)
	oldslice := (*reflect.SliceHeader)(unsafe.Pointer(&oldbytes))
	oldslice.Data = oldptr
	oldslice.Len = size
	oldslice.Cap = size
	newbytes := make([]byte, 0)
	newslice := (*reflect.SliceHeader)(unsafe.Pointer(&newbytes))
	newslice.Data = newptr
	newslice.Len = size
	newslice.Cap = size
	copy(newbytes, oldbytes)
}

func structPointer_Copy(oldptr structPointer, newptr structPointer, size int) {
	copyUintPtr(uintptr(oldptr), uintptr(newptr), size)
}

func appendStructPointer(base structPointer, f ***REMOVED***eld, typ reflect.Type) structPointer {
	size := typ.Elem().Size()

	oldHeader := structPointer_GetSliceHeader(base, f)
	oldSlice := reflect.NewAt(typ, unsafe.Pointer(oldHeader)).Elem()
	newLen := oldHeader.Len + 1
	newSlice := reflect.MakeSlice(typ, newLen, newLen)
	reflect.Copy(newSlice, oldSlice)
	bas := toStructPointer(newSlice)
	oldHeader.Data = uintptr(bas)
	oldHeader.Len = newLen
	oldHeader.Cap = newLen

	return structPointer(unsafe.Pointer(uintptr(unsafe.Pointer(bas)) + uintptr(uintptr(newLen-1)*size)))
}

func structPointer_FieldPointer(p structPointer, f ***REMOVED***eld) structPointer {
	return structPointer(unsafe.Pointer(uintptr(p) + uintptr(f)))
}

func structPointer_GetRefStructPointer(p structPointer, f ***REMOVED***eld) structPointer {
	return structPointer((*structPointer)(unsafe.Pointer(uintptr(p) + uintptr(f))))
}

func structPointer_GetSliceHeader(p structPointer, f ***REMOVED***eld) *reflect.SliceHeader {
	return (*reflect.SliceHeader)(unsafe.Pointer(uintptr(p) + uintptr(f)))
}

func structPointer_Add(p structPointer, size ***REMOVED***eld) structPointer {
	return structPointer(unsafe.Pointer(uintptr(p) + uintptr(size)))
}

func structPointer_Len(p structPointer, f ***REMOVED***eld) int {
	return len(*(*[]interface{})(unsafe.Pointer(structPointer_GetRefStructPointer(p, f))))
}

func structPointer_StructRefSlice(p structPointer, f ***REMOVED***eld, size uintptr) *structRefSlice {
	return &structRefSlice{p: p, f: f, size: size}
}

// A structRefSlice represents a slice of structs (themselves submessages or groups).
type structRefSlice struct {
	p    structPointer
	f    ***REMOVED***eld
	size uintptr
}

func (v *structRefSlice) Len() int {
	return structPointer_Len(v.p, v.f)
}

func (v *structRefSlice) Index(i int) structPointer {
	ss := structPointer_GetStructPointer(v.p, v.f)
	ss1 := structPointer_GetRefStructPointer(ss, 0)
	return structPointer_Add(ss1, ***REMOVED***eld(uintptr(i)*v.size))
}
