/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this ***REMOVED***le except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the speci***REMOVED***c language governing permissions and
limitations under the License.
*/

package namer

import (
	"path/***REMOVED***lepath"
	"strings"

	"k8s.io/gengo/types"
)

// Returns whether a name is a private Go name.
func IsPrivateGoName(name string) bool {
	return len(name) == 0 || strings.ToLower(name[:1]) == name[:1]
}

// NewPublicNamer is a helper function that returns a namer that makes
// CamelCase names. See the NameStrategy struct for an explanation of the
// arguments to this constructor.
func NewPublicNamer(prependPackageNames int, ignoreWords ...string) *NameStrategy {
	n := &NameStrategy{
		Join:                Joiner(IC, IC),
		IgnoreWords:         map[string]bool{},
		PrependPackageNames: prependPackageNames,
	}
	for _, w := range ignoreWords {
		n.IgnoreWords[w] = true
	}
	return n
}

// NewPrivateNamer is a helper function that returns a namer that makes
// camelCase names. See the NameStrategy struct for an explanation of the
// arguments to this constructor.
func NewPrivateNamer(prependPackageNames int, ignoreWords ...string) *NameStrategy {
	n := &NameStrategy{
		Join:                Joiner(IL, IC),
		IgnoreWords:         map[string]bool{},
		PrependPackageNames: prependPackageNames,
	}
	for _, w := range ignoreWords {
		n.IgnoreWords[w] = true
	}
	return n
}

// NewRawNamer will return a Namer that makes a name by which you would
// directly refer to a type, optionally keeping track of the import paths
// necessary to reference the names it provides. Tracker may be nil.
// The 'pkg' is the full package name, in which the Namer is used - all
// types from that package will be referenced by just type name without
// referencing the package.
//
// For example, if the type is map[string]int, a raw namer will literally
// return "map[string]int".
//
// Or if the type, in package foo, is "type Bar struct { ... }", then the raw
// namer will return "foo.Bar" as the name of the type, and if 'tracker' was
// not nil, will record that package foo needs to be imported.
func NewRawNamer(pkg string, tracker ImportTracker) *rawNamer {
	return &rawNamer{pkg: pkg, tracker: tracker}
}

// Names is a map from Type to name, as de***REMOVED***ned by some Namer.
type Names map[*types.Type]string

// Namer takes a type, and assigns a name.
//
// The purpose of this complexity is so that you can assign coherent
// side-by-side systems of names for the types. For example, you might want a
// public interface, a private implementation struct, and also to reference
// literally the type name.
//
// Note that it is safe to call your own Name() function recursively to ***REMOVED***nd
// the names of keys, elements, etc. This is because anonymous types can't have
// cycles in their names, and named types don't require the sort of recursion
// that would be problematic.
type Namer interface {
	Name(*types.Type) string
}

// NameSystems is a map of a system name to a namer for that system.
type NameSystems map[string]Namer

// NameStrategy is a general Namer. The easiest way to use it is to copy the
// Public/PrivateNamer variables, and modify the members you wish to change.
//
// The Name method produces a name for the given type, of the forms:
// Anonymous types: <Pre***REMOVED***x><Type description><Suf***REMOVED***x>
// Named types: <Pre***REMOVED***x><Optional Prepended Package name(s)><Original name><Suf***REMOVED***x>
//
// In all cases, every part of the name is run through the capitalization
// functions.
//
// The IgnoreWords map can be set if you have directory names that are
// semantically meaningless for naming purposes, e.g. "proto".
//
// Pre***REMOVED***x and Suf***REMOVED***x can be used to disambiguate parallel systems of type
// names. For example, if you want to generate an interface and an
// implementation, you might want to suf***REMOVED***x one with "Interface" and the other
// with "Implementation". Another common use-- if you want to generate private
// types, and one of your source types could be "string", you can't use the
// default lowercase private namer. You'll have to add a suf***REMOVED***x or pre***REMOVED***x.
type NameStrategy struct {
	Pre***REMOVED***x, Suf***REMOVED***x string
	Join           func(pre string, parts []string, post string) string

	// Add non-meaningful package directory names here (e.g. "proto") and
	// they will be ignored.
	IgnoreWords map[string]bool

	// If > 0, prepend exactly that many package directory names (or as
	// many as there are).  Package names listed in "IgnoreWords" will be
	// ignored.
	//
	// For example, if Ignore words lists "proto" and type Foo is in
	// pkg/server/frobbing/proto, then a value of 1 will give a type name
	// of FrobbingFoo, 2 gives ServerFrobbingFoo, etc.
	PrependPackageNames int

	// A cache of names thus far assigned by this namer.
	Names
}

