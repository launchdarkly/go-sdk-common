package lduser

// UserAttribute is a string type representing the name of a user attribute.
//
// Constants like KeyAttribute describe all of the built-in attributes; you may also cast any string to
// UserAttribute when referencing a custom attribute name.
type UserAttribute string

const (
	// KeyAttribute is the standard attribute name corresponding to User.GetKey().
	KeyAttribute UserAttribute = "key"
	// SecondaryKeyAttribute is the standard attribute name corresponding to User.GetSecondaryKey().
	SecondaryKeyAttribute UserAttribute = "secondary"
	// IPAttribute is the standard attribute name corresponding to User.GetIP().
	IPAttribute UserAttribute = "ip"
	// CountryAttribute is the standard attribute name corresponding to User.GetCountry().
	CountryAttribute UserAttribute = "country"
	// EmailAttribute is the standard attribute name corresponding to User.GetEmail().
	EmailAttribute UserAttribute = "email"
	// FirstNameAttribute is the standard attribute name corresponding to User.GetFirstName().
	FirstNameAttribute UserAttribute = "firstName"
	// LastNameAttribute is the standard attribute name corresponding to User.GetLastName().
	LastNameAttribute UserAttribute = "lastName"
	// AvatarAttribute is the standard attribute name corresponding to User.GetAvatar().
	AvatarAttribute UserAttribute = "avatar"
	// NameAttribute is the standard attribute name corresponding to User.GetName().
	NameAttribute UserAttribute = "name"
	// AnonymousAttribute is the standard attribute name corresponding to User.GetAnonymous().
	AnonymousAttribute UserAttribute = "anonymous"
)
