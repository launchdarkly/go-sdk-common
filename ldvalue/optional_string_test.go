package ldvalue

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyOptionalString(t *testing.T) {
	o := OptionalString{}
	assert.False(t, o.IsDefined())
	assert.Equal(t, "", o.StringValue())
	assert.Equal(t, "no", o.OrElse("no"))
	assert.Nil(t, o.AsPointer())
	assert.Equal(t, Null(), o.AsValue())
	assert.True(t, o == o)
}

func TestOptionalStringWithValue(t *testing.T) {
	o := NewOptionalString("value")
	assert.True(t, o.IsDefined())
	assert.Equal(t, "value", o.StringValue())
	assert.Equal(t, "value", o.OrElse("no"))
	assert.NotNil(t, o.AsPointer())
	assert.Equal(t, "value", *o.AsPointer())
	assert.Equal(t, String("value"), o.AsValue())
	assert.True(t, o == o)
	assert.False(t, o == OptionalString{})
}

func TestOptionalStringFromNilPointer(t *testing.T) {
	o := NewOptionalStringFromPointer(nil)
	assert.False(t, o.IsDefined())
	assert.Equal(t, "", o.StringValue())
	assert.Equal(t, "no", o.OrElse("no"))
	assert.Nil(t, o.AsPointer())
	assert.Equal(t, Null(), o.AsValue())
	assert.True(t, o == o)
	assert.True(t, o == OptionalString{})
}

func TestOptionalStringFromNonNilPointer(t *testing.T) {
	v := "value"
	p := &v
	o := NewOptionalStringFromPointer(p)
	assert.True(t, o.IsDefined())
	assert.Equal(t, "value", o.StringValue())
	assert.NotNil(t, o.AsPointer())
	assert.Equal(t, "value", *o.AsPointer())
	assert.False(t, p == o.AsPointer()) // should not be the same pointer, just the same underlying string
	assert.Equal(t, String("value"), o.AsValue())
	assert.True(t, o == o)
	assert.True(t, o == NewOptionalString("value"))
}

func TestOptionalStringAsStringer(t *testing.T) {
	assert.Equal(t, "[none]", OptionalString{}.String())
	assert.Equal(t, "[empty]", NewOptionalString("").String())
	assert.Equal(t, "x", NewOptionalString("x").String())
}

func TestOptionalStringMarshalling(t *testing.T) {
	bytes, err := json.Marshal(OptionalString{})
	assert.NoError(t, err)
	assert.Equal(t, "null", string(bytes))

	bytes, err = json.Marshal(NewOptionalString(`a "good" string`))
	assert.NoError(t, err)
	assert.Equal(t, `"a \"good\" string"`, string(bytes))

	swos := structWithOptionalStrings{S1: NewOptionalString("yes")}
	bytes, err = json.Marshal(swos)
	assert.NoError(t, err)
	assert.Equal(t, `{"s1":"yes","s2":null,"s3":null}`, string(bytes))
}

func TestOptionalStringUnmarshalling(t *testing.T) {
	var o OptionalString
	err := json.Unmarshal([]byte("null"), &o)
	assert.NoError(t, err)
	assert.False(t, o.IsDefined())

	err = json.Unmarshal([]byte(`"a \"good\" string"`), &o)
	assert.NoError(t, err)
	assert.True(t, o.IsDefined())
	assert.Equal(t, `a "good" string`, o.StringValue())

	err = json.Unmarshal([]byte("3"), &o)
	assert.Error(t, err)
	assert.False(t, o.IsDefined())

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
