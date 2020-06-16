package lduser

import (
	"encoding/json"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

// These temporary structs allow us to do JSON marshalling and unmarshalling as efficiently as possible
// while not requiring the User's internal representation to be constrained by the behavior of json.Marshal.
// When serializing, we will use pointers (that are only pointers to local variables, so the values won't
// escape to the heap) so unset attributes won't be serialized at all. When deserializing, we'll unmarshal
// directly to the ldvalue.Value type so no interim pointers are necessary.

type userForSerialization struct {
	Key                   string          `json:"key"`
	Secondary             *string         `json:"secondary,omitempty"`
	IP                    *string         `json:"ip,omitempty"`
	Country               *string         `json:"country,omitempty"`
	Email                 *string         `json:"email,omitempty"`
	FirstName             *string         `json:"firstName,omitempty"`
	LastName              *string         `json:"lastName,omitempty"`
	Avatar                *string         `json:"avatar,omitempty"`
	Name                  *string         `json:"name,omitempty"`
	Anonymous             *bool           `json:"anonymous,omitempty"`
	Custom                *ldvalue.Value  `json:"custom,omitempty"`
	PrivateAttributeNames []UserAttribute `json:"privateAttributeNames,omitempty"`
}

type userForDeserialization struct {
	Key                   string                 `json:"key"`
	Secondary             ldvalue.OptionalString `json:"secondary"`
	IP                    ldvalue.OptionalString `json:"ip"`
	Country               ldvalue.OptionalString `json:"country"`
	Email                 ldvalue.OptionalString `json:"email"`
	FirstName             ldvalue.OptionalString `json:"firstName"`
	LastName              ldvalue.OptionalString `json:"lastName"`
	Avatar                ldvalue.OptionalString `json:"avatar"`
	Name                  ldvalue.OptionalString `json:"name"`
	Anonymous             ldvalue.Value          `json:"anonymous"`
	Custom                ldvalue.Value          `json:"custom"`
	PrivateAttributeNames []UserAttribute        `json:"privateAttributeNames"`
}

// String returns a simple string representation of a user.
//
// This currently uses the same JSON string representation as User.MarshalJSON(). Do not rely on this
// specific behavior of String(); it is intended for convenience in debugging.
func (u User) String() string {
	bytes, _ := json.Marshal(u)
	return string(bytes)
}

// MarshalJSON provides JSON serialization for User when using json.MarshalJSON.
//
// This is LaunchDarkly's standard JSON representation for user properties, in which all of the built-in
// user attributes are at the top level along with a "custom" property that is an object containing all of
// the custom attributes.
//
// In order for the representation to be as compact as possible, any top-level attributes for which no
// value has been set (as opposed to being set to an empty string) will be completely omitted, rather
// than including "attributeName":null in the JSON output. Similarly, if there are no custom attributes,
// there will be no "custom" property (rather than "custom":{}). This distinction does not matter to
// LaunchDarkly services-- they will treat an explicit null value in JSON data the same as an unset
// attribute, and treat an omitted "custom" the same as an empty "custom" map.
func (u User) MarshalJSON() ([]byte, error) {
	// We want to be able to use string pointers here so we can take advantage of "omitempty", but we
	// don't want to cause any heap allocations. Go will allow us to take the address of a string without
	// having it escape to the heap if it's done within the same scope where it is used, so we need to go
	// through a somewhat repetitive process here to copy these values into local variables. If we used
	// OptionalString.AsPointer(), the strings would escape to the heap.
	secondary, hasSecondary := u.secondary.Get()
	ip, hasIP := u.ip.Get()
	country, hasCountry := u.country.Get()
	email, hasEmail := u.email.Get()
	firstName, hasFirstName := u.firstName.Get()
	lastName, hasLastName := u.lastName.Get()
	avatar, hasAvatar := u.avatar.Get()
	name, hasName := u.name.Get()
	anon := u.anonymous.BoolValue()
	// _, hasSecondary := u.secondary.Get()
	// _, hasIP := u.ip.Get()
	// _, hasCountry := u.country.Get()
	// _, hasEmail := u.email.Get()
	// _, hasFirstName := u.firstName.Get()
	// _, hasLastName := u.lastName.Get()
	// _, hasAvatar := u.avatar.Get()
	// _, hasName := u.name.Get()
	// anon := u.anonymous.BoolValue()
	custom := u.custom

	ufs := userForSerialization{Key: u.key}
	if hasSecondary {
		ufs.Secondary = &secondary
	}
	if hasIP {
		ufs.IP = &ip
	}
	if hasCountry {
		ufs.Country = &country
	}
	if hasEmail {
		ufs.Email = &email
	}
	if hasFirstName {
		ufs.FirstName = &firstName
	}
	if hasLastName {
		ufs.LastName = &lastName
	}
	if hasAvatar {
		ufs.Avatar = &avatar
	}
	if hasName {
		ufs.Name = &name
	}
	if !u.anonymous.IsNull() {
		ufs.Anonymous = &anon
	}
	if custom.Count() > 0 {
		ufs.Custom = &custom
	}
	for a := range u.privateAttributes {
		ufs.PrivateAttributeNames = append(ufs.PrivateAttributeNames, a)
	}
	return json.Marshal(ufs)
}

// UnmarshalJSON provides JSON deserialization for User when using json.UnmarshalJSON.
//
// This is LaunchDarkly's standard JSON representation for user properties, in which all of the built-in
// properties are at the top level along with a "custom" property that is an object containing all of
// the custom properties. Omitted properties are treated as unset.
func (u *User) UnmarshalJSON(data []byte) error {
	var ufs userForDeserialization
	if err := json.Unmarshal(data, &ufs); err != nil {
		return err
	}
	*u = User{
		key:       ufs.Key,
		secondary: ufs.Secondary,
		ip:        ufs.IP,
		country:   ufs.Country,
		email:     ufs.Email,
		firstName: ufs.FirstName,
		lastName:  ufs.LastName,
		avatar:    ufs.Avatar,
		name:      ufs.Name,
		anonymous: ufs.Anonymous,
		custom:    ufs.Custom,
	}
	if len(ufs.PrivateAttributeNames) > 0 {
		u.privateAttributes = make(map[UserAttribute]struct{})
		for _, a := range ufs.PrivateAttributeNames {
			u.privateAttributes[a] = struct{}{}
		}
	}
	return nil
}
