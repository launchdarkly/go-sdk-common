package ldcontext

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/lderrors"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

	"github.com/launchdarkly/go-test-helpers/v3/jsonhelpers"

	"github.com/stretchr/testify/assert"
)

// Note, matchers.JSONEqual is preferred in these tests when checking ldvalue.Value values, rather
// than assert.Equal or assert.JSONEq, because its failure output is easier to read.

func TestUninitializedContext(t *testing.T) {
	var c Context
	assert.False(t, c.IsDefined())
	assert.Equal(t, lderrors.ErrContextUninitialized{}, c.Err())
}

func TestMultiple(t *testing.T) {
	sc := New("my-key")
	mc := NewMulti(New("my-key"), NewWithKind("org", "my-key"))
	assert.False(t, sc.Multiple())
	assert.True(t, mc.Multiple())
}

func TestGetOptionalAttributeNames(t *testing.T) {
	sorted := func(values []string) []string {
		ret := append([]string(nil), values...)
		sort.Strings(ret)
		return ret
	}

	t.Run("none", func(t *testing.T) {
		c := New("my-key")
		an := c.GetOptionalAttributeNames(nil)
		assert.Len(t, an, 0)
	})

	t.Run("meta not included", func(t *testing.T) {
		c := NewBuilder("my-key").Private("x").Anonymous(true).Build()
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

	t.Run("none for multi-context", func(t *testing.T) {
		c := NewMulti(NewWithKind("kind1", "key1"), NewWithKind("otherkind", "otherkey"))
		an := c.GetOptionalAttributeNames(nil)
		assert.Len(t, an, 0)
	})

	t.Run("capacity of preallocated slice can be reused", func(t *testing.T) {
		c := NewBuilder("my-key").SetString("email", "x").SetBool("happy", true).Build()
		preallocSlice := make([]string, 2, 2)
		emptySlice := preallocSlice[0:0]
		an := c.GetOptionalAttributeNames(emptySlice)
		assert.Equal(t, []string{"email", "happy"}, sorted(an))
		preallocSlice[0] = "x"
		assert.Equal(t, "x", an[0])
	})

	t.Run("preallocated slice is overwritten rather than appended to", func(t *testing.T) {
		c := NewBuilder("my-key").SetString("email", "x").SetBool("happy", true).Build()
		preallocSlice := make([]string, 2, 2)
		an := c.GetOptionalAttributeNames(preallocSlice)
		assert.Equal(t, []string{"email", "happy"}, sorted(an))
		preallocSlice[0] = "x"
		assert.Equal(t, "x", an[0])
	})

	t.Run("preallocated slice without enough capacity is not reused", func(t *testing.T) {
		c := NewBuilder("my-key").SetString("email", "x").SetBool("happy", true).Build()
		preallocSlice := make([]string, 1, 1)
		emptySlice := preallocSlice[0:0]
		an := c.GetOptionalAttributeNames(emptySlice)
		assert.Equal(t, []string{"email", "happy"}, sorted(an))
		preallocSlice[0] = "x"
		assert.NotEqual(t, "x", an[0])
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

		mc := NewMulti(c, NewWithKind("otherkind", "otherkey"))
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
			c := NewMulti(New("my-key"), NewWithKind("otherkind", "otherkey"))
			expectAttributeFoundForRef(t, ldvalue.String("multi"), c, "kind")
		})
	})

	t.Run("key", func(t *testing.T) {
		t.Run("single-kind", func(t *testing.T) {
			c := New("my-key")
			expectAttributeFoundForRef(t, ldvalue.String("my-key"), c, "key")
		})

		t.Run("multi-kind", func(t *testing.T) {
			c := NewMulti(New("my-key"), NewWithKind("otherkind", "otherkey"))
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
			c := NewMulti(makeBasicBuilder().Name("my-name").Build(), NewWithKind("otherkind", "otherkey"))
			expectAttributeNotFoundForRef(t, c, "name")
		})
	})

	t.Run("anonymous", func(t *testing.T) {
		t.Run("single-kind, defined, true", func(t *testing.T) {
			c := makeBasicBuilder().Anonymous(true).Build()
			expectAttributeFoundForRef(t, ldvalue.Bool(true), c, "anonymous")
		})

		t.Run("single-kind, defined, false", func(t *testing.T) {
			c := makeBasicBuilder().Anonymous(false).Build()
			expectAttributeFoundForRef(t, ldvalue.Bool(false), c, "anonymous")
		})

		t.Run("single-kind, undefined", func(t *testing.T) {
			c := makeBasicBuilder().Build()
			expectAttributeFoundForRef(t, ldvalue.Bool(false), c, "anonymous")
		})

		t.Run("multi-kind", func(t *testing.T) {
			c := NewMulti(makeBasicBuilder().Anonymous(true).Build(), NewWithKind("otherkind", "otherkey"))
			expectAttributeNotFoundForRef(t, c, "anonymous")
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

	t.Run("property in raw JSON object", func(t *testing.T) {
		expected := ldvalue.String("abc")
		object := ldvalue.Raw(json.RawMessage(`{"my-prop": "abc"}`))
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

	t.Run("property whose name is a numeric string", func(t *testing.T) {
		expected := ldvalue.String("good")
		object := ldvalue.ObjectBuild().Set("1", expected).Build()
		c := makeBasicBuilder().SetValue("my-attr", object).Build()
		expectAttributeFoundForRef(t, expected, c, "/my-attr/1")
	})

	t.Run("property not applicable to array", func(t *testing.T) {
		array := ldvalue.ArrayOf(ldvalue.String("bad"), ldvalue.String("worse"))
		c := makeBasicBuilder().SetValue("my-attr", array).Build()
		expectAttributeNotFoundForRef(t, c, "/my-attr/1")
	})

	t.Run("property not applicable to simple value", func(t *testing.T) {
		c := makeBasicBuilder().SetValue("my-attr", ldvalue.String("xyz")).Build()
		expectAttributeNotFoundForRef(t, c, "/my-attr/a")
	})
}

func TestContextString(t *testing.T) {
	c := makeBasicBuilder().Name("x").Anonymous(true).SetString("attr", "value").Build()
	j, _ := json.Marshal(c)
	s := c.String()
	jsonhelpers.AssertEqual(t, j, s)
}

func TestGetValueForInvalidRef(t *testing.T) {
	c := makeBasicBuilder().Build()
	expectAttributeNotFoundForRef(t, c, "/")
}

func TestIndividualContextCount(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		c := New("my-key")
		assert.Equal(t, 1, c.IndividualContextCount())
	})

	t.Run("multi", func(t *testing.T) {
		sub1, sub2 := NewWithKind("kind1", "key1"), NewWithKind("kind2", "key2")
		c := NewMulti(sub1, sub2)
		assert.Equal(t, 2, c.IndividualContextCount())
	})
}

func TestIndividualContextByIndex(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		c := New("my-key")

		assert.Equal(t, c, c.IndividualContextByIndex(0))
		assert.Equal(t, Context{}, c.IndividualContextByIndex(1))
		assert.Equal(t, Context{}, c.IndividualContextByIndex(-1))
	})

	t.Run("multi", func(t *testing.T) {
		sub1, sub2 := NewWithKind("kind1", "key1"), NewWithKind("kind2", "key2")
		c := NewMulti(sub1, sub2)
		// We know that these are internally sorted by kind

		assert.Equal(t, sub1, c.IndividualContextByIndex(0))
		assert.Equal(t, sub2, c.IndividualContextByIndex(1))
		assert.Equal(t, Context{}, c.IndividualContextByIndex(2))
		assert.Equal(t, Context{}, c.IndividualContextByIndex(-1))
	})
}

func TestIndividualContextByKind(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		c := NewWithKind("kind1", "my-key")

		assert.Equal(t, c, c.IndividualContextByKind("kind1"))
		assert.Equal(t, Context{}, c.IndividualContextByKind("other"))
	})

	t.Run("multi", func(t *testing.T) {
		sub1, sub2 := NewWithKind("kind1", "key1"), NewWithKind("kind2", "key2")
		c := NewMulti(sub1, sub2)

		assert.Equal(t, sub1, c.IndividualContextByKind("kind1"))
		assert.Equal(t, sub2, c.IndividualContextByKind("kind2"))
		assert.Equal(t, Context{}, c.IndividualContextByKind("other"))
	})

	t.Run("default", func(t *testing.T) {
		sub1, sub2 := New("userkey"), NewWithKind("kind2", "key2")
		c := NewMulti(sub1, sub2)

		assert.Equal(t, sub1, c.IndividualContextByKind(""))
	})
}

