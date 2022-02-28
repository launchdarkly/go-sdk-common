package ldcontext

import (
	"strconv"
	"strings"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

// AttrRef is an attribute name or path expression identifying an attribute within a Context.
//
// This can be passed to Context.GetValue() (which is also the method that the SDK uses to look up
// attributes when evaluating a flag), or when identifying attributes with methods like Builder.Private().
// Having a special type with a constructor, instead of just using a string parameter for such methods,
// allows the SDK to do any necessary parsing and validation just once when the AttrRef is created/
//
// Call NewAttrRef() to create an AttrRef.  An uninitialized AttrRef struct is not valid for use in any
// SDK operations. Also, an AttrRef can be in an error state if it was built from an invalid string. See
// AttrRef.Err().
type AttrRef struct {
	err                 error
	rawPath             string
	singlePathComponent string
	components          []attrRefComponent
}

type attrRefComponent struct {
	name     string
	intValue ldvalue.OptionalInt
}

// NewAttrRef creates an AttrRef from a string.
//
// This is used in methods like Builder.Private(), where the input could be either a simple attribute
// name like "name", or a more complex reference like "/zoo/animals/1". See AttrRef for more
// information about why this has its own type.
//
// (TKTK - some conceptual text here, and/or a link to a docs page)
//
// For convenience in using AttrRef within expressions, this function always returns a single value,
// rather than two values of (AttrRef, error). Any invalid input causes error state to be stored
// within the AttrRef; see AttrRef.Err().
func NewAttrRef(referenceString string) AttrRef {
	if referenceString == "" || referenceString == "/" {
		return AttrRef{err: errAttributeEmpty, rawPath: referenceString}
	}
	if referenceString[0] != '/' {
		// When there is no leading slash, this is a simple attribute reference with no character escaping.
		return AttrRef{singlePathComponent: referenceString, rawPath: referenceString}
	}
	path := referenceString[1:]
	if !strings.Contains(path, "/") {
		// There's only one segment, so this is still a simple attribute reference. However, we still may
		// need to unescape special characters.
		return AttrRef{singlePathComponent: unescapePath(path), rawPath: referenceString}
	}
	parts := strings.Split(path, "/")
	ret := AttrRef{rawPath: referenceString, components: make([]attrRefComponent, 0, len(parts))}
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

// NewAttrRefForName is similar to NewAttrRef except that it always interprets the string as a single
// attribute name, never as a slash-delimited path expression. Use this in cases where you need an
// AttrRef but might be referencing an attribute whose name actually starts with a slash.
func NewAttrRefForName(attrName string) AttrRef {
	if attrName == "" || attrName == "/" {
		return AttrRef{err: errAttributeEmpty, rawPath: attrName}
	}
	if attrName[0] != '/' {
		// When there is no leading slash, this is a simple attribute reference with no character escaping.
		return AttrRef{singlePathComponent: attrName, rawPath: attrName}
	}
	// If there is a leading slash, then the attribute name actually starts with a slash. To represent it
	// as an AttrRef, it'll need to be escaped.
	escapedPath := "/" + strings.ReplaceAll(strings.ReplaceAll(attrName, "~", "~0"), "/", "~1")
	return AttrRef{singlePathComponent: attrName, rawPath: escapedPath}
}

// Err returns nil for a valid AttrRef, or a non-nil error value for an invalid AttrRef.
//
// An AttrRef can only be invalid for the following reasons:
//
// - The input string was empty, or consisted only of "/".
//
// - A slash-delimited string had a double slash causing one component to be empty, such as "/a//b".
//
// Otherwise, the AttrRef is valid, but that does not guarantee that such an attribute
// exists in any given Context. For instance, NewAttrRef("name") is a valid AttrRef,
// but a specific Context might or might not have a name.
func (a AttrRef) Err() error {
	if a.err == nil && a.singlePathComponent == "" && a.components == nil {
		return errAttributeEmpty
	}
	return a.err
}

// Depth returns the number of path components in the AttrRef.
//
// For a simple attribute reference such as "name" with no leading slash, this returns 1.
//
// For an attribute reference with a leading slash, it is the number of slash-delimited path
// components after the initial slash. For instance, Attribute("/a/b").Depth() returns 2.
func (a AttrRef) Depth() int {
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
//     Attribute("a").Component(0)      // returns ("a", ldvalue.OptionalInt{})
//     Attribute("/a/b").Component(1)   // returns ("b", ldvalue.OptionalInt{})
//     Attribute("/a/3").Component(1)   // returns ("3", ldvalue.NewOptionalInt(3))
func (a AttrRef) Component(index int) (string, ldvalue.OptionalInt) {
	if index == 0 && len(a.components) == 0 {
		return a.singlePathComponent, ldvalue.OptionalInt{}
	}
	if index < 0 || index >= len(a.components) {
		return "", ldvalue.OptionalInt{}
	}
	c := a.components[index]
	return c.name, c.intValue
}

// String returns the attribute reference as a string. This is always identical to the
// string that was originally passed to Attribute().
func (a AttrRef) String() string {
	return a.rawPath
}

func unescapePath(path string) string {
	// If there are no tildes then there's definitely nothing to do
	if !strings.Contains(path, "~") {
		return path
	}
	return strings.ReplaceAll(strings.ReplaceAll(path, "~1", "/"), "~0", "~")
}
