package ldattr

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

	"github.com/launchdarkly/go-jsonstream/v2/jreader"
)

// Ref is an attribute name or path expression identifying a value within a Context.
//
// This type is mainly intended to be used internally by LaunchDarkly SDK and service code, where
// efficiency is a major concern so it's desirable to do any parsing or preprocessing just once.
// Applications are unlikely to need to use the Ref type directly.
//
// It can be used to retrieve a value with Context.GetValueForRef(), or to identify an attribute or
// nested value that should be considered private with Builder.Private() (the SDK configuration can also
// have a list of private attribute references).
//
// Parsing and validation are done at the time that the NewRef or NewLiteralRef constructor is called.
// If a Ref instance was created from an invalid string, or if it is an uninitialized Ref{}, it is
// considered invalid and its Err() method will return a non-nil error.
//
// Syntax
//
// The string representation of an attribute reference in LaunchDarkly JSON data uses the following
// syntax:
//
// If the first character is not a slash, the string is interpreted literally as an attribute name.
// An attribute name can contain any characters, but must not be empty.
//
// If the first character is a slash, the string is interpreted as a slash-delimited path where the
// first path component is an attribute name, and each subsequent path component is either the name of
// a property in a JSON object, or a decimal numeric string that is the index of an element in a JSON
// array. Any instances of the characters "/" or "~" in a path component are escaped as "~1" or "~0"
// respectively. This syntax deliberately resembles JSON Pointer, but no JSON Pointer behaviors other
// than those mentioned here are supported.
//
// Examples
//
// Suppose there is a context whose JSON implementation looks like this:
//
//     {
//       "kind": "user",
//       "key": "value1",
//       "address": {
//         "street": "value2",
//         "city": "value3"
//       },
//       "groups": [ "value4", "value5" ],
//       "good/bad": "value6"
//     }
//
// The attribute references "key" and "/key" would both point to "value1".
//
// The attribute reference "/address/street" would point to "value2".
//
// The attribute reference "/groups/0" would point to "value4".
//
// The attribute references "good/bad" and "/good~1bad" would both point to "value6".
type Ref struct {
	err                 error
	rawPath             string
	singlePathComponent string
	components          []attrRefComponent
}

type attrRefComponent struct {
	name     string
	intValue ldvalue.OptionalInt
}

// NewRef creates a Ref from a string. For the supported syntax and examples, see comments on the Ref type.
//
// This constructor always returns a Ref that preserves the original string, even if validation fails,
// so that calling String() (or serializing the Ref to JSON) will produce the original string. If
// validation fails, Err() will return a non-nil error and any SDK method that takes this Ref as a
// parameter will consider it invalid.
func NewRef(referenceString string) Ref {
	if referenceString == "" || referenceString == "/" {
		return Ref{err: errAttributeEmpty, rawPath: referenceString}
	}
	if referenceString[0] != '/' {
		// When there is no leading slash, this is a simple attribute reference with no character escaping.
		return Ref{singlePathComponent: referenceString, rawPath: referenceString}
	}
	path := referenceString[1:]
	if !strings.Contains(path, "/") {
		// There's only one segment, so this is still a simple attribute reference. However, we still may
		// need to unescape special characters.
		return Ref{singlePathComponent: unescapePath(path), rawPath: referenceString}
	}
	parts := strings.Split(path, "/")
	ret := Ref{rawPath: referenceString, components: make([]attrRefComponent, 0, len(parts))}
	for _, p := range parts {
		if p == "" {
			ret.err = errAttributeExtraSlash
			return ret
		}
		component := attrRefComponent{name: unescapePath(p)}
		if p[0] >= '0' && p[0] <= '9' {
			if n, err := strconv.Atoi(p); err == nil {
				component.intValue = ldvalue.NewOptionalInt(n)
			}
		}
		ret.components = append(ret.components, component)
	}
	return ret
}

// NewLiteralRef is similar to NewRef except that it always interprets the string as a literal
// attribute name, never as a slash-delimited path expression. There is no escaping or unescaping,
// even if the name contains literal '/' or '~' characters. Since an attribute name can contain
// any characters, this method always returns a valid Ref unless the name is empty.
//
// For example: ldattr.NewLiteralRef("name") is exactly equivalent to ldattr.NewRef("name").
// ldattr.NewLiteralRef("a/b") is exactly equivalent to ldattr.NewRef("a/b") (since the syntax
// used by NewRef treats the whole string as a literal as long as it does not start with a slash),
// or to ldattr.NewRef("/a~1b").
func NewLiteralRef(attrName string) Ref {
	if attrName == "" {
		return Ref{err: errAttributeEmpty, rawPath: attrName}
	}
	if attrName[0] != '/' {
		// When there is no leading slash, this is a simple attribute reference with no character escaping.
		return Ref{singlePathComponent: attrName, rawPath: attrName}
	}
	// If there is a leading slash, then the attribute name actually starts with a slash. To represent it
	// as an Ref, it'll need to be escaped.
	escapedPath := "/" + strings.ReplaceAll(strings.ReplaceAll(attrName, "~", "~0"), "/", "~1")
	return Ref{singlePathComponent: attrName, rawPath: escapedPath}
}

