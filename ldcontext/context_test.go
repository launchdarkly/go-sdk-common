package ldcontext

import (
	"encoding/json"
	"sort"
	"testing"

	"gopkg.in/launchdarkly/go-sdk-common.v3/ldattr"
	"gopkg.in/launchdarkly/go-sdk-common.v3/ldvalue"

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

func TestGetValue(t *testing.T) {
	t.Run("equivalent to GetValueForRef for simple attribute name", func(t *testing.T) {
		c := NewBuilder("my-key").Kind("org").Name("x").SetString("my-attr", "y").SetString("/starts-with-slash", "z").Build()
		expectAttributeFoundForName(t, ldvalue.String("org"), c, "kind")
		expectAttributeFoundForName(t, ldvalue.String("my-key"), c, "key")
		expectAttributeFoundForName(t, ldvalue.String("x"), c, "name")
		expectAttributeFoundForName(t, ldvalue.String("y"), c, "my-attr")
		expectAttributeFoundForName(t, ldvalue.String("z"), c, "/starts-with-slash")
		expectAttributeNotFoundForName(t, c, "/kind")
		expectAttributeNotFoundForName(t, c, "/key")
		expectAttributeNotFoundForName(t, c, "/name")
		expectAttributeNotFoundForName(t, c, "/my-attr")
		expectAttributeNotFoundForName(t, c, "other")

		expectAttributeNotFoundForName(t, c, "")
		expectAttributeNotFoundForName(t, c, "/")

		mc := NewMulti(c)
		expectAttributeFoundForName(t, ldvalue.String("multi"), mc, "kind")
		expectAttributeNotFoundForName(t, mc, "/kind")
		expectAttributeNotFoundForName(t, mc, "key")
	})

	t.Run("does not allow querying of subpath/element", func(t *testing.T) {
		objValue := ldvalue.ObjectBuild().Set("a", ldvalue.Int(1)).Build()
		arrayValue := ldvalue.ArrayOf(ldvalue.Int(1))
		c := makeBasicBuilder().SetValue("obj-attr", objValue).SetValue("array-attr", arrayValue).Build()
		expectAttributeFoundForName(t, objValue, c, "obj-attr")
		expectAttributeFoundForName(t, arrayValue, c, "array-attr")
		expectAttributeNotFoundForName(t, c, "/obj-attr/a")
		expectAttributeNotFoundForName(t, c, "/array-attr/0")
	})
}
func TestGetValueForRefSpecialTopLevelAttributes(t *testing.T) {
	t.Run("kind", func(t *testing.T) {
		t.Run("single-kind", func(t *testing.T) {
			c := NewWithKind("org", "my-key")
			expectAttributeFoundForRef(t, ldvalue.String("org"), c, "kind")
		})

		t.Run("multi-kind", func(t *testing.T) {
			c := NewMultiBuilder().Add(New("my-key")).Build()
			expectAttributeFoundForRef(t, ldvalue.String("multi"), c, "kind")
		})
	})

	t.Run("key", func(t *testing.T) {
		t.Run("single-kind", func(t *testing.T) {
			c := New("my-key")
			expectAttributeFoundForRef(t, ldvalue.String("my-key"), c, "key")
		})

		t.Run("multi-kind", func(t *testing.T) {
			c := NewMultiBuilder().Add(New("my-key")).Build()
			expectAttributeNotFoundForRef(t, c, "key")
		})
	})

	t.Run("name", func(t *testing.T) {
		t.Run("single-kind, defined", func(t *testing.T) {
			c := makeBasicBuilder().Name("my-name").Build()
			expectAttributeFoundForRef(t, ldvalue.String("my-name"), c, "name")
		})

		t.Run("single-kind, undefined", func(t *testing.T) {
			c := makeBasicBuilder().Build()
			expectAttributeNotFoundForRef(t, c, "name")
		})

		t.Run("multi-kind", func(t *testing.T) {
			c := NewMultiBuilder().Add(makeBasicBuilder().Name("my-name").Build()).Build()
			expectAttributeNotFoundForRef(t, c, "name")
		})
	})

	t.Run("transient", func(t *testing.T) {
		t.Run("single-kind, defined, true", func(t *testing.T) {
			c := makeBasicBuilder().Transient(true).Build()
			expectAttributeFoundForRef(t, ldvalue.Bool(true), c, "transient")
		})

		t.Run("single-kind, defined, false", func(t *testing.T) {
			c := makeBasicBuilder().Transient(false).Build()
			expectAttributeFoundForRef(t, ldvalue.Bool(false), c, "transient")
		})

		t.Run("single-kind, undefined", func(t *testing.T) {
			c := makeBasicBuilder().Build()
			expectAttributeFoundForRef(t, ldvalue.Bool(false), c, "transient")
		})

		t.Run("multi-kind", func(t *testing.T) {
			c := NewMultiBuilder().Add(makeBasicBuilder().Transient(true).Build()).Build()
			expectAttributeNotFoundForRef(t, c, "transient")
		})
	})
}

func TestGetValueForRefCannotGetMetaProperties(t *testing.T) {
	t.Run("privateAttributes", func(t *testing.T) {
		t.Run("single-kind, defined", func(t *testing.T) {
			c := makeBasicBuilder().Private("attr").Build()
			expectAttributeNotFoundForRef(t, c, "privateAttributes")
		})

		t.Run("single-kind, undefined", func(t *testing.T) {
			c := makeBasicBuilder().Build()
			expectAttributeNotFoundForRef(t, c, "privateAttributes")
		})

		t.Run("multi-kind", func(t *testing.T) {
			c := NewMultiBuilder().Add(makeBasicBuilder().Private("attr").Build()).Build()
			expectAttributeNotFoundForRef(t, c, "privateAttributes")
		})
	})

	t.Run("secondary", func(t *testing.T) {
		t.Run("single-kind, defined", func(t *testing.T) {
			c := makeBasicBuilder().Secondary("my-value").Build()
			expectAttributeNotFoundForRef(t, c, "secondary")
		})

		t.Run("single-kind, undefined", func(t *testing.T) {
			c := makeBasicBuilder().Build()
			expectAttributeNotFoundForRef(t, c, "secondary")
		})

		t.Run("multi-kind", func(t *testing.T) {
			c := NewMultiBuilder().Add(makeBasicBuilder().Secondary("my-value").Build()).Build()
			expectAttributeNotFoundForRef(t, c, "secondary")
		})
	})
}

func TestGetValueForRefCustomAttributeSingleKind(t *testing.T) {
	t.Run("simple attribute name", func(t *testing.T) {
		expected := ldvalue.String("abc")
		c := makeBasicBuilder().SetValue("my-attr", expected).Build()
		expectAttributeFoundForRef(t, expected, c, "my-attr")
	})

	t.Run("simple attribute name not found", func(t *testing.T) {
		c := makeBasicBuilder().Build()
		expectAttributeNotFoundForRef(t, c, "my-attr")
	})

	t.Run("property in object", func(t *testing.T) {
		expected := ldvalue.String("abc")
		object := ldvalue.ObjectBuild().Set("my-prop", expected).Build()
		c := makeBasicBuilder().SetValue("my-attr", object).Build()
		expectAttributeFoundForRef(t, expected, c, "/my-attr/my-prop")
	})

	t.Run("property in object not found", func(t *testing.T) {
		expected := ldvalue.String("abc")
		object := ldvalue.ObjectBuild().Set("my-prop", expected).Build()
		c := makeBasicBuilder().SetValue("my-attr", object).Build()
		expectAttributeNotFoundForRef(t, c, "/my-attr/other-prop")
	})

	t.Run("property in nested object", func(t *testing.T) {
		expected := ldvalue.String("abc")
		object := ldvalue.ObjectBuild().Set("my-prop", ldvalue.ObjectBuild().Set("sub-prop", expected).Build()).Build()
		c := makeBasicBuilder().SetValue("my-attr", object).Build()
		expectAttributeFoundForRef(t, expected, c, "/my-attr/my-prop/sub-prop")
	})

	t.Run("property in value that is not an object", func(t *testing.T) {
		c := makeBasicBuilder().SetValue("my-attr", ldvalue.String("xyz")).Build()
		expectAttributeNotFoundForRef(t, c, "/my-attr/my-prop")
	})

	t.Run("element in array", func(t *testing.T) {
		expected := ldvalue.String("good")
		array := ldvalue.ArrayOf(ldvalue.String("bad"), expected, ldvalue.String("worse"))
		c := makeBasicBuilder().SetValue("my-attr", array).Build()
		expectAttributeFoundForRef(t, expected, c, "/my-attr/1")
	})

	t.Run("element in nested array in object", func(t *testing.T) {
		expected := ldvalue.String("good")
		array := ldvalue.ArrayOf(ldvalue.String("bad"), expected, ldvalue.String("worse"))
		object := ldvalue.ObjectBuild().Set("my-prop", array).Build()
		c := makeBasicBuilder().SetValue("my-attr", object).Build()
		expectAttributeFoundForRef(t, expected, c, "/my-attr/my-prop/1")
	})

	t.Run("index too low in array", func(t *testing.T) {
		expected := ldvalue.String("good")
		array := ldvalue.ArrayOf(ldvalue.String("bad"), expected, ldvalue.String("worse"))
		c := makeBasicBuilder().SetValue("my-attr", array).Build()
		expectAttributeNotFoundForRef(t, c, "/my-attr/-1")
	})

	t.Run("index too high in array", func(t *testing.T) {
		expected := ldvalue.String("good")
		array := ldvalue.ArrayOf(ldvalue.String("bad"), expected, ldvalue.String("worse"))
		c := makeBasicBuilder().SetValue("my-attr", array).Build()
		expectAttributeNotFoundForRef(t, c, "/my-attr/3")
	})

	t.Run("index in value that is not an object", func(t *testing.T) {
		c := makeBasicBuilder().SetValue("my-attr", ldvalue.String("xyz")).Build()
		expectAttributeNotFoundForRef(t, c, "/my-attr/1")
	})
}

func TestContextString(t *testing.T) {
	c := makeBasicBuilder().Name("x").Transient(true).SetString("attr", "value").Build()
	j, _ := json.Marshal(c)
	s := c.String()
	m.In(t).Assert(json.RawMessage(s), m.JSONEqual(json.RawMessage(j)))
}

func TestGetValueForInvalidRef(t *testing.T) {
	c := makeBasicBuilder().Build()
	expectAttributeNotFoundForRef(t, c, "/")
}

func expectAttributeFoundForName(t *testing.T, expected ldvalue.Value, c Context, attrName string) {
	t.Helper()
	value, ok := c.GetValue(attrName)
	assert.True(t, ok, "attribute %q should have been found, but was not", attrName)
	m.In(t).Assert(value, m.JSONEqual(expected))
}

func expectAttributeNotFoundForName(t *testing.T, c Context, attrName string) {
	t.Helper()
	value, ok := c.GetValue(attrName)
	assert.False(t, ok, "attribute %q should not have been found, but was", attrName)
	m.In(t).Assert(value, m.JSONEqual(nil))
}

func expectAttributeFoundForRef(t *testing.T, expected ldvalue.Value, c Context, attrRefString string) {
	t.Helper()
	value, ok := c.GetValueForRef(ldattr.NewRef(attrRefString))
	assert.True(t, ok, "attribute %q should have been found, but was not", attrRefString)
	m.In(t).Assert(value, m.JSONEqual(expected))
}

func expectAttributeNotFoundForRef(t *testing.T, c Context, attrRefString string) {
	t.Helper()
	value, ok := c.GetValueForRef(ldattr.NewRef(attrRefString))
	assert.False(t, ok, "attribute %q should not have been found, but was", attrRefString)
	m.In(t).Assert(value, m.JSONEqual(nil))
}