func TestIndividualContextKeyByKind(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		c := NewWithKind("kind1", "my-key")

		assert.Equal(t, "my-key", c.IndividualContextKeyByKind("kind1"))
		assert.Equal(t, "", c.IndividualContextKeyByKind("other"))
	})

	t.Run("multi", func(t *testing.T) {
		sub1, sub2 := NewWithKind("kind1", "key1"), NewWithKind("kind2", "key2")
		c := NewMulti(sub1, sub2)

		assert.Equal(t, "key1", c.IndividualContextKeyByKind("kind1"))
		assert.Equal(t, "key2", c.IndividualContextKeyByKind("kind2"))
		assert.Equal(t, "", c.IndividualContextKeyByKind("other"))
	})

	t.Run("default", func(t *testing.T) {
		sub1, sub2 := New("userkey"), NewWithKind("kind2", "key2")
		c := NewMulti(sub1, sub2)

		assert.Equal(t, "userkey", c.IndividualContextKeyByKind(""))
	})
}

func TestGetAllIndividualContexts(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		c := NewWithKind("kind1", "my-key")

		assert.Equal(t, []Context{c}, c.GetAllIndividualContexts(nil))
	})

	t.Run("multi", func(t *testing.T) {
		sub1, sub2 := NewWithKind("kind1", "key1"), NewWithKind("kind2", "key2")
		c := NewMulti(sub1, sub2)
		// We know that these are internally sorted by kind

		assert.Equal(t, []Context{sub1, sub2}, c.GetAllIndividualContexts(nil))
	})

	t.Run("capacity of preallocated slice can be reused", func(t *testing.T) {
		sub1, sub2 := NewWithKind("kind1", "key1"), NewWithKind("kind2", "key2")
		c := NewMulti(sub1, sub2)

		preallocSlice := make([]Context, 2, 2)
		emptySlice := preallocSlice[0:0]
		all := c.GetAllIndividualContexts(emptySlice)
		assert.Equal(t, []Context{sub1, sub2}, all)
		preallocSlice[0] = New("different")
		assert.Equal(t, preallocSlice[0], all[0])
	})

	t.Run("preallocated slice is overwritten rather than appended to", func(t *testing.T) {
		sub1, sub2 := NewWithKind("kind1", "key1"), NewWithKind("kind2", "key2")
		c := NewMulti(sub1, sub2)

		preallocSlice := make([]Context, 2, 2)
		all := c.GetAllIndividualContexts(preallocSlice)
		assert.Equal(t, []Context{sub1, sub2}, all)
		preallocSlice[0] = New("different")
		assert.Equal(t, preallocSlice[0], all[0])
	})

	t.Run("preallocated slice without enough capacity is not reused", func(t *testing.T) {
		sub1, sub2 := NewWithKind("kind1", "key1"), NewWithKind("kind2", "key2")
		c := NewMulti(sub1, sub2)

		preallocSlice := make([]Context, 1, 1)
		emptySlice := preallocSlice[0:0]
		all := c.GetAllIndividualContexts(emptySlice)
		assert.Equal(t, []Context{sub1, sub2}, all)
		preallocSlice[0] = New("different")
		assert.NotEqual(t, preallocSlice[0], all[0])
	})
}

