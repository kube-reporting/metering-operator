// Package yaml implements YAML support for the Go language.
//
// Source code and other details for the project are available at GitHub:
//
//   https://github.com/go-yaml/yaml
//
package yaml

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// MapSlice encodes and decodes as a YAML map.
// The order of keys is preserved when encoding and decoding.
type MapSlice []MapItem

// MapItem is an item in a MapSlice.
type MapItem struct {
	Key, Value interface{}
}

// The Unmarshaler interface may be implemented by types to customize their
// behavior when being unmarshaled from a YAML document. The UnmarshalYAML
// method receives a function that may be called to unmarshal the original
// YAML value into a ***REMOVED***eld or variable. It is safe to call the unmarshal
// function parameter more than once if necessary.
type Unmarshaler interface {
	UnmarshalYAML(unmarshal func(interface{}) error) error
}

// The Marshaler interface may be implemented by types to customize their
// behavior when being marshaled into a YAML document. The returned value
// is marshaled in place of the original value implementing Marshaler.
//
// If an error is returned by MarshalYAML, the marshaling procedure stops
// and returns with the provided error.
type Marshaler interface {
	MarshalYAML() (interface{}, error)
}

// Unmarshal decodes the ***REMOVED***rst document found within the in byte slice
// and assigns decoded values into the out value.
//
// Maps and pointers (to a struct, string, int, etc) are accepted as out
// values. If an internal pointer within a struct is not initialized,
// the yaml package will initialize it if necessary for unmarshalling
// the provided data. The out parameter must not be nil.
//
// The type of the decoded values should be compatible with the respective
// values in out. If one or more values cannot be decoded due to a type
// mismatches, decoding continues partially until the end of the YAML
// content, and a *yaml.TypeError is returned with details for all
// missed values.
//
// Struct ***REMOVED***elds are only unmarshalled if they are exported (have an
// upper case ***REMOVED***rst letter), and are unmarshalled using the ***REMOVED***eld name
// lowercased as the default key. Custom keys may be de***REMOVED***ned via the
// "yaml" name in the ***REMOVED***eld tag: the content preceding the ***REMOVED***rst comma
// is used as the key, and the following comma-separated options are
// used to tweak the marshalling process (see Marshal).
// Conflicting names result in a runtime error.
//
// For example:
//
//     type T struct {
//         F int `yaml:"a,omitempty"`
//         B int
//     }
//     var t T
//     yaml.Unmarshal([]byte("a: 1\nb: 2"), &t)
//
// See the documentation of Marshal for the format of tags and a list of
// supported tag options.
//
func Unmarshal(in []byte, out interface{}) (err error) {
	defer handleErr(&err)
	d := newDecoder()
	p := newParser(in)
	defer p.destroy()
	node := p.parse()
	if node != nil {
		v := reflect.ValueOf(out)
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			v = v.Elem()
		}
		d.unmarshal(node, v)
	}
	if len(d.terrors) > 0 {
		return &TypeError{d.terrors}
	}
	return nil
}

// Marshal serializes the value provided into a YAML document. The structure
// of the generated document will reflect the structure of the value itself.
// Maps and pointers (to struct, string, int, etc) are accepted as the in value.
//
// Struct ***REMOVED***elds are only unmarshalled if they are exported (have an upper case
// ***REMOVED***rst letter), and are unmarshalled using the ***REMOVED***eld name lowercased as the
// default key. Custom keys may be de***REMOVED***ned via the "yaml" name in the ***REMOVED***eld
// tag: the content preceding the ***REMOVED***rst comma is used as the key, and the
// following comma-separated options are used to tweak the marshalling process.
// Conflicting names result in a runtime error.
//
// The ***REMOVED***eld tag format accepted is:
//
//     `(...) yaml:"[<key>][,<flag1>[,<flag2>]]" (...)`
//
// The following flags are currently supported:
//
//     omitempty    Only include the ***REMOVED***eld if it's not set to the zero
//                  value for the type or to empty slices or maps.
//                  Does not apply to zero valued structs.
//
//     flow         Marshal using a flow style (useful for structs,
//                  sequences and maps).
//
//     inline       Inline the ***REMOVED***eld, which must be a struct or a map,
//                  causing all of its ***REMOVED***elds or keys to be processed as if
//                  they were part of the outer struct. For maps, keys must
//                  not conflict with the yaml keys of other struct ***REMOVED***elds.
//
// In addition, if the key is "-", the ***REMOVED***eld is ignored.
//
// For example:
//
//     type T struct {
//         F int "a,omitempty"
//         B int
//     }
//     yaml.Marshal(&T{B: 2}) // Returns "b: 2\n"
//     yaml.Marshal(&T{F: 1}} // Returns "a: 1\nb: 0\n"
//
func Marshal(in interface{}) (out []byte, err error) {
	defer handleErr(&err)
	e := newEncoder()
	defer e.destroy()
	e.marshal("", reflect.ValueOf(in))
	e.***REMOVED***nish()
	out = e.out
	return
}

func handleErr(err *error) {
	if v := recover(); v != nil {
		if e, ok := v.(yamlError); ok {
			*err = e.err
		} ***REMOVED*** {
			panic(v)
		}
	}
}

type yamlError struct {
	err error
}

func fail(err error) {
	panic(yamlError{err})
}

func failf(format string, args ...interface{}) {
	panic(yamlError{fmt.Errorf("yaml: "+format, args...)})
}

