package ldcontext

import (
	"encoding/json"
	"sort"

	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/lderrors"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

	"golang.org/x/exp/slices"
)

// Context is a collection of attributes that can be referenced in flag evaluations and analytics events.
//
// To create a Context of a single kind, such as a user, you may use the New() or NewWithKind()
// constructors. Or, to specify other attributes, use NewBuilder().
//
// To create a Context with multiple kinds, use NewMultiBuilder().
//
// An uninitialized Context struct is not valid for use in any SDK operations. Also, a Context can
// be in an error state if it was built with invalid attributes. See Context.Err().
type Context struct {
	defined           bool
	err               error
	kind              Kind
	multiContexts     []Context
	key               string
	fullyQualifiedKey string
	name              ldvalue.OptionalString
	attributes        ldvalue.ValueMap
	secondary         ldvalue.OptionalString
	anonymous         bool
	privateAttrs      []ldattr.Ref
}

// IsDefined returns true if this is a Context that was created with a constructor or builder
// (regardless of whether its properties are valid), or false if it is an empty uninitialized
// Context{}.
func (c Context) IsDefined() bool {
	return c.defined
}

// Err returns nil for a valid Context, or a non-nil error value for an invalid Context.
//
// A valid Context is one that can be used in SDK operations. An invalid Context is one that is
// missing necessary attributes or has invalid attributes, indicating an incorrect usage of the
// SDK API. For a complete list of the ways a Context can be invalid, see the ErrContext___
// types in the lderrors package.
//
// Since in normal usage it is easy for applications to be sure they are using context kinds
// correctly (so that having to constantly check error return values would be needlessly
// inconvenient), and because some states such as the empty value are impossible to prevent in the
// Go language, the SDK stores the error state in the Context itself and checks for such errors
// at the time the Context is used, such as in a flag evaluation. At that point, if the Context is
// invalid, the operation will fail in some well-defined way as described in the documentation for
// that method, and the SDK will generally log a warning as well. But in any situation where you
// are not sure if you have a valid Context, you can call Err() to check.
func (c Context) Err() error {
	if !c.defined && c.err == nil {
		return lderrors.ErrContextUninitialized{}
	}
	return c.err
}

// Kind returns the Context's kind attribute.
//
// Every valid Context has a non-empty kind. For multi-kind contexts, this value is "multi" and the
// kinds within the Context can be inspected with IndividualContextCount(), IndividualContextByIndex(),
// IndividualContextByKind(), or GetAllIndividualContexts().
//
// For rules regarding the kind value, see Builder.Kind().
func (c Context) Kind() Kind {
	return c.kind
}

// Multiple returns true for a multi-kind Context, or false for a single-kind Context.
//
// If this value is true, then Kind() is guaranteed to return "multi", and you can inspect the
// individual Contexts for each kind by calling IndividualContextCount(), IndividualContextByIndex(),
// IndividualContextByKind() or GetAllIndividualContexts().
//
// If this value is false, then Kind() is guaranteed to return a value that is not "multi", and
// IndividualContextCount() is guaranteed to return 1.
func (c Context) Multiple() bool {
	return len(c.multiContexts) != 0
}

// Key returns the Context's key attribute.
//
// For a single-kind context, this value is set by the Context constructors or the Builder methods.
//
// For a multi-kind context, there is no single value, so Key() returns an empty name; use
// IndividualContextByIndex(), IndividualContextByName(), or GetAllIndividualContexts() to inspect
// a Context for a particular kind and call Key() on it.
func (c Context) Key() string {
	return c.key
}

// FullyQualifiedKey returns a string that describes the entire Context based on Kind and Key values.
//
// This value is used whenever LaunchDarkly needs a string identifier based on all of the Kind and
// Key values in the context; the SDK may use this for caching previously seen contexts, for instance.
func (c Context) FullyQualifiedKey() string {
	return c.fullyQualifiedKey
}

