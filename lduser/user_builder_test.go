package lduser

import (
	"testing"

	"github.com/launchdarkly/go-sdk-common/v3/ldcontext"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

	"github.com/stretchr/testify/assert"
)

func TestConstructors(t *testing.T) {
	assert.Equal(t, ldcontext.New("some-key"), NewUser("some-key"))
	assert.Equal(t, ldcontext.NewBuilder("some-key").Anonymous(true).Build(), NewAnonymousUser("some-key"))
}

func TestUserBuilderSetsOnlyKeyByDefault(t *testing.T) {
	c := NewUserBuilder("some-key").Build()

	assert.Equal(t, ldcontext.Kind("user"), c.Kind())
	assert.Equal(t, "some-key", c.Key())
	assert.False(t, c.Secondary().IsDefined())
	assert.False(t, c.Anonymous())
	assert.Len(t, c.GetOptionalAttributeNames(nil), 0)
	assert.Equal(t, 0, c.PrivateAttributeCount())
}

func TestUserBuilderCanSetStringAttributes(t *testing.T) {
	for a, setter := range optionalStringSetters {
		t.Run(string(a), func(t *testing.T) {
			builder := NewUserBuilder("some-key")
			setter(builder, "value")
			c := builder.Build()

			assert.Equal(t, ldcontext.Kind("user"), c.Kind())
			assert.Equal(t, "some-key", c.Key())

			assert.Equal(t, ldvalue.NewOptionalString("value"), optionalStringGetters[a](c), a)

			for a1 := range optionalStringSetters {
				if a1 != a {
					assertStringAttrNotSet(t, a1, c)
				}
			}
		})
	}
}

func TestUserBuilderCanSetAnonymous(t *testing.T) {
	user0 := NewUserBuilder("some-key").Build()
	assert.False(t, user0.Anonymous())

	user1 := NewUserBuilder("some-key").Anonymous(true).Build()
	assert.True(t, user1.Anonymous())

	user2 := NewUserBuilder("some-key").Anonymous(false).Build()
	assert.False(t, user2.Anonymous())
}

func TestUserBuilderCanSetPrivateAttributes(t *testing.T) {
	for a, setter := range optionalStringSetters {
		t.Run(string(a), func(t *testing.T) {
			builder := NewUserBuilder("some-key")
			setter(builder, "value").AsPrivateAttribute()
			c := builder.Build()

			assert.Equal(t, "some-key", c.Key())

			assert.Equal(t, ldvalue.NewOptionalString("value"), optionalStringGetters[a](c))
			value := c.GetValue(string(a))
			if a == SecondaryKeyAttribute {
				// "secondary" is no longer addressable as an attribute in evaluations
				assert.False(t, value.IsDefined())
			} else {
				assert.Equal(t, ldvalue.String("value"), value)
			}

			for a1 := range optionalStringSetters {
				if a1 != a {
					assertStringAttrNotSet(t, a1, c)
				}
			}

			if string(a) == "secondary" {
				assert.Len(t, sortedOptionalAttributes(c), 0)
			} else {
				assert.Equal(t, []string{string(a)}, sortedOptionalAttributes(c))
			}
			assert.Equal(t, []string{string(a)}, sortedPrivateAttributes(c))
		})
	}

	t.Run("custom", func(t *testing.T) {
		builder := NewUserBuilder("some-key")
		builder.Custom("my-attr", ldvalue.String("value")).AsPrivateAttribute()
		c := builder.Build()

		value := c.GetValue("my-attr")
		assert.Equal(t, ldvalue.String("value"), value)

		assert.Equal(t, []string{"my-attr"}, sortedPrivateAttributes(c))
	})

	t.Run("custom with leading slash", func(t *testing.T) {
		builder := NewUserBuilder("some-key")
		builder.Custom("/my-attr", ldvalue.String("value")).AsPrivateAttribute()
		c := builder.Build()

		value := c.GetValue("/my-attr")
		assert.Equal(t, ldvalue.String("value"), value)

		assert.Equal(t, []string{"/~1my-attr"}, sortedPrivateAttributes(c))
	})
}

func TestUserBuilderCanMakeAttributeNonPrivate(t *testing.T) {
	builder := NewUserBuilder("some-key")
	builder.Country("us").AsNonPrivateAttribute()
	builder.Email("e").AsPrivateAttribute()
	builder.Name("n").AsPrivateAttribute()
	builder.Email("f").AsNonPrivateAttribute()

	c := builder.Build()

	value := c.GetValue("email")
	assert.Equal(t, ldvalue.String("f"), value)

	assert.Equal(t, []string{"name"}, sortedPrivateAttributes(c))
}

func TestUserBuilderCanSetCustomAttributes(t *testing.T) {
	c := NewUserBuilder("some-key").Custom("first", ldvalue.Int(1)).Custom("second", ldvalue.String("two")).Build()

	value := c.GetValue("first")
	assert.Equal(t, 1, value.IntValue())

	value = c.GetValue("second")
	assert.Equal(t, "two", value.StringValue())

	value = c.GetValue("no")
	assert.Equal(t, ldvalue.Null(), value)

	assert.Equal(t, []string{"first", "second"}, sortedOptionalAttributes(c))
	assert.Len(t, sortedPrivateAttributes(c), 0)
}

func TestUserBuilderCanSetCustomAttributesAsMap(t *testing.T) {
	valueMap := ldvalue.ValueMapBuild().Set("first", ldvalue.Int(1)).Set("second", ldvalue.String("two")).Build()
	c := NewUserBuilder("some-key").CustomAll(valueMap).Build()

	value := c.GetValue("first")
	assert.Equal(t, ldvalue.Int(1), value)

	value = c.GetValue("second")
	assert.Equal(t, ldvalue.String("two"), value)

	assert.Equal(t, []string{"first", "second"}, sortedOptionalAttributes(c))
}

