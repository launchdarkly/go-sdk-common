package lduser

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

func TestUserBuilderSetsOnlyKeyByDefault(t *testing.T) {
	user := NewUserBuilder("some-key").Build()

	assert.Equal(t, "some-key", user.GetKey())
	assert.Equal(t, ldvalue.String("some-key"), user.GetAttribute(KeyAttribute))

	assertStringPropertiesNotSet(t, user)

	anon, found := user.GetAnonymousOptional()
	assert.False(t, anon)
	assert.False(t, found)
	assert.Equal(t, ldvalue.Null(), user.GetAttribute(AnonymousAttribute))

	assert.Nil(t, getCustomAttrs(user))
	assert.Nil(t, getPrivateAttrs(user))
	assert.False(t, user.HasPrivateAttributes())
}

func TestUserBuilderCanSetStringAttributes(t *testing.T) {
	for a, setter := range optionalStringSetters {
		t.Run(string(a), func(t *testing.T) {
			builder := NewUserBuilder("some-key")
			setter(builder, "value")
			user := builder.Build()

			assert.Equal(t, "some-key", user.GetKey())
			assert.Equal(t, ldvalue.String("some-key"), user.GetAttribute(KeyAttribute))

			assert.Equal(t, ldvalue.NewOptionalString("value"), optionalStringGetters[a](user), a)
			assert.Equal(t, ldvalue.String("value"), user.GetAttribute(a))

			for a1, _ := range optionalStringSetters {
				if a1 != a {
					assertStringAttrNotSet(t, a1, user)
				}
			}
		})
	}
}

func TestUserBuilderCanSetAnonymous(t *testing.T) {
	user0 := NewUserBuilder("some-key").Build()
	assert.False(t, user0.GetAnonymous())
	value, ok := user0.GetAnonymousOptional()
	assert.False(t, ok)
	assert.False(t, value)
	assert.Equal(t, ldvalue.Null(), user0.GetAttribute(AnonymousAttribute))

	user1 := NewUserBuilder("some-key").Anonymous(true).Build()
	assert.True(t, user1.GetAnonymous())
	value, ok = user1.GetAnonymousOptional()
	assert.True(t, ok)
	assert.True(t, value)
	assert.Equal(t, ldvalue.Bool(true), user1.GetAttribute(AnonymousAttribute))

	user2 := NewUserBuilder("some-key").Anonymous(false).Build()
	assert.False(t, user2.GetAnonymous())
	value, ok = user2.GetAnonymousOptional()
	assert.True(t, ok)
	assert.False(t, value)
	assert.Equal(t, ldvalue.Bool(false), user2.GetAttribute(AnonymousAttribute))
}

func TestUserBuilderCanSetPrivateStringAttributes(t *testing.T) {
	for a, setter := range optionalStringSetters {
		t.Run(string(a), func(t *testing.T) {
			builder := NewUserBuilder("some-key")
			setter(builder, "value").AsPrivateAttribute()
			user := builder.Build()

			assert.Equal(t, "some-key", user.GetKey())

			assert.Equal(t, ldvalue.NewOptionalString("value"), optionalStringGetters[a](user))
			assert.Equal(t, ldvalue.String("value"), user.GetAttribute(a))

			for a1, _ := range optionalStringSetters {
				if a1 != a {
					assertStringAttrNotSet(t, a1, user)
				}
			}

			assert.Nil(t, getCustomAttrs(user))
			assert.Equal(t, []string{string(a)}, getPrivateAttrs(user))
			assert.True(t, user.HasPrivateAttributes())
		})
	}
}

func TestUserBuilderCanMakeAttributeNonPrivate(t *testing.T) {
	builder := NewUserBuilder("some-key")
	builder.Country("us").AsNonPrivateAttribute()
	builder.Email("e").AsPrivateAttribute()
	builder.Name("n").AsPrivateAttribute()
	builder.Email("f").AsNonPrivateAttribute()
	user := builder.Build()
	assert.Equal(t, "f", user.GetEmail().StringValue())
	assert.Equal(t, []string{"name"}, getPrivateAttrs(user))
	assert.True(t, user.HasPrivateAttributes())
}

func TestUserBuilderCanSetCustomAttributes(t *testing.T) {
	user := NewUserBuilder("some-key").Custom("first", ldvalue.Int(1)).Custom("second", ldvalue.String("two")).Build()

	value, ok := user.GetCustom("first")
	assert.True(t, ok)
	assert.Equal(t, 1, value.IntValue())

	value, ok = user.GetCustom("second")
	assert.True(t, ok)
	assert.Equal(t, "two", value.StringValue())

	value, ok = user.GetCustom("no")
	assert.False(t, ok)
	assert.Equal(t, ldvalue.Null(), value)

	valueMap := ldvalue.ValueMapBuild().Set("first", ldvalue.Int(1)).Set("second", ldvalue.String("two")).Build()
	assert.Equal(t, valueMap, user.GetAllCustomMap())
	assert.Equal(t, valueMap.AsValue(), user.GetAllCustom())

	assert.Equal(t, []string{"first", "second"}, getCustomAttrs(user))

	assert.Nil(t, getPrivateAttrs(user))
}

