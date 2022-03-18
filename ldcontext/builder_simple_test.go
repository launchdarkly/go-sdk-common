package ldcontext

import (
	"fmt"
	"testing"

	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

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
	var boolFalse, boolTrue, stringEmpty, stringNonEmpty = ldvalue.Bool(false), ldvalue.Bool(true),
		ldvalue.String("x"), ldvalue.String("")
	var nullValue, intValue, floatValue, arrayValue, objectValue = ldvalue.Null(),
		ldvalue.Int(1), ldvalue.Float64(1.5), ldvalue.ArrayOf(), ldvalue.ObjectBuild().Build()

	type params struct {
		name             string
		equivalentSetter func(*Builder, ldvalue.Value)
		good, bad        []ldvalue.Value
	}

	for _, p := range []params{
		{
			name:             "kind",
			equivalentSetter: func(b *Builder, v ldvalue.Value) { b.Kind(Kind(v.StringValue())) },
			good:             []ldvalue.Value{stringNonEmpty, stringEmpty},
			bad:              []ldvalue.Value{nullValue, boolFalse, intValue, floatValue, arrayValue, objectValue},
		},
		{
			name:             "key",
			equivalentSetter: func(b *Builder, v ldvalue.Value) { b.Key(v.StringValue()) },
			good:             []ldvalue.Value{stringNonEmpty, stringEmpty},
			bad:              []ldvalue.Value{nullValue, boolFalse, intValue, floatValue, arrayValue, objectValue},
		},
		{
			name:             "name",
			equivalentSetter: func(b *Builder, v ldvalue.Value) { b.OptName(v.AsOptionalString()) },
			good:             []ldvalue.Value{stringNonEmpty, stringEmpty, nullValue},
			bad:              []ldvalue.Value{boolFalse, intValue, floatValue, arrayValue, objectValue},
		},
		{
			name:             "transient",
			equivalentSetter: func(b *Builder, v ldvalue.Value) { b.Transient(v.BoolValue()) },
			good:             []ldvalue.Value{boolTrue, boolFalse},
			bad:              []ldvalue.Value{nullValue, intValue, floatValue, stringEmpty, stringNonEmpty, arrayValue, objectValue},
		},
	} {
		t.Run(p.name, func(t *testing.T) {
			builder := makeBasicBuilder() // we will reuse this to prove that SetValue overwrites previous values
			var lastGoodNonNullValue ldvalue.Value

			for _, goodValue := range p.good {
				t.Run(fmt.Sprintf("can set to %s", goodValue.JSONString()), func(t *testing.T) {
					previousState := *builder

					if !goodValue.IsNull() {
						lastGoodNonNullValue = goodValue
					}
					expected := makeBasicBuilder()
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
					startingState := func() *Builder {
						if lastGoodNonNullValue.IsDefined() {
							return makeBasicBuilder().SetValue(p.name, lastGoodNonNullValue)
						}
						return makeBasicBuilder()
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
						assert.Equal(t, startingState(), makeBasicBuilder().SetString(p.name, badValue.StringValue()))
					}
				})
			}
		})
	}
}

func TestBuilderSetValueCannotSetMetaProperties(t *testing.T) {
	t.Run("privateAttributes", func(t *testing.T) {
		assert.Equal(t,
			makeBasicBuilder(),
			makeBasicBuilder().SetValue("privateAttributes", ldvalue.ArrayOf(ldvalue.String("x"))))
	})

	t.Run("secondary", func(t *testing.T) {
		assert.Equal(t,
			makeBasicBuilder(),
			makeBasicBuilder().SetValue("secondary", ldvalue.String("x")))
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
		c := makeBasicBuilder().
			PrivateRef(attrRef1, attrRef2).PrivateRef(attrRef3).
			Build()

		expectPrivateRefsToBe(t, c, attrRef1, attrRef2, attrRef3)
	})

	t.Run("using strings", func(t *testing.T) {
		s1, s2, s3 := "a", "/b/c", "d"
		b0 := makeBasicBuilder().
			PrivateRef(ldattr.NewRef(s1), ldattr.NewRef(s2)).PrivateRef(ldattr.NewRef(s3))
		b1 := makeBasicBuilder().
			Private(s1, s2, s3)
		assert.Equal(t, b0, b1)
	})

	t.Run("RemovePrivate", func(t *testing.T) {
		b := makeBasicBuilder().Private("a", "/b/c", "d", "/b/c")
		b.RemovePrivate("/b/c")
		c := b.Build()

		expectPrivateRefsToBe(t, c, ldattr.NewRef("a"), ldattr.NewRef("d"))
	})

	t.Run("RemovePrivateRef", func(t *testing.T) {
		b := makeBasicBuilder().Private("a", "/b/c", "d", "/b/c")
		b.RemovePrivateRef(ldattr.NewRef("/b/c"))
		c := b.Build()

		expectPrivateRefsToBe(t, c, ldattr.NewRef("a"), ldattr.NewRef("d"))
	})

	t.Run("copy on write", func(t *testing.T) {
		b0 := makeBasicBuilder().Private("a")

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
