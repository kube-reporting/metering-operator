package rest

import "reflect"

// PayloadMember returns the payload ***REMOVED***eld member of i if there is one, or nil.
func PayloadMember(i interface{}) interface{} {
	if i == nil {
		return nil
	}

	v := reflect.ValueOf(i).Elem()
	if !v.IsValid() {
		return nil
	}
	if ***REMOVED***eld, ok := v.Type().FieldByName("_"); ok {
		if payloadName := ***REMOVED***eld.Tag.Get("payload"); payloadName != "" {
			***REMOVED***eld, _ := v.Type().FieldByName(payloadName)
			if ***REMOVED***eld.Tag.Get("type") != "structure" {
				return nil
			}

			payload := v.FieldByName(payloadName)
			if payload.IsValid() || (payload.Kind() == reflect.Ptr && !payload.IsNil()) {
				return payload.Interface()
			}
		}
	}
	return nil
}

// PayloadType returns the type of a payload ***REMOVED***eld member of i if there is one, or "".
func PayloadType(i interface{}) string {
	v := reflect.Indirect(reflect.ValueOf(i))
	if !v.IsValid() {
		return ""
	}
	if ***REMOVED***eld, ok := v.Type().FieldByName("_"); ok {
		if payloadName := ***REMOVED***eld.Tag.Get("payload"); payloadName != "" {
			if member, ok := v.Type().FieldByName(payloadName); ok {
				return member.Tag.Get("type")
			}
		}
	}
	return ""
}
