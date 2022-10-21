package ldvalue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyOptionalInt(t *testing.T) {
	o := OptionalInt{}
	assert.False(t, o.IsDefined())
	assert.Equal(t, 0, o.IntValue())

	n, ok := o.Get()
	assert.Equal(t, 0, n)
	assert.False(t, ok)

	assert.Equal(t, 2, o.OrElse(2))
	assert.Nil(t, o.AsPointer())
	assert.Equal(t, Null(), o.AsValue())
	assert.True(t, o == o)
}

func TestOptionalIntWithValue(t *testing.T) {
	o := NewOptionalInt(3)
	assert.True(t, o.IsDefined())
	assert.Equal(t, 3, o.IntValue())

	n, ok := o.Get()
	assert.Equal(t, 3, n)
	assert.True(t, ok)

	assert.Equal(t, 3, o.OrElse(2))
	assert.NotNil(t, o.AsPointer())
	assert.Equal(t, 3, *o.AsPointer())
	assert.Equal(t, Int(3), o.AsValue())
	assert.True(t, o == o)
	assert.False(t, o == OptionalInt{})
}

func TestOptionalIntFromNilPointer(t *testing.T) {
	o := NewOptionalIntFromPointer(nil)
	assert.True(t, o == OptionalInt{})
}

func TestOptionalIntFromNonNilPointer(t *testing.T) {
	v := 3
	p := &v
	o := NewOptionalIntFromPointer(p)
	assert.True(t, o == NewOptionalInt(3))

	assert.Equal(t, 3, *o.AsPointer())
	assert.False(t, p == o.AsPointer()) // should not be the same pointer, just the same underlying value
}
