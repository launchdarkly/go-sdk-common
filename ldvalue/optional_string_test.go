package ldvalue

import (
	"testing"

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
