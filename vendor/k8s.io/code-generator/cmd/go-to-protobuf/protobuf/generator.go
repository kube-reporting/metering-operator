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

package protobuf

import (
	"fmt"
	"io"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"k8s.io/klog"

	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

// genProtoIDL produces a .proto IDL.
type genProtoIDL struct {
	generator.DefaultGen
	localPackage   types.Name
	localGoPackage types.Name
	imports        namer.ImportTracker

	generateAll    bool
	omitGogo       bool
	omitFieldTypes map[types.Name]struct{}
}

func (g *genProtoIDL) PackageVars(c *generator.Context) []string {
	if g.omitGogo {
		return []string{
			fmt.Sprintf("option go_package = %q;", g.localGoPackage.Name),
		}
	}
	return []string{
		"option (gogoproto.marshaler_all) = true;",
		"option (gogoproto.stable_marshaler_all) = true;",
		"option (gogoproto.sizer_all) = true;",
		"option (gogoproto.goproto_stringer_all) = false;",
		"option (gogoproto.stringer_all) = true;",
		"option (gogoproto.unmarshaler_all) = true;",
		"option (gogoproto.goproto_unrecognized_all) = false;",
		"option (gogoproto.goproto_enum_pre***REMOVED***x_all) = false;",
		"option (gogoproto.goproto_getters_all) = false;",
		fmt.Sprintf("option go_package = %q;", g.localGoPackage.Name),
	}
}
func (g *genProtoIDL) Filename() string { return g.OptionalName + ".proto" }
func (g *genProtoIDL) FileType() string { return "protoidl" }
func (g *genProtoIDL) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		// The local namer returns the correct protobuf name for a proto type
		// in the context of a package
		"local": localNamer{g.localPackage},
	}
}

// Filter ignores types that are identi***REMOVED***ed as not exportable.
func (g *genProtoIDL) Filter(c *generator.Context, t *types.Type) bool {
	tagVals := types.ExtractCommentTags("+", t.CommentLines)["protobuf"]
	if tagVals != nil {
		if tagVals[0] == "false" {
			// Type speci***REMOVED***ed "false".
			return false
		}
		if tagVals[0] == "true" {
			// Type speci***REMOVED***ed "true".
			return true
		}
		klog.Fatalf(`Comment tag "protobuf" must be true or false, found: %q`, tagVals[0])
	}
	if !g.generateAll {
		// We're not generating everything.
		return false
	}
	seen := map[*types.Type]bool{}
	ok := isProtoable(seen, t)
	return ok
}

func isProtoable(seen map[*types.Type]bool, t *types.Type) bool {
	if seen[t] {
		// be optimistic in the case of type cycles.
		return true
	}
	seen[t] = true
	switch t.Kind {
	case types.Builtin:
		return true
	case types.Alias:
		return isProtoable(seen, t.Underlying)
	case types.Slice, types.Pointer:
		return isProtoable(seen, t.Elem)
	case types.Map:
		return isProtoable(seen, t.Key) && isProtoable(seen, t.Elem)
	case types.Struct:
		if len(t.Members) == 0 {
			return true
		}
		for _, m := range t.Members {
			if isProtoable(seen, m.Type) {
				return true
			}
		}
		return false
	case types.Func, types.Chan:
		return false
	case types.DeclarationOf, types.Unknown, types.Unsupported:
		return false
	case types.Interface:
		return false
	default:
		log.Printf("WARNING: type %q is not portable: %s", t.Kind, t.Name)
		return false
	}
}

// isOptionalAlias should return true if the speci***REMOVED***ed type has an underlying type
// (is an alias) of a map or slice and has the comment tag protobuf.nullable=true,
// indicating that the type should be nullable in protobuf.
func isOptionalAlias(t *types.Type) bool {
	if t.Underlying == nil || (t.Underlying.Kind != types.Map && t.Underlying.Kind != types.Slice) {
		return false
	}
	if extractBoolTagOrDie("protobuf.nullable", t.CommentLines) == false {
		return false
	}
	return true
}

