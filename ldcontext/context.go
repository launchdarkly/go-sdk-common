package ldcontext

import (
	"encoding/json"
	"sort"

	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
)

// Context is a collection of attributes that can be referenced in flag evaluations and analytics events.
//
// (TKTK - some conceptual text here, and/or a link to a docs page)
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
	transient         bool
	privateAttrs      []ldattr.Ref
}

// Err returns nil for a valid Context, or a non-nil error value for an invalid Context.
//
// A valid Context is one that can be used in SDK operations. An invalid Context is one that is
// missing necessary attributes or has invalid attributes, indicating an incorrect usage of the
// SDK API. The only ways for a Context to be invalid are:
//
// - It has a disallowed value for the Kind property. See Builder.Kind().
// - It is a multi-kind Context that does not have any kinds. See MultiBuilder.
// - It is a multi-kind Context where the same kind appears more than once.
// - It is a multi-kind Context where at least one of the nested Contexts had an error.
// - It is an uninitialized empty Context{} value.
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
		return errContextUninitialized
	}
	return c.err
}

// Kind returns the Context's kind attribute.
//
// Every valid Context has a non-empty kind. For multi-kind contexts, this value is "multi" and the
// kinds within the Context can be inspected with MultiKindCount(), MultiKindByIndex, and MultiKindByName.
//
// For rules regarding the kind value, see Builder.Kind().
func (c Context) Kind() Kind {
	return c.kind
}

// Multiple returns true for a multi-kind Context, or false for a single-kind Context.
//
// If this value is true, then Kind() is guaranteed to return "multi", and you can inspect the
// individual Contexts for each kind by calling MultiKindCount(), MultiKindByIndex, and MultiKindByKey().
//
// If this value is false, then Kind() is guaranteed to return a value that is not "multi", and
// MultiKindCount() is guaranteed to return zero.
func (c Context) Multiple() bool {
	return len(c.multiContexts) != 0
}

// Key returns the Context's key attribute.
//
// For a single-kind context, this value is set by the Context constructors or the Builder methods.
//
// For a multi-kind context, there is no single value, so Key() returns an empty name; use
// MultiKindByIndex or MultiKindByName to inspect a Context for a particular kind and call Key() on it.
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
// For a multi-kind context, there is no single value, so Name() returns an empty string. Use
// MultiKindByIndex or MultiKindByName to inspect a Context for a particular kind and get its Name().
func (c Context) Name() ldvalue.OptionalString {
	return c.name
}

// GetOptionalAttributeNames returns a slice containing the names of all regular optional attributes defined
// on this Context. These do not include the mandatory Kind and Key, or the metadata attributes Secondary,
// Transient, and Private.
//
// If a non-nil slice is passed in, it will be reused to hold the return values if it has enough capacity.
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
// For a multi-kind context, the only supported attribute name is "kind". Use MultiKindByIndex() or
// MultiKindByName() to inspect a Context for a particular kind and then get its attributes.
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
	return c.GetValueForRef(ldattr.NewNameRef(attrName))
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
// For a multi-kind context, the only supported attribute name is "kind". Use MultiKindByIndex() or
// MultiKindByName() to inspect a Context for a particular kind and then get its attributes.
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

	firstPathComponent, _ := ref.Component(0)

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
		name, index := ref.Component(i)
		if index.IsDefined() && value.Type() == ldvalue.ArrayType {
			value, ok = value.TryGetByIndex(index.IntValue())
		} else {
			value, ok = value.TryGetByKey(name)
		}
		if !ok {
			return ldvalue.Null()
		}
	}
	return value
}

// Transient returns true if this Context is only intended for flag evaluations and will not be indexed by
// LaunchDarkly.
//
// For a single-kind context, this value can be set by Builder.Transient(), and is false if not specified.
//
// For a multi-kind context, there is no single value, so Transient() always returns false; use
// MultiKindByIndex or MultiKindByName to inspect a Context for a particular kind and call Transient() on it.
func (c Context) Transient() bool {
	return c.transient
}

// Secondary returns the secondary key attribute for the Context, if any.
//
// For a single-kind context, this value can be set by Builder.Secondary(), and is an empty
// ldvalue.OptionalString{} value if not specified.
//
// For a multi-kind context, there is no single value, so Secondary() always returns an empty value; use
// MultiKindByIndex or MultiKindByName to inspect a Context for a particular kind and call Secondary() on it.
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

// MultiKindCount returns the number of Kinds if this is a multi-kind Context created with NewMulti()
// or NewMultiBuilder().
func (c Context) MultiKindCount() int {
	return len(c.multiContexts)
}

// MultiKindByIndex returns one of the individual Contexts in a multi-kind Context.
func (c Context) MultiKindByIndex(index int) (Context, bool) {
	if index < 0 || index >= len(c.multiContexts) {
		return Context{}, false
	}
	return c.multiContexts[index], true
}

// MultiKindByName finds one of the individual Contexts in a multi-kind Context.
func (c Context) MultiKindByName(kind Kind) (Context, bool) {
	for _, mc := range c.multiContexts {
		if mc.kind == kind {
			return mc, true
		}
	}
	return Context{}, false
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
	case ldattr.TransientAttr:
		return ldvalue.Bool(c.transient), true
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
	if c.kind != other.kind {
		return false
	}

	if c.Multiple() {
		if len(c.multiContexts) != len(other.multiContexts) {
			return false
		}
		for _, mc1 := range c.multiContexts {
			if mc2, ok := other.MultiKindByName(mc1.kind); !ok || !mc1.Equal(mc2) {
				return false
			}
		}
		return true
	}

	if c.key != other.key ||
		c.name != other.name ||
		c.transient != other.transient ||
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
	for i, a := range attrs1 {
		if a != attrs2[i] {
			return false
		}
	}
	return true
}
