package log

import (
	"fmt"
	"math"
)

type ***REMOVED***eldType int

const (
	stringType ***REMOVED***eldType = iota
	boolType
	intType
	int32Type
	uint32Type
	int64Type
	uint64Type
	float32Type
	float64Type
	errorType
	objectType
	lazyLoggerType
	noopType
)

// Field instances are constructed via LogBool, LogString, and so on.
// Tracing implementations may then handle them via the Field.Marshal
// method.
//
// "heavily influenced by" (i.e., partially stolen from)
// https://github.com/uber-go/zap
type Field struct {
	key          string
	***REMOVED***eldType    ***REMOVED***eldType
	numericVal   int64
	stringVal    string
	interfaceVal interface{}
}

// String adds a string-valued key:value pair to a Span.LogFields() record
func String(key, val string) Field {
	return Field{
		key:       key,
		***REMOVED***eldType: stringType,
		stringVal: val,
	}
}

// Bool adds a bool-valued key:value pair to a Span.LogFields() record
func Bool(key string, val bool) Field {
	var numericVal int64
	if val {
		numericVal = 1
	}
	return Field{
		key:        key,
		***REMOVED***eldType:  boolType,
		numericVal: numericVal,
	}
}

// Int adds an int-valued key:value pair to a Span.LogFields() record
func Int(key string, val int) Field {
	return Field{
		key:        key,
		***REMOVED***eldType:  intType,
		numericVal: int64(val),
	}
}

// Int32 adds an int32-valued key:value pair to a Span.LogFields() record
func Int32(key string, val int32) Field {
	return Field{
		key:        key,
		***REMOVED***eldType:  int32Type,
		numericVal: int64(val),
	}
}

// Int64 adds an int64-valued key:value pair to a Span.LogFields() record
func Int64(key string, val int64) Field {
	return Field{
		key:        key,
		***REMOVED***eldType:  int64Type,
		numericVal: val,
	}
}

// Uint32 adds a uint32-valued key:value pair to a Span.LogFields() record
func Uint32(key string, val uint32) Field {
	return Field{
		key:        key,
		***REMOVED***eldType:  uint32Type,
		numericVal: int64(val),
	}
}

// Uint64 adds a uint64-valued key:value pair to a Span.LogFields() record
func Uint64(key string, val uint64) Field {
	return Field{
		key:        key,
		***REMOVED***eldType:  uint64Type,
		numericVal: int64(val),
	}
}

// Float32 adds a float32-valued key:value pair to a Span.LogFields() record
func Float32(key string, val float32) Field {
	return Field{
		key:        key,
		***REMOVED***eldType:  float32Type,
		numericVal: int64(math.Float32bits(val)),
	}
}

// Float64 adds a float64-valued key:value pair to a Span.LogFields() record
func Float64(key string, val float64) Field {
	return Field{
		key:        key,
		***REMOVED***eldType:  float64Type,
		numericVal: int64(math.Float64bits(val)),
	}
}

// Error adds an error with the key "error" to a Span.LogFields() record
func Error(err error) Field {
	return Field{
		key:          "error",
		***REMOVED***eldType:    errorType,
		interfaceVal: err,
	}
}

// Object adds an object-valued key:value pair to a Span.LogFields() record
func Object(key string, obj interface{}) Field {
	return Field{
		key:          key,
		***REMOVED***eldType:    objectType,
		interfaceVal: obj,
	}
}

// LazyLogger allows for user-de***REMOVED***ned, late-bound logging of arbitrary data
type LazyLogger func(fv Encoder)

// Lazy adds a LazyLogger to a Span.LogFields() record; the tracing
// implementation will call the LazyLogger function at an inde***REMOVED***nite time in
// the future (after Lazy() returns).
func Lazy(ll LazyLogger) Field {
	return Field{
		***REMOVED***eldType:    lazyLoggerType,
		interfaceVal: ll,
	}
}

