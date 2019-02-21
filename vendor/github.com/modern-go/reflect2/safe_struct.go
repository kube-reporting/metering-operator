package reflect2

type safeStructType struct {
	safeType
}

func (type2 *safeStructType) FieldByName(name string) StructField {
	***REMOVED***eld, found := type2.Type.FieldByName(name)
	if !found {
		panic("***REMOVED***eld " + name + " not found")
	}
	return &safeField{StructField: ***REMOVED***eld}
}

func (type2 *safeStructType) Field(i int) StructField {
	return &safeField{StructField: type2.Type.Field(i)}
}

func (type2 *safeStructType) FieldByIndex(index []int) StructField {
	return &safeField{StructField: type2.Type.FieldByIndex(index)}
}

func (type2 *safeStructType) FieldByNameFunc(match func(string) bool) StructField {
	***REMOVED***eld, found := type2.Type.FieldByNameFunc(match)
	if !found {
		panic("***REMOVED***eld match condition not found in " + type2.Type.String())
	}
	return &safeField{StructField: ***REMOVED***eld}
}
