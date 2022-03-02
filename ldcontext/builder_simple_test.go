package ldcontext

import (
	"testing"

	"gopkg.in/launchdarkly/go-sdk-common.v3/ldvalue"

	m "github.com/launchdarkly/go-test-helpers/v2/matchers"

	"github.com/stretchr/testify/assert"
)

type invalidKindTestParams struct {
	kind string
	err  error
}

func makeInvalidKindTestParams() []invalidKindTestParams {
	return []invalidKindTestParams{
		{"kind", errContextKindCannotBeKind},
		{"multi", errContextKindMultiWithSimpleBuilder},
		{"örg", errContextKindInvalidChars},
		{"o~rg", errContextKindInvalidChars},
		{"😀rg", errContextKindInvalidChars},
		{"o\trg", errContextKindInvalidChars},
	}
}

func makeBasicBuilder() *Builder {
	// for test cases where the kind and key are unimportant
	return NewBuilder("my-key")
}

func TestBuilderDefaultProperties(t *testing.T) {
	c := NewBuilder("my-key").Build()
	assert.NoError(t, c.Err())
	assert.Equal(t, DefaultKind, c.Kind())
	assert.Equal(t, "my-key", c.Key())

	assert.Equal(t, ldvalue.OptionalString{}, c.Name())
	assert.False(t, c.Transient())
	assert.Equal(t, ldvalue.OptionalString{}, c.Secondary())
	assert.Len(t, c.GetOptionalAttributeNames(nil), 0)
}

func TestBuilderKindValidation(t *testing.T) {
	for _, p := range makeInvalidKindTestParams() {
		t.Run(p.kind, func(t *testing.T) {
			b := NewBuilder("my-key").Kind(Kind(p.kind))

			c0 := b.Build()
			assert.Equal(t, p.err, c0.Err())

			c1, err := b.TryBuild()
			assert.Equal(t, p.err, c1.Err())
			assert.Equal(t, p.err, err)
		})
	}
}

func TestBuilderKeyValidation(t *testing.T) {
	b := NewBuilder("")

	c0 := b.Build()
	assert.Equal(t, errContextKeyEmpty, c0.Err())

	c1, err := b.TryBuild()
	assert.Equal(t, errContextKeyEmpty, c1.Err())
	assert.Equal(t, errContextKeyEmpty, err)
}

func TestBuilderFullyQualifiedKey(t *testing.T) {
	t.Run("kind is user", func(t *testing.T) {
		c := New("my-user-key")
		assert.Equal(t, "my-user-key", c.FullyQualifiedKey())
	})

	t.Run("kind is not user", func(t *testing.T) {
		c := NewWithKind("org", "my-org-key")
		assert.Equal(t, "org:my-org-key", c.FullyQualifiedKey())
	})
}

func TestBuilderBasicSetters(t *testing.T) {
	t.Run("Kind", func(t *testing.T) {
		assert.Equal(t, Kind("org"), NewBuilder("my-key").Kind("org").Build().Kind())

		assert.Equal(t, DefaultKind, NewBuilder("my-key").Kind("").Build().Kind())
	})

	t.Run("Key", func(t *testing.T) {
		assert.Equal(t, "other-key", NewBuilder("my-key").Key("other-key").Build().Key())
	})

	t.Run("Name", func(t *testing.T) {
		c0 := makeBasicBuilder().Build()
		assert.Equal(t, ldvalue.OptionalString{}, c0.Name())

		c1 := makeBasicBuilder().Name("my-name").Build()
		assert.Equal(t, ldvalue.NewOptionalString("my-name"), c1.Name())

		c2 := makeBasicBuilder().OptName(ldvalue.OptionalString{}).Build()
		assert.Equal(t, ldvalue.OptionalString{}, c2.Name())

		c3 := makeBasicBuilder().OptName(ldvalue.NewOptionalString("my-name")).Build()
		assert.Equal(t, ldvalue.NewOptionalString("my-name"), c3.Name())
	})

	t.Run("Secondary", func(t *testing.T) {
		c0 := makeBasicBuilder().Build()
		assert.Equal(t, ldvalue.OptionalString{}, c0.Secondary())

		c1 := makeBasicBuilder().Secondary("value").Build()
		assert.Equal(t, ldvalue.NewOptionalString("value"), c1.Secondary())

		c2 := makeBasicBuilder().OptSecondary(ldvalue.OptionalString{}).Build()
		assert.Equal(t, ldvalue.OptionalString{}, c2.Secondary())

		c3 := makeBasicBuilder().OptSecondary(ldvalue.NewOptionalString("value")).Build()
		assert.Equal(t, ldvalue.NewOptionalString("value"), c3.Secondary())
	})

	t.Run("Transient", func(t *testing.T) {
		c0 := makeBasicBuilder().Build()
		assert.False(t, c0.Transient())

		c1 := makeBasicBuilder().Transient(false).Build()
		assert.False(t, c1.Transient())

		c2 := makeBasicBuilder().Transient(true).Build()
		assert.True(t, c2.Transient())
	})
}

