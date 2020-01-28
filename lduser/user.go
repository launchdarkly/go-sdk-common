package lduser

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

type UserAttribute string

const (
	KeyAttribute          UserAttribute = "key"
	SecondaryKeyAttribute UserAttribute = "secondary"
	IPAttribute           UserAttribute = "ip"
	CountryAttribute      UserAttribute = "country"
	EmailAttribute        UserAttribute = "email"
	FirstNameAttribute    UserAttribute = "firstName"
	LastNameAttribute     UserAttribute = "lastName"
	AvatarAttribute       UserAttribute = "avatar"
	NameAttribute         UserAttribute = "name"
	AnonymousAttribute    UserAttribute = "anonymous"
)

// A User contains specific attributes of a user browsing your site. The only mandatory property is the Key,
// which must uniquely identify each user. For authenticated users, this may be a username or e-mail address.
// For anonymous users, this could be an IP address or session ID.
//
// Besides the mandatory key, User supports two kinds of optional attributes: interpreted attributes (e.g.
// IP and Country) and custom attributes.  LaunchDarkly can parse interpreted attributes and attach meaning
// to them. For example, from an IP address, LaunchDarkly can do a geo IP lookup and determine the user's
// country.
//
// Custom attributes are not parsed by LaunchDarkly. They can be used in custom rules-- for example, a custom
// attribute such as "customer_ranking" can be used to launch a feature to the top 10% of users on a site.
//
// User fields are immutable and can be accessed only via getter methods. To construct a User, use either
// a simple constructor (NewUser, NewAnonymousUser) or the builder pattern with NewUserBuilder.
type User struct {
	key               string
	secondary         ldvalue.OptionalString
	ip                ldvalue.OptionalString
	country           ldvalue.OptionalString
	email             ldvalue.OptionalString
	firstName         ldvalue.OptionalString
	lastName          ldvalue.OptionalString
	avatar            ldvalue.OptionalString
	name              ldvalue.OptionalString
	anonymous         ldvalue.Value
	custom            map[UserAttribute]ldvalue.Value
	privateAttributes map[UserAttribute]struct{}
}

// GetAttribute returns one of the user's attributes.
//
// The attribute parameter specifies which attribute to get. To get a custom attribute rather than one
// of the built-in ones identified by the UserAttribute constants, simply cast any string to the
// UserAttribute type.
//
// If no value has been set for this attribute, GetAttribute returns ldvalue.Null().
func (u User) GetAttribute(attribute UserAttribute) ldvalue.Value {
	switch attribute {
	case KeyAttribute:
		return ldvalue.String(u.key)
	case SecondaryKeyAttribute:
		return u.secondary.AsValue()
	case IPAttribute:
		return u.ip.AsValue()
	case CountryAttribute:
		return u.country.AsValue()
	case EmailAttribute:
		return u.email.AsValue()
	case FirstNameAttribute:
		return u.firstName.AsValue()
	case LastNameAttribute:
		return u.lastName.AsValue()
	case AvatarAttribute:
		return u.avatar.AsValue()
	case NameAttribute:
		return u.name.AsValue()
	case AnonymousAttribute:
		return u.anonymous
	default:
		value, _ := u.GetCustom(attribute)
		return value
	}
}

// GetKey gets the unique key of the user.
func (u User) GetKey() string {
	return u.key
}

// GetSecondaryKey returns the secondary key of the user, if any.
//
// This affects feature flag targeting
// (https://docs.launchdarkly.com/docs/targeting-users#section-targeting-rules-based-on-user-attributes)
// as follows: if you have chosen to bucket users by a specific attribute, the secondary key (if set)
// is used to further distinguish between users who are otherwise identical according to that attribute.
func (u User) GetSecondaryKey() ldvalue.OptionalString {
	return u.secondary
}

// GetIP returns the IP address attribute of the user, if any.
func (u User) GetIP() ldvalue.OptionalString {
	return u.ip
}

// GetCountry returns the country attribute of the user, if any.
func (u User) GetCountry() ldvalue.OptionalString {
	return u.country
}

// GetEmail returns the email address attribute of the user, if any.
func (u User) GetEmail() ldvalue.OptionalString {
	return u.email
}

// GetFirstName returns the first name attribute of the user, if any.
func (u User) GetFirstName() ldvalue.OptionalString {
	return u.firstName
}

// GetLastName returns the last name attribute of the user, if any.
func (u User) GetLastName() ldvalue.OptionalString {
	return u.lastName
}

// GetAvatar returns the avatar URL attribute of the user, if any.
func (u User) GetAvatar() ldvalue.OptionalString {
	return u.avatar
}

// GetName returns the full name attribute of the user, if any.
func (u User) GetName() ldvalue.OptionalString {
	return u.name
}

// GetAnonymous returns the anonymous attribute of the user.
//
// If a user is anonymous, the user key will not appear on your LaunchDarkly dashboard.
func (u User) GetAnonymous() bool {
	return u.anonymous.BoolValue()
}

// GetAnonymousOptional returns the anonymous attribute of the user, with a second value indicating
// whether that attribute was defined for the user or not.
func (u User) GetAnonymousOptional() (bool, bool) {
	return u.anonymous.BoolValue(), !u.anonymous.IsNull()
}

// GetCustom returns a custom attribute of the user by name. The boolean second return value indicates
// whether any value was set for this attribute or not.
//
// The value is returned using the ldvalue.Value type, which can contain any type supported by JSON:
// boolean, number, string, array (slice), or object (map). Use Value methods to access the value as
// the desired type, rather than casting it. If the attribute did not exist, the value will be
// ldvalue.Null() and the second return value will be false.
func (u User) GetCustom(attribute UserAttribute) (ldvalue.Value, bool) {
	value, found := u.custom[attribute]
	return value, found
}

// ForCustom iterates through all custom attributes that have been set on this user.
//
// The specified function receives each custom attribute name and the corresponding value. This avoids
// the overhead of returning the attributes as a map, since to preserve immutability the map would
// have to be copied.
func (u User) ForCustom(fn func(string, ldvalue.Value)) {
	for key, value := range u.custom {
		fn(string(key), value)
	}
}

// IsPrivateAttribute tests whether the given attribute is private for this user.
//
// The attribute name can either be a built-in attribute like NameAttribute or a custom one.
func (u User) IsPrivateAttribute(attribute UserAttribute) bool {
	_, ok := u.privateAttributes[attribute]
	return ok
}

// Equal tests whether two users have equal attributes.
//
// Regular struct equality comparison is not allowed for User because it can contain slices and
// maps. This method is faster than using reflect.DeepEqual(), and also correctly ignores
// insignificant differences in the internal representation of the attributes.
func (u User) Equal(other User) bool {
	if u.key != other.key ||
		u.secondary != other.secondary ||
		u.ip != other.ip ||
		u.country != other.country ||
		u.email != other.email ||
		u.firstName != other.firstName ||
		u.lastName != other.lastName ||
		u.avatar != other.avatar ||
		u.name != other.name ||
		!u.anonymous.Equal(other.anonymous) {
		return false
	}
	if len(u.custom) != len(other.custom) {
		return false
	}
	for k, v := range u.custom {
		v1, ok := other.custom[k]
		if !ok || !v.Equal(v1) {
			return false
		}
	}
	if len(u.privateAttributes) != len(other.privateAttributes) {
		return false
	}
	for k, _ := range u.privateAttributes {
		if _, ok := other.privateAttributes[k]; !ok {
			return false
		}
	}
	return true
}