func (g *genProtoIDL) Imports(c *generator.Context) (imports []string) {
	lines := []string{}
	// TODO: this could be expressed more cleanly
	for _, line := range g.imports.ImportLines() {
		if g.omitGogo && line == "github.com/gogo/protobuf/gogoproto/gogo.proto" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

// GenerateType makes the body of a ***REMOVED***le implementing a set for type t.
func (g *genProtoIDL) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "$", "$")
	b := bodyGen{
		locator: &protobufLocator{
			namer:    c.Namers["proto"].(ProtobufFromGoNamer),
			tracker:  g.imports,
			universe: c.Universe,

			localGoPackage: g.localGoPackage.Package,
		},
		localPackage: g.localPackage,

		omitGogo:       g.omitGogo,
		omitFieldTypes: g.omitFieldTypes,

		t: t,
	}
	switch t.Kind {
	case types.Alias:
		return b.doAlias(sw)
	case types.Struct:
		return b.doStruct(sw)
	default:
		return b.unknown(sw)
	}
}

// ProtobufFromGoNamer ***REMOVED***nds the protobuf name of a type (and its package, and
// the package path) from its Go name.
type ProtobufFromGoNamer interface {
	GoNameToProtoName(name types.Name) types.Name
}

type ProtobufLocator interface {
	ProtoTypeFor(t *types.Type) (*types.Type, error)
	GoTypeForName(name types.Name) *types.Type
	CastTypeName(name types.Name) string
}

type protobufLocator struct {
	namer    ProtobufFromGoNamer
	tracker  namer.ImportTracker
	universe types.Universe

	localGoPackage string
}

// CastTypeName returns the cast type name of a Go type
// TODO: delegate to a new localgo namer?
func (p protobufLocator) CastTypeName(name types.Name) string {
	if name.Package == p.localGoPackage {
		return name.Name
	}
	return name.String()
}

func (p protobufLocator) GoTypeForName(name types.Name) *types.Type {
	if len(name.Package) == 0 {
		name.Package = p.localGoPackage
	}
	return p.universe.Type(name)
}

// ProtoTypeFor locates a Protobuf type for the provided Go type (if possible).
func (p protobufLocator) ProtoTypeFor(t *types.Type) (*types.Type, error) {
	switch {
	// we've already converted the type, or it's a map
	case t.Kind == types.Protobuf || t.Kind == types.Map:
		p.tracker.AddType(t)
		return t, nil
	}
	// it's a fundamental type
	if t, ok := isFundamentalProtoType(t); ok {
		p.tracker.AddType(t)
		return t, nil
	}
	// it's a message
	if t.Kind == types.Struct || isOptionalAlias(t) {
		t := &types.Type{
			Name: p.namer.GoNameToProtoName(t.Name),
			Kind: types.Protobuf,

			CommentLines: t.CommentLines,
		}
		p.tracker.AddType(t)
		return t, nil
	}
	return nil, errUnrecognizedType
}

type bodyGen struct {
	locator        ProtobufLocator
	localPackage   types.Name
	omitGogo       bool
	omitFieldTypes map[types.Name]struct{}

	t *types.Type
}

func (b bodyGen) unknown(sw *generator.SnippetWriter) error {
	return fmt.Errorf("not sure how to generate: %#v", b.t)
}

func (b bodyGen) doAlias(sw *generator.SnippetWriter) error {
	if !isOptionalAlias(b.t) {
		return nil
	}

	var kind string
	switch b.t.Underlying.Kind {
	case types.Map:
		kind = "map"
	default:
		kind = "slice"
	}
	optional := &types.Type{
		Name: b.t.Name,
		Kind: types.Struct,

		CommentLines:              b.t.CommentLines,
		SecondClosestCommentLines: b.t.SecondClosestCommentLines,
		Members: []types.Member{
			{
				Name:         "Items",
				CommentLines: []string{fmt.Sprintf("items, if empty, will result in an empty %s\n", kind)},
				Type:         b.t.Underlying,
			},
		},
	}
	nested := b
	nested.t = optional
	return nested.doStruct(sw)
}

