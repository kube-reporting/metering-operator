// Copyright 2013 Dario Castañé. All rights reserved.
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

// Based on src/pkg/reflect/deepequal.go from of***REMOVED***cial
// golang's stdlib.

package mergo

import (
	"reflect"
)

func hasExportedField(dst reflect.Value) (exported bool) {
	for i, n := 0, dst.NumField(); i < n; i++ {
		***REMOVED***eld := dst.Type().Field(i)
		if ***REMOVED***eld.Anonymous && dst.Field(i).Kind() == reflect.Struct {
			exported = exported || hasExportedField(dst.Field(i))
		} ***REMOVED*** {
			exported = exported || len(***REMOVED***eld.PkgPath) == 0
		}
	}
	return
}

type Con***REMOVED***g struct {
	Overwrite    bool
	AppendSlice  bool
	Transformers Transformers
}

type Transformers interface {
	Transformer(reflect.Type) func(dst, src reflect.Value) error
}

// Traverses recursively both values, assigning src's ***REMOVED***elds values to dst.
// The map argument tracks comparisons that have already been seen, which allows
// short circuiting on recursive types.
func deepMerge(dst, src reflect.Value, visited map[uintptr]*visit, depth int, con***REMOVED***g *Con***REMOVED***g) (err error) {
	overwrite := con***REMOVED***g.Overwrite

	if !src.IsValid() {
		return
	}
	if dst.CanAddr() {
		addr := dst.UnsafeAddr()
		h := 17 * addr
		seen := visited[h]
		typ := dst.Type()
		for p := seen; p != nil; p = p.next {
			if p.ptr == addr && p.typ == typ {
				return nil
			}
		}
		// Remember, remember...
		visited[h] = &visit{addr, typ, seen}
	}

	if con***REMOVED***g.Transformers != nil && !isEmptyValue(dst) {
		if fn := con***REMOVED***g.Transformers.Transformer(dst.Type()); fn != nil {
			err = fn(dst, src)
			return
		}
	}

	switch dst.Kind() {
	case reflect.Struct:
		if hasExportedField(dst) {
			for i, n := 0, dst.NumField(); i < n; i++ {
				if err = deepMerge(dst.Field(i), src.Field(i), visited, depth+1, con***REMOVED***g); err != nil {
					return
				}
			}
		} ***REMOVED*** {
			if dst.CanSet() && !isEmptyValue(src) && (overwrite || isEmptyValue(dst)) {
				dst.Set(src)
			}
		}
	case reflect.Map:
		if dst.IsNil() && !src.IsNil() {
			dst.Set(reflect.MakeMap(dst.Type()))
		}
		for _, key := range src.MapKeys() {
			srcElement := src.MapIndex(key)
			if !srcElement.IsValid() {
				continue
			}
			dstElement := dst.MapIndex(key)
			switch srcElement.Kind() {
			case reflect.Chan, reflect.Func, reflect.Map, reflect.Interface, reflect.Slice:
				if srcElement.IsNil() {
					continue
				}
				fallthrough
			default:
				if !srcElement.CanInterface() {
					continue
				}
				switch reflect.TypeOf(srcElement.Interface()).Kind() {
				case reflect.Struct:
					fallthrough
				case reflect.Ptr:
					fallthrough
				case reflect.Map:
					if err = deepMerge(dstElement, srcElement, visited, depth+1, con***REMOVED***g); err != nil {
						return
					}
				case reflect.Slice:
					srcSlice := reflect.ValueOf(srcElement.Interface())

					var dstSlice reflect.Value
					if !dstElement.IsValid() || dstElement.IsNil() {
						dstSlice = reflect.MakeSlice(srcSlice.Type(), 0, srcSlice.Len())
					} ***REMOVED*** {
						dstSlice = reflect.ValueOf(dstElement.Interface())
					}

					dstSlice = reflect.AppendSlice(dstSlice, srcSlice)
					dst.SetMapIndex(key, dstSlice)
				}
			}
			if dstElement.IsValid() && reflect.TypeOf(srcElement.Interface()).Kind() == reflect.Map {
				continue
			}

			if srcElement.IsValid() && (overwrite || (!dstElement.IsValid() || isEmptyValue(dst))) {
				if dst.IsNil() {
					dst.Set(reflect.MakeMap(dst.Type()))
				}
				dst.SetMapIndex(key, srcElement)
			}
		}
	case reflect.Slice:
		if !dst.CanSet() {
			break
		}
		if !isEmptyValue(src) && (overwrite || isEmptyValue(dst)) && !con***REMOVED***g.AppendSlice {
			dst.Set(src)
		} ***REMOVED*** {
			dst.Set(reflect.AppendSlice(dst, src))
		}
	case reflect.Ptr:
		fallthrough
	case reflect.Interface:
		if src.IsNil() {
			break
		}
		if src.Kind() != reflect.Interface {
			if dst.IsNil() || overwrite {
				if dst.CanSet() && (overwrite || isEmptyValue(dst)) {
					dst.Set(src)
				}
			} ***REMOVED*** if src.Kind() == reflect.Ptr {
				if err = deepMerge(dst.Elem(), src.Elem(), visited, depth+1, con***REMOVED***g); err != nil {
					return
				}
			} ***REMOVED*** if dst.Elem().Type() == src.Type() {
				if err = deepMerge(dst.Elem(), src, visited, depth+1, con***REMOVED***g); err != nil {
					return
				}
			} ***REMOVED*** {
				return ErrDifferentArgumentsTypes
			}
			break
		}
		if dst.IsNil() || overwrite {
			if dst.CanSet() && (overwrite || isEmptyValue(dst)) {
				dst.Set(src)
			}
		} ***REMOVED*** if err = deepMerge(dst.Elem(), src.Elem(), visited, depth+1, con***REMOVED***g); err != nil {
			return
		}
	default:
		if dst.CanSet() && !isEmptyValue(src) && (overwrite || isEmptyValue(dst)) {
			dst.Set(src)
		}
	}
	return
}

