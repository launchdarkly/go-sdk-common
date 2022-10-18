package ldvalue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionalBoolAsStringer(t *testing.T) {
	assert.Equal(t, "[none]", OptionalBool{}.String())
	assert.Equal(t, "false", NewOptionalBool(false).String())
	assert.Equal(t, "true", NewOptionalBool(true).String())
}

func TestOptionalIntAsStringer(t *testing.T) {
	assert.Equal(t, "[none]", OptionalInt{}.String())
	assert.Equal(t, "3", NewOptionalInt(3).String())
}

func TestOptionalStringAsStringer(t *testing.T) {
	assert.Equal(t, "[none]", OptionalString{}.String())
	assert.Equal(t, "[empty]", NewOptionalString("").String())
	assert.Equal(t, "x", NewOptionalString("x").String())
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

func TestOptionalIntTextMarshalling(t *testing.T) {
	b, e := NewOptionalInt(3).MarshalText()
	assert.NoError(t, e)
	assert.Equal(t, []byte("3"), b)

	b, e = OptionalInt{}.MarshalText()
	assert.NoError(t, e)
	assert.Len(t, b, 0)
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

func TestOptionalIntTextUnmarshalling(t *testing.T) {
	var o1 OptionalInt
	assert.NoError(t, o1.UnmarshalText([]byte("3")))
	assert.Equal(t, NewOptionalInt(3), o1)

	var o2 OptionalInt
	assert.NoError(t, o2.UnmarshalText([]byte("")))
	assert.Equal(t, OptionalInt{}, o2)

	var o3 OptionalInt
	assert.NoError(t, o3.UnmarshalText(nil))
	assert.Equal(t, OptionalInt{}, o3)

	var o4 OptionalInt
	assert.Error(t, o4.UnmarshalText([]byte("x")))
	assert.Equal(t, OptionalInt{}, o4)
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