func TestCanGetSecondaryKeyIfPrivatelySet(t *testing.T) {
	c1, c2 := New("key1"), New("key2")
	c2.secondary = ldvalue.NewOptionalString("x")
	assert.Equal(t, ldvalue.OptionalString{}, c1.Secondary())
	assert.Equal(t, c2.secondary, c2.Secondary())
}

func TestContextEqual(t *testing.T) {
	// Each top-level element in makeInstances is a slice of factories that should produce contexts equal to
	// each other, and unequal to the contexts produced by the factories in any other slice.
	makeInstances := [][]func() Context{
		{func() Context { return Context{} }},
		{func() Context { return New("a") }},
		{func() Context { return New("b") }},
		{func() Context { return NewWithKind("k1", "a") }},
		{func() Context { return NewWithKind("k2", "a") }},
		{func() Context { return NewBuilder("a").Name("b").Build() }},
		{func() Context { return NewBuilder("a").Name("c").Build() }},
		{func() Context { return NewBuilder("a").Anonymous(true).Build() }},
		{func() Context { return NewBuilder("a").SetBool("b", true).Build() }},
		{func() Context { return NewBuilder("a").SetBool("b", false).Build() }},
		{func() Context { return NewBuilder("a").SetInt("b", 0).Build() }},
		{func() Context { return NewBuilder("a").SetInt("b", 1).Build() }},
		{func() Context { return NewBuilder("a").SetString("b", "").Build() }},
		{func() Context { return NewBuilder("a").SetString("b", "c").Build() }},
		{func() Context { return NewBuilder("a").SetBool("b", true).SetBool("c", false).Build() },
			func() Context { return NewBuilder("a").SetBool("c", false).SetBool("b", true).Build() }},
		{func() Context { return NewBuilder("a").Name("b").Private("name").Build() }},
		{func() Context { return NewBuilder("a").Name("b").SetBool("c", true).Private("name").Build() }},
		{func() Context { return NewBuilder("a").Name("b").SetBool("c", true).Private("name", "c").Build() },
			func() Context { return NewBuilder("a").Name("b").SetBool("c", true).Private("c", "name").Build() }},
		{func() Context { return NewBuilder("a").Name("b").SetBool("c", true).Private("name", "d").Build() }},
		{func() Context { return NewMulti(NewWithKind("k1", "a"), NewWithKind("k2", "b")) },
			func() Context { return NewMulti(NewWithKind("k2", "b"), NewWithKind("k1", "a")) }},
		{func() Context { return NewMulti(NewWithKind("k1", "a"), NewWithKind("k2", "c")) }},
		{func() Context { return NewMulti(NewWithKind("k1", "a"), NewWithKind("k3", "b")) }},
		{func() Context {
			return NewMulti(NewWithKind("k1", "a"), NewWithKind("k2", "b"), NewWithKind("k3", "c"))
		}},
	}
	for i, equalGroup := range makeInstances {
		for _, factory1 := range equalGroup {
			c1 := factory1()
			for _, factory2 := range equalGroup {
				c2 := factory2()
				assert.True(t, c1.Equal(c2), "%s should have equaled %s", c1, c2)
			}
			for j, unequalGroup := range makeInstances {
				if i == j {
					continue
				}
				c2 := unequalGroup[0]()
				assert.False(t, c1.Equal(c2), "%s should not have equaled %s", c1, c2)
			}
		}
	}
}

