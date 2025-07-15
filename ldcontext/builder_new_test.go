package ldcontext

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/lderrors"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
	"github.com/launchdarkly/go-test-helpers/v3/jsonhelpers"
)

// Tests are adapted from builder_simple_test.go and builder_multi_test.go.

// SINGLE KIND TESTS (adapted from builder_simple_test.go)

func makeKindBuilder() *KindBuilder {
	// for test cases where the kind and key are unimportant
	return NewContextBuilder().Kind("user", "my-key")
}

func TestKindBuilderDefaultProperties(t *testing.T) {
	c := makeKindBuilder().Build()
	assert.True(t, c.IsDefined())
	assert.NoError(t, c.Err())
	assert.Equal(t, DefaultKind, c.Kind())
	assert.Equal(t, "my-key", c.Key())

	assert.Equal(t, ldvalue.OptionalString{}, c.Name())
	assert.False(t, c.Anonymous())
	assert.Equal(t, ldvalue.OptionalString{}, c.Secondary())
	assert.Len(t, c.GetOptionalAttributeNames(nil), 0)
}

func TestKindBuilderKindValidation(t *testing.T) {
	for _, p := range makeInvalidKindTestParams() {
		t.Run(p.kind, func(t *testing.T) {
			b := NewContextBuilder().Kind(Kind(p.kind), "my-key")

			c0 := b.Build()
			assert.True(t, c0.IsDefined())
			assert.Equal(t, p.err, c0.Err())

			c1, err := b.TryBuild()
			assert.True(t, c1.IsDefined())
			assert.Equal(t, p.err, c1.Err())
			assert.Equal(t, p.err, err)
		})
	}
}

func TestKindBuilderKeyValidation(t *testing.T) {
	b := NewContextBuilder().Kind("user", "")

	c0 := b.Build()
	assert.True(t, c0.IsDefined())
	assert.Equal(t, lderrors.ErrContextKeyEmpty{}, c0.Err())

	c1, err := b.TryBuild()
	assert.True(t, c1.IsDefined())
	assert.Equal(t, lderrors.ErrContextKeyEmpty{}, c1.Err())
	assert.Equal(t, lderrors.ErrContextKeyEmpty{}, err)
}

func TestKindBuilderFullyQualifiedKey(t *testing.T) {
	t.Run("kind is user", func(t *testing.T) {
		c := New("my-user-key")
		assert.Equal(t, "my-user-key", c.FullyQualifiedKey())
	})

	t.Run("kind is not user", func(t *testing.T) {
		c := NewWithKind("org", "my-org-key")
		assert.Equal(t, "org:my-org-key", c.FullyQualifiedKey())
	})

	t.Run("key is escaped", func(t *testing.T) {
		c := NewWithKind("org", "my:key%x/y")
		assert.Equal(t, "org:my%3Akey%25x/y", c.FullyQualifiedKey())
	})
}

func TestKindBuilderBasicSetters(t *testing.T) {
	t.Run("Key", func(t *testing.T) {
		assert.Equal(t, "other-key", makeKindBuilder().Key("other-key").Build().Key())
	})

	t.Run("Name", func(t *testing.T) {
		c0 := makeKindBuilder().Build()
		assert.Equal(t, ldvalue.OptionalString{}, c0.Name())

		c1 := makeKindBuilder().Name("my-name").Build()
		assert.Equal(t, ldvalue.NewOptionalString("my-name"), c1.Name())

		c2 := makeKindBuilder().OptName(ldvalue.OptionalString{}).Build()
		assert.Equal(t, ldvalue.OptionalString{}, c2.Name())

		c3 := makeKindBuilder().OptName(ldvalue.NewOptionalString("my-name")).Build()
		assert.Equal(t, ldvalue.NewOptionalString("my-name"), c3.Name())
	})

	t.Run("Secondary", func(t *testing.T) {
		c0 := makeKindBuilder().Build()
		assert.Equal(t, ldvalue.OptionalString{}, c0.Secondary())
	})

	t.Run("Anonymous", func(t *testing.T) {
		c0 := makeKindBuilder().Build()
		assert.False(t, c0.Anonymous())

		c1 := makeKindBuilder().Anonymous(false).Build()
		assert.False(t, c1.Anonymous())

		c2 := makeKindBuilder().Anonymous(true).Build()
		assert.True(t, c2.Anonymous())
	})
}