// Name returns the Context's optional name attribute.
//
// For a single-kind context, this value is set by Builder.Name() or Builder.OptName(). If no value was
// specified, it returns the empty value ldvalue.OptionalString{}. The name attribute is treated
// differently from other user attributes in that its value, if specified, can only be a string, and
// it is used as the display name for the Context on the LaunchDarkly dashboard.
//
// For a multi-kind context, there is no single value, so Name() returns an empty string; use
// IndividualContextByIndex(), IndividualContextByName(), or GetAllIndividualContexts() to inspect
// a Context for a particular kind and call Name() on it.
func (c Context) Name() ldvalue.OptionalString {
	return c.name
}

// GetOptionalAttributeNames returns a slice containing the names of all regular optional attributes defined
// on this Context. These do not include the mandatory Kind and Key, or the metadata attributes Secondary,
// Anonymous, and Private.
//
// If a non-nil slice is passed in, it will be reused to hold the return values if it has enough capacity.
// For instance, in the following example, no heap allocations will happen unless there are more than 10
// optional attribute names; if there are more than 10, the slice will be allocated on the stack:
//
//	preallocNames := make([]string, 0, 10)
//	names := c.GetOptionalAttributeNames(preallocNames)
func (c Context) GetOptionalAttributeNames(sliceIn []string) []string {
	if c.Multiple() {
		return nil
	}
	ret := c.attributes.Keys(sliceIn)
	if c.name.IsDefined() {
		ret = append(ret, ldattr.NameAttr)
	}
	return ret
}

// GetValue looks up the value of any attribute of the Context by name. This includes only attributes
// that are addressable in evaluations-- not metadata such as Secondary() or Private().
//
// For a single-kind context, the attribute name can be any custom attribute that was set by methods
// like Builder.SetString(). It can also be one of the built-in ones like "kind", "key", or "name"; in
// such cases, it is equivalent to calling Kind(), Key(), or Name(), except that the value is returned
// using the general-purpose ldvalue.Value type.
//
// For a multi-kind context, the only supported attribute name is "kind". Use
// IndividualContextByIndex(), IndividualContextByName(), or GetAllIndividualContexts() to inspect
// a Context for a particular kind and then get its attributes.
//
// This method does not support complex expressions for getting individual values out of JSON objects
// or arrays, such as "/address/street". Use GetValueForRef() for that purpose.
//
// If the value is found, the return value is the attribute value, using the type ldvalue.Value to
// represent a value of any JSON type).
//
// If there is no such attribute, the return value is ldvalue.Null(). An attribute that actually
// exists cannot have a null value.
func (c Context) GetValue(attrName string) ldvalue.Value {
	return c.GetValueForRef(ldattr.NewLiteralRef(attrName))
}

// GetValueForRef looks up the value of any attribute of the Context, or a value contained within an
// attribute, based on an ldattr.Ref. This includes only attributes that are addressable in evaluations--
// not metadata such as Secondary() or Private().
//
// This implements the same behavior that the SDK uses to resolve attribute references during a flag
// evaluation. In a single-kind context, the ldattr.Ref can represent a simple attribute name-- either a
// built-in one like "name" or "key", or a custom attribute that was set by methods like
// Builder.SetString()-- or, it can be a slash-delimited path using a JSON-Pointer-like syntax. See
// ldattr.Ref for more details.
//
// For a multi-kind context, the only supported attribute name is "kind". Use
// IndividualContextByIndex(), IndividualContextByName(), or GetAllIndividualContexts() to inspect
// a Context for a particular kind and then get its attributes.
//
// If the value is found, the return value is the attribute value, using the type ldvalue.Value to
// represent a value of any JSON type).
//
// If there is no such attribute, or if the ldattr.Ref is invalid, the return value is ldvalue.Null().
// An attribute that actually exists cannot have a null value.
func (c Context) GetValueForRef(ref ldattr.Ref) ldvalue.Value {
	if ref.Err() != nil {
		return ldvalue.Null()
	}

	firstPathComponent := ref.Component(0)

	if c.Multiple() {
		if ref.Depth() == 1 && firstPathComponent == ldattr.KindAttr {
			return ldvalue.String(string(c.kind))
		}
		return ldvalue.Null() // multi-kind context has no other addressable attributes
	}

	// Look up attribute in single-kind context
	value, ok := c.getTopLevelAddressableAttributeSingleKind(firstPathComponent)
	if !ok {
		return ldvalue.Null()
	}
	for i := 1; i < ref.Depth(); i++ {
		name := ref.Component(i)
		value, ok = value.TryGetByKey(name)
		if !ok { // ok is false if either the name is not found or the value was not an object
			return ldvalue.Null()
		}
	}
	return value
}

