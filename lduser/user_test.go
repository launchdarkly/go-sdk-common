package lduser

import (
	"sort"
	"testing"

	"github.com/launchdarkly/go-sdk-common/v3/ldcontext"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

	"github.com/stretchr/testify/assert"
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

func makeStringGetter(name string) func(c ldcontext.Context) ldvalue.OptionalString {
	return func(c ldcontext.Context) ldvalue.OptionalString {
		return c.GetValue(name).AsOptionalString()
	}
}

var optionalStringGetters = map[UserAttribute]func(ldcontext.Context) ldvalue.OptionalString{
	SecondaryKeyAttribute: ldcontext.Context.Secondary, // Context doesn't consider this to be an addressable attribute
	IPAttribute:           makeStringGetter("ip"),
	CountryAttribute:      makeStringGetter("country"),
	EmailAttribute:        makeStringGetter("email"),
	FirstNameAttribute:    makeStringGetter("firstName"),
	LastNameAttribute:     makeStringGetter("lastName"),
	AvatarAttribute:       makeStringGetter("avatar"),
	NameAttribute:         makeStringGetter("name"),
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

func assertStringAttrNotSet(t *testing.T, a UserAttribute, c ldcontext.Context) {
	assert.Equal(t, ldvalue.OptionalString{}, optionalStringGetters[a](c), "should not have had a value for %s", a)
}

func sortedOptionalAttributes(c ldcontext.Context) []string {
	ret := c.GetOptionalAttributeNames(nil)
	sort.Strings(ret)
	return ret
}

func sortedPrivateAttributes(c ldcontext.Context) []string {
	ret := make([]string, 0, c.PrivateAttributeCount())
	for i := 0; i < c.PrivateAttributeCount(); i++ {
		a, _ := c.PrivateAttributeByIndex(i)
		ret = append(ret, a.String())
	}
	return ret
}
