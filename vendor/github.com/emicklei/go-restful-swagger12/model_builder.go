package swagger

import (
	"encoding/json"
	"reflect"
	"strings"
)

// ModelBuildable is used for extending Structs that need more control over
// how the Model appears in the Swagger api declaration.
type ModelBuildable interface {
	PostBuildModel(m *Model) *Model
}

type modelBuilder struct {
	Models *ModelList
	Con***REMOVED***g *Con***REMOVED***g
}

type documentable interface {
	SwaggerDoc() map[string]string
}

// Check if this structure has a method with signature func (<theModel>) SwaggerDoc() map[string]string
// If it exists, retrive the documentation and overwrite all struct tag descriptions
func getDocFromMethodSwaggerDoc2(model reflect.Type) map[string]string {
	if docable, ok := reflect.New(model).Elem().Interface().(documentable); ok {
		return docable.SwaggerDoc()
	}
	return make(map[string]string)
}

// addModelFrom creates and adds a Model to the builder and detects and calls
// the post build hook for customizations
func (b modelBuilder) addModelFrom(sample interface{}) {
	if modelOrNil := b.addModel(reflect.TypeOf(sample), ""); modelOrNil != nil {
		// allow customizations
		if buildable, ok := sample.(ModelBuildable); ok {
			modelOrNil = buildable.PostBuildModel(modelOrNil)
			b.Models.Put(modelOrNil.Id, *modelOrNil)
		}
	}
}

func (b modelBuilder) addModel(st reflect.Type, nameOverride string) *Model {
	// Turn pointers into simpler types so further checks are
	// correct.
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
	}

	modelName := b.keyFrom(st)
	if nameOverride != "" {
		modelName = nameOverride
	}
	// no models needed for primitive types
	if b.isPrimitiveType(modelName) {
		return nil
	}
	// golang encoding/json packages says array and slice values encode as
	// JSON arrays, except that []byte encodes as a base64-encoded string.
	// If we see a []byte here, treat it at as a primitive type (string)
	// and deal with it in buildArrayTypeProperty.
	if (st.Kind() == reflect.Slice || st.Kind() == reflect.Array) &&
		st.Elem().Kind() == reflect.Uint8 {
		return nil
	}
	// see if we already have visited this model
	if _, ok := b.Models.At(modelName); ok {
		return nil
	}
	sm := Model{
		Id:         modelName,
		Required:   []string{},
		Properties: ModelPropertyList{}}

	// reference the model before further initializing (enables recursive structs)
	b.Models.Put(modelName, sm)

	// check for slice or array
	if st.Kind() == reflect.Slice || st.Kind() == reflect.Array {
		b.addModel(st.Elem(), "")
		return &sm
	}
	// check for structure or primitive type
	if st.Kind() != reflect.Struct {
		return &sm
	}

	fullDoc := getDocFromMethodSwaggerDoc2(st)
	modelDescriptions := []string{}

	for i := 0; i < st.NumField(); i++ {
		***REMOVED***eld := st.Field(i)
		jsonName, modelDescription, prop := b.buildProperty(***REMOVED***eld, &sm, modelName)
		if len(modelDescription) > 0 {
			modelDescriptions = append(modelDescriptions, modelDescription)
		}

		// add if not omitted
		if len(jsonName) != 0 {
			// update description
			if ***REMOVED***eldDoc, ok := fullDoc[jsonName]; ok {
				prop.Description = ***REMOVED***eldDoc
			}
			// update Required
			if b.isPropertyRequired(***REMOVED***eld) {
				sm.Required = append(sm.Required, jsonName)
			}
			sm.Properties.Put(jsonName, prop)
		}
	}

	// We always overwrite documentation if SwaggerDoc method exists
	// "" is special for documenting the struct itself
	if modelDoc, ok := fullDoc[""]; ok {
		sm.Description = modelDoc
	} ***REMOVED*** if len(modelDescriptions) != 0 {
		sm.Description = strings.Join(modelDescriptions, "\n")
	}

	// update model builder with completed model
	b.Models.Put(modelName, sm)

	return &sm
}