func (b bodyGen) doStruct(sw *generator.SnippetWriter) error {
	if len(b.t.Name.Name) == 0 {
		return nil
	}
	if namer.IsPrivateGoName(b.t.Name.Name) {
		return nil
	}

	var alias *types.Type
	var ***REMOVED***elds []protoField
	options := []string{}
	allOptions := types.ExtractCommentTags("+", b.t.CommentLines)
	for k, v := range allOptions {
		switch {
		case strings.HasPre***REMOVED***x(k, "protobuf.options."):
			key := strings.TrimPre***REMOVED***x(k, "protobuf.options.")
			switch key {
			case "marshal":
				if v[0] == "false" {
					if !b.omitGogo {
						options = append(options,
							"(gogoproto.marshaler) = false",
							"(gogoproto.unmarshaler) = false",
							"(gogoproto.sizer) = false",
						)
					}
				}
			default:
				if !b.omitGogo || !strings.HasPre***REMOVED***x(key, "(gogoproto.") {
					if key == "(gogoproto.goproto_stringer)" && v[0] == "false" {
						options = append(options, "(gogoproto.stringer) = false")
					}
					options = append(options, fmt.Sprintf("%s = %s", key, v[0]))
				}
			}
		// protobuf.as allows a type to have the same message contents as another Go type
		case k == "protobuf.as":
			***REMOVED***elds = nil
			if alias = b.locator.GoTypeForName(types.Name{Name: v[0]}); alias == nil {
				return fmt.Errorf("type %v references alias %q which does not exist", b.t, v[0])
			}
		// protobuf.embed instructs the generator to use the named type in this package
		// as an embedded message.
		case k == "protobuf.embed":
			***REMOVED***elds = []protoField{
				{
					Tag:  1,
					Name: v[0],
					Type: &types.Type{
						Name: types.Name{
							Name:    v[0],
							Package: b.localPackage.Package,
							Path:    b.localPackage.Path,
						},
					},
				},
			}
		}
	}
	if alias == nil {
		alias = b.t
	}

	// If we don't explicitly embed anything, generate ***REMOVED***elds by traversing ***REMOVED***elds.
	if ***REMOVED***elds == nil {
		memberFields, err := membersToFields(b.locator, alias, b.localPackage, b.omitFieldTypes)
		if err != nil {
			return fmt.Errorf("type %v cannot be converted to protobuf: %v", b.t, err)
		}
		***REMOVED***elds = memberFields
	}

	out := sw.Out()
	genComment(out, b.t.CommentLines, "")
	sw.Do(`message $.Name.Name$ {
`, b.t)

	if len(options) > 0 {
		sort.Sort(sort.StringSlice(options))
		for _, s := range options {
			fmt.Fprintf(out, "  option %s;\n", s)
		}
		fmt.Fprintln(out)
	}

	for i, ***REMOVED***eld := range ***REMOVED***elds {
		genComment(out, ***REMOVED***eld.CommentLines, "  ")
		fmt.Fprintf(out, "  ")
		switch {
		case ***REMOVED***eld.Map:
		case ***REMOVED***eld.Repeated:
			fmt.Fprintf(out, "repeated ")
		case ***REMOVED***eld.Required:
			fmt.Fprintf(out, "required ")
		default:
			fmt.Fprintf(out, "optional ")
		}
		sw.Do(`$.Type|local$ $.Name$ = $.Tag$`, ***REMOVED***eld)
		if len(***REMOVED***eld.Extras) > 0 {
			extras := []string{}
			for k, v := range ***REMOVED***eld.Extras {
				if b.omitGogo && strings.HasPre***REMOVED***x(k, "(gogoproto.") {
					continue
				}
				extras = append(extras, fmt.Sprintf("%s = %s", k, v))
			}
			sort.Sort(sort.StringSlice(extras))
			if len(extras) > 0 {
				fmt.Fprintf(out, " [")
				fmt.Fprint(out, strings.Join(extras, ", "))
				fmt.Fprintf(out, "]")
			}
		}
		fmt.Fprintf(out, ";\n")
		if i != len(***REMOVED***elds)-1 {
			fmt.Fprintf(out, "\n")
		}
	}
	fmt.Fprintf(out, "}\n\n")
	return nil
}

type protoField struct {
	LocalPackage types.Name

	Tag      int
	Name     string
	Type     *types.Type
	Map      bool
	Repeated bool
	Optional bool
	Required bool
	Nullable bool
	Extras   map[string]string

	CommentLines []string
}

var (
	errUnrecognizedType = fmt.Errorf("did not recognize the provided type")
)

