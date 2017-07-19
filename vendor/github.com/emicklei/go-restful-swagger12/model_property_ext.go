package swagger

import (
	"reflect"
	"strings"
)

func (prop *ModelProperty) setDescription(***REMOVED***eld reflect.StructField) {
	if tag := ***REMOVED***eld.Tag.Get("description"); tag != "" {
		prop.Description = tag
	}
}

func (prop *ModelProperty) setDefaultValue(***REMOVED***eld reflect.StructField) {
	if tag := ***REMOVED***eld.Tag.Get("default"); tag != "" {
		prop.DefaultValue = Special(tag)
	}
}

func (prop *ModelProperty) setEnumValues(***REMOVED***eld reflect.StructField) {
	// We use | to separate the enum values.  This value is chosen
	// since its unlikely to be useful in actual enumeration values.
	if tag := ***REMOVED***eld.Tag.Get("enum"); tag != "" {
		prop.Enum = strings.Split(tag, "|")
	}
}

func (prop *ModelProperty) setMaximum(***REMOVED***eld reflect.StructField) {
	if tag := ***REMOVED***eld.Tag.Get("maximum"); tag != "" {
		prop.Maximum = tag
	}
}

func (prop *ModelProperty) setType(***REMOVED***eld reflect.StructField) {
	if tag := ***REMOVED***eld.Tag.Get("type"); tag != "" {
		// Check if the ***REMOVED***rst two characters of the type tag are
		// intended to emulate slice/array behaviour.
		//
		// If type is intended to be a slice/array then add the
		// overriden type to the array item instead of the main property
		if len(tag) > 2 && tag[0:2] == "[]" {
			pType := "array"
			prop.Type = &pType
			prop.Items = new(Item)

			iType := tag[2:]
			prop.Items.Type = &iType
			return
		}

		prop.Type = &tag
	}
}

func (prop *ModelProperty) setMinimum(***REMOVED***eld reflect.StructField) {
	if tag := ***REMOVED***eld.Tag.Get("minimum"); tag != "" {
		prop.Minimum = tag
	}
}

func (prop *ModelProperty) setUniqueItems(***REMOVED***eld reflect.StructField) {
	tag := ***REMOVED***eld.Tag.Get("unique")
	switch tag {
	case "true":
		v := true
		prop.UniqueItems = &v
	case "false":
		v := false
		prop.UniqueItems = &v
	}
}

func (prop *ModelProperty) setPropertyMetadata(***REMOVED***eld reflect.StructField) {
	prop.setDescription(***REMOVED***eld)
	prop.setEnumValues(***REMOVED***eld)
	prop.setMinimum(***REMOVED***eld)
	prop.setMaximum(***REMOVED***eld)
	prop.setUniqueItems(***REMOVED***eld)
	prop.setDefaultValue(***REMOVED***eld)
	prop.setType(***REMOVED***eld)
}
