// Copyright 2014 Unknwon
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this ***REMOVED***le except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the speci***REMOVED***c language governing permissions and limitations
// under the License.

package ini

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"
)

// NameMapper represents a ini tag name mapper.
type NameMapper func(string) string

// Built-in name getters.
var (
	// AllCapsUnderscore converts to format ALL_CAPS_UNDERSCORE.
	AllCapsUnderscore NameMapper = func(raw string) string {
		newstr := make([]rune, 0, len(raw))
		for i, chr := range raw {
			if isUpper := 'A' <= chr && chr <= 'Z'; isUpper {
				if i > 0 {
					newstr = append(newstr, '_')
				}
			}
			newstr = append(newstr, unicode.ToUpper(chr))
		}
		return string(newstr)
	}
	// TitleUnderscore converts to format title_underscore.
	TitleUnderscore NameMapper = func(raw string) string {
		newstr := make([]rune, 0, len(raw))
		for i, chr := range raw {
			if isUpper := 'A' <= chr && chr <= 'Z'; isUpper {
				if i > 0 {
					newstr = append(newstr, '_')
				}
				chr -= ('A' - 'a')
			}
			newstr = append(newstr, chr)
		}
		return string(newstr)
	}
)

func (s *Section) parseFieldName(raw, actual string) string {
	if len(actual) > 0 {
		return actual
	}
	if s.f.NameMapper != nil {
		return s.f.NameMapper(raw)
	}
	return raw
}

func parseDelim(actual string) string {
	if len(actual) > 0 {
		return actual
	}
	return ","
}

var reflectTime = reflect.TypeOf(time.Now()).Kind()

// setSliceWithProperType sets proper values to slice based on its type.
func setSliceWithProperType(key *Key, ***REMOVED***eld reflect.Value, delim string, allowShadow, isStrict bool) error {
	var strs []string
	if allowShadow {
		strs = key.StringsWithShadows(delim)
	} ***REMOVED*** {
		strs = key.Strings(delim)
	}

	numVals := len(strs)
	if numVals == 0 {
		return nil
	}

	var vals interface{}
	var err error

	sliceOf := ***REMOVED***eld.Type().Elem().Kind()
	switch sliceOf {
	case reflect.String:
		vals = strs
	case reflect.Int:
		vals, err = key.parseInts(strs, true, false)
	case reflect.Int64:
		vals, err = key.parseInt64s(strs, true, false)
	case reflect.Uint:
		vals, err = key.parseUints(strs, true, false)
	case reflect.Uint64:
		vals, err = key.parseUint64s(strs, true, false)
	case reflect.Float64:
		vals, err = key.parseFloat64s(strs, true, false)
	case reflectTime:
		vals, err = key.parseTimesFormat(time.RFC3339, strs, true, false)
	default:
		return fmt.Errorf("unsupported type '[]%s'", sliceOf)
	}
	if isStrict {
		return err
	}

	slice := reflect.MakeSlice(***REMOVED***eld.Type(), numVals, numVals)
	for i := 0; i < numVals; i++ {
		switch sliceOf {
		case reflect.String:
			slice.Index(i).Set(reflect.ValueOf(vals.([]string)[i]))
		case reflect.Int:
			slice.Index(i).Set(reflect.ValueOf(vals.([]int)[i]))
		case reflect.Int64:
			slice.Index(i).Set(reflect.ValueOf(vals.([]int64)[i]))
		case reflect.Uint:
			slice.Index(i).Set(reflect.ValueOf(vals.([]uint)[i]))
		case reflect.Uint64:
			slice.Index(i).Set(reflect.ValueOf(vals.([]uint64)[i]))
		case reflect.Float64:
			slice.Index(i).Set(reflect.ValueOf(vals.([]float64)[i]))
		case reflectTime:
			slice.Index(i).Set(reflect.ValueOf(vals.([]time.Time)[i]))
		}
	}
	***REMOVED***eld.Set(slice)
	return nil
}

func wrapStrictError(err error, isStrict bool) error {
	if isStrict {
		return err
	}
	return nil
}

