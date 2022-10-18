package ldvalue

import (
	"fmt"
	"testing"

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