func TestUserBuilderCanSetCustomAttributesAsMap(t *testing.T) {
	valueMap := ldvalue.ValueMapBuild().Set("first", ldvalue.Int(1)).Set("second", ldvalue.String("two")).Build()
	user := NewUserBuilder("some-key").CustomAll(valueMap).Build()

	assert.Equal(t, valueMap, user.GetAllCustomMap())
	assert.Equal(t, valueMap.AsValue(), user.GetAllCustom())

	assert.Equal(t, []string{"first", "second"}, getCustomAttrs(user))

	assert.Nil(t, getPrivateAttrs(user))
}

func TestUserBuilderCustomAllReplacesAllCustomAttributes(t *testing.T) {
	valueMap := ldvalue.ValueMapBuild().Set("second", ldvalue.String("two")).Build()
	user1 := NewUserBuilder("some-key").Custom("first", ldvalue.Int(1)).
		CustomAll(valueMap).Build()

	assert.Equal(t, valueMap, user1.GetAllCustomMap())

	user2 := NewUserBuilder("some-key").Custom("first", ldvalue.Int(1)).
		CustomAll(ldvalue.ValueMap{}).Build()

	assert.Equal(t, ldvalue.ValueMap{}, user2.GetAllCustomMap())
}

func TestUserBuilderCanSetAttributesAfterSettingAttributeThatCanBePrivate(t *testing.T) {
	// This tests that chaining methods off of UserBuilderCanMakeAttributePrivate works correctly.
	builder := NewUserBuilder("some-key").Name("original-name")
	builder.Key("new-key")
	builder.Anonymous(true)
	builder.Custom("thing", ldvalue.String("custom-value"))
	user := builder.Build()

	assert.Equal(t, "new-key", user.GetKey())
	assert.Equal(t, ldvalue.Bool(true), user.GetAttribute(AnonymousAttribute))
	assert.Equal(t, ldvalue.String("custom-value"), user.GetAttribute(UserAttribute("thing")))

	for a, setter := range optionalStringSetters {
		t.Run(string(a), func(t *testing.T) {
			builder := NewUserBuilder("some-key").Name("original-name")
			// builder is now a UserBuilderCanMakeAttributePrivate
			setter(builder, "value")
			user := builder.Build()

			assert.Equal(t, ldvalue.NewOptionalString("value"), optionalStringGetters[a](user))
			assert.Equal(t, ldvalue.String("value"), user.GetAttribute(a))
		})
	}
}