func TestUserBuilderCustomAllReplacesAllCustomAttributes(t *testing.T) {
	valueMap := ldvalue.ValueMapBuild().Set("second", ldvalue.String("two")).Build()
	c1 := NewUserBuilder("some-key").Email("my-email").Custom("first", ldvalue.Int(1)).
		CustomAll(valueMap).Build()

	assert.Equal(t, []string{"email", "second"}, sortedOptionalAttributes(c1))

	c2 := NewUserBuilder("some-key").Email("my-email").Custom("first", ldvalue.Int(1)).
		CustomAll(ldvalue.ValueMap{}).Build()

	assert.Equal(t, []string{"email"}, sortedOptionalAttributes(c2))
}

func TestUserBuilderCanSetAttributesAfterSettingAttributeThatCanBePrivate(t *testing.T) {
	// This tests that chaining methods off of UserBuilderCanMakeAttributePrivate works correctly.
	builder := NewUserBuilder("some-key").Name("original-name").Key("new-key")
	c := builder.Build()

	assert.Equal(t, "new-key", c.Key())
}

func TestUserBuilderGenericSetAttribute(t *testing.T) {
	t.Run("key", func(t *testing.T) {
		builder := NewUserBuilder("some-key")
		value := "value"

		builder.SetAttribute(KeyAttribute, ldvalue.String(value))
		assert.Equal(t, value, builder.Build().Key())

		// setting key to wrong type is a no-op
		builder.SetAttribute(KeyAttribute, ldvalue.Null())
		assert.Equal(t, value, builder.Build().Key())
		builder.SetAttribute(KeyAttribute, ldvalue.Bool(true))
		assert.Equal(t, value, builder.Build().Key())

		builder.SetAttribute(KeyAttribute, ldvalue.String(value)).AsPrivateAttribute()
		assert.Len(t, sortedPrivateAttributes(builder.Build()), 0)
	})

	for a, getter := range optionalStringGetters {
		t.Run(string(a), func(t *testing.T) {
			builder := NewUserBuilder("some-key")
			valueStr := ldvalue.NewOptionalString("value")
			value := valueStr.AsValue()

			builder.SetAttribute(a, value)
			assert.Equal(t, valueStr, getter(builder.Build()))

			// setting optional string attribute to wrong type is a no-op
			builder.SetAttribute(a, ldvalue.Bool(true))
			assert.Equal(t, valueStr, getter(builder.Build()))

			for a1 := range optionalStringGetters {
				if a1 != a {
					assertStringAttrNotSet(t, a1, builder.Build())
				}
			}

			assert.Len(t, sortedPrivateAttributes(builder.Build()), 0)

			builder.SetAttribute(a, ldvalue.Null())
			assert.Equal(t, ldvalue.OptionalString{}, getter(builder.Build()))

			builder.SetAttribute(a, value).AsPrivateAttribute()
			assert.Equal(t, []string{string(a)}, sortedPrivateAttributes(builder.Build()))
		})
	}

	t.Run("anonymous", func(t *testing.T) {
		builder := NewUserBuilder("some-key")

		builder.SetAttribute(AnonymousAttribute, ldvalue.Bool(false))
		assert.False(t, builder.Build().Anonymous())

		builder.SetAttribute(AnonymousAttribute, ldvalue.Bool(true))
		assert.True(t, builder.Build().Anonymous())

		// setting anonymous to wrong type is a no-op
		builder.SetAttribute(AnonymousAttribute, ldvalue.String("x"))
		assert.True(t, builder.Build().Anonymous())

		builder.SetAttribute(AnonymousAttribute, ldvalue.Null())
		assert.False(t, builder.Build().Anonymous())
	})

	t.Run("custom", func(t *testing.T) {
		builder := NewUserBuilder("some-key")
		name := "thing"
		value := ldvalue.Int(2)

		builder.SetAttribute(UserAttribute(name), value)
		c0 := builder.Build()
		v := c0.GetValue(name)
		assert.Equal(t, value, v)
		assert.Equal(t, []string{name}, sortedOptionalAttributes(c0))
		assert.Len(t, sortedPrivateAttributes(c0), 0)

		builder.SetAttribute(UserAttribute(name), ldvalue.Null())
		c1 := builder.Build()
		v = c1.GetValue(name)
		assert.Equal(t, ldvalue.Null(), v)
		assert.Len(t, sortedOptionalAttributes(c1), 0)
		assert.Len(t, sortedPrivateAttributes(c1), 0)

		builder.SetAttribute(UserAttribute(name), value).AsPrivateAttribute()
		c2 := builder.Build()
		assert.Equal(t, []string{name}, sortedPrivateAttributes(c2))
	})
}

func TestUserBuilderCanCopyFromExistingUserWithOnlyKey(t *testing.T) {
	user0 := NewUser("some-key")
	user1 := NewUserBuilderFromUser(user0).Build()

	assert.Equal(t, user0, user1)
}

func TestUserBuilderCanCopyFromExistingUserWithAllAttributes(t *testing.T) {
	user0 := NewUserBuilder("some-key").
		Name("name").
		FirstName("firstName").
		LastName("lastName").
		Email("email").AsPrivateAttribute().
		Country("country").
		Avatar("avatar").
		IP("ip").
		Custom("attr", ldvalue.String("value")).
		Secondary("secondary").
		Anonymous(true).
		Build()
	user1 := NewUserBuilderFromUser(user0).Build()
	assert.Equal(t, user0, user1)
}
