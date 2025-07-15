package ldcontext

import (
	"testing"

	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/lderrors"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

	"github.com/stretchr/testify/assert"
)

func TestContextBuilderSingleKind(t *testing.T) {
	t.Run("basic properties", func(t *testing.T) {
		c := NewContextBuilder().Kind("user", "my-key").Build()
		assert.True(t, c.IsDefined())
		assert.NoError(t, c.Err())
		assert.Equal(t, DefaultKind, c.Kind())
		assert.Equal(t, "my-key", c.Key())
		assert.Equal(t, ldvalue.OptionalString{}, c.Name())
		assert.False(t, c.Anonymous())
	})

	t.Run("custom kind", func(t *testing.T) {
		c := NewContextBuilder().Kind("org", "my-org-key").Build()
		assert.Equal(t, Kind("org"), c.Kind())
		assert.Equal(t, "my-org-key", c.Key())
	})

	t.Run("with name", func(t *testing.T) {
		c := NewContextBuilder().Kind("user", "my-key").Name("my-name").Build()
		assert.Equal(t, ldvalue.NewOptionalString("my-name"), c.Name())
	})

	t.Run("with anonymous", func(t *testing.T) {
		c := NewContextBuilder().Kind("user", "my-key").Anonymous(true).Build()
		assert.True(t, c.Anonymous())
	})

	t.Run("with custom attributes", func(t *testing.T) {
		c := NewContextBuilder().Kind("user", "my-key").
			SetString("email", "test@example.com").
			SetBool("active", true).
			Build()

		assert.Equal(t, ldvalue.String("test@example.com"), c.GetValue("email"))
		assert.Equal(t, ldvalue.Bool(true), c.GetValue("active"))
	})

	t.Run("with private attributes", func(t *testing.T) {
		c := NewContextBuilder().Kind("user", "my-key").
			Name("my-name").
			Private("name", "email").
			Build()

		assert.Equal(t, 2, c.PrivateAttributeCount())
	})
}

func TestContextBuilderMultiKind(t *testing.T) {
	t.Run("two kinds", func(t *testing.T) {
		c := NewContextBuilder().
			Kind("user", "user-key").Name("User Name").
			Kind("org", "org-key").Name("Org Name").
			Build()

		assert.True(t, c.Multiple())
		assert.Equal(t, Kind("multi"), c.Kind())
		assert.Equal(t, 2, c.IndividualContextCount())

		userCtx := c.IndividualContextByKind("user")
		assert.Equal(t, "user-key", userCtx.Key())
		assert.Equal(t, ldvalue.NewOptionalString("User Name"), userCtx.Name())

		orgCtx := c.IndividualContextByKind("org")
		assert.Equal(t, "org-key", orgCtx.Key())
		assert.Equal(t, ldvalue.NewOptionalString("Org Name"), orgCtx.Name())
	})

	t.Run("three kinds", func(t *testing.T) {
		c := NewContextBuilder().
			Kind("user", "user-key").
			Kind("org", "org-key").
			Kind("device", "device-key").
			Build()

		assert.Equal(t, 3, c.IndividualContextCount())
		assert.NotEqual(t, Context{}, c.IndividualContextByKind("user"))
		assert.NotEqual(t, Context{}, c.IndividualContextByKind("org"))
		assert.NotEqual(t, Context{}, c.IndividualContextByKind("device"))
	})

	t.Run("updating existing kind", func(t *testing.T) {
		c := NewContextBuilder().
			Kind("user", "user-key-1").Name("Name 1").
			Kind("org", "org-key").
			Kind("user", "user-key-2").Name("Name 2").
			Build()

		assert.Equal(t, 2, c.IndividualContextCount())
		userCtx := c.IndividualContextByKind("user")
		assert.Equal(t, "user-key-2", userCtx.Key())
		assert.Equal(t, ldvalue.NewOptionalString("Name 2"), userCtx.Name())
	})
}

func TestContextBuilderKindValidation(t *testing.T) {
	for _, p := range makeInvalidKindTestParams() {
		t.Run(p.kind, func(t *testing.T) {
			c := NewContextBuilder().Kind(Kind(p.kind), "my-key").Build()
			assert.True(t, c.IsDefined())
			assert.Equal(t, p.err, c.Err())
		})
	}
}

func TestContextBuilderKeyValidation(t *testing.T) {
	c := NewContextBuilder().Kind("user", "").Build()
	assert.True(t, c.IsDefined())
	assert.Equal(t, lderrors.ErrContextKeyEmpty{}, c.Err())
}

func TestContextBuilderFullyQualifiedKey(t *testing.T) {
	t.Run("single kind user", func(t *testing.T) {
		c := NewContextBuilder().Kind("user", "my-user-key").Build()
		assert.Equal(t, "my-user-key", c.FullyQualifiedKey())
	})

	t.Run("single kind non-user", func(t *testing.T) {
		c := NewContextBuilder().Kind("org", "my-org-key").Build()
		assert.Equal(t, "org:my-org-key", c.FullyQualifiedKey())
	})

	t.Run("multi kind", func(t *testing.T) {
		c := NewContextBuilder().
			Kind("kind-c", "key-1").
			Kind("kind-a", "key-2").
			Kind("kind-d", "key-3").
			Kind("kind-b", "key-4").
			Build()
		assert.Equal(t, "kind-a:key-2:kind-b:key-4:kind-c:key-1:kind-d:key-3", c.FullyQualifiedKey())
	})

	t.Run("keys are escaped", func(t *testing.T) {
		c := NewContextBuilder().
			Kind("kind-a", "key-1").
			Kind("kind-b", "key:2").
			Build()
		assert.Equal(t, "kind-a:key-1:kind-b:key%3A2", c.FullyQualifiedKey())
	})
}