// setWithProperType sets proper value to ***REMOVED***eld based on its type,
// but it does not return error for failing parsing,
// because we want to use default value that is already assigned to strcut.
func setWithProperType(t reflect.Type, key *Key, ***REMOVED***eld reflect.Value, delim string, allowShadow, isStrict bool) error {
	switch t.Kind() {
	case reflect.String:
		if len(key.String()) == 0 {
			return nil
		}
		***REMOVED***eld.SetString(key.String())
	case reflect.Bool:
		boolVal, err := key.Bool()
		if err != nil {
			return wrapStrictError(err, isStrict)
		}
		***REMOVED***eld.SetBool(boolVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		durationVal, err := key.Duration()
		// Skip zero value
		if err == nil && int(durationVal) > 0 {
			***REMOVED***eld.Set(reflect.ValueOf(durationVal))
			return nil
		}

		intVal, err := key.Int64()
		if err != nil {
			return wrapStrictError(err, isStrict)
		}
		***REMOVED***eld.SetInt(intVal)
	//	byte is an alias for uint8, so supporting uint8 breaks support for byte
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		durationVal, err := key.Duration()
		// Skip zero value
		if err == nil && int(durationVal) > 0 {
			***REMOVED***eld.Set(reflect.ValueOf(durationVal))
			return nil
		}

		uintVal, err := key.Uint64()
		if err != nil {
			return wrapStrictError(err, isStrict)
		}
		***REMOVED***eld.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := key.Float64()
		if err != nil {
			return wrapStrictError(err, isStrict)
		}
		***REMOVED***eld.SetFloat(floatVal)
	case reflectTime:
		timeVal, err := key.Time()
		if err != nil {
			return wrapStrictError(err, isStrict)
		}
		***REMOVED***eld.Set(reflect.ValueOf(timeVal))
	case reflect.Slice:
		return setSliceWithProperType(key, ***REMOVED***eld, delim, allowShadow, isStrict)
	default:
		return fmt.Errorf("unsupported type '%s'", t)
	}
	return nil
}

func parseTagOptions(tag string) (rawName string, omitEmpty bool, allowShadow bool) {
	opts := strings.SplitN(tag, ",", 3)
	rawName = opts[0]
	if len(opts) > 1 {
		omitEmpty = opts[1] == "omitempty"
	}
	if len(opts) > 2 {
		allowShadow = opts[2] == "allowshadow"
	}
	return rawName, omitEmpty, allowShadow
}

func (s *Section) mapTo(val reflect.Value, isStrict bool) error {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		***REMOVED***eld := val.Field(i)
		tpField := typ.Field(i)

		tag := tpField.Tag.Get("ini")
		if tag == "-" {
			continue
		}

		rawName, _, allowShadow := parseTagOptions(tag)
		***REMOVED***eldName := s.parseFieldName(tpField.Name, rawName)
		if len(***REMOVED***eldName) == 0 || !***REMOVED***eld.CanSet() {
			continue
		}

		isAnonymous := tpField.Type.Kind() == reflect.Ptr && tpField.Anonymous
		isStruct := tpField.Type.Kind() == reflect.Struct
		if isAnonymous {
			***REMOVED***eld.Set(reflect.New(tpField.Type.Elem()))
		}

		if isAnonymous || isStruct {
			if sec, err := s.f.GetSection(***REMOVED***eldName); err == nil {
				if err = sec.mapTo(***REMOVED***eld, isStrict); err != nil {
					return fmt.Errorf("error mapping ***REMOVED***eld(%s): %v", ***REMOVED***eldName, err)
				}
				continue
			}
		}

		if key, err := s.GetKey(***REMOVED***eldName); err == nil {
			delim := parseDelim(tpField.Tag.Get("delim"))
			if err = setWithProperType(tpField.Type, key, ***REMOVED***eld, delim, allowShadow, isStrict); err != nil {
				return fmt.Errorf("error mapping ***REMOVED***eld(%s): %v", ***REMOVED***eldName, err)
			}
		}
	}
	return nil
}

// MapTo maps section to given struct.
func (s *Section) MapTo(v interface{}) error {
	typ := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	} ***REMOVED*** {
		return errors.New("cannot map to non-pointer struct")
	}

	return s.mapTo(val, false)
}

// MapTo maps section to given struct in strict mode,
// which returns all possible error including value parsing error.
func (s *Section) StrictMapTo(v interface{}) error {
	typ := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	} ***REMOVED*** {
		return errors.New("cannot map to non-pointer struct")
	}

	return s.mapTo(val, true)
}

// MapTo maps ***REMOVED***le to given struct.
func (f *File) MapTo(v interface{}) error {
	return f.Section("").MapTo(v)
}

// MapTo maps ***REMOVED***le to given struct in strict mode,
// which returns all possible error including value parsing error.
func (f *File) StrictMapTo(v interface{}) error {
	return f.Section("").StrictMapTo(v)
}

// MapTo maps data sources to given struct with name mapper.
func MapToWithMapper(v interface{}, mapper NameMapper, source interface{}, others ...interface{}) error {
	cfg, err := Load(source, others...)
	if err != nil {
		return err
	}
	cfg.NameMapper = mapper
	return cfg.MapTo(v)
}

// StrictMapToWithMapper maps data sources to given struct with name mapper in strict mode,
// which returns all possible error including value parsing error.
func StrictMapToWithMapper(v interface{}, mapper NameMapper, source interface{}, others ...interface{}) error {
	cfg, err := Load(source, others...)
	if err != nil {
		return err
	}
	cfg.NameMapper = mapper
	return cfg.StrictMapTo(v)
}

// MapTo maps data sources to given struct.
func MapTo(v, source interface{}, others ...interface{}) error {
	return MapToWithMapper(v, nil, source, others...)
}

// StrictMapTo maps data sources to given struct in strict mode,
// which returns all possible error including value parsing error.
func StrictMapTo(v, source interface{}, others ...interface{}) error {
	return StrictMapToWithMapper(v, nil, source, others...)
}

