package ldvalue

import (
	"encoding/json"
	"testing"

	"github.com/launchdarkly/go-jsonstream/jwriter"
	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream"

	"github.com/stretchr/testify/assert"
)

func TestEmptyOptionalString(t *testing.T) {
	o := OptionalString{}
	assert.False(t, o.IsDefined())
	assert.Equal(t, "", o.StringValue())

	s, ok := o.Get()
	assert.Equal(t, "", s)
	assert.False(t, ok)

	assert.Equal(t, "no", o.OrElse("no"))
	assert.Nil(t, o.AsPointer())
	assert.Equal(t, Null(), o.AsValue())
	assert.True(t, o == o)
}

func TestOptionalStringWithValue(t *testing.T) {
	o := NewOptionalString("value")
	assert.True(t, o.IsDefined())
	assert.Equal(t, "value", o.StringValue())

	s, ok := o.Get()
	assert.Equal(t, "value", s)
	assert.True(t, ok)

	assert.Equal(t, "value", o.OrElse("no"))
	assert.NotNil(t, o.AsPointer())
	assert.Equal(t, "value", *o.AsPointer())
	assert.Equal(t, String("value"), o.AsValue())
	assert.True(t, o == o)
	assert.False(t, o == OptionalString{})
}

func TestOptionalStringFromNilPointer(t *testing.T) {
	o := NewOptionalStringFromPointer(nil)
	assert.True(t, o == OptionalString{})
}

func TestOptionalStringFromNonNilPointer(t *testing.T) {
	v := "value"
	p := &v
	o := NewOptionalStringFromPointer(p)
	assert.True(t, o == NewOptionalString("value"))

	assert.Equal(t, "value", *o.AsPointer())
	assert.False(t, p == o.AsPointer()) // should not be the same pointer, just the same underlying string
}

func TestOptionalStringOnlyIfNonEmptyString(t *testing.T) {
	assert.Equal(t, OptionalString{}, OptionalString{}.OnlyIfNonEmptyString())
	assert.Equal(t, OptionalString{}, NewOptionalString("").OnlyIfNonEmptyString())
	assert.Equal(t, NewOptionalString("x"), NewOptionalString("x").OnlyIfNonEmptyString())
}

func TestOptionalStringAsStringer(t *testing.T) {
	assert.Equal(t, "[none]", OptionalString{}.String())
	assert.Equal(t, "[empty]", NewOptionalString("").String())
	assert.Equal(t, "x", NewOptionalString("x").String())
}

func TestOptionalStringJSONMarshalling(t *testing.T) {
	testWithMarshaler := func(t *testing.T, marshal func(OptionalString) ([]byte, error)) {
		bytes, err := marshal(OptionalString{})
		assert.NoError(t, err)
		assert.Equal(t, nullAsJSON, string(bytes))

		bytes, err = marshal(NewOptionalString(`a "good" string`))
		assert.NoError(t, err)
		assert.Equal(t, `"a \"good\" string"`, string(bytes))
	}

	t.Run("with json.Marshal", func(t *testing.T) {
		testWithMarshaler(t, func(o OptionalString) ([]byte, error) {
			return json.Marshal(o)
		})

		swos := structWithOptionalStrings{S1: NewOptionalString("yes")}
		bytes, err := json.Marshal(swos)
		assert.NoError(t, err)
		assert.Equal(t, `{"s1":"yes","s2":null,"s3":null}`, string(bytes))
	})

	t.Run("with WriteToJSONWriter", func(t *testing.T) {
		testWithMarshaler(t, func(o OptionalString) ([]byte, error) {
			w := jwriter.NewWriter()
			o.WriteToJSONWriter(&w)
			return w.Bytes(), w.Error()
		})
	})

	t.Run("with WriteToJSONBuffer", func(t *testing.T) {
		testWithMarshaler(t, func(o OptionalString) ([]byte, error) {
			var b jsonstream.JSONBuffer
			o.WriteToJSONBuffer(&b)
			return b.Get()
		})
	})
}

func TestOptionalStringJSONUnmarshalling(t *testing.T) {
	var o OptionalString
	err := json.Unmarshal([]byte(nullAsJSON), &o)
	assert.NoError(t, err)
	assert.False(t, o.IsDefined())

	err = json.Unmarshal([]byte(`"a \"good\" string"`), &o)
	assert.NoError(t, err)
	assert.True(t, o.IsDefined())
	assert.Equal(t, `a "good" string`, o.StringValue())

	err = json.Unmarshal([]byte("3"), &o)
	assert.Error(t, err)
	assert.IsType(t, &json.UnmarshalTypeError{}, err)

	var swos structWithOptionalStrings
	err = json.Unmarshal([]byte(`{"s1":"yes","s3":null}`), &swos)
	assert.NoError(t, err)
	assert.Equal(t, NewOptionalString("yes"), swos.S1)
	assert.Equal(t, OptionalString{}, swos.S2)
	assert.Equal(t, OptionalString{}, swos.S3)
}

type structWithOptionalStrings struct {
	S1 OptionalString `json:"s1"`
	S2 OptionalString `json:"s2"`
	S3 OptionalString `json:"s3"`
}

func TestOptionalStringTextMarshalling(t *testing.T) {
	b, e := NewOptionalString("x").MarshalText()
	assert.NoError(t, e)
	assert.Equal(t, []byte("x"), b)

	b, e = NewOptionalString("").MarshalText()
	assert.NoError(t, e)
	assert.Equal(t, []byte{}, b)

	b, e = OptionalString{}.MarshalText()
	assert.NoError(t, e)
	assert.Nil(t, b)
}

func TestOptionalStringTextUnmarshalling(t *testing.T) {
	var o1 OptionalString
	assert.NoError(t, o1.UnmarshalText([]byte("x")))
	assert.Equal(t, NewOptionalString("x"), o1)

	var o2 OptionalString
	assert.NoError(t, o2.UnmarshalText([]byte("")))
	assert.Equal(t, NewOptionalString(""), o2)

	var o3 OptionalString
	assert.NoError(t, o3.UnmarshalText(nil))
	assert.Equal(t, OptionalString{}, o3)
}
