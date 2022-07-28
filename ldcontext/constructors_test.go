package ldcontext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c := New("my-key")
	assert.Equal(t, NewBuilder("my-key").Build(), c)
	// More detailed tests of the default state of the Context are in the tests for Builder. Here we have
	// just verified that the constructor gives us the same result as Builder.
}

func TestNewErrors(t *testing.T) {
	c := New("")

	assert.Error(t, c.Err())
}

func TestNewWithKind(t *testing.T) {
	c0 := NewWithKind("org", "my-key")
	assert.Equal(t, NewBuilder("my-key").Kind("org").Build(), c0)
	// More detailed tests of the default state of the Context are in the tests for Builder. Here we have
	// just verified that the constructor gives us the same result as Builder.

	c1 := NewWithKind("", "my-key")
	assert.Equal(t, NewBuilder("my-key").Kind(DefaultKind).Build(), c1)
}

func TestNewWithKindErrors(t *testing.T) {
	for _, p := range makeInvalidKindTestParams() {
		t.Run(p.kind, func(t *testing.T) {
			c := NewWithKind(Kind(p.kind), "my-key")
			assert.Equal(t, p.err, c.Err())
		})
	}
}

func TestNewMulti(t *testing.T) {
	sub1 := NewWithKind("kind1", "key1")
	sub2 := NewWithKind("kind2", "key2")
	sub3 := NewWithKind("kind3", "key3")
	sub4 := NewWithKind("kind3", "key3")

	assert.Equal(t, NewMultiBuilder().Add(sub1).Add(sub2).Build(), NewMulti(sub1, sub2))
	assert.Equal(t, NewMultiBuilder().Add(sub1).Add(sub2).Add(sub3).Add(sub4).Build(),
		NewMulti(sub1, NewMulti(sub2, sub3), sub4))
}