func isFundamentalProtoType(t *types.Type) (*types.Type, bool) {
	// TODO: when we enable proto3, also include other fundamental types in the google.protobuf package
	// switch {
	// case t.Kind == types.Struct && t.Name == types.Name{Package: "time", Name: "Time"}:
	// 	return &types.Type{
	// 		Kind: types.Protobuf,
	// 		Name: types.Name{Path: "google/protobuf/timestamp.proto", Package: "google.protobuf", Name: "Timestamp"},
	// 	}, true
	// }
	switch t.Kind {
	case types.Slice:
		if t.Elem.Name.Name == "byte" && len(t.Elem.Name.Package) == 0 {
			return &types.Type{Name: types.Name{Name: "bytes"}, Kind: types.Protobuf}, true
		}
	case types.Builtin:
		switch t.Name.Name {
		case "string", "uint32", "int32", "uint64", "int64", "bool":
			return &types.Type{Name: types.Name{Name: t.Name.Name}, Kind: types.Protobuf}, true
		case "int":
			return &types.Type{Name: types.Name{Name: "int64"}, Kind: types.Protobuf}, true
		case "uint":
			return &types.Type{Name: types.Name{Name: "uint64"}, Kind: types.Protobuf}, true
		case "float64", "float":
			return &types.Type{Name: types.Name{Name: "double"}, Kind: types.Protobuf}, true
		case "float32":
			return &types.Type{Name: types.Name{Name: "float"}, Kind: types.Protobuf}, true
		case "uintptr":
			return &types.Type{Name: types.Name{Name: "uint64"}, Kind: types.Protobuf}, true
		}
		// TODO: complex?
	}
	return t, false
}

func memberTypeToProtobufField(locator ProtobufLocator, ***REMOVED***eld *protoField, t *types.Type) error {
	var err error
	switch t.Kind {
	case types.Protobuf:
		***REMOVED***eld.Type, err = locator.ProtoTypeFor(t)
	case types.Builtin:
		***REMOVED***eld.Type, err = locator.ProtoTypeFor(t)
	case types.Map:
		valueField := &protoField{}
		if err := memberTypeToProtobufField(locator, valueField, t.Elem); err != nil {
			return err
		}
		keyField := &protoField{}
		if err := memberTypeToProtobufField(locator, keyField, t.Key); err != nil {
			return err
		}
		// All other protobuf types have kind types.Protobuf, so setting types.Map
		// here would be very misleading.
		***REMOVED***eld.Type = &types.Type{
			Kind: types.Protobuf,
			Key:  keyField.Type,
			Elem: valueField.Type,
		}
		if !strings.HasPre***REMOVED***x(t.Name.Name, "map[") {
			***REMOVED***eld.Extras["(gogoproto.casttype)"] = strconv.Quote(locator.CastTypeName(t.Name))
		}
		if k, ok := keyField.Extras["(gogoproto.casttype)"]; ok {
			***REMOVED***eld.Extras["(gogoproto.castkey)"] = k
		}
		if v, ok := valueField.Extras["(gogoproto.casttype)"]; ok {
			***REMOVED***eld.Extras["(gogoproto.castvalue)"] = v
		}
		***REMOVED***eld.Map = true
	case types.Pointer:
		if err := memberTypeToProtobufField(locator, ***REMOVED***eld, t.Elem); err != nil {
			return err
		}
		***REMOVED***eld.Nullable = true
	case types.Alias:
		if isOptionalAlias(t) {
			***REMOVED***eld.Type, err = locator.ProtoTypeFor(t)
			***REMOVED***eld.Nullable = true
		} ***REMOVED*** {
			if err := memberTypeToProtobufField(locator, ***REMOVED***eld, t.Underlying); err != nil {
				log.Printf("failed to alias: %s %s: err %v", t.Name, t.Underlying.Name, err)
				return err
			}
			// If this is not an alias to a slice, cast to the alias
			if !***REMOVED***eld.Repeated {
				if ***REMOVED***eld.Extras == nil {
					***REMOVED***eld.Extras = make(map[string]string)
				}
				***REMOVED***eld.Extras["(gogoproto.casttype)"] = strconv.Quote(locator.CastTypeName(t.Name))
			}
		}
	case types.Slice:
		if t.Elem.Name.Name == "byte" && len(t.Elem.Name.Package) == 0 {
			***REMOVED***eld.Type = &types.Type{Name: types.Name{Name: "bytes"}, Kind: types.Protobuf}
			return nil
		}
		if err := memberTypeToProtobufField(locator, ***REMOVED***eld, t.Elem); err != nil {
			return err
		}
		***REMOVED***eld.Repeated = true
	case types.Struct:
		if len(t.Name.Name) == 0 {
			return errUnrecognizedType
		}
		***REMOVED***eld.Type, err = locator.ProtoTypeFor(t)
		***REMOVED***eld.Nullable = false
	default:
		return errUnrecognizedType
	}
	return err
}