func (b modelBuilder) isPropertyRequired(***REMOVED***eld reflect.StructField) bool {
	required := true
	if jsonTag := ***REMOVED***eld.Tag.Get("json"); jsonTag != "" {
		s := strings.Split(jsonTag, ",")
		if len(s) > 1 && s[1] == "omitempty" {
			return false
		}
	}
	return required
}

func (b modelBuilder) buildProperty(***REMOVED***eld reflect.StructField, model *Model, modelName string) (jsonName, modelDescription string, prop ModelProperty) {
	jsonName = b.jsonNameOfField(***REMOVED***eld)
	if len(jsonName) == 0 {
		// empty name signals skip property
		return "", "", prop
	}

	if ***REMOVED***eld.Name == "XMLName" && ***REMOVED***eld.Type.String() == "xml.Name" {
		// property is metadata for the xml.Name attribute, can be skipped
		return "", "", prop
	}

	if tag := ***REMOVED***eld.Tag.Get("modelDescription"); tag != "" {
		modelDescription = tag
	}

	prop.setPropertyMetadata(***REMOVED***eld)
	if prop.Type != nil {
		return jsonName, modelDescription, prop
	}
	***REMOVED***eldType := ***REMOVED***eld.Type

	// check if type is doing its own marshalling
	marshalerType := reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	if ***REMOVED***eldType.Implements(marshalerType) {
		var pType = "string"
		if prop.Type == nil {
			prop.Type = &pType
		}
		if prop.Format == "" {
			prop.Format = b.jsonSchemaFormat(b.keyFrom(***REMOVED***eldType))
		}
		return jsonName, modelDescription, prop
	}

	// check if annotation says it is a string
	if jsonTag := ***REMOVED***eld.Tag.Get("json"); jsonTag != "" {
		s := strings.Split(jsonTag, ",")
		if len(s) > 1 && s[1] == "string" {
			stringt := "string"
			prop.Type = &stringt
			return jsonName, modelDescription, prop
		}
	}

	***REMOVED***eldKind := ***REMOVED***eldType.Kind()
	switch {
	case ***REMOVED***eldKind == reflect.Struct:
		jsonName, prop := b.buildStructTypeProperty(***REMOVED***eld, jsonName, model)
		return jsonName, modelDescription, prop
	case ***REMOVED***eldKind == reflect.Slice || ***REMOVED***eldKind == reflect.Array:
		jsonName, prop := b.buildArrayTypeProperty(***REMOVED***eld, jsonName, modelName)
		return jsonName, modelDescription, prop
	case ***REMOVED***eldKind == reflect.Ptr:
		jsonName, prop := b.buildPointerTypeProperty(***REMOVED***eld, jsonName, modelName)
		return jsonName, modelDescription, prop
	case ***REMOVED***eldKind == reflect.String:
		stringt := "string"
		prop.Type = &stringt
		return jsonName, modelDescription, prop
	case ***REMOVED***eldKind == reflect.Map:
		// if it's a map, it's unstructured, and swagger 1.2 can't handle it
		objectType := "object"
		prop.Type = &objectType
		return jsonName, modelDescription, prop
	}

	***REMOVED***eldTypeName := b.keyFrom(***REMOVED***eldType)
	if b.isPrimitiveType(***REMOVED***eldTypeName) {
		mapped := b.jsonSchemaType(***REMOVED***eldTypeName)
		prop.Type = &mapped
		prop.Format = b.jsonSchemaFormat(***REMOVED***eldTypeName)
		return jsonName, modelDescription, prop
	}
	modelType := b.keyFrom(***REMOVED***eldType)
	prop.Ref = &modelType

	if ***REMOVED***eldType.Name() == "" { // override type of anonymous structs
		nestedTypeName := modelName + "." + jsonName
		prop.Ref = &nestedTypeName
		b.addModel(***REMOVED***eldType, nestedTypeName)
	}
	return jsonName, modelDescription, prop
}

func hasNamedJSONTag(***REMOVED***eld reflect.StructField) bool {
	parts := strings.Split(***REMOVED***eld.Tag.Get("json"), ",")
	if len(parts) == 0 {
		return false
	}
	for _, s := range parts[1:] {
		if s == "inline" {
			return false
		}
	}
	return len(parts[0]) > 0
}