func TestKindBuilderSetCustomAttributes(t *testing.T) {
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
				c := makeKindBuilder().
					SetValue("my-attr", value).
					SetValue("other-attr", otherValue).
					Build()
				assert.Len(t, c.attributes.Keys(nil), 2)
				jsonhelpers.AssertEqual(t, value, c.attributes.Get("my-attr"))
				jsonhelpers.AssertEqual(t, otherValue, c.attributes.Get("other-attr"))
			})
		}
	})

	t.Run("typed setters", func(t *testing.T) {
		// For the typed setters, just verify that they produce the same builder state as SetValue
		assert.Equal(t,
			makeKindBuilder().SetValue("my-attr", ldvalue.Bool(true)),
			makeKindBuilder().SetBool("my-attr", true))
		assert.Equal(t,
			makeKindBuilder().SetValue("my-attr", ldvalue.Int(100)),
			makeKindBuilder().SetInt("my-attr", 100))
		assert.Equal(t,
			makeKindBuilder().SetValue("my-attr", ldvalue.Float64(1.5)),
			makeKindBuilder().SetFloat64("my-attr", 1.5))
		assert.Equal(t,
			makeKindBuilder().SetValue("my-attr", ldvalue.String("x")),
			makeKindBuilder().SetString("my-attr", "x"))
	})

	t.Run("setting to null does not add attribute", func(t *testing.T) {
		assert.Equal(t,
			makeKindBuilder().SetString("attr1", "value1").SetString("attr3", "value3"),
			makeKindBuilder().SetString("attr1", "value1").SetValue("attr2", ldvalue.Null()).SetString("attr3", "value3"))
	})

	t.Run("setting to null removes existing attribute", func(t *testing.T) {
		assert.Equal(t,
			makeKindBuilder().SetString("attr1", "value1").SetString("attr3", "value3"),
			makeKindBuilder().SetString("attr1", "value1").SetString("attr2", "value2").SetString("attr3", "value3").
				SetValue("attr2", ldvalue.Null()))
	})

	t.Run("cannot add attribute with empty name", func(t *testing.T) {
		assert.Equal(t, makeKindBuilder().Build(), makeKindBuilder().SetBool("", true).Build())
		assert.Equal(t, makeKindBuilder().Build(), makeKindBuilder().SetInt("", 1).Build())
		assert.Equal(t, makeKindBuilder().Build(), makeKindBuilder().SetFloat64("", 1).Build())
		assert.Equal(t, makeKindBuilder().Build(), makeKindBuilder().SetString("", "x").Build())
		assert.Equal(t, makeKindBuilder().Build(), makeKindBuilder().SetValue("", ldvalue.ArrayOf()).Build())
	})
}

func TestKindBuilderSetBuiltInAttributesByName(t *testing.T) {
	var boolFalse, boolTrue, stringEmpty, stringNonEmpty = ldvalue.Bool(false), ldvalue.Bool(true),
		ldvalue.String("x"), ldvalue.String("")
	var nullValue, intValue, floatValue, arrayValue, objectValue = ldvalue.Null(),
		ldvalue.Int(1), ldvalue.Float64(1.5), ldvalue.ArrayOf(), ldvalue.ObjectBuild().Build()

	type params struct {
		name             string
		equivalentSetter func(*KindBuilder, ldvalue.Value)
		good, bad        []ldvalue.Value
	}

	for _, p := range []params{
		{
			name:             "key",
			equivalentSetter: func(b *KindBuilder, v ldvalue.Value) { b.Key(v.StringValue()) },
			good:             []ldvalue.Value{stringNonEmpty, stringEmpty},
			bad:              []ldvalue.Value{nullValue, boolFalse, intValue, floatValue, arrayValue, objectValue},
		},
		{
			name:             "name",
			equivalentSetter: func(b *KindBuilder, v ldvalue.Value) { b.OptName(v.AsOptionalString()) },
			good:             []ldvalue.Value{stringNonEmpty, stringEmpty, nullValue},
			bad:              []ldvalue.Value{boolFalse, intValue, floatValue, arrayValue, objectValue},
		},
		{
			name:             "anonymous",
			equivalentSetter: func(b *KindBuilder, v ldvalue.Value) { b.Anonymous(v.BoolValue()) },
			good:             []ldvalue.Value{boolTrue, boolFalse},
			bad:              []ldvalue.Value{nullValue, intValue, floatValue, stringEmpty, stringNonEmpty, arrayValue, objectValue},
		},
	} {
		t.Run(p.name, func(t *testing.T) {
			builder := makeKindBuilder() // we will reuse this to prove that SetValue overwrites previous values
			var lastGoodNonNullValue ldvalue.Value

			for _, goodValue := range p.good {
				t.Run(fmt.Sprintf("can set to %s", goodValue.JSONString()), func(t *testing.T) {
					previousState := *builder

					if !goodValue.IsNull() {
						lastGoodNonNullValue = goodValue
					}
					expected := makeKindBuilder()
					p.equivalentSetter(expected, goodValue)

					builder.SetValue(p.name, goodValue)
					assert.Equal(t, expected, builder)

					b1 := previousState
					assert.True(t, b1.TrySetValue(p.name, goodValue))
					assert.Equal(t, *expected, b1)

					b2 := previousState
					switch goodValue.Type() {
					case ldvalue.BoolType:
						assert.Equal(t, expected, b2.SetBool(p.name, goodValue.BoolValue()))
					case ldvalue.StringType:
						assert.Equal(t, expected, b2.SetString(p.name, goodValue.StringValue()))
					}
				})
			}
			for _, badValue := range p.bad {
				t.Run(fmt.Sprintf("cannot set to %s", badValue.JSONString()), func(t *testing.T) {
					startingState := func() *KindBuilder {
						if lastGoodNonNullValue.IsDefined() {
							return makeKindBuilder().SetValue(p.name, lastGoodNonNullValue)
						}
						return makeKindBuilder()
					}

					assert.Equal(t, startingState(), startingState().SetValue(p.name, badValue))

					b := startingState()
					assert.False(t, b.TrySetValue(p.name, badValue))
					assert.Equal(t, startingState(), b)

					switch badValue.Type() {
					case ldvalue.BoolType:
						assert.Equal(t, startingState(), startingState().SetBool(p.name, badValue.BoolValue()))
					case ldvalue.NumberType:
						if badValue.IsInt() {
							assert.Equal(t, startingState(), startingState().SetInt(p.name, badValue.IntValue()))
						} else {
							assert.Equal(t, startingState(), startingState().SetFloat64(p.name, badValue.Float64Value()))
						}
					case ldvalue.StringType:
						assert.Equal(t, startingState(), makeKindBuilder().SetString(p.name, badValue.StringValue()))
					}
				})
			}
		})
	}
}

