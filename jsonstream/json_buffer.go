package jsonstream

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
)

type streamState int

const (
	stateValueFirst     = iota // no top-level values have been written
	stateValueNext      = iota // at least one top-level value has been written
	stateArrayStart     = iota // array has been started, has no values yet
	stateArrayNext      = iota // array has been started and has at least one value
	stateObjectStart    = iota // object has been started, has no names or values yet
	stateObjectNameNext = iota // object has been started, has name-value pairs, can write a new name
	stateObjectValue    = iota // object has been started, name has been written, needs a value
)

var (
	tokenNull  = []byte("null")  //nolint:gochecknoglobals
	tokenTrue  = []byte("true")  //nolint:gochecknoglobals
	tokenFalse = []byte("false") //nolint:gochecknoglobals
)

// JSONBuffer is a fast JSON encoder for manual generation of sequential output, writing one token at a
// time. Output is written to an in-memory buffer.
//
// Any invalid operation (such as trying to write a property name when a value is expected, or vice versa)
// causes the JSONBuffer to enter a failed state where all subsequent write operations are ignored. The
// error will be returned by Get() or GetError().
//
// If the caller write smore than one JSON value at the top level (that is, not inside an array or object),
// the values will be separated by whatever byte sequence was specified with SetSeparator(), or, if not
// specified, by a newline.
//
// JSONBuffer is not safe for concurrent access by multiple goroutines.
//
//     var buf jsonstream.JSONBuffer
//     buf.BeginObject()
//     buf.WriteName("a")
//     buf.WriteInt(2)
//     buf.EndObject()
//     bytes, err := buf.Get() // bytes == []byte(`{"a":2}`)
type JSONBuffer struct {
	buf                bytes.Buffer
	state              streamState
	stateStackArray    [20]streamState // we use this fixed-size array whenever possible to avoid heap allocations
	stateStackArrayLen int
	stateStackSlice    []streamState // this slice gets allocated on the heap only if necessary
	separator          []byte
	err                error
}

// NewJSONBuffer creates a new JSONBuffer on the heap. This is not strictly necessary; declaring a local
// value of JSONBuffer{} will also work.
func NewJSONBuffer() *JSONBuffer {
	return &JSONBuffer{}
}

// Get returns the full encoded byte slice.
//
// If the buffer is in a failed state from a previous invalid operation, or cannot be ended at this point
// because of an unfinished array or object, Get() returns a nil slice and the error. In that case, the
// data written so far can be accessed with GetPartial().
func (j *JSONBuffer) Get() ([]byte, error) {
	if j.err != nil {
		return nil, j.err
	}
	if j.getStateStackCount() > 0 {
		j.err = errors.New("array or object not ended")
		return nil, j.err
	}
	return j.buf.Bytes(), nil
}

// GetError returns an error if the buffer is in a failed state from a previous invalid operation, or
// nil otherwise.
func (j *JSONBuffer) GetError() error {
	return j.err
}

// GetPartial returns the data written to the buffer so far, regardless of whether it is in a failed or
// incomplete state.
func (j *JSONBuffer) GetPartial() []byte {
	return j.buf.Bytes()
}

// Grow expands the internal buffer by the specified number of bytes. It is the same as calling Grow
// on a bytes.Buffer.
func (j *JSONBuffer) Grow(n int) {
	j.buf.Grow(n)
}

// SetSeparator specifies a byte sequence that should be added to the buffer in between values if more
// than one value is written outside of an array or object. If not specified, a newline is used.
//
//     var buf jsonstream.JSONBuffer
//     buf.SetSeparator([]byte("! "))
//     buf.WriteInt(1)
//     buf.WriteInt(2)
//     buf.WriteInt(3)
//     bytes, err := buf.Get() // bytes == []byte(`1! 2! 3`)
func (j *JSONBuffer) SetSeparator(separator []byte) {
	if separator == nil {
		j.separator = nil
	} else {
		j.separator = make([]byte, len(separator))
		copy(j.separator, separator)
	}
}

// WriteNull writes a JSON null value to the output.
func (j *JSONBuffer) WriteNull() {
	if !j.beforeValue() {
		return
	}
	j.buf.Write(tokenNull)
	j.afterValue()
}

// WriteBool writes a JSON boolean value to the output.
func (j *JSONBuffer) WriteBool(value bool) {
	if !j.beforeValue() {
		return
	}
	if value {
		j.buf.Write(tokenTrue)
	} else {
		j.buf.Write(tokenFalse)
	}
	j.afterValue()
}

