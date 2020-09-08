package ldvalue

import (
	"errors"
	"strconv"

	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream"
)

// OptionalInt represents an int that may or may not have a value. This is similar to using an
// int pointer to distinguish between a zero value and nil, but it is safer because it does not
// expose a pointer to any mutable value.
//
// To create an instance with an int value, use NewOptionalInt. There is no corresponding method
// for creating an instance with no value; simply use the empty literal OptionalInt{}.
//
//     oi1 := NewOptionalInt(1)
//     oi2 := NewOptionalInt(0) // this has a value which is zero
//     oi3 := OptionalInt{}     // this does not have a value
//
// This can also be used as a convenient way to construct an int pointer within an expression.
// For instance, this example causes myIntPointer to point to the int value 2:
//
//     var myIntPointer *int = NewOptionalInt("x").AsPointer()
//
// This type is used in ldreason.EvaluationDetail.VariationIndex, and for other similar fields
// in the LaunchDarkly Go SDK where an int value may or may not be defined.
type OptionalInt struct {
	value    int
	hasValue bool
}

// NewOptionalInt constructs an OptionalInt that has an int value.
//
// There is no corresponding method for creating an OptionalInt with no value; simply use the
// empty literal OptionalInt{}.
func NewOptionalInt(value int) OptionalInt {
	return OptionalInt{value: value, hasValue: true}
}

// NewOptionalIntFromPointer constructs an OptionalInt from an int pointer. If the pointer is
// non-nil, then the OptionalInt copies its value; otherwise the OptionalInt has no value.
func NewOptionalIntFromPointer(valuePointer *int) OptionalInt {
	if valuePointer == nil {
		return OptionalInt{hasValue: false}
	}
	return OptionalInt{value: *valuePointer, hasValue: true}
}

// IsDefined returns true if the OptionalInt contains an int value, or false if it has no value.
func (o OptionalInt) IsDefined() bool {
	return o.hasValue
}

// IntValue returns the OptionalInt's value, or zero if it has no value.
func (o OptionalInt) IntValue() int {
	return o.value
}

// Get is a combination of IntValue and IsDefined. If the OptionalInt contains an int value, it
// returns that value and true; otherwise it returns zero and false.
func (o OptionalInt) Get() (int, bool) {
	return o.value, o.hasValue
}

// OrElse returns the OptionalInt's value if it has one, or else the specified fallback value.
func (o OptionalInt) OrElse(valueIfEmpty int) int {
	if o.hasValue {
		return o.value
	}
	return valueIfEmpty
}

// AsPointer returns the OptionalInt's value as an int pointer if it has a value, or nil
// otherwise.
//
// The int value, if any, is copied rather than returning to a pointer to the internal field.
func (o OptionalInt) AsPointer() *int {
	if o.hasValue {
		v := o.value
		return &v
	}
	return nil
}

// AsValue converts the OptionalInt to a Value, which is either Null() or a number value.
func (o OptionalInt) AsValue() Value {
	if o.hasValue {
		return Int(o.value)
	}
	return Null()
}

// String is a debugging convenience method that returns a description of the OptionalInt. This
// is either a numeric string, or "[none]" if it has no value.
func (o OptionalInt) String() string {
	if o.hasValue {
		return strconv.Itoa(o.value)
	}
	return noneDescription
}

// MarshalJSON converts the OptionalInt to its JSON representation.
//
// The output will be either a JSON number or null. Note that the "omitempty" tag for a struct
// field will not cause an empty OptionalInt field to be omitted; it will be output as null.
// If you want to completely omit a JSON property when there is no value, it must be an int
// pointer instead of an OptionalInt; use the AsPointer() method to get a pointer.
func (o OptionalInt) MarshalJSON() ([]byte, error) {
	return o.AsValue().MarshalJSON()
}

// UnmarshalJSON parses an OptionalInt from JSON.
//
// The input must be either a JSON number that is an integer or null.
func (o *OptionalInt) UnmarshalJSON(data []byte) error {
	var v Value
	if err := v.UnmarshalJSON(data); err != nil {
		return err // COVERAGE: should not be possible, parser normally doesn't pass malformed content to UnmarshalJSON
	}
	switch {
	case v.IsNull():
		*o = OptionalInt{}
	case v.IsInt():
		*o = NewOptionalInt(v.IntValue())
	default:
		*o = OptionalInt{}
		return errors.New("expected integer or null")
	}
	return nil
}

// WriteToJSONBuffer provides JSON serialization for OptionalInt with the jsonstream API.
//
// The JSON output format is identical to what is produced by json.Marshal, but this implementation is
// more efficient when building output with JSONBuffer. See the jsonstream package for more details.
func (o OptionalInt) WriteToJSONBuffer(j *jsonstream.JSONBuffer) {
	o.AsValue().WriteToJSONBuffer(j)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (o OptionalInt) MarshalText() ([]byte, error) {
	if o.hasValue {
		return []byte(strconv.Itoa(o.value)), nil
	}
	return []byte(""), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// This allows OptionalInt to be used with packages that can parse text content, such as gcfg.
func (o *OptionalInt) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*o = OptionalInt{}
		return nil
	}
	n, err := strconv.Atoi(string(text))
	if err != nil {
		return err
	}
	*o = NewOptionalInt(n)
	return nil
}