package ldvalue

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/launchdarkly/go-jsonstream/v3/jreader"
	"github.com/launchdarkly/go-jsonstream/v3/jwriter"

	"github.com/stretchr/testify/assert"
)

func TestEmptyOptionalBool(t *testing.T) {
	o := OptionalBool{}
	assert.False(t, o.IsDefined())
	assert.False(t, o.BoolValue())

	b, ok := o.Get()
	assert.False(t, b)
	assert.False(t, ok)

	assert.Equal(t, true, o.OrElse(true))
	assert.Nil(t, o.AsPointer())
	assert.Equal(t, Null(), o.AsValue())
	assert.True(t, o == o)
}

func TestOptionalBoolWithValue(t *testing.T) {
	for _, v := range []bool{false, true} {
		t.Run(fmt.Sprintf("%t", v), func(t *testing.T) {
			o := NewOptionalBool(v)
			assert.True(t, o.IsDefined())
			assert.Equal(t, v, o.BoolValue())

			b, ok := o.Get()
			assert.Equal(t, v, b)
			assert.True(t, ok)

			assert.Equal(t, v, o.OrElse(false))
			assert.Equal(t, v, o.OrElse(true))
			assert.NotNil(t, o.AsPointer())
			assert.Equal(t, v, *o.AsPointer())
			assert.Equal(t, Bool(v), o.AsValue())
			assert.True(t, o == o)
			assert.False(t, o == OptionalBool{})

		})
	}
}

func TestOptionalBoolFromNilPointer(t *testing.T) {
	o := NewOptionalBoolFromPointer(nil)
	assert.True(t, o == OptionalBool{})
}

func TestOptionalBoolFromNonNilPointer(t *testing.T) {
	v := true
	p := &v
	o := NewOptionalBoolFromPointer(p)
	assert.True(t, o == NewOptionalBool(true))

	assert.Equal(t, true, *o.AsPointer())
	assert.False(t, p == o.AsPointer()) // should not be the same pointer, just the same underlying value
}

func TestOptionalBoolAsStringer(t *testing.T) {
	assert.Equal(t, "[none]", OptionalBool{}.String())
	assert.Equal(t, "false", NewOptionalBool(false).String())
	assert.Equal(t, "true", NewOptionalBool(true).String())
}

func TestOptionalBoolJSONMarshalling(t *testing.T) {
	testWithMarshaler := func(t *testing.T, marshal func(OptionalBool) ([]byte, error)) {
		bytes, err := marshal(OptionalBool{})
		assert.NoError(t, err)
		assert.Equal(t, nullAsJSON, string(bytes))

		bytes, err = marshal(NewOptionalBool(true))
		assert.NoError(t, err)
		assert.Equal(t, `true`, string(bytes))

		bytes, err = marshal(NewOptionalBool(false))
		assert.NoError(t, err)
		assert.Equal(t, `false`, string(bytes))
	}

	t.Run("with json.Marshal", func(t *testing.T) {
		testWithMarshaler(t, func(o OptionalBool) ([]byte, error) {
			return json.Marshal(o)
		})

		swos := structWithOptionalBools{B1: NewOptionalBool(true), B2: NewOptionalBool(false)}
		bytes, err := json.Marshal(swos)
		assert.NoError(t, err)
		assert.Equal(t, `{"b1":true,"b2":false,"b3":null}`, string(bytes))
	})

	t.Run("with WriteToJSONWriter", func(t *testing.T) {
		testWithMarshaler(t, func(o OptionalBool) ([]byte, error) {
			w := jwriter.NewWriter()
			o.WriteToJSONWriter(&w)
			return w.Bytes(), w.Error()
		})
	})
}

func TestOptionalBoolJSONUnmarshalling(t *testing.T) {
	testWithUnmarshaler := func(t *testing.T, unmarshal func([]byte, *OptionalBool) error) {
		var o OptionalBool
		err := unmarshal([]byte(nullAsJSON), &o)
		assert.NoError(t, err)
		assert.False(t, o.IsDefined())

		err = unmarshal([]byte(`true`), &o)
		assert.NoError(t, err)
		assert.Equal(t, NewOptionalBool(true), o)

		err = unmarshal([]byte(`false`), &o)
		assert.NoError(t, err)
		assert.Equal(t, NewOptionalBool(false), o)

		err = unmarshal([]byte(`3`), &o)
		assert.Error(t, err)
		assert.IsType(t, &json.UnmarshalTypeError{}, err)

		err = unmarshal([]byte(`x`), &o)
		assert.Error(t, err)
		assert.IsType(t, &json.SyntaxError{}, err)
	}

	t.Run("with json.Unmarshal", func(t *testing.T) {
		testWithUnmarshaler(t, func(data []byte, o *OptionalBool) error {
			return json.Unmarshal(data, o)
		})

		var swos structWithOptionalBools
		err := json.Unmarshal([]byte(`{"b1":true,"b2":false,"b3":null}`), &swos)
		assert.NoError(t, err)
		assert.Equal(t, NewOptionalBool(true), swos.B1)
		assert.Equal(t, NewOptionalBool(false), swos.B2)
		assert.Equal(t, OptionalBool{}, swos.B3)
	})

	t.Run("with ReadFromJSONReader", func(t *testing.T) {
		testWithUnmarshaler(t, func(data []byte, o *OptionalBool) error {
			r := jreader.NewReader(data)
			o.ReadFromJSONReader(&r)
			return jreader.ToJSONError(r.Error(), o)
		})
	})
}

type structWithOptionalBools struct {
	B1 OptionalBool `json:"b1"`
	B2 OptionalBool `json:"b2"`
	B3 OptionalBool `json:"b3"`
}

func TestOptionalBoolTextMarshalling(t *testing.T) {
	b, e := NewOptionalBool(true).MarshalText()
	assert.NoError(t, e)
	assert.Equal(t, []byte("true"), b)

	b, e = NewOptionalBool(false).MarshalText()
	assert.NoError(t, e)
	assert.Equal(t, []byte("false"), b)

	b, e = OptionalBool{}.MarshalText()
	assert.NoError(t, e)
	assert.Len(t, b, 0)
}

func TestOptionalBoolTextUnmarshalling(t *testing.T) {
	var o1 OptionalBool
	assert.NoError(t, o1.UnmarshalText([]byte("true")))
	assert.Equal(t, NewOptionalBool(true), o1)

	var o2 OptionalBool
	assert.NoError(t, o2.UnmarshalText([]byte("false")))
	assert.Equal(t, NewOptionalBool(false), o2)

	var o3 OptionalBool
	assert.NoError(t, o3.UnmarshalText([]byte("")))
	assert.Equal(t, OptionalBool{}, o3)

	var o4 OptionalBool
	assert.NoError(t, o4.UnmarshalText(nil))
	assert.Equal(t, OptionalBool{}, o4)

	var o5 OptionalBool
	assert.Error(t, o5.UnmarshalText([]byte("x")))
	assert.Equal(t, OptionalBool{}, o5)
}