// Anonymous returns true if this Context is only intended for flag evaluations and will not be indexed by
// LaunchDarkly.
//
// For a single-kind context, this value can be set by Builder.Anonymous(), and is false if not specified.
//
// For a multi-kind context, there is no single value, so Anonymous() always returns false; use
// IndividualContextByIndex(), IndividualContextByName(), or GetAllIndividualContexts() to inspect
// a Context for a particular kind and then call Anonymous() on it.
func (c Context) Anonymous() bool {
	return c.anonymous
}

// Secondary returns the secondary key attribute for the Context, if any.
//
// For a single-kind context, this value can be set by Builder.Secondary(), and is an empty
// ldvalue.OptionalString{} value if not specified.
//
// For a multi-kind context, there is no single value, so Secondary() always returns an empty value; use
// IndividualContextByIndex(), IndividualContextByName(), or GetAllIndividualContexts() to inspect
// a Context for a particular kind and then call Secondary() on it.
func (c Context) Secondary() ldvalue.OptionalString {
	return c.secondary
}

// PrivateAttributeCount returns the number of attributes that were marked as private for this Context
// with Builder.Private().
func (c Context) PrivateAttributeCount() int {
	return len(c.privateAttrs)
}

// PrivateAttributeByIndex returns one of the attributes that were marked as private for thie Context
// with Builder.Private().
func (c Context) PrivateAttributeByIndex(index int) (ldattr.Ref, bool) {
	if index < 0 || index >= len(c.privateAttrs) {
		return ldattr.Ref{}, false
	}
	return c.privateAttrs[index], true
}

// IndividualContextCount returns the number of Kinds in the context. If this is a single-kind
// context, the return value is 1. If it is a multi-kind context, it is the number of individual
// Kinds within.
func (c Context) IndividualContextCount() int {
	if n := len(c.multiContexts); n != 0 {
		return n
	}
	return 1
}

// IndividualContextByIndex returns the single-kind context corresponding to one of the Kinds in
// this context. If the method is called on a single-kind context, then the only allowable value
// for index is zero, and the return value on success is the same context. If the method is called
// on a multi-kind context, then index must be >= zero and < the number of kinds, and the return
// value on success is one of the individual contexts within.
//
// If the index is out of range, then the return value is an uninitialized Context{}. You can
// detect this condition because its IsDefined() method will return false.
//
// In a multi-kind context, the ordering of the individual contexts is not guaranteed to be the
// same order that was passed into the builder or constructor.
func (c Context) IndividualContextByIndex(index int) Context {
	if n := len(c.multiContexts); n != 0 {
		if index < 0 || index >= n {
			return Context{}
		}
		return c.multiContexts[index]
	}
	if index != 0 {
		return Context{}
	}
	return c
}

// IndividualContextByKind returns the single-kind context, if any, whose Kind matches the
// specified value exactly. If the method is called on a single-kind context, then the specified
// Kind must match the kind of that context. If the method is called on a multi-kind context,
// then the Kind can match any of the individual contexts within.
//
// If the kind parameter is an empty string, ldcontext.DefaultKind is used instead.
//
// If no matching Kind is found, then the return value is an uninitialized Context{}. You can
// detect this condition because its IsDefined() method will return false.
func (c Context) IndividualContextByKind(kind Kind) Context {
	if kind == "" {
		kind = DefaultKind
	}
	if len(c.multiContexts) == 0 {
		if c.kind == kind {
			return c
		}
	} else {
		for _, mc := range c.multiContexts {
			if mc.kind == kind {
				return mc
			}
		}
	}
	return Context{}
}

