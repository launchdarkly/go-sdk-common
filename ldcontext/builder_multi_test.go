package ldcontext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiBuilder(t *testing.T) {
	t.Run("single kind", func(t *testing.T) {
		sub1 := NewWithKind("org", "my-org-key")
		c := NewMultiBuilder().Add(sub1).Build()
		assert.Equal(t, sub1, c)
	})

	t.Run("multiple kinds", func(t *testing.T) {
		sub1 := NewWithKind("org", "my-org-key")
		sub2 := NewWithKind("user", "my-user-key")
		c0 := NewMultiBuilder().Add(sub1).Add(sub2).Build()

		assert.NoError(t, c0.Err())
		assert.Equal(t, Kind("multi"), c0.Kind())
		assert.Equal(t, "", c0.Key())

		assert.Equal(t, 2, c0.MultiKindCount())

		c1a, ok := c0.MultiKindByIndex(0)
		assert.True(t, ok)
		assert.Equal(t, sub1, c1a)

		c1b, ok := c0.MultiKindByName("org")
		assert.True(t, ok)
		assert.Equal(t, sub1, c1b)

		c2a, ok := c0.MultiKindByIndex(1)
		assert.True(t, ok)
		assert.Equal(t, sub2, c2a)

		c2b, ok := c0.MultiKindByName("user")
		assert.True(t, ok)
		assert.Equal(t, sub2, c2b)

		_, ok = c0.MultiKindByIndex(-1)
		assert.False(t, ok)

		_, ok = c0.MultiKindByIndex(2)
		assert.False(t, ok)

		_, ok = c0.MultiKindByName("notfound")
		assert.False(t, ok)
	})
}

func TestMultiBuilderFullyQualifiedKey(t *testing.T) {
	t.Run("multiple kinds", func(t *testing.T) {
		c := NewMultiBuilder().
			// The following ordering is deliberate because we want to verify that these items are
			// sorted by kind, not by key.
			Add(NewWithKind("kind-c", "key-1")).
			Add(NewWithKind("kind-a", "key-2")).
			Add(NewWithKind("kind-d", "key-3")).
			Add(NewWithKind("kind-b", "key-4")).
			Build()
		assert.Equal(t, "kind-a:key-2:kind-b:key-4:kind-c:key-1:kind-d:key-3", c.FullyQualifiedKey())
	})
}

func TestMultiBuilderErrors(t *testing.T) {
	verifyError := func(t *testing.T, builder *MultiBuilder, expectedErr error) {
		c0 := builder.Build()
		assert.Equal(t, expectedErr, c0.Err())

		c1, err := builder.TryBuild()
		assert.Equal(t, expectedErr, c1.Err())
		assert.Equal(t, expectedErr, err)
	}

	t.Run("empty", func(t *testing.T) {
		verifyError(t, NewMultiBuilder(), errContextKindMultiWithNoKinds)
	})

	t.Run("nested multi", func(t *testing.T) {
		sub1 := NewWithKind("org", "my-key")
		sub2 := NewMulti(New("user-key"), NewWithKind("org", "other"))
		b0 := NewMultiBuilder().Add(sub1).Add(sub2)
		verifyError(t, b0, errContextKindMultiWithinMulti)

		b1 := NewMultiBuilder().Add(sub2)
		verifyError(t, b1, errContextKindMultiWithinMulti)
	})

	t.Run("duplicate kind", func(t *testing.T) {
		sub1 := NewWithKind("org", "my-org-key")
		sub2 := NewWithKind("user", "my-user-key")
		sub3 := NewWithKind("org", "other-org-key")
		b := NewMultiBuilder().Add(sub1).Add(sub2).Add(sub3)
		verifyError(t, b, errContextKindMultiDuplicates)
	})

	t.Run("error in individual contexts", func(t *testing.T) {
		sub1 := NewWithKind("kind1", "")
		sub2 := NewWithKind("kind2", "my-key")
		sub3 := NewWithKind("kind3!", "other-key")
		b := NewMultiBuilder().Add(sub1).Add(sub2).Add(sub3)
		c0 := b.Build()
		assert.Error(t, c0.Err())
		assert.Regexp(t, "\\(kind1\\).*must not be empty, \\(kind3!\\).*disallowed characters", c0.Err().Error())
		c1, err := b.TryBuild()
		assert.Equal(t, c0.Err(), c1.Err())
		assert.Equal(t, c0.Err(), err)
	})
}

func TestMultiBuilderCopyOnWrite(t *testing.T) {
	c1 := NewWithKind("org", "my-org-key")
	c2 := NewWithKind("user", "my-user-key")

	b := NewMultiBuilder()
	b.Add(c1).Add(c2)

	multi1 := b.Build()
	assert.Equal(t, 2, multi1.MultiKindCount())

	c3 := NewWithKind("thing", "stuff")
	b.Add(c3)

	multi2 := b.Build()
	assert.Equal(t, 3, multi2.MultiKindCount())
	assert.Equal(t, 2, multi1.MultiKindCount()) // unchanged
}
