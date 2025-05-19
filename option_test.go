package ldcommon

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionSome(t *testing.T) {
	opt := Some(42)
	assert.True(t, opt.IsSome())
	assert.False(t, opt.IsNone())
	assert.Equal(t, 42, opt.Unwrap())
}

func TestOptionNone(t *testing.T) {
	opt := None[int]()
	assert.False(t, opt.IsSome())
	assert.True(t, opt.IsNone())
	assert.Panics(t, func() { opt.Unwrap() })
}

func TestOptionUnwrapOr(t *testing.T) {
	someOpt := Some(42)
	noneOpt := None[int]()

	assert.Equal(t, 42, someOpt.UnwrapOr(100))
	assert.Equal(t, 100, noneOpt.UnwrapOr(100))
}

func TestOptionMap(t *testing.T) {
	someOpt := Some(42)
	noneOpt := None[int]()

	mappedSome := Map(someOpt, func(x int) string { return "Value: " + fmt.Sprint(x) })
	mappedNone := Map(noneOpt, func(x int) string { return "Value: " + fmt.Sprint(x) })

	assert.True(t, mappedSome.IsSome())
	assert.False(t, mappedNone.IsSome())
	assert.Equal(t, "Value: 42", mappedSome.Unwrap())
}

func TestOptionAsPtr(t *testing.T) {
	someOpt := Some(42)
	noneOpt := None[int]()

	somePtr := someOpt.AsPtr()
	nonePtr := noneOpt.AsPtr()

	assert.NotNil(t, somePtr)
	assert.Equal(t, 42, *somePtr)
	assert.Nil(t, nonePtr)
}

func TestOptionOrElse(t *testing.T) {
	someOpt := Some(42)
	noneOpt := None[int]()

	orElseSome := someOpt.OrElse(func() Option[int] { return Some(100) })
	orElseNone := noneOpt.OrElse(func() Option[int] { return Some(100) })

	assert.Equal(t, 42, orElseSome.Unwrap())
	assert.Equal(t, 100, orElseNone.Unwrap())
}
