package jsoniter

import (
	"bytes"
	"io"
)

// RawMessage to make replace json with jsoniter
type RawMessage []byte

// Unmarshal adapts to json/encoding Unmarshal API
//
// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
// Refer to https://godoc.org/encoding/json#Unmarshal for more information
func Unmarshal(data []byte, v interface{}) error {
	return Con***REMOVED***gDefault.Unmarshal(data, v)
}

// UnmarshalFromString convenient method to read from string instead of []byte
func UnmarshalFromString(str string, v interface{}) error {
	return Con***REMOVED***gDefault.UnmarshalFromString(str, v)
}

// Get quick method to get value from deeply nested JSON structure
func Get(data []byte, path ...interface{}) Any {
	return Con***REMOVED***gDefault.Get(data, path...)
}

// Marshal adapts to json/encoding Marshal API
//
// Marshal returns the JSON encoding of v, adapts to json/encoding Marshal API
// Refer to https://godoc.org/encoding/json#Marshal for more information
func Marshal(v interface{}) ([]byte, error) {
	return Con***REMOVED***gDefault.Marshal(v)
}

// MarshalIndent same as json.MarshalIndent. Pre***REMOVED***x is not supported.
func MarshalIndent(v interface{}, pre***REMOVED***x, indent string) ([]byte, error) {
	return Con***REMOVED***gDefault.MarshalIndent(v, pre***REMOVED***x, indent)
}

// MarshalToString convenient method to write as string instead of []byte
func MarshalToString(v interface{}) (string, error) {
	return Con***REMOVED***gDefault.MarshalToString(v)
}

// NewDecoder adapts to json/stream NewDecoder API.
//
// NewDecoder returns a new decoder that reads from r.
//
// Instead of a json/encoding Decoder, an Decoder is returned
// Refer to https://godoc.org/encoding/json#NewDecoder for more information
func NewDecoder(reader io.Reader) *Decoder {
	return Con***REMOVED***gDefault.NewDecoder(reader)
}

// Decoder reads and decodes JSON values from an input stream.
// Decoder provides identical APIs with json/stream Decoder (Token() and UseNumber() are in progress)
type Decoder struct {
	iter *Iterator
}

// Decode decode JSON into interface{}
func (adapter *Decoder) Decode(obj interface{}) error {
	if adapter.iter.head == adapter.iter.tail && adapter.iter.reader != nil {
		if !adapter.iter.loadMore() {
			return io.EOF
		}
	}
	adapter.iter.ReadVal(obj)
	err := adapter.iter.Error
	if err == io.EOF {
		return nil
	}
	return adapter.iter.Error
}

// More is there more?
func (adapter *Decoder) More() bool {
	return adapter.iter.head != adapter.iter.tail
}

// Buffered remaining buffer
func (adapter *Decoder) Buffered() io.Reader {
	remaining := adapter.iter.buf[adapter.iter.head:adapter.iter.tail]
	return bytes.NewReader(remaining)
}

// UseNumber causes the Decoder to unmarshal a number into an interface{} as a
// Number instead of as a float64.
func (adapter *Decoder) UseNumber() {
	cfg := adapter.iter.cfg.con***REMOVED***gBeforeFrozen
	cfg.UseNumber = true
	adapter.iter.cfg = cfg.frozeWithCacheReuse()
}

// DisallowUnknownFields causes the Decoder to return an error when the destination
// is a struct and the input contains object keys which do not match any
// non-ignored, exported ***REMOVED***elds in the destination.
func (adapter *Decoder) DisallowUnknownFields() {
	cfg := adapter.iter.cfg.con***REMOVED***gBeforeFrozen
	cfg.DisallowUnknownFields = true
	adapter.iter.cfg = cfg.frozeWithCacheReuse()
}

// NewEncoder same as json.NewEncoder
func NewEncoder(writer io.Writer) *Encoder {
	return Con***REMOVED***gDefault.NewEncoder(writer)
}

// Encoder same as json.Encoder
type Encoder struct {
	stream *Stream
}

// Encode encode interface{} as JSON to io.Writer
func (adapter *Encoder) Encode(val interface{}) error {
	adapter.stream.WriteVal(val)
	adapter.stream.WriteRaw("\n")
	adapter.stream.Flush()
	return adapter.stream.Error
}

// SetIndent set the indention. Pre***REMOVED***x is not supported
func (adapter *Encoder) SetIndent(pre***REMOVED***x, indent string) {
	con***REMOVED***g := adapter.stream.cfg.con***REMOVED***gBeforeFrozen
	con***REMOVED***g.IndentionStep = len(indent)
	adapter.stream.cfg = con***REMOVED***g.frozeWithCacheReuse()
}

// SetEscapeHTML escape html by default, set to false to disable
func (adapter *Encoder) SetEscapeHTML(escapeHTML bool) {
	con***REMOVED***g := adapter.stream.cfg.con***REMOVED***gBeforeFrozen
	con***REMOVED***g.EscapeHTML = escapeHTML
	adapter.stream.cfg = con***REMOVED***g.frozeWithCacheReuse()
}

// Valid reports whether data is a valid JSON encoding.
func Valid(data []byte) bool {
	return Con***REMOVED***gDefault.Valid(data)
}
