package lduser

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

var allBuiltInAttributes = []UserAttribute{
	KeyAttribute,
	SecondaryKeyAttribute,
	IPAttribute,
	CountryAttribute,
	EmailAttribute,
	FirstNameAttribute,
	LastNameAttribute,
	AvatarAttribute,
	NameAttribute,
	AnonymousAttribute,
}

var optionalStringGetters = map[UserAttribute]func(User) ldvalue.OptionalString{
	SecondaryKeyAttribute: User.GetSecondaryKey,
	IPAttribute:           User.GetIP,
	CountryAttribute:      User.GetCountry,
	EmailAttribute:        User.GetEmail,
	FirstNameAttribute:    User.GetFirstName,
	LastNameAttribute:     User.GetLastName,
	AvatarAttribute:       User.GetAvatar,
	NameAttribute:         User.GetName,
}

var optionalStringSetters = map[UserAttribute]func(UserBuilder, string) UserBuilderCanMakeAttributePrivate{
	SecondaryKeyAttribute: UserBuilder.Secondary,
	IPAttribute:           UserBuilder.IP,
	CountryAttribute:      UserBuilder.Country,
	EmailAttribute:        UserBuilder.Email,
	FirstNameAttribute:    UserBuilder.FirstName,
	LastNameAttribute:     UserBuilder.LastName,
	AvatarAttribute:       UserBuilder.Avatar,
	NameAttribute:         UserBuilder.Name,
}

func assertStringAttrNotSet(t *testing.T, a UserAttribute, user User) {
	assert.Equal(t, ldvalue.OptionalString{}, optionalStringGetters[a](user), "should not have had a value for %s", a)
	assert.Equal(t, ldvalue.Null(), user.GetAttribute(a), "should not have had a value for %s", a)
}

func assertStringPropertiesNotSet(t *testing.T, user User) {
	for a, _ := range optionalStringGetters {
		assertStringAttrNotSet(t, a, user)
	}
}

func getCustomAttrs(user User) []string {
	var ret []string
	user.GetAllCustom().Enumerate(func(i int, a string, v ldvalue.Value) bool {
		ret = append(ret, a)
		return true
	})
	sort.Strings(ret)
	return ret
}

func getPrivateAttrs(user User) []string {
	var ret []string
	for _, a := range allBuiltInAttributes {
		if user.IsPrivateAttribute(a) {
			ret = append(ret, string(a))
		}
	}
	user.GetAllCustom().Enumerate(func(i int, a string, v ldvalue.Value) bool {
		if user.IsPrivateAttribute(UserAttribute(a)) {
			ret = append(ret, a)
		}
		return true
	})
	sort.Strings(ret)
	return ret
}

func TestNewUser(t *testing.T) {
	user := NewUser("some-key")

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

func TestNewAnonymousUser(t *testing.T) {
	user := NewAnonymousUser("some-key")

	assert.Equal(t, "some-key", user.GetKey())
	assert.Equal(t, ldvalue.String("some-key"), user.GetAttribute(KeyAttribute))

	assertStringPropertiesNotSet(t, user)

	anon, found := user.GetAnonymousOptional()
	assert.True(t, anon)
	assert.True(t, found)
	assert.Equal(t, ldvalue.Bool(true), user.GetAttribute(AnonymousAttribute))

	assert.Nil(t, getCustomAttrs(user))
	assert.Nil(t, getPrivateAttrs(user))
	assert.False(t, user.HasPrivateAttributes())
}

func TestGetUnknownAttribute(t *testing.T) {
	assert.Equal(t, ldvalue.Null(), NewUser("some-key").GetAttribute(UserAttribute("no-such-thing")))
}

func TestEnumerateCustomAttributes(t *testing.T) {
	user := NewUserBuilder("some-key").Custom("first", ldvalue.Int(1)).Custom("second", ldvalue.String("two")).Build()
	m := make(map[string]ldvalue.Value)
	user.GetAllCustom().Enumerate(func(i int, a string, v ldvalue.Value) bool {
		m[a] = v
		return true
	})
	assert.Equal(t, map[string]ldvalue.Value{"first": ldvalue.Int(1), "second": ldvalue.String("two")}, m)
}

func TestUserWithNoCustomAttributes(t *testing.T) {
	user := NewUser("some-key")

	assert.Nil(t, getCustomAttrs(user))

	value, ok := user.GetCustom("attr")
	assert.False(t, ok)
	assert.Equal(t, ldvalue.Null(), value)

	assert.Nil(t, getCustomAttrs(user))
}

func TestUserEqualsComparesAllAttributes(t *testing.T) {
	shouldEqual := func(a User, b User) {
		assert.True(t, b.Equal(a), "%s should equal %s", b, a)
		assert.True(t, a.Equal(b), "%s should equal %s", a, b)
	}
	shouldNotEqual := func(a User, b User) {
		assert.False(t, b.Equal(a), "%s should not equal %s", b, a)
		assert.False(t, a.Equal(b), "%s should not equal %s", a, b)
	}

	user0 := NewUser("some-key")
	assert.True(t, user0.Equal(user0), "%s should equal itself", user0)

	user1 := newUserBuilderWithAllPropertiesSet("some-key").Build()
	assert.True(t, user1.Equal(user1), "%s should equal itself", user1)
	user2 := NewUserBuilderFromUser(user1).Build()
	shouldEqual(user1, user2)

	for a, setter := range optionalStringSetters {
		builder3 := NewUserBuilderFromUser(user1)
		setter(builder3, "different-value")
		user3 := builder3.Build()
		shouldNotEqual(user1, user3)

		builder4 := NewUserBuilderFromUser(user1)
		setter(builder4, fmt.Sprintf("value-%s", a)).AsPrivateAttribute()
		user4 := builder4.Build()
		shouldNotEqual(user1, user4)
	}

	shouldNotEqual(user1, NewUserBuilderFromUser(user1).Key("other-key").Build())

	shouldNotEqual(user0, NewUserBuilderFromUser(user0).Anonymous(true).Build())
	shouldNotEqual(NewUserBuilderFromUser(user0).Anonymous(true).Build(), NewUserBuilderFromUser(user0).Anonymous(false).Build())

	// modifying an existing custom attribute
	shouldNotEqual(user1, NewUserBuilderFromUser(user1).Custom("thing1", ldvalue.String("other-value")).Build())

	// adding a new custom attribute
	shouldNotEqual(user1, NewUserBuilderFromUser(user1).Custom("thing3", ldvalue.String("other-value")).Build())

	// adding an extra private attribute
	shouldNotEqual(user1, NewUserBuilderFromUser(user1).Custom("thing1", ldvalue.String("value1")).AsPrivateAttribute().Build())

	// having the same number of private attributes, but not the same ones
	shouldNotEqual(
		NewUserBuilderFromUser(user1).
			Custom("thing3", ldvalue.Bool(true)).AsPrivateAttribute().
			Custom("thing4", ldvalue.Bool(true)).
			Build(),
		NewUserBuilderFromUser(user1).
			Custom("thing3", ldvalue.Bool(true)).
			Custom("thing4", ldvalue.Bool(true)).AsPrivateAttribute().
			Build())
}

func newUserBuilderWithAllPropertiesSet(key string) UserBuilder {
	builder := NewUserBuilder(key)
	for a, setter := range optionalStringSetters {
		setter(builder, fmt.Sprintf("value-%s", a))
	}
	builder.Anonymous(true)
	builder.Custom("thing1", ldvalue.String("value1"))
	builder.Custom("thing2", ldvalue.String("value2")).AsPrivateAttribute()
	return builder
}