func (b modelBuilder) buildStructTypeProperty(***REMOVED***eld reflect.StructField, jsonName string, model *Model) (nameJson string, prop ModelProperty) {
	prop.setPropertyMetadata(***REMOVED***eld)
	// Check for type override in tag
	if prop.Type != nil {
		return jsonName, prop
	}
	***REMOVED***eldType := ***REMOVED***eld.Type
	// check for anonymous
	if len(***REMOVED***eldType.Name()) == 0 {
		// anonymous
		anonType := model.Id + "." + jsonName
		b.addModel(***REMOVED***eldType, anonType)
		prop.Ref = &anonType
		return jsonName, prop
	}

	if ***REMOVED***eld.Name == ***REMOVED***eldType.Name() && ***REMOVED***eld.Anonymous && !hasNamedJSONTag(***REMOVED***eld) {
		// embedded struct
		sub := modelBuilder{new(ModelList), b.Con***REMOVED***g}
		sub.addModel(***REMOVED***eldType, "")
		subKey := sub.keyFrom(***REMOVED***eldType)
		// merge properties from sub
		subModel, _ := sub.Models.At(subKey)
		subModel.Properties.Do(func(k string, v ModelProperty) {
			model.Properties.Put(k, v)
			// if subModel says this property is required then include it
			required := false
			for _, each := range subModel.Required {
				if k == each {
					required = true
					break
				}
			}
			if required {
				model.Required = append(model.Required, k)
			}
		})
		// add all new referenced models
		sub.Models.Do(func(key string, sub Model) {
			if key != subKey {
				if _, ok := b.Models.At(key); !ok {
					b.Models.Put(key, sub)
				}
			}
		})
		// empty name signals skip property
		return "", prop
	}
	// simple struct
	b.addModel(***REMOVED***eldType, "")
	var pType = b.keyFrom(***REMOVED***eldType)
	prop.Ref = &pType
	return jsonName, prop
}

func (b modelBuilder) buildArrayTypeProperty(***REMOVED***eld reflect.StructField, jsonName, modelName string) (nameJson string, prop ModelProperty) {
	// check for type override in tags
	prop.setPropertyMetadata(***REMOVED***eld)
	if prop.Type != nil {
		return jsonName, prop
	}
	***REMOVED***eldType := ***REMOVED***eld.Type
	if ***REMOVED***eldType.Elem().Kind() == reflect.Uint8 {
		stringt := "string"
		prop.Type = &stringt
		return jsonName, prop
	}
	var pType = "array"
	prop.Type = &pType
	isPrimitive := b.isPrimitiveType(***REMOVED***eldType.Elem().Name())
	elemTypeName := b.getElementTypeName(modelName, jsonName, ***REMOVED***eldType.Elem())
	prop.Items = new(Item)
	if isPrimitive {
		mapped := b.jsonSchemaType(elemTypeName)
		prop.Items.Type = &mapped
	} ***REMOVED*** {
		prop.Items.Ref = &elemTypeName
	}
	// add|overwrite model for element type
	if ***REMOVED***eldType.Elem().Kind() == reflect.Ptr {
		***REMOVED***eldType = ***REMOVED***eldType.Elem()
	}
	if !isPrimitive {
		b.addModel(***REMOVED***eldType.Elem(), elemTypeName)
	}
	return jsonName, prop
}

