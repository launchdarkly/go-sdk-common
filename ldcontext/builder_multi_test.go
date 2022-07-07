package ldcontext

import (
	"testing"

	"github.com/launchdarkly/go-sdk-common/v3/lderrors"
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

		assert.True(t, c0.IsDefined())
		assert.NoError(t, c0.Err())
		assert.Equal(t, Kind("multi"), c0.Kind())
		assert.Equal(t, "", c0.Key())

		assert.Equal(t, 2, c0.IndividualContextCount())

		assert.Equal(t, []Context{sub1, sub2}, c0.GetAllIndividualContexts(nil))
		// other accessors are tested in context_test.go
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
		assert.True(t, c0.IsDefined())
		assert.Equal(t, expectedErr, c0.Err())

		c1, err := builder.TryBuild()
		assert.True(t, c1.IsDefined())
		assert.Equal(t, expectedErr, c1.Err())
		assert.Equal(t, expectedErr, err)
	}

	t.Run("empty", func(t *testing.T) {
		verifyError(t, NewMultiBuilder(), lderrors.ErrContextKindMultiWithNoKinds{})
	})

	t.Run("nested multi", func(t *testing.T) {
		sub1 := NewWithKind("org", "my-key")
		sub2 := NewMulti(New("user-key"), NewWithKind("org", "other"))
		b0 := NewMultiBuilder().Add(sub1).Add(sub2)
		verifyError(t, b0, lderrors.ErrContextKindMultiWithinMulti{})

		b1 := NewMultiBuilder().Add(sub2)
		verifyError(t, b1, lderrors.ErrContextKindMultiWithinMulti{})
	})

	t.Run("duplicate kind", func(t *testing.T) {
		sub1 := NewWithKind("org", "my-org-key")
		sub2 := NewWithKind("user", "my-user-key")
		sub3 := NewWithKind("org", "other-org-key")
		b := NewMultiBuilder().Add(sub1).Add(sub2).Add(sub3)
		verifyError(t, b, lderrors.ErrContextKindMultiDuplicates{})
	})

	t.Run("error in individual contexts", func(t *testing.T) {
		sub1 := NewWithKind("kind1", "")
		sub2 := NewWithKind("kind2", "my-key")
		sub3 := NewWithKind("kind3!", "other-key")
		b := NewMultiBuilder().Add(sub1).Add(sub2).Add(sub3)
		c0 := b.Build()
		assert.Error(t, c0.Err())
		if assert.IsType(t, lderrors.ErrContextPerKindErrors{}, c0.Err()) {
			e := c0.Err().(lderrors.ErrContextPerKindErrors)
			assert.Len(t, e.Errors, 2)
			assert.Equal(t, lderrors.ErrContextKeyEmpty{}, e.Errors["kind1"])
			assert.Equal(t, lderrors.ErrContextKindInvalidChars{}, e.Errors["kind3!"])
		}
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
	assert.Equal(t, 2, multi1.IndividualContextCount())

	c3 := NewWithKind("thing", "stuff")
	b.Add(c3)

	multi2 := b.Build()
	assert.Equal(t, 3, multi2.IndividualContextCount())
	assert.Equal(t, 2, multi1.IndividualContextCount()) // unchanged
}