// IC ensures the ***REMOVED***rst character is uppercase.
func IC(in string) string {
	if in == "" {
		return in
	}
	return strings.ToUpper(in[:1]) + in[1:]
}

// IL ensures the ***REMOVED***rst character is lowercase.
func IL(in string) string {
	if in == "" {
		return in
	}
	return strings.ToLower(in[:1]) + in[1:]
}

// Joiner lets you specify functions that preprocess the various components of
// a name before joining them. You can construct e.g. camelCase or CamelCase or
// any other way of joining words. (See the IC and IL convenience functions.)
func Joiner(***REMOVED***rst, others func(string) string) func(pre string, in []string, post string) string {
	return func(pre string, in []string, post string) string {
		tmp := []string{others(pre)}
		for i := range in {
			tmp = append(tmp, others(in[i]))
		}
		tmp = append(tmp, others(post))
		return ***REMOVED***rst(strings.Join(tmp, ""))
	}
}

func (ns *NameStrategy) removePre***REMOVED***xAndSuf***REMOVED***x(s string) string {
	// The join function may have changed capitalization.
	lowerIn := strings.ToLower(s)
	lowerP := strings.ToLower(ns.Pre***REMOVED***x)
	lowerS := strings.ToLower(ns.Suf***REMOVED***x)
	b, e := 0, len(s)
	if strings.HasPre***REMOVED***x(lowerIn, lowerP) {
		b = len(ns.Pre***REMOVED***x)
	}
	if strings.HasSuf***REMOVED***x(lowerIn, lowerS) {
		e -= len(ns.Suf***REMOVED***x)
	}
	return s[b:e]
}

var (
	importPathNameSanitizer = strings.NewReplacer("-", "_", ".", "")
)

// ***REMOVED***lters out unwanted directory names and sanitizes remaining names.
func (ns *NameStrategy) ***REMOVED***lterDirs(path string) []string {
	allDirs := strings.Split(path, string(***REMOVED***lepath.Separator))
	dirs := make([]string, 0, len(allDirs))
	for _, p := range allDirs {
		if ns.IgnoreWords == nil || !ns.IgnoreWords[p] {
			dirs = append(dirs, importPathNameSanitizer.Replace(p))
		}
	}
	return dirs
}

// See the comment on NameStrategy.
func (ns *NameStrategy) Name(t *types.Type) string {
	if ns.Names == nil {
		ns.Names = Names{}
	}
	if s, ok := ns.Names[t]; ok {
		return s
	}

	if t.Name.Package != "" {
		dirs := append(ns.***REMOVED***lterDirs(t.Name.Package), t.Name.Name)
		i := ns.PrependPackageNames + 1
		dn := len(dirs)
		if i > dn {
			i = dn
		}
		name := ns.Join(ns.Pre***REMOVED***x, dirs[dn-i:], ns.Suf***REMOVED***x)
		ns.Names[t] = name
		return name
	}

	// Only anonymous types remain.
	var name string
	switch t.Kind {
	case types.Builtin:
		name = ns.Join(ns.Pre***REMOVED***x, []string{t.Name.Name}, ns.Suf***REMOVED***x)
	case types.Map:
		name = ns.Join(ns.Pre***REMOVED***x, []string{
			"Map",
			ns.removePre***REMOVED***xAndSuf***REMOVED***x(ns.Name(t.Key)),
			"To",
			ns.removePre***REMOVED***xAndSuf***REMOVED***x(ns.Name(t.Elem)),
		}, ns.Suf***REMOVED***x)
	case types.Slice:
		name = ns.Join(ns.Pre***REMOVED***x, []string{
			"Slice",
			ns.removePre***REMOVED***xAndSuf***REMOVED***x(ns.Name(t.Elem)),
		}, ns.Suf***REMOVED***x)
	case types.Pointer:
		name = ns.Join(ns.Pre***REMOVED***x, []string{
			"Pointer",
			ns.removePre***REMOVED***xAndSuf***REMOVED***x(ns.Name(t.Elem)),
		}, ns.Suf***REMOVED***x)
	case types.Struct:
		names := []string{"Struct"}
		for _, m := range t.Members {
			names = append(names, ns.removePre***REMOVED***xAndSuf***REMOVED***x(ns.Name(m.Type)))
		}
		name = ns.Join(ns.Pre***REMOVED***x, names, ns.Suf***REMOVED***x)
	case types.Chan:
		name = ns.Join(ns.Pre***REMOVED***x, []string{
			"Chan",
			ns.removePre***REMOVED***xAndSuf***REMOVED***x(ns.Name(t.Elem)),
		}, ns.Suf***REMOVED***x)
	case types.Interface:
		// TODO: add to name test
		names := []string{"Interface"}
		for _, m := range t.Methods {
			// TODO: include function signature
			names = append(names, m.Name.Name)
		}
		name = ns.Join(ns.Pre***REMOVED***x, names, ns.Suf***REMOVED***x)
	case types.Func:
		// TODO: add to name test
		parts := []string{"Func"}
		for _, pt := range t.Signature.Parameters {
			parts = append(parts, ns.removePre***REMOVED***xAndSuf***REMOVED***x(ns.Name(pt)))
		}
		parts = append(parts, "Returns")
		for _, rt := range t.Signature.Results {
			parts = append(parts, ns.removePre***REMOVED***xAndSuf***REMOVED***x(ns.Name(rt)))
		}
		name = ns.Join(ns.Pre***REMOVED***x, parts, ns.Suf***REMOVED***x)
	default:
		name = "unnameable_" + string(t.Kind)
	}
	ns.Names[t] = name
	return name
}