// IndividualContextKeyByKind returns the Key of the single-kind context, if any, whose Kind
// matches the specified value exactly. If the method is called on a single-kind context, then
// the specified Kind must match the Kind of that context. If the method is called on a multi-kind
// context, then the Kind can match any of the individual contexts within.
//
// If the kind parameter is an empty string, ldcontext.DefaultKind is used instead.
//
// If no matching Kind is found, the return value is an empty string.
//
// This method is equivalent to calling IndividualContextByKind and then Key, but is slightly
// more efficient (since it does not require copying an entire Context struct by value).
func (c Context) IndividualContextKeyByKind(kind Kind) string {
	if kind == "" {
		kind = DefaultKind
	}
	if len(c.multiContexts) == 0 {
		if c.kind == kind {
			return c.key
		}
	} else {
		for _, mc := range c.multiContexts {
			if mc.kind == kind {
				return mc.key
			}
		}
	}
	return ""
}

// GetAllIndividualContexts converts this context to a slice of individual contexts. If the method
// is called on a single-kind context, then the resulting slice has exactly one element, which
// is the same context. If the method is called on a multi-kind context, then the resulting
// slice contains each individual context within.
//
// If a non-nil slice is passed in, it will be reused to hold the return values if it has enough
// capacity. For instance, in the following example, no heap allocations will happen unless there
// are more than 10 individual contexts; if there are more than 10, the slice will be allocated on
// the stack:
//
//	preallocContexts := make([]ldcontext.Context, 0, 10)
//	contexts := c.GetAllIndividualContexts(preallocContexts)
func (c Context) GetAllIndividualContexts(sliceIn []Context) []Context {
	ret := sliceIn[0:0]
	if len(c.multiContexts) == 0 {
		return append(ret, c)
	}
	return append(ret, c.multiContexts...)
}

// String returns a string representation of the Context.
//
// This is currently defined as being the same as the JSON representation, since that is the simplest
// way to represent all of the Context properties. However, Go's fmt.Stringer interface is deliberately
// nonspecific about what format a type may use for its string representation, and application code
// should not rely on String() always being the same as the JSON representation. If you specifically
// want the latter, use json.Marshal. However, if you do use String() for convenience in debugging or
// logging, you should assume that the output may contain any and all properties of the Context, so if
// there is anything you do not want to be visible, you should write your own formatting logic.
func (c Context) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func (c Context) getTopLevelAddressableAttributeSingleKind(name string) (ldvalue.Value, bool) {
	switch name {
	case ldattr.KindAttr:
		return ldvalue.String(string(c.kind)), true
	case ldattr.KeyAttr:
		return ldvalue.String(c.key), true
	case ldattr.NameAttr:
		return c.name.AsValue(), c.name.IsDefined()
	case ldattr.AnonymousAttr:
		return ldvalue.Bool(c.anonymous), true
	default:
		return c.attributes.TryGet(name)
	}
}

// Equal tests whether two contexts are logically equal.
//
// Two single-kind contexts are logically equal if they have the same attribute names and values.
// Two multi-kind contexts are logically equal if they contain the same kinds (in any order) and
// the individual contexts are equal. A single-kind context is never equal to a multi-kind context.
func (c Context) Equal(other Context) bool {
	if !c.defined || !other.defined {
		return c.defined == other.defined
	}

	if c.kind != other.kind {
		return false
	}

	if c.Multiple() {
		if len(c.multiContexts) != len(other.multiContexts) {
			return false
		}
		for _, mc1 := range c.multiContexts {
			if mc2 := other.IndividualContextByKind(mc1.kind); !mc1.Equal(mc2) {
				return false
			}
		}
		return true
	}

	if c.key != other.key ||
		c.name != other.name ||
		c.anonymous != other.anonymous ||
		c.secondary != other.secondary {
		return false
	}
	if !c.attributes.Equal(other.attributes) {
		return false
	}
	if len(c.privateAttrs) != len(other.privateAttrs) {
		return false
	}
	sortedPrivateAttrs := func(attrs []ldattr.Ref) []string {
		ret := make([]string, 0, len(attrs))
		for _, a := range attrs {
			ret = append(ret, a.String())
		}
		sort.Strings(ret)
		return ret
	}
	attrs1, attrs2 := sortedPrivateAttrs(c.privateAttrs), sortedPrivateAttrs(other.privateAttrs)
	return slices.Equal(attrs1, attrs2)
}