// A TypeError is returned by Unmarshal when one or more ***REMOVED***elds in
// the YAML document cannot be properly decoded into the requested
// types. When this error is returned, the value is still
// unmarshaled partially.
type TypeError struct {
	Errors []string
}

func (e *TypeError) Error() string {
	return fmt.Sprintf("yaml: unmarshal errors:\n  %s", strings.Join(e.Errors, "\n  "))
}

// --------------------------------------------------------------------------
// Maintain a mapping of keys to structure ***REMOVED***eld indexes

// The code in this section was copied from mgo/bson.

// structInfo holds details for the serialization of ***REMOVED***elds of
// a given struct.
type structInfo struct {
	FieldsMap  map[string]***REMOVED***eldInfo
	FieldsList []***REMOVED***eldInfo

	// InlineMap is the number of the ***REMOVED***eld in the struct that
	// contains an ,inline map, or -1 if there's none.
	InlineMap int
}

type ***REMOVED***eldInfo struct {
	Key       string
	Num       int
	OmitEmpty bool
	Flow      bool

	// Inline holds the ***REMOVED***eld index if the ***REMOVED***eld is part of an inlined struct.
	Inline []int
}

var structMap = make(map[reflect.Type]*structInfo)
var ***REMOVED***eldMapMutex sync.RWMutex

func getStructInfo(st reflect.Type) (*structInfo, error) {
	***REMOVED***eldMapMutex.RLock()
	sinfo, found := structMap[st]
	***REMOVED***eldMapMutex.RUnlock()
	if found {
		return sinfo, nil
	}

	n := st.NumField()
	***REMOVED***eldsMap := make(map[string]***REMOVED***eldInfo)
	***REMOVED***eldsList := make([]***REMOVED***eldInfo, 0, n)
	inlineMap := -1
	for i := 0; i != n; i++ {
		***REMOVED***eld := st.Field(i)
		if ***REMOVED***eld.PkgPath != "" {
			continue // Private ***REMOVED***eld
		}

		info := ***REMOVED***eldInfo{Num: i}

		tag := ***REMOVED***eld.Tag.Get("yaml")
		if tag == "" && strings.Index(string(***REMOVED***eld.Tag), ":") < 0 {
			tag = string(***REMOVED***eld.Tag)
		}
		if tag == "-" {
			continue
		}

		inline := false
		***REMOVED***elds := strings.Split(tag, ",")
		if len(***REMOVED***elds) > 1 {
			for _, flag := range ***REMOVED***elds[1:] {
				switch flag {
				case "omitempty":
					info.OmitEmpty = true
				case "flow":
					info.Flow = true
				case "inline":
					inline = true
				default:
					return nil, errors.New(fmt.Sprintf("Unsupported flag %q in tag %q of type %s", flag, tag, st))
				}
			}
			tag = ***REMOVED***elds[0]
		}

		if inline {
			switch ***REMOVED***eld.Type.Kind() {
			case reflect.Map:
				if inlineMap >= 0 {
					return nil, errors.New("Multiple ,inline maps in struct " + st.String())
				}
				if ***REMOVED***eld.Type.Key() != reflect.TypeOf("") {
					return nil, errors.New("Option ,inline needs a map with string keys in struct " + st.String())
				}
				inlineMap = info.Num
			case reflect.Struct:
				sinfo, err := getStructInfo(***REMOVED***eld.Type)
				if err != nil {
					return nil, err
				}
				for _, ***REMOVED***nfo := range sinfo.FieldsList {
					if _, found := ***REMOVED***eldsMap[***REMOVED***nfo.Key]; found {
						msg := "Duplicated key '" + ***REMOVED***nfo.Key + "' in struct " + st.String()
						return nil, errors.New(msg)
					}
					if ***REMOVED***nfo.Inline == nil {
						***REMOVED***nfo.Inline = []int{i, ***REMOVED***nfo.Num}
					} ***REMOVED*** {
						***REMOVED***nfo.Inline = append([]int{i}, ***REMOVED***nfo.Inline...)
					}
					***REMOVED***eldsMap[***REMOVED***nfo.Key] = ***REMOVED***nfo
					***REMOVED***eldsList = append(***REMOVED***eldsList, ***REMOVED***nfo)
				}
			default:
				//return nil, errors.New("Option ,inline needs a struct value or map ***REMOVED***eld")
				return nil, errors.New("Option ,inline needs a struct value ***REMOVED***eld")
			}
			continue
		}

		if tag != "" {
			info.Key = tag
		} ***REMOVED*** {
			info.Key = strings.ToLower(***REMOVED***eld.Name)
		}

		if _, found = ***REMOVED***eldsMap[info.Key]; found {
			msg := "Duplicated key '" + info.Key + "' in struct " + st.String()
			return nil, errors.New(msg)
		}

		***REMOVED***eldsList = append(***REMOVED***eldsList, info)
		***REMOVED***eldsMap[info.Key] = info
	}

	sinfo = &structInfo{***REMOVED***eldsMap, ***REMOVED***eldsList, inlineMap}

	***REMOVED***eldMapMutex.Lock()
	structMap[st] = sinfo
	***REMOVED***eldMapMutex.Unlock()
	return sinfo, nil
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return len(v.String()) == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Slice:
		return v.Len() == 0
	case reflect.Map:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Struct:
		vt := v.Type()
		for i := v.NumField() - 1; i >= 0; i-- {
			if vt.Field(i).PkgPath != "" {
				continue // Private ***REMOVED***eld
			}
			if !isZero(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return false
}