func TestBuilderSetCustomAttributes(t *testing.T) {
	t.Run("SetValue", func(t *testing.T) {
		otherValue := ldvalue.String("other-value")
		for _, value := range []ldvalue.Value{
			ldvalue.Bool(true),
			ldvalue.Bool(false),
			ldvalue.Int(0),
			ldvalue.Int(1),
			ldvalue.String(""),
			ldvalue.String("x"),
			ldvalue.ArrayOf(ldvalue.Int(1), ldvalue.Int(2)),
			ldvalue.ObjectBuild().Set("a", ldvalue.Int(1)).Build(),
		} {
			t.Run(value.JSONString(), func(t *testing.T) {
				c := makeBasicBuilder().
					SetValue("my-attr", value).
					SetValue("other-attr", otherValue).
					Build()
				assert.Len(t, c.attributes, 2)
				m.In(t).Assert(c.attributes["my-attr"], m.JSONEqual(value))
				m.In(t).Assert(c.attributes["other-attr"], m.JSONEqual(otherValue))
			})
		}
	})

	t.Run("typed setters", func(t *testing.T) {
		// For the typed setters, just verify that they produce the same builder state as SetValue
		assert.Equal(t,
			makeBasicBuilder().SetValue("my-attr", ldvalue.Bool(true)),
			makeBasicBuilder().SetBool("my-attr", true))
		assert.Equal(t,
			makeBasicBuilder().SetValue("my-attr", ldvalue.Int(100)),
			makeBasicBuilder().SetInt("my-attr", 100))
		assert.Equal(t,
			makeBasicBuilder().SetValue("my-attr", ldvalue.Float64(1.5)),
			makeBasicBuilder().SetFloat64("my-attr", 1.5))
		assert.Equal(t,
			makeBasicBuilder().SetValue("my-attr", ldvalue.String("x")),
			makeBasicBuilder().SetString("my-attr", "x"))
	})

	t.Run("setting to null does not add attribute", func(t *testing.T) {
		assert.Equal(t,
			makeBasicBuilder().SetString("attr1", "value1").SetString("attr3", "value3"),
			makeBasicBuilder().SetString("attr1", "value1").SetValue("attr2", ldvalue.Null()).SetString("attr3", "value3"))
	})

	t.Run("setting to null removes existing attribute", func(t *testing.T) {
		assert.Equal(t,
			makeBasicBuilder().SetString("attr1", "value1").SetString("attr3", "value3"),
			makeBasicBuilder().SetString("attr1", "value1").SetString("attr2", "value2").SetString("attr3", "value3").
				SetValue("attr2", ldvalue.Null()))
	})
}