// Merge will ***REMOVED***ll any empty for value type attributes on the dst struct using corresponding
// src attributes if they themselves are not empty. dst and src must be valid same-type structs
// and dst must be a pointer to struct.
// It won't merge unexported (private) ***REMOVED***elds and will do recursively any exported ***REMOVED***eld.
func Merge(dst, src interface{}, opts ...func(*Con***REMOVED***g)) error {
	return merge(dst, src, opts...)
}

// MergeWithOverwrite will do the same as Merge except that non-empty dst attributes will be overriden by
// non-empty src attribute values.
// Deprecated: use Merge(…) with WithOverride
func MergeWithOverwrite(dst, src interface{}, opts ...func(*Con***REMOVED***g)) error {
	return merge(dst, src, append(opts, WithOverride)...)
}

// WithTransformers adds transformers to merge, allowing to customize the merging of some types.
func WithTransformers(transformers Transformers) func(*Con***REMOVED***g) {
	return func(con***REMOVED***g *Con***REMOVED***g) {
		con***REMOVED***g.Transformers = transformers
	}
}

// WithOverride will make merge override non-empty dst attributes with non-empty src attributes values.
func WithOverride(con***REMOVED***g *Con***REMOVED***g) {
	con***REMOVED***g.Overwrite = true
}

// WithAppendSlice will make merge append slices instead of overwriting it
func WithAppendSlice(con***REMOVED***g *Con***REMOVED***g) {
	con***REMOVED***g.AppendSlice = true
}

func merge(dst, src interface{}, opts ...func(*Con***REMOVED***g)) error {
	var (
		vDst, vSrc reflect.Value
		err        error
	)

	con***REMOVED***g := &Con***REMOVED***g{}

	for _, opt := range opts {
		opt(con***REMOVED***g)
	}

	if vDst, vSrc, err = resolveValues(dst, src); err != nil {
		return err
	}
	if vDst.Type() != vSrc.Type() {
		return ErrDifferentArgumentsTypes
	}
	return deepMerge(vDst, vSrc, make(map[uintptr]*visit), 0, con***REMOVED***g)
}