func TestKindBuilderSetters(t *testing.T) {
	t.Run("Name", func(t *testing.T) {
		c := NewContextBuilder().Kind("user", "my-key").Name("my-name").Build()
		assert.Equal(t, ldvalue.NewOptionalString("my-name"), c.Name())
	})

	t.Run("Anonymous", func(t *testing.T) {
		c1 := NewContextBuilder().Kind("user", "my-key").Anonymous(false).Build()
		assert.False(t, c1.Anonymous())

		c2 := NewContextBuilder().Kind("user", "my-key").Anonymous(true).Build()
		assert.True(t, c2.Anonymous())
	})

	t.Run("SetValue", func(t *testing.T) {
		c := NewContextBuilder().Kind("user", "my-key").
			SetValue("my-attr", ldvalue.Bool(true)).
			SetValue("other-attr", ldvalue.String("other-value")).
			Build()

		assert.Equal(t, ldvalue.Bool(true), c.GetValue("my-attr"))
		assert.Equal(t, ldvalue.String("other-value"), c.GetValue("other-attr"))
	})

	t.Run("typed setters", func(t *testing.T) {
		c := NewContextBuilder().Kind("user", "my-key").
			SetBool("bool-attr", true).
			SetInt("int-attr", 100).
			SetFloat64("float-attr", 1.5).
			SetString("string-attr", "x").
			Build()

		assert.Equal(t, ldvalue.Bool(true), c.GetValue("bool-attr"))
		assert.Equal(t, ldvalue.Int(100), c.GetValue("int-attr"))
		assert.Equal(t, ldvalue.Float64(1.5), c.GetValue("float-attr"))
		assert.Equal(t, ldvalue.String("x"), c.GetValue("string-attr"))
	})

	t.Run("Private", func(t *testing.T) {
		c := NewContextBuilder().Kind("user", "my-key").
			Name("my-name").
			Private("name", "email").
			Build()

		assert.Equal(t, 2, c.PrivateAttributeCount())
		ref1, ok1 := c.PrivateAttributeByIndex(0)
		ref2, ok2 := c.PrivateAttributeByIndex(1)
		assert.True(t, ok1)
		assert.True(t, ok2)

		refs := []ldattr.Ref{ref1, ref2}
		assert.Contains(t, refs, ldattr.NewRef("name"))
		assert.Contains(t, refs, ldattr.NewRef("email"))
	})
}

func TestContextBuilderErrors(t *testing.T) {
	t.Run("empty builder", func(t *testing.T) {
		c := NewContextBuilder().Build()
		assert.True(t, c.IsDefined())
		assert.Equal(t, lderrors.ErrContextKindMultiWithNoKinds{}, c.Err())
	})

	t.Run("duplicate kind error not applicable", func(t *testing.T) {
		// With ContextBuilder, setting the same kind twice just updates it
		c := NewContextBuilder().
			Kind("org", "key1").
			Kind("org", "key2").
			Build()
		assert.NoError(t, c.Err())
		assert.Equal(t, "key2", c.Key())
	})

	t.Run("error in individual contexts", func(t *testing.T) {
		c := NewContextBuilder().
			Kind("kind1", "").
			Kind("kind2", "my-key").
			Kind("kind3!", "other-key").
			Build()

		assert.Error(t, c.Err())
		if assert.IsType(t, lderrors.ErrContextPerKindErrors{}, c.Err()) {
			e := c.Err().(lderrors.ErrContextPerKindErrors)
			assert.Len(t, e.Errors, 2)
			assert.Equal(t, lderrors.ErrContextKeyEmpty{}, e.Errors["kind1"])
			assert.Equal(t, lderrors.ErrContextKindInvalidChars{}, e.Errors["kind3!"])
		}
	})
}

func TestContextBuilderChaining(t *testing.T) {
	t.Run("method chaining works", func(t *testing.T) {
		c := NewContextBuilder().
			Kind("user", "user-key").Name("User Name").Anonymous(true).
			Kind("org", "org-key").SetString("type", "enterprise").
			Build()

		assert.Equal(t, 2, c.IndividualContextCount())

		userCtx := c.IndividualContextByKind("user")
		assert.Equal(t, "user-key", userCtx.Key())
		assert.Equal(t, ldvalue.NewOptionalString("User Name"), userCtx.Name())
		assert.True(t, userCtx.Anonymous())

		orgCtx := c.IndividualContextByKind("org")
		assert.Equal(t, "org-key", orgCtx.Key())
		assert.Equal(t, ldvalue.String("enterprise"), orgCtx.GetValue("type"))
	})
}