// IsDefined returns true if the Ref has a value, meaning that it is not an uninitialized Ref{}.
// That does not guarantee that the value is valid; use Err() to test that.
func (a Ref) IsDefined() bool {
	return a.rawPath != "" || a.err != nil
}

// Equal returns true if the two Ref instances have the same value.
//
// You cannot compare Ref instances with the == operator, because the struct may contain a slice;
// reflect.DeepEqual will work, but is less efficient.
func (a Ref) Equal(other Ref) bool {
	if a.err != other.err || a.rawPath != other.rawPath || a.singlePathComponent != other.singlePathComponent {
		return false
	}
	return true
	// We don't need to check the components slice, because it's impossible for the components to be different
	// if rawPath is the same.
}

// Err returns nil for a valid Ref, or a non-nil error value for an invalid Ref.
//
// A Ref can only be invalid for the following reasons:
//
// - The input string was empty, or consisted only of "/".
//
// - A slash-delimited string had a double slash causing one component to be empty, such as "/a//b".
//
// Otherwise, the Ref is valid, but that does not guarantee that such an attribute exists in any
// given Context. For instance, NewRef("name") is a valid Ref, but a specific Context might or might
// not have a name.
func (a Ref) Err() error {
	if a.err == nil && a.rawPath == "" {
		return errAttributeEmpty
	}
	return a.err
}

// Depth returns the number of path components in the Ref.
//
// For a simple attribute reference such as "name" with no leading slash, this returns 1.
//
// For an attribute reference with a leading slash, it is the number of slash-delimited path
// components after the initial slash. For instance, NewRef("/a/b").Depth() returns 2.
func (a Ref) Depth() int {
	if a.err != nil || (a.singlePathComponent == "" && a.components == nil) {
		return 0
	}
	if a.components == nil {
		return 1
	}
	return len(a.components)
}

// Component retrieves a single path component from the attribute reference.
//
// For a simple attribute reference such as "name" with no leading slash, if index is zero,
// Component returns the attribute name and an empty ldvalue.OptionalInt{}.
//
// For an attribute reference with a leading slash, if index is non-negative and less than
// a.Depth(), Component returns the path component as a string for its first value. The
// second value is an ldvalue.OptionalInt that is the integer value of that string as returned
// by strconv.Atoi() if applicable, or an empty ldvalue.OptionalInt{} if the string does not
// represent an integer; this is used to implement a "find a value by index within a JSON
// array" behavior similar to JSON Pointer.
//
// If index is out of range, it returns "" and an empty ldvalue.OptionalInt{}.
//
//     NewRef("a").Component(0)      // returns ("a", ldvalue.OptionalInt{})
//     NewRef("/a/b").Component(1)   // returns ("b", ldvalue.OptionalInt{})
//     NewRef("/a/3").Component(1)   // returns ("3", ldvalue.NewOptionalInt(3))
func (a Ref) Component(index int) (string, ldvalue.OptionalInt) {
	if index == 0 && len(a.components) == 0 {
		return a.singlePathComponent, ldvalue.OptionalInt{}
	}
	if index < 0 || index >= len(a.components) {
		return "", ldvalue.OptionalInt{}
	}
	c := a.components[index]
	return c.name, c.intValue
}

// String returns the attribute reference as a string, in the same format used by NewRef().
// If the Ref was created with NewRef(), this value is identical to the original string. If it
// was created with NewRefForName(), the value may be different due to unescaping (for instance,
// an attribute whose name is "/a" would be represented as "~1a".
func (a Ref) String() string {
	return a.rawPath
}

// MarshalJSON produces a JSON representation of the Ref. If it is an uninitialized Ref{}, this
// is a JSON null token. Otherwise, it is a JSON string using the same value returned by String().
func (a Ref) MarshalJSON() ([]byte, error) {
	if !a.IsDefined() {
		return []byte(`null`), nil
	}
	return json.Marshal(a.String())
}

// UnmarshalJSON parses a Ref from a JSON value. If the value is null, the result is an
// uninitialized Ref(). If the value is a string, it is passed to NewRef(). Any other type
// causes an error.
//
// A valid JSON string that is not valid as a Ref path (such as "" or "///") does not cause
// UnmarshalJSON to return an error; instead, it stores the string in the Ref and the error
// can be obtained from Ref.Err(). This is deliberate, so that the LaunchDarkly SDK will be
// able to parse a set of feature flag data even if one of the flags contains an invalid Ref.
func (a *Ref) UnmarshalJSON(data []byte) error {
	r := jreader.NewReader(data)
	s, nonNull := r.StringOrNull()
	if err := r.Error(); err != nil {
		return err
	}
	if nonNull {
		*a = NewRef(s)
	} else {
		*a = Ref{}
	}
	return nil
}

func unescapePath(path string) string {
	// If there are no tildes then there's definitely nothing to do
	if !strings.Contains(path, "~") {
		return path
	}
	return strings.ReplaceAll(strings.ReplaceAll(path, "~1", "/"), "~0", "~")
}
