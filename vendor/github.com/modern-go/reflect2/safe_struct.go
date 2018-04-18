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