// ImportTracker allows a raw namer to keep track of the packages needed for
// import. You can implement yourself or use the one in the generation package.
type ImportTracker interface {
	AddType(*types.Type)
	LocalNameOf(packagePath string) string
	PathOf(localName string) (string, bool)
	ImportLines() []string
}

type rawNamer struct {
	pkg     string
	tracker ImportTracker
	Names
}

// Name makes a name the way you'd write it to literally refer to type t,
// making ordinary assumptions about how you've imported t's package (or using
// r.tracker to speci***REMOVED***cally track the package imports).
func (r *rawNamer) Name(t *types.Type) string {
	if r.Names == nil {
		r.Names = Names{}
	}
	if name, ok := r.Names[t]; ok {
		return name
	}
	if t.Name.Package != "" {
		var name string
		if r.tracker != nil {
			r.tracker.AddType(t)
			if t.Name.Package == r.pkg {
				name = t.Name.Name
			} ***REMOVED*** {
				name = r.tracker.LocalNameOf(t.Name.Package) + "." + t.Name.Name
			}
		} ***REMOVED*** {
			if t.Name.Package == r.pkg {
				name = t.Name.Name
			} ***REMOVED*** {
				name = ***REMOVED***lepath.Base(t.Name.Package) + "." + t.Name.Name
			}
		}
		r.Names[t] = name
		return name
	}
	var name string
	switch t.Kind {
	case types.Builtin:
		name = t.Name.Name
	case types.Map:
		name = "map[" + r.Name(t.Key) + "]" + r.Name(t.Elem)
	case types.Slice:
		name = "[]" + r.Name(t.Elem)
	case types.Pointer:
		name = "*" + r.Name(t.Elem)
	case types.Struct:
		elems := []string{}
		for _, m := range t.Members {
			elems = append(elems, m.Name+" "+r.Name(m.Type))
		}
		name = "struct{" + strings.Join(elems, "; ") + "}"
	case types.Chan:
		// TODO: include directionality
		name = "chan " + r.Name(t.Elem)
	case types.Interface:
		// TODO: add to name test
		elems := []string{}
		for _, m := range t.Methods {
			// TODO: include function signature
			elems = append(elems, m.Name.Name)
		}
		name = "interface{" + strings.Join(elems, "; ") + "}"
	case types.Func:
		// TODO: add to name test
		params := []string{}
		for _, pt := range t.Signature.Parameters {
			params = append(params, r.Name(pt))
		}
		results := []string{}
		for _, rt := range t.Signature.Results {
			results = append(results, r.Name(rt))
		}
		name = "func(" + strings.Join(params, ",") + ")"
		if len(results) == 1 {
			name += " " + results[0]
		} ***REMOVED*** if len(results) > 1 {
			name += " (" + strings.Join(results, ",") + ")"
		}
	default:
		name = "unnameable_" + string(t.Kind)
	}
	r.Names[t] = name
	return name
}