func TestBuilderSetBuiltInAttributesByName(t *testing.T) {
	const nonEmptyString = "x"
	nonEmptyStringValue := ldvalue.String(nonEmptyString)

	t.Run("Kind", func(t *testing.T) {
		assert.Equal(t,
			makeBasicBuilder().Kind(nonEmptyString),
			makeBasicBuilder().SetValue("kind", nonEmptyStringValue))

		assert.Equal(t,
			makeBasicBuilder().Kind(nonEmptyString),
			makeBasicBuilder().SetString("kind", nonEmptyString))

		assert.Equal(t,
			makeBasicBuilder().Kind(nonEmptyString).Kind(""),                         // set it and then clear it
			makeBasicBuilder().Kind(nonEmptyString).SetValue("kind", ldvalue.Null())) // using wrong type clears it

		assert.Equal(t,
			makeBasicBuilder().Kind(nonEmptyString).Kind(""),                             // set it and then clear it
			makeBasicBuilder().Kind(nonEmptyString).SetValue("kind", ldvalue.Bool(true))) // using wrong type clears it
	})

	t.Run("Key", func(t *testing.T) {
		assert.Equal(t,
			makeBasicBuilder().Key(nonEmptyString),
			makeBasicBuilder().SetValue("key", nonEmptyStringValue))

		assert.Equal(t,
			makeBasicBuilder().Key(nonEmptyString),
			makeBasicBuilder().SetString("key", nonEmptyString))

		assert.Equal(t,
			makeBasicBuilder().Key(nonEmptyString).Key(""),                         // set it and then clear it
			makeBasicBuilder().Key(nonEmptyString).SetValue("key", ldvalue.Null())) // using wrong type clears it

		assert.Equal(t,
			makeBasicBuilder().Key(nonEmptyString).Key(""),                             // set it and then clear it
			makeBasicBuilder().Key(nonEmptyString).SetValue("key", ldvalue.Bool(true))) // using wrong type clears it
	})

	testNullableStringAttr := func(
		t *testing.T,
		attrName string,
		setter func(*Builder, string) *Builder,
		optSetter func(*Builder, ldvalue.OptionalString) *Builder,
	) {
		assert.Equal(t,
			setter(makeBasicBuilder(), nonEmptyString),
			makeBasicBuilder().SetValue(attrName, nonEmptyStringValue))

		assert.Equal(t,
			setter(makeBasicBuilder(), nonEmptyString),
			makeBasicBuilder().SetString(attrName, nonEmptyString))

		assert.Equal(t,
			makeBasicBuilder(), // attribute not set, defaults to null
			setter(makeBasicBuilder(), nonEmptyString).SetValue(attrName, ldvalue.Null())) // null value clears previous value

		assert.Equal(t,
			makeBasicBuilder(), // attribute not set, defaults to null
			setter(makeBasicBuilder(), nonEmptyString).SetValue(attrName, ldvalue.Bool(true))) // wrong type clears previous value

		assert.Equal(t,
			setter(makeBasicBuilder(), ""), // "" is distinct from null
			makeBasicBuilder().SetValue(attrName, ldvalue.String("")))

		assert.Equal(t,
			setter(makeBasicBuilder(), ""),
			makeBasicBuilder().SetString(attrName, ""))
	}

	t.Run("Name", func(t *testing.T) {
		testNullableStringAttr(t, "name", (*Builder).Name, (*Builder).OptName)
	})

	t.Run("Secondary", func(t *testing.T) {
		testNullableStringAttr(t, "secondary", (*Builder).Secondary, (*Builder).OptSecondary)
	})

	t.Run("Transient", func(t *testing.T) {
		assert.Equal(t,
			makeBasicBuilder().Transient(true),
			makeBasicBuilder().SetValue("transient", ldvalue.Bool(true)))

		assert.Equal(t,
			makeBasicBuilder().Transient(false),                                           // for clarity, but it defaults to false
			makeBasicBuilder().Transient(true).SetValue("transient", ldvalue.Bool(false))) // overwrites previous value

		assert.Equal(t,
			makeBasicBuilder().Transient(false),
			makeBasicBuilder().Transient(true).SetValue("transient", ldvalue.Null()))

		assert.Equal(t,
			makeBasicBuilder().Transient(false),
			makeBasicBuilder().Transient(true).SetValue("transient", ldvalue.String("x"))) // wrong type sets it to false
	})
}

func TestBuilderAttributesCopyOnWrite(t *testing.T) {
	value1, value2 := ldvalue.String("value1"), ldvalue.String("value2")

	b := makeBasicBuilder().SetValue("attr", value1)

	c1 := b.Build()
	m.In(t).Assert(c1.attributes["attr"], m.JSONEqual(value1))

	b.SetValue("attr", value2)

	c2 := b.Build()
	m.In(t).Assert(c2.attributes["attr"], m.JSONEqual(value2))
	m.In(t).Assert(c1.attributes["attr"], m.JSONEqual(value1)) // unchanged
}

