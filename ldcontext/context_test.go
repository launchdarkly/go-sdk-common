package ldcontext

import (
	"sort"
	"testing"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"

	m "github.com/launchdarkly/go-test-helpers/v2/matchers"

	"github.com/stretchr/testify/assert"
)

// Note, matchers.JSONEqual is preferred in these tests when checking ldvalue.Value values, rather
// than assert.Equal or assert.JSONEq, because its failure output is easier to read.

func TestUninitializedContextIsInvalid(t *testing.T) {
	var c Context
	assert.Equal(t, errContextUninitialized, c.Err())
}

func TestGetOptionalAttributeNames(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		c := New("my-key")
		an := c.GetOptionalAttributeNames(nil)
		assert.Len(t, an, 0)
	})

	t.Run("meta not included", func(t *testing.T) {
		c := NewBuilder("my-key").Secondary("x").Transient(true).Build()
		an := c.GetOptionalAttributeNames(nil)
		assert.Len(t, an, 0)
	})

	t.Run("name", func(t *testing.T) {
		c := NewBuilder("my-key").Name("x").Build()
		an := c.GetOptionalAttributeNames(nil)
		assert.Equal(t, []string{"name"}, an)
	})

	t.Run("others", func(t *testing.T) {
		c := NewBuilder("my-key").SetString("email", "x").SetBool("happy", true).Build()
		an := c.GetOptionalAttributeNames(nil)
		sort.Strings(an)
		assert.Equal(t, []string{"email", "happy"}, an)
	})

	t.Run("none for multi-kind context", func(t *testing.T) {
		c := NewMulti(NewWithKind("kind1", "key1"))
		an := c.GetOptionalAttributeNames(nil)
		assert.Len(t, an, 0)
	})
}

func TestGetValueSpecialTopLevelAttributes(t *testing.T) {
	t.Run("kind", func(t *testing.T) {
		t.Run("single-kind", func(t *testing.T) {
			c := NewWithKind("org", "my-key")
			expectAttributeFound(t, ldvalue.String("org"), c, "kind")
		})

		t.Run("multi-kind", func(t *testing.T) {
			c := NewMultiBuilder().Add(New("my-key")).Build()
			expectAttributeFound(t, ldvalue.String("multi"), c, "kind")
		})
	})

	t.Run("key", func(t *testing.T) {
		t.Run("single-kind", func(t *testing.T) {
			c := New("my-key")
			expectAttributeFound(t, ldvalue.String("my-key"), c, "key")
		})

		t.Run("multi-kind", func(t *testing.T) {
			c := NewMultiBuilder().Add(New("my-key")).Build()
			expectAttributeNotFound(t, c, "key")
		})
	})

	t.Run("name", func(t *testing.T) {
		t.Run("single-kind, defined", func(t *testing.T) {
			c := makeBasicBuilder().Name("my-name").Build()
			expectAttributeFound(t, ldvalue.String("my-name"), c, "name")
		})

		t.Run("single-kind, undefined", func(t *testing.T) {
			c := makeBasicBuilder().Build()
			expectAttributeNotFound(t, c, "name")
		})

		t.Run("multi-kind", func(t *testing.T) {
			c := NewMultiBuilder().Add(makeBasicBuilder().Name("my-name").Build()).Build()
			expectAttributeNotFound(t, c, "name")
		})
	})

	t.Run("secondary", func(t *testing.T) {
		t.Run("single-kind, defined", func(t *testing.T) {
			c := makeBasicBuilder().Secondary("my-value").Build()
			expectAttributeFound(t, ldvalue.String("my-value"), c, "secondary")
		})

		t.Run("single-kind, undefined", func(t *testing.T) {
			c := makeBasicBuilder().Build()
			expectAttributeNotFound(t, c, "secondary")
		})

		t.Run("multi-kind", func(t *testing.T) {
			c := NewMultiBuilder().Add(makeBasicBuilder().Secondary("my-value").Build()).Build()
			expectAttributeNotFound(t, c, "secondary")
		})
	})

	t.Run("transient", func(t *testing.T) {
		t.Run("single-kind, defined, true", func(t *testing.T) {
			c := makeBasicBuilder().Transient(true).Build()
			expectAttributeFound(t, ldvalue.Bool(true), c, "transient")
		})

		t.Run("single-kind, defined, false", func(t *testing.T) {
			c := makeBasicBuilder().Transient(false).Build()
			expectAttributeFound(t, ldvalue.Bool(false), c, "transient")
		})

		t.Run("single-kind, undefined", func(t *testing.T) {
			c := makeBasicBuilder().Build()
			expectAttributeFound(t, ldvalue.Bool(false), c, "transient")
		})

		t.Run("multi-kind", func(t *testing.T) {
			c := NewMultiBuilder().Add(makeBasicBuilder().Transient(true).Build()).Build()
			expectAttributeNotFound(t, c, "transient")
		})
	})
}

func TestGetValueCustomAttributeSingleKind(t *testing.T) {
	t.Run("simple attribute name", func(t *testing.T) {
		expected := ldvalue.String("abc")
		c := makeBasicBuilder().SetValue("my-attr", expected).Build()
		expectAttributeFound(t, expected, c, "my-attr")
	})

	t.Run("simple attribute name not found", func(t *testing.T) {
		c := makeBasicBuilder().Build()
		expectAttributeNotFound(t, c, "my-attr")
	})
}

func expectAttributeFound(t *testing.T, expected ldvalue.Value, c Context, attrName string) {
	value, ok := c.GetValue(attrName)
	assert.True(t, ok, "attribute %q should have been found, but was not", attrName)
	m.In(t).Assert(value, m.JSONEqual(expected))
}

func expectAttributeNotFound(t *testing.T, c Context, attrName string) {
	value, ok := c.GetValue(attrName)
	assert.False(t, ok, "attribute %q should not have been found, but was", attrName)
	m.In(t).Assert(value, m.JSONEqual(nil))
}