func expectAttributeFoundForName(t *testing.T, expected ldvalue.Value, c Context, attrName string) {
	t.Helper()
	value := c.GetValue(attrName)
	assert.True(t, value.IsDefined(), "attribute %q should have been found, but was not", attrName)
	jsonhelpers.AssertEqual(t, expected, value)
}

func expectAttributeNotFoundForName(t *testing.T, c Context, attrName string) {
	t.Helper()
	value := c.GetValue(attrName)
	assert.False(t, value.IsDefined(), "attribute %q should not have been found, but was", attrName)
	jsonhelpers.AssertEqual(t, `null`, value)
}

func expectAttributeFoundForRef(t *testing.T, expected ldvalue.Value, c Context, attrRefString string) {
	t.Helper()
	value := c.GetValueForRef(ldattr.NewRef(attrRefString))
	assert.True(t, value.IsDefined(), "attribute %q should have been found, but was not", attrRefString)
	jsonhelpers.AssertEqual(t, expected, value)
}

func expectAttributeNotFoundForRef(t *testing.T, c Context, attrRefString string) {
	t.Helper()
	value := c.GetValueForRef(ldattr.NewRef(attrRefString))
	assert.False(t, value.IsDefined(), "attribute %q should not have been found, but was", attrRefString)
	jsonhelpers.AssertEqual(t, `null`, value)
}