// Noop creates a no-op log ***REMOVED***eld that should be ignored by the tracer.
// It can be used to capture optional ***REMOVED***elds, for example those that should
// only be logged in non-production environment:
//
//     func customerField(order *Order) log.Field {
//          if os.Getenv("ENVIRONMENT") == "dev" {
//              return log.String("customer", order.Customer.ID)
//          }
//          return log.Noop()
//     }
//
//     span.LogFields(log.String("event", "purchase"), customerField(order))
//
func Noop() Field {
	return Field{
		***REMOVED***eldType: noopType,
	}
}

// Encoder allows access to the contents of a Field (via a call to
// Field.Marshal).
//
// Tracer implementations typically provide an implementation of Encoder;
// OpenTracing callers typically do not need to concern themselves with it.
type Encoder interface {
	EmitString(key, value string)
	EmitBool(key string, value bool)
	EmitInt(key string, value int)
	EmitInt32(key string, value int32)
	EmitInt64(key string, value int64)
	EmitUint32(key string, value uint32)
	EmitUint64(key string, value uint64)
	EmitFloat32(key string, value float32)
	EmitFloat64(key string, value float64)
	EmitObject(key string, value interface{})
	EmitLazyLogger(value LazyLogger)
}

// Marshal passes a Field instance through to the appropriate
// ***REMOVED***eld-type-speci***REMOVED***c method of an Encoder.
func (lf Field) Marshal(visitor Encoder) {
	switch lf.***REMOVED***eldType {
	case stringType:
		visitor.EmitString(lf.key, lf.stringVal)
	case boolType:
		visitor.EmitBool(lf.key, lf.numericVal != 0)
	case intType:
		visitor.EmitInt(lf.key, int(lf.numericVal))
	case int32Type:
		visitor.EmitInt32(lf.key, int32(lf.numericVal))
	case int64Type:
		visitor.EmitInt64(lf.key, int64(lf.numericVal))
	case uint32Type:
		visitor.EmitUint32(lf.key, uint32(lf.numericVal))
	case uint64Type:
		visitor.EmitUint64(lf.key, uint64(lf.numericVal))
	case float32Type:
		visitor.EmitFloat32(lf.key, math.Float32frombits(uint32(lf.numericVal)))
	case float64Type:
		visitor.EmitFloat64(lf.key, math.Float64frombits(uint64(lf.numericVal)))
	case errorType:
		if err, ok := lf.interfaceVal.(error); ok {
			visitor.EmitString(lf.key, err.Error())
		} ***REMOVED*** {
			visitor.EmitString(lf.key, "<nil>")
		}
	case objectType:
		visitor.EmitObject(lf.key, lf.interfaceVal)
	case lazyLoggerType:
		visitor.EmitLazyLogger(lf.interfaceVal.(LazyLogger))
	case noopType:
		// intentionally left blank
	}
}

// Key returns the ***REMOVED***eld's key.
func (lf Field) Key() string {
	return lf.key
}

// Value returns the ***REMOVED***eld's value as interface{}.
func (lf Field) Value() interface{} {
	switch lf.***REMOVED***eldType {
	case stringType:
		return lf.stringVal
	case boolType:
		return lf.numericVal != 0
	case intType:
		return int(lf.numericVal)
	case int32Type:
		return int32(lf.numericVal)
	case int64Type:
		return int64(lf.numericVal)
	case uint32Type:
		return uint32(lf.numericVal)
	case uint64Type:
		return uint64(lf.numericVal)
	case float32Type:
		return math.Float32frombits(uint32(lf.numericVal))
	case float64Type:
		return math.Float64frombits(uint64(lf.numericVal))
	case errorType, objectType, lazyLoggerType:
		return lf.interfaceVal
	case noopType:
		return nil
	default:
		return nil
	}
}

// String returns a string representation of the key and value.
func (lf Field) String() string {
	return fmt.Sprint(lf.key, ":", lf.Value())
}