// protobufTagToField extracts information from an existing protobuf tag
func protobufTagToField(tag string, ***REMOVED***eld *protoField, m types.Member, t *types.Type, localPackage types.Name) error {
	if len(tag) == 0 || tag == "-" {
		return nil
	}

	// protobuf:"bytes,3,opt,name=Id,customtype=github.com/gogo/protobuf/test.Uuid"
	parts := strings.Split(tag, ",")
	if len(parts) < 3 {
		return fmt.Errorf("member %q of %q malformed 'protobuf' tag, not enough segments\n", m.Name, t.Name)
	}
	protoTag, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("member %q of %q malformed 'protobuf' tag, ***REMOVED***eld ID is %q which is not an integer: %v\n", m.Name, t.Name, parts[1], err)
	}
	***REMOVED***eld.Tag = protoTag

	// In general there is doesn't make sense to parse the protobuf tags to get the type,
	// as all auto-generated once will have wire type "bytes", "varint" or "***REMOVED***xed64".
	// However, sometimes we explicitly set them to have a custom serialization, e.g.:
	//   type Time struct {
	//     time.Time `protobuf:"Timestamp,1,req,name=time"`
	//   }
	// to force the generator to use a given type (that we manually wrote serialization &
	// deserialization methods for).
	switch parts[0] {
	case "varint", "***REMOVED***xed32", "***REMOVED***xed64", "bytes", "group":
	default:
		name := types.Name{}
		if last := strings.LastIndex(parts[0], "."); last != -1 {
			pre***REMOVED***x := parts[0][:last]
			name = types.Name{
				Name:    parts[0][last+1:],
				Package: pre***REMOVED***x,
				Path:    strings.Replace(pre***REMOVED***x, ".", "/", -1),
			}
		} ***REMOVED*** {
			name = types.Name{
				Name:    parts[0],
				Package: localPackage.Package,
				Path:    localPackage.Path,
			}
		}
		***REMOVED***eld.Type = &types.Type{
			Name: name,
			Kind: types.Protobuf,
		}
	}

	protoExtra := make(map[string]string)
	for i, extra := range parts[3:] {
		parts := strings.SplitN(extra, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("member %q of %q malformed 'protobuf' tag, tag %d should be key=value, got %q\n", m.Name, t.Name, i+4, extra)
		}
		switch parts[0] {
		case "name":
			protoExtra[parts[0]] = parts[1]
		case "casttype", "castkey", "castvalue":
			parts[0] = fmt.Sprintf("(gogoproto.%s)", parts[0])
			protoExtra[parts[0]] = strconv.Quote(parts[1])
		}
	}

	***REMOVED***eld.Extras = protoExtra
	if name, ok := protoExtra["name"]; ok {
		***REMOVED***eld.Name = name
		delete(protoExtra, "name")
	}

	return nil
}