// reflectSliceWithProperType does the opposite thing as setSliceWithProperType.
func reflectSliceWithProperType(key *Key, ***REMOVED***eld reflect.Value, delim string) error {
	slice := ***REMOVED***eld.Slice(0, ***REMOVED***eld.Len())
	if ***REMOVED***eld.Len() == 0 {
		return nil
	}

	var buf bytes.Buffer
	sliceOf := ***REMOVED***eld.Type().Elem().Kind()
	for i := 0; i < ***REMOVED***eld.Len(); i++ {
		switch sliceOf {
		case reflect.String:
			buf.WriteString(slice.Index(i).String())
		case reflect.Int, reflect.Int64:
			buf.WriteString(fmt.Sprint(slice.Index(i).Int()))
		case reflect.Uint, reflect.Uint64:
			buf.WriteString(fmt.Sprint(slice.Index(i).Uint()))
		case reflect.Float64:
			buf.WriteString(fmt.Sprint(slice.Index(i).Float()))
		case reflectTime:
			buf.WriteString(slice.Index(i).Interface().(time.Time).Format(time.RFC3339))
		default:
			return fmt.Errorf("unsupported type '[]%s'", sliceOf)
		}
		buf.WriteString(delim)
	}
	key.SetValue(buf.String()[:buf.Len()-1])
	return nil
}

// reflectWithProperType does the opposite thing as setWithProperType.
func reflectWithProperType(t reflect.Type, key *Key, ***REMOVED***eld reflect.Value, delim string) error {
	switch t.Kind() {
	case reflect.String:
		key.SetValue(***REMOVED***eld.String())
	case reflect.Bool:
		key.SetValue(fmt.Sprint(***REMOVED***eld.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		key.SetValue(fmt.Sprint(***REMOVED***eld.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		key.SetValue(fmt.Sprint(***REMOVED***eld.Uint()))
	case reflect.Float32, reflect.Float64:
		key.SetValue(fmt.Sprint(***REMOVED***eld.Float()))
	case reflectTime:
		key.SetValue(fmt.Sprint(***REMOVED***eld.Interface().(time.Time).Format(time.RFC3339)))
	case reflect.Slice:
		return reflectSliceWithProperType(key, ***REMOVED***eld, delim)
	default:
		return fmt.Errorf("unsupported type '%s'", t)
	}
	return nil
}

// CR: copied from encoding/json/encode.go with modi***REMOVED***cations of time.Time support.
// TODO: add more test coverage.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflectTime:
		return v.Interface().(time.Time).IsZero()
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func (s *Section) reflectFrom(val reflect.Value) error {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		***REMOVED***eld := val.Field(i)
		tpField := typ.Field(i)

		tag := tpField.Tag.Get("ini")
		if tag == "-" {
			continue
		}

		opts := strings.SplitN(tag, ",", 2)
		if len(opts) == 2 && opts[1] == "omitempty" && isEmptyValue(***REMOVED***eld) {
			continue
		}

		***REMOVED***eldName := s.parseFieldName(tpField.Name, opts[0])
		if len(***REMOVED***eldName) == 0 || !***REMOVED***eld.CanSet() {
			continue
		}

		if (tpField.Type.Kind() == reflect.Ptr && tpField.Anonymous) ||
			(tpField.Type.Kind() == reflect.Struct && tpField.Type.Name() != "Time") {
			// Note: The only error here is section doesn't exist.
			sec, err := s.f.GetSection(***REMOVED***eldName)
			if err != nil {
				// Note: ***REMOVED***eldName can never be empty here, ignore error.
				sec, _ = s.f.NewSection(***REMOVED***eldName)
			}
			if err = sec.reflectFrom(***REMOVED***eld); err != nil {
				return fmt.Errorf("error reflecting ***REMOVED***eld (%s): %v", ***REMOVED***eldName, err)
			}
			continue
		}

		// Note: Same reason as secion.
		key, err := s.GetKey(***REMOVED***eldName)
		if err != nil {
			key, _ = s.NewKey(***REMOVED***eldName, "")
		}
		if err = reflectWithProperType(tpField.Type, key, ***REMOVED***eld, parseDelim(tpField.Tag.Get("delim"))); err != nil {
			return fmt.Errorf("error reflecting ***REMOVED***eld (%s): %v", ***REMOVED***eldName, err)
		}

	}
	return nil
}

// ReflectFrom reflects secion from given struct.
func (s *Section) ReflectFrom(v interface{}) error {
	typ := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	} ***REMOVED*** {
		return errors.New("cannot reflect from non-pointer struct")
	}

	return s.reflectFrom(val)
}

// ReflectFrom reflects ***REMOVED***le from given struct.
func (f *File) ReflectFrom(v interface{}) error {
	return f.Section("").ReflectFrom(v)
}

// ReflectFrom reflects data sources from given struct with name mapper.
func ReflectFromWithMapper(cfg *File, v interface{}, mapper NameMapper) error {
	cfg.NameMapper = mapper
	return cfg.ReflectFrom(v)
}

// ReflectFrom reflects data sources from given struct.
func ReflectFrom(cfg *File, v interface{}) error {
	return ReflectFromWithMapper(cfg, v, nil)
}