// WriteInt writes a JSON numeric value to the output.
func (j *JSONBuffer) WriteInt(value int) {
	if !j.beforeValue() {
		return
	}

	if value == 0 {
		j.buf.WriteRune('0')
		return
	}

	byteSlice := make([]byte, 0, 11) // preallocate on stack with room for any numeric string of this size
	byteSlice = strconv.AppendInt(byteSlice, int64(value), 10)
	j.buf.Write(byteSlice)

	j.afterValue()
}

// WriteUint64 writes a JSON numeric value to the output.
func (j *JSONBuffer) WriteUint64(value uint64) {
	if !j.beforeValue() {
		return
	}

	if value == 0 {
		j.buf.WriteRune('0')
		return
	}

	byteSlice := make([]byte, 0, 25) // preallocate on stack with room for any numeric string of this size
	byteSlice = strconv.AppendUint(byteSlice, value, 10)
	j.buf.Write(byteSlice)

	j.afterValue()
}

// WriteFloat64 writes a JSON numeric value to the output.
func (j *JSONBuffer) WriteFloat64(value float64) {
	if !j.beforeValue() {
		return
	}

	if value == 0 {
		j.buf.WriteRune('0')
		return
	}

	byteSlice := make([]byte, 0, 30) // preallocate on stack with room for most numeric strings of this size
	// (due to how append works, if it happens *not* to be big enough, byteSlice will just escape to the heap)

	byteSlice = strconv.AppendFloat(byteSlice, value, 'f', -1, 64)
	j.buf.Write(byteSlice)

	j.afterValue()
}

// WriteString writes a JSON string value to the output, with quotes and escaping.
func (j *JSONBuffer) WriteString(value string) {
	if !j.beforeValue() {
		return
	}
	j.writeQuotedString(value)
	j.afterValue()
}

// WriteRaw writes a pre-encoded JSON value to the output as-is. Its format is assumed to be correct;
// this operation will not fail unless it is not permitted to write a value at this point.
func (j *JSONBuffer) WriteRaw(value json.RawMessage) {
	if !j.beforeValue() {
		return
	}
	j.buf.Write(value)
	j.afterValue()
}

// BeginArray begins writing a JSON array.
//
// All subsequent values written will be delimited by commas. Call EndArray to finish the array. The
// array may contain any types of values, including nested arrays or objects.
//
//     buf.BeginArray()
//     buf.WriteInt(1)
//     buf.WriteString("b")
//     buf.EndArray() // produces [1,"b"]
func (j *JSONBuffer) BeginArray() {
	if !j.beforeValue() {
		return
	}
	j.buf.WriteRune('[')
	j.pushState(stateArrayStart)
}

// EndArray finishes writing the current JSON array.
func (j *JSONBuffer) EndArray() {
	if j.err != nil {
		return
	}
	if j.state != stateArrayStart && j.state != stateArrayNext {
		j.fail("called EndArray when not inside an array")
		return
	}
	j.buf.WriteRune(']')
	j.popState()
	j.afterValue()
}

// BeginObject begins writing a JSON object.
//
// Until this object is ended, you must call WriteName before each value. Call EndObject to finish
// the object. The object may contain any types of values, including nested objects or arrays.
//
//     buf.BeginObject()
//     buf.WriteName("a")
//     buf.WriteInt(1)
//     buf.WriteName("b")
//     buf.WriteBool(true)
//     buf.EndObject() // produces {"a":1,"b":true}
func (j *JSONBuffer) BeginObject() {
	if !j.beforeValue() {
		return
	}
	j.buf.WriteRune('{')
	j.pushState(stateObjectStart)
}

// WriteName writes a property name within an object.
//
// It is an error to call this method outside of an object, or immediately after another WriteName.
// Each WriteName should be followed by some JSON value (WriteBool, WriteString, BeginArray, etc.).
func (j *JSONBuffer) WriteName(name string) {
	if j.err != nil {
		return
	}
	if j.state != stateObjectStart && j.state != stateObjectNameNext {
		j.fail("called WriteName when a value was expected")
		return
	}
	if j.state == stateObjectNameNext {
		j.buf.WriteRune(',')
	}
	j.writeQuotedString(name)
	j.buf.WriteRune(':')
	j.state = stateObjectValue
}

