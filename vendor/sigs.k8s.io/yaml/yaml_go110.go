// This ***REMOVED***le contains changes that are only compatible with go 1.10 and onwards.

// +build go1.10

package yaml

import "encoding/json"

// DisallowUnknownFields con***REMOVED***gures the JSON decoder to error out if unknown
// ***REMOVED***elds come along, instead of dropping them by default.
func DisallowUnknownFields(d *json.Decoder) *json.Decoder {
	d.DisallowUnknownFields()
	return d
}