func TestUserBuilderGenericSetAttribute(t *testing.T) {
	t.Run("key", func(t *testing.T) {
		builder := NewUserBuilder("some-key")
		value := "value"

		builder.SetAttribute(KeyAttribute, ldvalue.String(value))
		assert.Equal(t, value, builder.Build().GetKey())

		builder.SetAttribute(KeyAttribute, ldvalue.Null())
		assert.Equal(t, value, builder.Build().GetKey())

		builder.SetAttribute(KeyAttribute, ldvalue.Bool(true))
		assert.Equal(t, value, builder.Build().GetKey())

		builder.SetAttribute(KeyAttribute, ldvalue.Int(1))
		assert.Equal(t, value, builder.Build().GetKey())

		builder.SetAttribute(KeyAttribute, ldvalue.String(value)).AsPrivateAttribute()
		assert.Len(t, getPrivateAttrs(builder.Build()), 0)
	})

	for a, getter := range optionalStringGetters {
		t.Run(string(a), func(t *testing.T) {
			builder := NewUserBuilder("some-key")
			valueStr := ldvalue.NewOptionalString("value")
			value := valueStr.AsValue()

			builder.SetAttribute(a, value)
			assert.Equal(t, valueStr, getter(builder.Build()))

			for a1, _ := range optionalStringGetters {
				if a1 != a {
					assertStringAttrNotSet(t, a1, builder.Build())
				}
			}

			assert.Len(t, getPrivateAttrs(builder.Build()), 0)

			builder.SetAttribute(a, ldvalue.Bool(true))
			assert.Equal(t, valueStr, getter(builder.Build()))

			builder.SetAttribute(a, ldvalue.Int(1))
			assert.Equal(t, valueStr, getter(builder.Build()))

			builder.SetAttribute(a, ldvalue.Null())
			assert.Equal(t, ldvalue.OptionalString{}, getter(builder.Build()))

			builder.SetAttribute(a, value).AsPrivateAttribute()

			assert.Equal(t, []string{string(a)}, getPrivateAttrs(builder.Build()))
		})
	}

	t.Run("anonymous", func(t *testing.T) {
		builder := NewUserBuilder("some-key")

		builder.SetAttribute(AnonymousAttribute, ldvalue.Bool(false))
		value, ok := builder.Build().GetAnonymousOptional()
		assert.True(t, ok)
		assert.False(t, value)
		assert.False(t, builder.Build().GetAnonymous())
		builder.SetAttribute(AnonymousAttribute, ldvalue.Int(1))
		assert.False(t, builder.Build().GetAnonymous())

		builder.SetAttribute(AnonymousAttribute, ldvalue.Bool(true))
		value, ok = builder.Build().GetAnonymousOptional()
		assert.True(t, ok)
		assert.True(t, value)
		assert.True(t, builder.Build().GetAnonymous())
		builder.SetAttribute(AnonymousAttribute, ldvalue.Int(1))
		assert.True(t, builder.Build().GetAnonymous())

		builder.SetAttribute(AnonymousAttribute, ldvalue.Null())
		assert.False(t, builder.Build().GetAnonymous())
		value, ok = builder.Build().GetAnonymousOptional()
		assert.False(t, ok)
		assert.False(t, value)

		builder.SetAttribute(AnonymousAttribute, ldvalue.Bool(true)).AsPrivateAttribute()
		assert.Len(t, getPrivateAttrs(builder.Build()), 0)
	})

	t.Run("custom", func(t *testing.T) {
		builder := NewUserBuilder("some-key")
		a := UserAttribute("thing")
		value := ldvalue.Int(2)

		builder.SetAttribute(a, value)
		assert.Equal(t, value, builder.Build().GetAttribute(a))

		assert.Len(t, getPrivateAttrs(builder.Build()), 0)

		builder.SetAttribute(a, ldvalue.Null())
		assert.Equal(t, ldvalue.Null(), builder.Build().GetAttribute(a))

		builder.SetAttribute(a, value).AsPrivateAttribute()
		assert.Equal(t, []string{string(a)}, getPrivateAttrs(builder.Build()))
	})
}

func TestUserBuilderCanSetPrivateCustomAttributes(t *testing.T) {
	user := NewUserBuilder("some-key").Custom("first", ldvalue.Int(1)).AsPrivateAttribute().
		Custom("second", ldvalue.String("two")).Build()

	value, ok := user.GetCustom("first")
	assert.True(t, ok)
	assert.Equal(t, 1, value.IntValue())

	value, ok = user.GetCustom("second")
	assert.True(t, ok)
	assert.Equal(t, "two", value.StringValue())

	value, ok = user.GetCustom("no")
	assert.False(t, ok)
	assert.Equal(t, ldvalue.Null(), value)

	assert.Equal(t, []string{"first", "second"}, getCustomAttrs(user))

	assert.Equal(t, []string{"first"}, getPrivateAttrs(user))
	assert.True(t, user.HasPrivateAttributes())
}

func TestUserBuilderCanCopyFromExistingUserWithOnlyKey(t *testing.T) {
	user0 := NewUser("some-key")
	user1 := NewUserBuilderFromUser(user0).Build()

	assert.Equal(t, "some-key", user1.GetKey())

	assertStringPropertiesNotSet(t, user1)
	assert.Nil(t, getCustomAttrs(user1))
	assert.Nil(t, getPrivateAttrs(user1))
	assert.False(t, user1.HasPrivateAttributes())
}

func TestUserBuilderCanCopyFromExistingUserWithAllAttributes(t *testing.T) {
	user0 := newUserBuilderWithAllPropertiesSet("some-key").Build()
	user1 := NewUserBuilderFromUser(user0).Build()
	assert.Equal(t, user0, user1)
}

func TestUserBuilderPrivateAttributesCopyOnWrite(t *testing.T) {
	user0 := NewUserBuilder("userkey").Name("n").AsPrivateAttribute().Build()
	user1 := NewUserBuilderFromUser(user0).Build()

	assert.Equal(t, map[UserAttribute]struct{}{NameAttribute: struct{}{}}, user0.privateAttributes)
	assert.Equal(t, user0.privateAttributes, user1.privateAttributes)
	user0.privateAttributes[UserAttribute("temp-test")] = struct{}{}
	assert.Equal(t, user0.privateAttributes, user1.privateAttributes, "users should have shared same private attr map")
	delete(user0.privateAttributes, UserAttribute("temp-test"))

	user2 := NewUserBuilderFromUser(user0).Email("e").AsPrivateAttribute().Build()
	assert.NotEqual(t, user0.privateAttributes, user2.privateAttributes)
	assert.Equal(t, map[UserAttribute]struct{}{NameAttribute: struct{}{}, EmailAttribute: struct{}{}}, user2.privateAttributes)
}