// EndObject finishes writing the current JSON object.
func (j *JSONBuffer) EndObject() {
	if j.err != nil {
		return
	}
	if j.state == stateObjectValue {
		j.fail("called EndObject when a value was expected")
		return
	}
	if j.state != stateObjectStart && j.state != stateObjectNameNext {
		j.fail("called EndObject when not inside an object")
		return
	}
	j.buf.WriteRune('}')
	j.popState()
	j.afterValue()
}

func (j *JSONBuffer) beforeValue() bool {
	if j.err != nil {
		return false
	}
	switch j.state {
	case stateValueNext:
		if j.separator == nil {
			j.buf.WriteByte('\n')
		} else {
			j.buf.Write(j.separator)
		}
	case stateArrayNext:
		j.buf.WriteByte(',')
	case stateObjectStart:
		j.fail("wrote value when property name was expected")
		return false
	case stateObjectNameNext:
		j.fail("wrote value when property name was expected")
		return false
	}
	return true
}

func (j *JSONBuffer) afterValue() {
	switch j.state {
	case stateValueFirst:
		j.state = stateValueNext
	case stateArrayStart:
		j.state = stateArrayNext
	case stateObjectValue:
		j.state = stateObjectNameNext
	}
}

func (j *JSONBuffer) writeQuotedString(s string) {
	j.buf.WriteRune('"')
	if s == "" {
		j.buf.WriteRune('"')
		return
	}
	foundSpecial := false

	var i int
	var ch rune
	for i, ch = range s {
		if ch == '"' || ch == '\\' || ch < ' ' {
			foundSpecial = true
			break
		}
	}

	if !foundSpecial {
		j.buf.WriteString(s)
	} else {
		start := 0
		len := len(s)
		for i < len {
			if i > start {
				j.buf.WriteString(s[start:i])
			}
			j.writeEscapedChar(ch)
			i++
			start = i
			for i < len {
				ch := s[i]
				if ch == '"' || ch == '\\' || ch < ' ' {
					break
				}
				i++
			}
		}
		if i > start {
			j.buf.WriteString(s[start:i])
		}
	}

	j.buf.WriteRune('"')
}

func (j *JSONBuffer) writeEscapedChar(ch rune) {
	j.buf.WriteRune('\\')
	switch ch {
	case '\b':
		j.buf.WriteRune('b')
	case '\t':
		j.buf.WriteRune('t')
	case '\n':
		j.buf.WriteRune('n')
	case '\f':
		j.buf.WriteRune('f')
	case '\r':
		j.buf.WriteRune('r')
	case '"':
		j.buf.WriteRune('"')
	case '\\':
		j.buf.WriteRune('\\')
	}
}

func (j *JSONBuffer) pushState(s streamState) {
	// The separate logic here for the slice and the array allows us to avoid allocating a backing array for a
	// slice on the heap unless we run out of room in our local array. The original implementation of this tried
	// to be clever by initializing the slice to refer to stateStackArray[0:0]; that would work if we were only
	// using the slice within the same scope where it was initialized or a deeper scope, but since it stays
	// around after this method returns, the compiler would consider it suspicious enough to cause escaping.
	if j.stateStackSlice == nil {
		if j.stateStackArrayLen < len(j.stateStackArray) {
			j.stateStackArray[j.stateStackArrayLen] = j.state
			j.stateStackArrayLen++
		} else {
			j.stateStackSlice = make([]streamState, j.stateStackArrayLen, j.stateStackArrayLen*2)
			copy(j.stateStackSlice, j.stateStackArray[:])
			j.stateStackSlice = append(j.stateStackSlice, j.state)
		}
	} else {
		j.stateStackSlice = append(j.stateStackSlice, j.state)
	}
	j.state = s
}

func (j *JSONBuffer) popState() {
	if j.stateStackSlice != nil {
		n := len(j.stateStackSlice)
		j.state = j.stateStackSlice[n-1]
		j.stateStackSlice = j.stateStackSlice[0 : n-1]
	} else {
		j.state = j.stateStackArray[j.stateStackArrayLen-1]
		j.stateStackArrayLen--
	}
}

func (j *JSONBuffer) getStateStackCount() int {
	if j.stateStackSlice == nil {
		return j.stateStackArrayLen
	}
	return len(j.stateStackSlice)
}

func (j *JSONBuffer) fail(message string) {
	j.err = errors.New(message)
}
