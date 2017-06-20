package log

import "fmt"

// InterleavedKVToFields converts keyValues a la Span.LogKV() to a Field slice
// a la Span.LogFields().
func InterleavedKVToFields(keyValues ...interface{}) ([]Field, error) {
	if len(keyValues)%2 != 0 {
		return nil, fmt.Errorf("non-even keyValues len: %d", len(keyValues))
	}
	***REMOVED***elds := make([]Field, len(keyValues)/2)
	for i := 0; i*2 < len(keyValues); i++ {
		key, ok := keyValues[i*2].(string)
		if !ok {
			return nil, fmt.Errorf(
				"non-string key (pair #%d): %T",
				i, keyValues[i*2])
		}
		switch typedVal := keyValues[i*2+1].(type) {
		case bool:
			***REMOVED***elds[i] = Bool(key, typedVal)
		case string:
			***REMOVED***elds[i] = String(key, typedVal)
		case int:
			***REMOVED***elds[i] = Int(key, typedVal)
		case int8:
			***REMOVED***elds[i] = Int32(key, int32(typedVal))
		case int16:
			***REMOVED***elds[i] = Int32(key, int32(typedVal))
		case int32:
			***REMOVED***elds[i] = Int32(key, typedVal)
		case int64:
			***REMOVED***elds[i] = Int64(key, typedVal)
		case uint:
			***REMOVED***elds[i] = Uint64(key, uint64(typedVal))
		case uint64:
			***REMOVED***elds[i] = Uint64(key, typedVal)
		case uint8:
			***REMOVED***elds[i] = Uint32(key, uint32(typedVal))
		case uint16:
			***REMOVED***elds[i] = Uint32(key, uint32(typedVal))
		case uint32:
			***REMOVED***elds[i] = Uint32(key, typedVal)
		case float32:
			***REMOVED***elds[i] = Float32(key, typedVal)
		case float64:
			***REMOVED***elds[i] = Float64(key, typedVal)
		default:
			// When in doubt, coerce to a string
			***REMOVED***elds[i] = String(key, fmt.Sprint(typedVal))
		}
	}
	return ***REMOVED***elds, nil
}