func TestKindBuilderSetValueCannotSetMetaProperties(t *testing.T) {
	for _, p := range []struct {
		name  string
		value ldvalue.Value
	}{
		{"secondary", ldvalue.String("x")},
		{"privateAttributes", ldvalue.ArrayOf(ldvalue.String("x"))},
	} {
		t.Run(p.name, func(t *testing.T) {
			c := makeKindBuilder().SetValue(p.name, p.value).Build()
			assert.Equal(t, p.value, c.attributes.Get(p.name))
			assert.Equal(t, ldvalue.OptionalString{}, c.secondary)
			assert.Len(t, c.privateAttrs, 0)
		})
	}

	t.Run("_meta", func(t *testing.T) {
		b := makeKindBuilder()
		assert.False(t, b.TrySetValue("_meta", ldvalue.String("hi")))
		assert.Equal(t, 0, b.Build().attributes.Count())
	})
}

func TestKindBuilderAttributesCopyOnWrite(t *testing.T) {
	value1, value2 := ldvalue.String("value1"), ldvalue.String("value2")

	b := makeKindBuilder().SetValue("attr", value1)

	c1 := b.Build()
	jsonhelpers.AssertEqual(t, value1, c1.attributes.Get("attr"))

	b.SetValue("attr", value2)

	c2 := b.Build()
	jsonhelpers.AssertEqual(t, value2, c2.attributes.Get("attr"))
	jsonhelpers.AssertEqual(t, value1, c1.attributes.Get("attr")) // unchanged
}

func TestKindBuilderPrivate(t *testing.T) {
	expectPrivateRefsToBe := func(t *testing.T, c Context, expectedRefs ...ldattr.Ref) {
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

	t.Run("using Refs", func(t *testing.T) {
		attrRef1, attrRef2, attrRef3 := ldattr.NewRef("a"), ldattr.NewRef("/b/c"), ldattr.NewRef("d")
		c := makeKindBuilder().
			PrivateRef(attrRef1, attrRef2).PrivateRef(attrRef3).
			Build()

		expectPrivateRefsToBe(t, c, attrRef1, attrRef2, attrRef3)
	})

	t.Run("using strings", func(t *testing.T) {
		s1, s2, s3 := "a", "/b/c", "d"
		b0 := makeKindBuilder().
			PrivateRef(ldattr.NewRef(s1), ldattr.NewRef(s2)).PrivateRef(ldattr.NewRef(s3))
		b1 := makeKindBuilder().
			Private(s1, s2, s3)
		assert.Equal(t, b0, b1)
	})

	t.Run("RemovePrivate", func(t *testing.T) {
		b := makeKindBuilder().Private("a", "/b/c", "d", "/b/c")
		b.RemovePrivate("/b/c")
		c := b.Build()

		expectPrivateRefsToBe(t, c, ldattr.NewRef("a"), ldattr.NewRef("d"))
	})

	t.Run("RemovePrivateRef", func(t *testing.T) {
		b := makeKindBuilder().Private("a", "/b/c", "d", "/b/c")
		b.RemovePrivateRef(ldattr.NewRef("/b/c"))
		c := b.Build()

		expectPrivateRefsToBe(t, c, ldattr.NewRef("a"), ldattr.NewRef("d"))
	})

	t.Run("copy on write", func(t *testing.T) {
		b0 := makeKindBuilder().Private("a")

		c0 := b0.Build()
		expectPrivateRefsToBe(t, c0, ldattr.NewRef("a"))

		b0.Private("b")
		c1 := b0.Build()
		expectPrivateRefsToBe(t, c1, ldattr.NewRef("a"), ldattr.NewRef("b"))
		expectPrivateRefsToBe(t, c0, ldattr.NewRef("a")) // unchanged

		b0.RemovePrivateRef(ldattr.NewRef("a"))
		c2 := b0.Build()
		expectPrivateRefsToBe(t, c2, ldattr.NewRef("b"))
		expectPrivateRefsToBe(t, c1, ldattr.NewRef("a"), ldattr.NewRef("b")) // unchanged
		expectPrivateRefsToBe(t, c0, ldattr.NewRef("a"))                     // unchanged
	})
}

// MULTI KIND TESTS (adapted from builder_multi_test.go)

// NEW TESTS (specific to ContextBuilder)