func TestBuilderPrivate(t *testing.T) {
	expectPrivateRefsToBe := func(t *testing.T, c Context, expectedRefs ...AttrRef) {
		if assert.Equal(t, len(expectedRefs), c.PrivateAttributeCount()) {
			for i, expectedRef := range expectedRefs {
				a, ok := c.PrivateAttributeByIndex(i)
				assert.True(t, ok)
				assert.Equal(t, expectedRef, a)
			}
			_, ok := c.PrivateAttributeByIndex(len(expectedRefs))
			assert.False(t, ok)
		}
		_, ok := c.PrivateAttributeByIndex(-1)
		assert.False(t, ok)
	}

	t.Run("using AttrRefs", func(t *testing.T) {
		attrRef1, attrRef2, attrRef3 := NewAttrRef("a"), NewAttrRef("/b/c"), NewAttrRef("d")
		c := makeBasicBuilder().
			PrivateRef(attrRef1, attrRef2).PrivateRef(attrRef3).
			Build()

		expectPrivateRefsToBe(t, c, attrRef1, attrRef2, attrRef3)
	})

	t.Run("using strings", func(t *testing.T) {
		s1, s2, s3 := "a", "/b/c", "d"
		b0 := makeBasicBuilder().
			PrivateRef(NewAttrRef(s1), NewAttrRef(s2)).PrivateRef(NewAttrRef(s3))
		b1 := makeBasicBuilder().
			Private(s1, s2, s3)
		assert.Equal(t, b0, b1)
	})

	t.Run("RemovePrivate", func(t *testing.T) {
		b := makeBasicBuilder().Private("a", "/b/c", "d", "/b/c")
		b.RemovePrivate("/b/c")
		c := b.Build()

		expectPrivateRefsToBe(t, c, NewAttrRef("a"), NewAttrRef("d"))
	})

	t.Run("RemovePrivateRef", func(t *testing.T) {
		b := makeBasicBuilder().Private("a", "/b/c", "d", "/b/c")
		b.RemovePrivateRef(NewAttrRef("/b/c"))
		c := b.Build()

		expectPrivateRefsToBe(t, c, NewAttrRef("a"), NewAttrRef("d"))
	})

	t.Run("copy on write", func(t *testing.T) {
		b0 := makeBasicBuilder().Private("a")

		c0 := b0.Build()
		expectPrivateRefsToBe(t, c0, NewAttrRef("a"))

		b0.Private("b")
		c1 := b0.Build()
		expectPrivateRefsToBe(t, c1, NewAttrRef("a"), NewAttrRef("b"))
		expectPrivateRefsToBe(t, c0, NewAttrRef("a")) // unchanged

		b0.RemovePrivateRef(NewAttrRef("a"))
		c2 := b0.Build()
		expectPrivateRefsToBe(t, c2, NewAttrRef("b"))
		expectPrivateRefsToBe(t, c1, NewAttrRef("a"), NewAttrRef("b")) // unchanged
		expectPrivateRefsToBe(t, c0, NewAttrRef("a"))                  // unchanged
	})
}

func TestNewBuilderFromContext(t *testing.T) {
	value1, value2 := ldvalue.String("value1"), ldvalue.String("value2")

	b1 := NewBuilder("key1").Kind("kind1").Name("name1").Secondary("sec1").Transient(true).SetValue("attr", value1)
	b1.Private("private1")
	c1 := b1.Build()
	m.In(t).Assert(c1.attributes["attr"], m.JSONEqual(value1))
	assert.Len(t, c1.privateAttrs, 1)

	b2 := NewBuilderFromContext(c1)
	c2 := b2.Build()
	assert.Equal(t, Kind("kind1"), c2.Kind())
	assert.Equal(t, "key1", c2.Key())
	assert.Equal(t, ldvalue.NewOptionalString("sec1"), c2.Secondary())
	assert.True(t, c2.Transient())
	m.In(t).Assert(c2.attributes["attr"], m.JSONEqual(value1))
	assert.Equal(t, c1.privateAttrs, c2.privateAttrs)

	b3 := NewBuilderFromContext(c1)
	b3.SetValue("attr", value2)
	b3.Private("private2")
	c3 := b3.Build()
	m.In(t).Assert(c3.attributes["attr"], m.JSONEqual(value2))
	m.In(t).Assert(c1.attributes["attr"], m.JSONEqual(value1)) // unchanged
	assert.Len(t, c3.privateAttrs, 2)
	assert.Len(t, c1.privateAttrs, 1) // unchanged

	multi := NewMulti(NewWithKind("kind1", "key1"), NewWithKind("kind2", "key2"))
	assert.NoError(t, multi.Err())
	c4 := NewBuilderFromContext(multi).Build()
	assert.Error(t, c4.Err()) // can't copy Builder from multi-kind context
}