func membersToFields(locator ProtobufLocator, t *types.Type, localPackage types.Name, omitFieldTypes map[types.Name]struct{}) ([]protoField, error) {
	***REMOVED***elds := []protoField{}

	for _, m := range t.Members {
		if namer.IsPrivateGoName(m.Name) {
			// skip private ***REMOVED***elds
			continue
		}
		if _, ok := omitFieldTypes[types.Name{Name: m.Type.Name.Name, Package: m.Type.Name.Package}]; ok {
			continue
		}
		tags := reflect.StructTag(m.Tags)
		***REMOVED***eld := protoField{
			LocalPackage: localPackage,

			Tag:    -1,
			Extras: make(map[string]string),
		}

		protobufTag := tags.Get("protobuf")
		if protobufTag == "-" {
			continue
		}

		if err := protobufTagToField(protobufTag, &***REMOVED***eld, m, t, localPackage); err != nil {
			return nil, err
		}

		// extract information from JSON ***REMOVED***eld tag
		if tag := tags.Get("json"); len(tag) > 0 {
			parts := strings.Split(tag, ",")
			if len(***REMOVED***eld.Name) == 0 && len(parts[0]) != 0 {
				***REMOVED***eld.Name = parts[0]
			}
			if ***REMOVED***eld.Tag == -1 && ***REMOVED***eld.Name == "-" {
				continue
			}
		}

		if ***REMOVED***eld.Type == nil {
			if err := memberTypeToProtobufField(locator, &***REMOVED***eld, m.Type); err != nil {
				return nil, fmt.Errorf("unable to embed type %q as ***REMOVED***eld %q in %q: %v", m.Type, ***REMOVED***eld.Name, t.Name, err)
			}
		}
		if len(***REMOVED***eld.Name) == 0 {
			***REMOVED***eld.Name = namer.IL(m.Name)
		}

		if ***REMOVED***eld.Map && ***REMOVED***eld.Repeated {
			// maps cannot be repeated
			***REMOVED***eld.Repeated = false
			***REMOVED***eld.Nullable = true
		}

		if !***REMOVED***eld.Nullable {
			***REMOVED***eld.Extras["(gogoproto.nullable)"] = "false"
		}
		if (***REMOVED***eld.Type.Name.Name == "bytes" && ***REMOVED***eld.Type.Name.Package == "") || (***REMOVED***eld.Repeated && ***REMOVED***eld.Type.Name.Package == "" && namer.IsPrivateGoName(***REMOVED***eld.Type.Name.Name)) {
			delete(***REMOVED***eld.Extras, "(gogoproto.nullable)")
		}
		if ***REMOVED***eld.Name != m.Name {
			***REMOVED***eld.Extras["(gogoproto.customname)"] = strconv.Quote(m.Name)
		}
		***REMOVED***eld.CommentLines = m.CommentLines
		***REMOVED***elds = append(***REMOVED***elds, ***REMOVED***eld)
	}

	// assign tags
	highest := 0
	byTag := make(map[int]*protoField)
	// ***REMOVED***elds are in Go struct order, which we preserve
	for i := range ***REMOVED***elds {
		***REMOVED***eld := &***REMOVED***elds[i]
		tag := ***REMOVED***eld.Tag
		if tag != -1 {
			if existing, ok := byTag[tag]; ok {
				return nil, fmt.Errorf("***REMOVED***eld %q and %q both have tag %d", ***REMOVED***eld.Name, existing.Name, tag)
			}
			byTag[tag] = ***REMOVED***eld
		}
		if tag > highest {
			highest = tag
		}
	}
	// starting from the highest observed tag, assign new ***REMOVED***eld tags
	for i := range ***REMOVED***elds {
		***REMOVED***eld := &***REMOVED***elds[i]
		if ***REMOVED***eld.Tag != -1 {
			continue
		}
		highest++
		***REMOVED***eld.Tag = highest
		byTag[***REMOVED***eld.Tag] = ***REMOVED***eld
	}
	return ***REMOVED***elds, nil
}

func genComment(out io.Writer, lines []string, indent string) {
	for {
		l := len(lines)
		if l == 0 || len(lines[l-1]) != 0 {
			break
		}
		lines = lines[:l-1]
	}
	for _, c := range lines {
		if len(c) == 0 {
			fmt.Fprintf(out, "%s//\n", indent) // avoid trailing whitespace
			continue
		}
		fmt.Fprintf(out, "%s// %s\n", indent, c)
	}
}

func formatProtoFile(source []byte) ([]byte, error) {
	// TODO; Is there any protobuf formatter?
	return source, nil
}

func assembleProtoFile(w io.Writer, f *generator.File) {
	w.Write(f.Header)

	fmt.Fprint(w, "syntax = 'proto2';\n\n")

	if len(f.PackageName) > 0 {
		fmt.Fprintf(w, "package %s;\n\n", f.PackageName)
	}

	if len(f.Imports) > 0 {
		imports := []string{}
		for i := range f.Imports {
			imports = append(imports, i)
		}
		sort.Strings(imports)
		for _, s := range imports {
			fmt.Fprintf(w, "import %q;\n", s)
		}
		fmt.Fprint(w, "\n")
	}

	if f.Vars.Len() > 0 {
		fmt.Fprintf(w, "%s\n", f.Vars.String())
	}

	w.Write(f.Body.Bytes())
}

func NewProtoFile() *generator.DefaultFileType {
	return &generator.DefaultFileType{
		Format:   formatProtoFile,
		Assemble: assembleProtoFile,
	}
}