func (b modelBuilder) buildPointerTypeProperty(***REMOVED***eld reflect.StructField, jsonName, modelName string) (nameJson string, prop ModelProperty) {
	prop.setPropertyMetadata(***REMOVED***eld)
	// Check for type override in tags
	if prop.Type != nil {
		return jsonName, prop
	}
	***REMOVED***eldType := ***REMOVED***eld.Type

	// override type of pointer to list-likes
	if ***REMOVED***eldType.Elem().Kind() == reflect.Slice || ***REMOVED***eldType.Elem().Kind() == reflect.Array {
		var pType = "array"
		prop.Type = &pType
		isPrimitive := b.isPrimitiveType(***REMOVED***eldType.Elem().Elem().Name())
		elemName := b.getElementTypeName(modelName, jsonName, ***REMOVED***eldType.Elem().Elem())
		if isPrimitive {
			primName := b.jsonSchemaType(elemName)
			prop.Items = &Item{Ref: &primName}
		} ***REMOVED*** {
			prop.Items = &Item{Ref: &elemName}
		}
		if !isPrimitive {
			// add|overwrite model for element type
			b.addModel(***REMOVED***eldType.Elem().Elem(), elemName)
		}
	} ***REMOVED*** {
		// non-array, pointer type
		***REMOVED***eldTypeName := b.keyFrom(***REMOVED***eldType.Elem())
		var pType = b.jsonSchemaType(***REMOVED***eldTypeName) // no star, include pkg path
		if b.isPrimitiveType(***REMOVED***eldTypeName) {
			prop.Type = &pType
			prop.Format = b.jsonSchemaFormat(***REMOVED***eldTypeName)
			return jsonName, prop
		}
		prop.Ref = &pType
		elemName := ""
		if ***REMOVED***eldType.Elem().Name() == "" {
			elemName = modelName + "." + jsonName
			prop.Ref = &elemName
		}
		b.addModel(***REMOVED***eldType.Elem(), elemName)
	}
	return jsonName, prop
}

func (b modelBuilder) getElementTypeName(modelName, jsonName string, t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Name() == "" {
		return modelName + "." + jsonName
	}
	return b.keyFrom(t)
}

func (b modelBuilder) keyFrom(st reflect.Type) string {
	key := st.String()
	if b.Con***REMOVED***g != nil && b.Con***REMOVED***g.ModelTypeNameHandler != nil {
		if name, ok := b.Con***REMOVED***g.ModelTypeNameHandler(st); ok {
			key = name
		}
	}
	if len(st.Name()) == 0 { // unnamed type
		// Swagger UI has special meaning for [
		key = strings.Replace(key, "[]", "||", -1)
	}
	return key
}

// see also https://golang.org/ref/spec#Numeric_types
func (b modelBuilder) isPrimitiveType(modelName string) bool {
	if len(modelName) == 0 {
		return false
	}
	return strings.Contains("uint uint8 uint16 uint32 uint64 int int8 int16 int32 int64 float32 float64 bool string byte rune time.Time", modelName)
}

// jsonNameOfField returns the name of the ***REMOVED***eld as it should appear in JSON format
// An empty string indicates that this ***REMOVED***eld is not part of the JSON representation
func (b modelBuilder) jsonNameOfField(***REMOVED***eld reflect.StructField) string {
	if jsonTag := ***REMOVED***eld.Tag.Get("json"); jsonTag != "" {
		s := strings.Split(jsonTag, ",")
		if s[0] == "-" {
			// empty name signals skip property
			return ""
		} ***REMOVED*** if s[0] != "" {
			return s[0]
		}
	}
	return ***REMOVED***eld.Name
}

// see also http://json-schema.org/latest/json-schema-core.html#anchor8
func (b modelBuilder) jsonSchemaType(modelName string) string {
	schemaMap := map[string]string{
		"uint":   "integer",
		"uint8":  "integer",
		"uint16": "integer",
		"uint32": "integer",
		"uint64": "integer",

		"int":   "integer",
		"int8":  "integer",
		"int16": "integer",
		"int32": "integer",
		"int64": "integer",

		"byte":      "integer",
		"float64":   "number",
		"float32":   "number",
		"bool":      "boolean",
		"time.Time": "string",
	}
	mapped, ok := schemaMap[modelName]
	if !ok {
		return modelName // use as is (custom or struct)
	}
	return mapped
}

func (b modelBuilder) jsonSchemaFormat(modelName string) string {
	if b.Con***REMOVED***g != nil && b.Con***REMOVED***g.SchemaFormatHandler != nil {
		if mapped := b.Con***REMOVED***g.SchemaFormatHandler(modelName); mapped != "" {
			return mapped
		}
	}
	schemaMap := map[string]string{
		"int":        "int32",
		"int32":      "int32",
		"int64":      "int64",
		"byte":       "byte",
		"uint":       "integer",
		"uint8":      "byte",
		"float64":    "double",
		"float32":    "float",
		"time.Time":  "date-time",
		"*time.Time": "date-time",
	}
	mapped, ok := schemaMap[modelName]
	if !ok {
		return "" // no format
	}
	return mapped
}
