package lduser

import (
	"encoding/json"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

// A User contains specific attributes of a user browsing your site. The only mandatory property property is the Key,
// which must uniquely identify each user. For authenticated users, this may be a username or e-mail address. For anonymous users,
// this could be an IP address or session ID.
//
// Besides the mandatory Key, User supports two kinds of optional attributes: interpreted attributes (e.g. Ip and Country)
// and custom attributes.  LaunchDarkly can parse interpreted attributes and attach meaning to them. For example, from an IP address, LaunchDarkly can
// do a geo IP lookup and determine the user's country.
//
// Custom attributes are not parsed by LaunchDarkly. They can be used in custom rules-- for example, a custom attribute such as "customer_ranking" can be used to
// launch a feature to the top 10% of users on a site.
//
// User fields will be made private in the future, accessible only via getter methods, to prevent unsafe
// modification of users after they are created. The preferred method of constructing a User is to use either
// a simple constructor (NewUser, NewAnonymousUser) or the builder pattern with NewUserBuilder. If you do set
// the User fields directly, it is important not to change any map/slice elements, and not change a string
// that is pointed to by an existing pointer, after the User has been passed to any SDK methods; otherwise,
// flag evaluations and analytics events may refer to the wrong user properties (or, in the case of a map, you
// may even cause a concurrent modification panic).
type User struct {
	// Key is the unique key of the user.
	//
	// Deprecated: Direct access to User fields is now deprecated in favor of UserBuilder. In a future version,
	// User fields will be private and only accessible via getter methods.
	Key string `json:"key" bson:"key"`
	// SecondaryKey is the secondary key of the user.
	//
	// This affects feature flag targeting (https://docs.launchdarkly.com/docs/targeting-users#section-targeting-rules-based-on-user-attributes)
	// as follows: if you have chosen to bucket users by a specific attribute, the secondary key (if set)
	// is used to further distinguish between users who are otherwise identical according to that attribute.
	//
	// Deprecated: Direct access to User fields is now deprecated in favor of UserBuilder. In a future version,
	// User fields will be private and only accessible via getter methods.
	Secondary ldvalue.OptionalString `json:"secondary" bson:"secondary"`
	// Ip is the IP address attribute of the user.
	//
	// Deprecated: Direct access to User fields is now deprecated in favor of UserBuilder. In a future version,
	// User fields will be private and only accessible via getter methods.
	Ip ldvalue.OptionalString `json:"ip" bson:"ip"`
	// Country is the country attribute of the user.
	//
	// Deprecated: Direct access to User fields is now deprecated in favor of UserBuilder. In a future version,
	// User fields will be private and only accessible via getter methods.
	Country ldvalue.OptionalString `json:"country" bson:"country"`
	// Email is the email address attribute of the user.
	//
	// Deprecated: Direct access to User fields is now deprecated in favor of UserBuilder. In a future version,
	// User fields will be private and only accessible via getter methods.
	Email ldvalue.OptionalString `json:"email" bson:"email"`
	// FirstName is the first name attribute of the user.
	//
	// Deprecated: Direct access to User fields is now deprecated in favor of UserBuilder. In a future version,
	// User fields will be private and only accessible via getter methods.
	FirstName ldvalue.OptionalString `json:"firstName" bson:"firstName"`
	// LastName is the last name attribute of the user.
	//
	// Deprecated: Direct access to User fields is now deprecated in favor of UserBuilder. In a future version,
	// User fields will be private and only accessible via getter methods.
	LastName ldvalue.OptionalString `json:"lastName" bson:"lastName"`
	// Avatar is the avatar URL attribute of the user.
	//
	// Deprecated: Direct access to User fields is now deprecated in favor of UserBuilder. In a future version,
	// User fields will be private and only accessible via getter methods.
	Avatar ldvalue.OptionalString `json:"avatar" bson:"avatar"`
	// Name is the name attribute of the user.
	//
	// Deprecated: Direct access to User fields is now deprecated in favor of UserBuilder. In a future version,
	// User fields will be private and only accessible via getter methods.
	Name ldvalue.OptionalString `json:"name" bson:"name"`
	// Anonymous indicates whether the user is anonymous.
	//
	// If a user is anonymous, the user key will not appear on your LaunchDarkly dashboard.
	//
	// Deprecated: Direct access to User fields is now deprecated in favor of UserBuilder. In a future version,
	// User fields will be private and only accessible via getter methods.
	Anonymous ldvalue.Value `json:"anonymous,omitempty" bson:"anonymous,omitempty"`
	// Custom is the user's map of custom attribute names and values.
	//
	// Deprecated: Direct access to User fields is now deprecated in favor of UserBuilder. In a future version,
	// User fields will be private and only accessible via getter methods.
	Custom map[string]ldvalue.Value `json:"custom,omitempty" bson:"custom,omitempty"`

	// PrivateAttributes contains a list of attribute names that were included in the user,
	// but were marked as private. As such, these attributes are not included in the fields above.
	//
	// Deprecated: Direct access to User fields is now deprecated in favor of UserBuilder. In a future version,
	// User fields will be private and only accessible via getter methods.
	PrivateAttributes []string `json:"privateAttrs,omitempty" bson:"privateAttrs,omitempty"`

	// This contains list of attributes to keep private, whether they appear at the top-level or Custom
	// The attribute "key" is always sent regardless of whether it is in this list, and "custom" cannot be used to
	// eliminate all custom attributes
	//
	// Deprecated: Direct access to User fields is now deprecated in favor of UserBuilder. In a future version,
	// User fields will be private and only accessible via getter methods.
	PrivateAttributeNames []string `json:"-" bson:"-"`
}

// GetKey gets the unique key of the user.
func (u User) GetKey() string {
	return u.Key
}

// GetSecondaryKey returns the secondary key of the user, if any.
//
// This affects feature flag targeting (https://docs.launchdarkly.com/docs/targeting-users#section-targeting-rules-based-on-user-attributes)
// as follows: if you have chosen to bucket users by a specific attribute, the secondary key (if set)
// is used to further distinguish between users who are otherwise identical according to that attribute.
func (u User) GetSecondaryKey() ldvalue.OptionalString {
	return u.Secondary
}

// GetIP() returns the IP address attribute of the user, if any.
func (u User) GetIP() ldvalue.OptionalString {
	return u.Ip
}

// GetCountry() returns the country attribute of the user, if any.
func (u User) GetCountry() ldvalue.OptionalString {
	return u.Country
}

// GetEmail() returns the email address attribute of the user, if any.
func (u User) GetEmail() ldvalue.OptionalString {
	return u.Email
}

// GetFirstName() returns the first name attribute of the user, if any.
func (u User) GetFirstName() ldvalue.OptionalString {
	return u.FirstName
}

// GetLastName() returns the last name attribute of the user, if any.
func (u User) GetLastName() ldvalue.OptionalString {
	return u.LastName
}

// GetAvatar() returns the avatar URL attribute of the user, if any.
func (u User) GetAvatar() ldvalue.OptionalString {
	return u.Avatar
}

// GetName() returns the full name attribute of the user, if any.
func (u User) GetName() ldvalue.OptionalString {
	return u.Name
}

// GetAnonymous() returns the anonymous attribute of the user.
//
// If a user is anonymous, the user key will not appear on your LaunchDarkly dashboard.
func (u User) GetAnonymous() bool {
	return u.Anonymous.BoolValue()
}

// GetAnonymousOptional() returns the anonymous attribute of the user, with a second value indicating
// whether that attribute was defined for the user or not.
func (u User) GetAnonymousOptional() (bool, bool) {
	return u.Anonymous.BoolValue(), !u.Anonymous.IsNull()
}

// GetCustom() returns a custom attribute of the user by name. The boolean second return value indicates
// whether any value was set for this attribute or not.
//
// The value is returned using the ldvalue.Value type, which can contain any type supported by JSON:
// boolean, number, string, array (slice), or object (map). Use Value methods to access the value as
// the desired type, rather than casting it. If the attribute did not exist, the value will be
// ldvalue.Null() and the second return value will be false.
func (u User) GetCustom(attrName string) (ldvalue.Value, bool) {
	value, found := u.Custom[attrName]
	return value, found
}

// GetCustomKeys() returns the keys of all custom attributes that have been set on this user.
func (u User) GetCustomKeys() []string {
	if len(u.Custom) == 0 {
		return nil
	}
	keys := make([]string, 0, len(u.Custom))
	for key := range u.Custom {
		keys = append(keys, key)
	}
	return keys
}

// Equal tests whether two users have equal attributes.
//
// Regular struct equality comparison is not allowed for User because it can contain slices and
// maps. This method is faster than using reflect.DeepEqual(), and also correctly ignores
// insignificant differences in the internal representation of the attributes.
func (u User) Equal(other User) bool {
	if u.Key != other.Key ||
		u.Secondary != other.Secondary ||
		u.Ip != other.Ip ||
		u.Country != other.Country ||
		u.Email != other.Email ||
		u.FirstName != other.FirstName ||
		u.LastName != other.LastName ||
		u.Avatar != other.Avatar ||
		u.Name != other.Name ||
		!u.Anonymous.Equal(other.Anonymous) {
		return false
	}
	if len(u.Custom) != len(other.Custom) {
		return false
	}
	for k, v := range u.Custom {
		v1, ok := other.Custom[k]
		if !ok || !v.Equal(v1) {
			return false
		}
	}
	if !stringSlicesEqual(u.PrivateAttributeNames, other.PrivateAttributeNames) {
		return false
	}
	if !stringSlicesEqual(u.PrivateAttributes, other.PrivateAttributes) {
		return false
	}
	return true
}

// String returns a simple string representation of a user.
func (u User) String() string {
	bytes, _ := json.Marshal(u)
	return string(bytes)
}

// Used internally in evaluations. The second return value is true if the attribute exists for this user,
// false if not.
func (u User) valueOf(attr string) (ldvalue.Value, bool) {
	if attr == "key" {
		return ldvalue.String(u.Key), true
	} else if attr == "ip" {
		return u.Ip.AsValue(), u.Ip.IsDefined()
	} else if attr == "country" {
		return u.Country.AsValue(), u.Country.IsDefined()
	} else if attr == "email" {
		return u.Email.AsValue(), u.Email.IsDefined()
	} else if attr == "firstName" {
		return u.FirstName.AsValue(), u.FirstName.IsDefined()
	} else if attr == "lastName" {
		return u.LastName.AsValue(), u.LastName.IsDefined()
	} else if attr == "avatar" {
		return u.Avatar.AsValue(), u.Avatar.IsDefined()
	} else if attr == "name" {
		return u.Name.AsValue(), u.Name.IsDefined()
	} else if attr == "anonymous" {
		return u.Anonymous, !u.Anonymous.IsNull()
	}

	// Select a custom attribute
	return u.GetCustom(attr)
}

func stringSlicesEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, n0 := range a {
		ok := false
		for _, n1 := range b {
			if n1 == n0 {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}
