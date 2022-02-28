package ldcontext

import (
	"fmt"
	"net/url"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

// Builder is a mutable object that uses the builder pattern to specify properties for a Context.
//
// Use this type if you need to construct a Context that has only a single Kind. To define a
// multi-kind Context, use MultiBuilder instead.
//
// Obtain an instance of Builder by calling NewBuilder. Then, call setter methods such as Kind,
// Name, or SetString to specify any additional properties. Then, call Build() to create the
// Context. All of the Builder setters return a reference to the same builder, so they can be
// chained together:
//
//     context := ldcontext.NewBuilder("user-key").
//         Name("my-name").
//         SetString("country", "us").
//         Build()
//
// A Builder should not be accessed by multiple goroutines at once. Once you have called Build(),
// the resulting Context is immutable and is safe to use from multiple goroutines. Instances
// created with Build() are not affected by subsequent actions taken on the Builder.
type Builder struct {
	kind                  Kind
	key                   string
	allowEmptyKey         bool
	name                  ldvalue.OptionalString
	attributes            map[string]ldvalue.Value
	attributesCopyOnWrite bool
	secondary             ldvalue.OptionalString
	transient             bool
	privateAttrs          []AttrRef
	privateCopyOnWrite    bool
}

// NewBuilder creates a Builder for building a Context, initializing its Key property and
// setting Kind to DefaultKind.
//
// You may use Builder methods to set additional attributes and/or change the Kind before
// calling Build(). If you do not change any values, the defaults for the Context are that
// its Kind is DefaultKind ("user"), its Key is set to whatever value you passed to NewBuilder,
// its Transient attribute is false, and it has no values for any other attributes.
//
// This method is for building a Context that has only a single Kind. To define a multi-kind
// Context, use NewMultiBuilder() instead.
//
// If the key parameter is an empty string, there is no default. A Context must have a
// non-empty key, so if you call Build() in this state without using Key() to set the key, you
// will get an invalid Context.
//
// An empty Builder{} is valid as long as you call Key() to set a non-empty key. This means
// that in in performance-critical code paths where you want to minimize heap allocations, if
// you do not want to allocate a Builder on the heap with NewBuilder, you can declare one
// locally instead:
//
//     var b ldcontext.Builder
//     c := b.Kind("org").Key("my-key").Name("my-name").Build()
func NewBuilder(key string) *Builder {
	b := &Builder{}
	return b.Key(key)
}

// NewBuilderFromContext creates a Builder whose properties are the same as an existing
// single-kind Context. You may then change the Builder's state in any way and call Build()
// to create a new independent Context.
//
// If fromContext is a multi-kind Context, this method does nothing.
func NewBuilderFromContext(fromContext Context) *Builder {
	b := &Builder{}
	b.copyFrom(fromContext)
	return b
}

// Build creates a Context from the current Builder properties.
//
// The Context is immutable and will not be affected by any subsequent actions on the Builder.
//
// It is possible to specify invalid attributes for a Builder, such as an empty Key. Instead of
// returning two values (Context, error), the Builder always returns a Context and you can call
// Context.Err() to see if it has an error. See Context.Err() for more information about
// invalid Context conditions. Using a single-return-value syntax is more convenient for
// application code, since in normal usage an application will never build an invalid Context.
// If you pass an invalid Context to an SDK method, the SDK will detect this and will generally
// log a description of the error.
//
// You may call TryBuild instead of Build if you prefer to use two-value return semantics, but
// the validation behavior is the same for both.
func (b *Builder) Build() Context {
	actualKind, err := validateSingleKind(b.kind)
	if err != nil {
		return Context{err: err, kind: b.kind}
	}
	if b.key == "" && !b.allowEmptyKey {
		return Context{err: errContextKeyEmpty, kind: b.kind}
	}
	// We set the kind in the error cases above because that improves error reporting if this
	// context is used within a multi-kind context.

	ret := Context{
		defined:   true,
		kind:      actualKind,
		key:       b.key,
		name:      b.name,
		transient: b.transient,
		secondary: b.secondary,
	}

	ret.fullyQualifiedKey = makeFullyQualifiedKeySingleKind(actualKind, ret.key, true)

	if b.attributes != nil {
		ret.attributes = b.attributes
		b.attributesCopyOnWrite = true
		// The ___CopyOnWrite fields allow us to avoid the overhead of cloning maps/slices in
		// the typical case where Builder properties do not get modified after calling Build().
		// To guard against concurrent modification if someone does continue to modify the
		// Builder after calling Build(), we will clone the data later if and only if someone
		// tries to modify it when ___CopyOnWrite is true. That is safe as long as no one is
		// trying to modify Builder from two goroutines at once, which (per our documentation)
		// is not supported anyway.
	}
	if b.privateAttrs != nil {
		ret.privateAttrs = b.privateAttrs
		b.privateCopyOnWrite = true
	}

	return ret
}

// TryBuild is an alternative to Build that returns any validation errors as a second value.
//
// As described in Build(), there are several ways the state of a Context could be invalid.
// Since in normal usage it is possible to be confident that these will not occur, the Build()
// method is designed for convenient use within expressions by returning a single Context
// value, and any validation problems are contained within that value where they can be
// detected by calling the context's Err() method. But, if you prefer to use the two-value
// pattern that is common in Go, you can call TryBuild instead:
//
//     c, err := ldcontext.NewBuilder("my-key").
//         Name("my-name").
//         TryBuild()
//     if err != nil {
//         // do whatever is appropriate if building the context failed
//     }
//
// The two return values are the same as to 1. the Context that would be returned by Build(),
// and 2. the result of calling Err() on that Context. So, the above example is exactly
// equivalent to:
//
//     c := ldcontext.NewBuilder("my-key").
//         Name("my-name").
//         Build()
//     if c.Err() != nil {
//         // do whatever is appropriate if building the context failed
//     }
//
// Note that unlike some Go methods where the first return value is normally an
// uninitialized zero value if the error is non-nil, the Context returned by TryBuild in case
// of an error is not completely uninitialized: it does contain the error information as well,
// so that if it is mistakenly passed to an SDK method, the SDK can tell what the error was.
func (b *Builder) TryBuild() (Context, error) {
	c := b.Build()
	return c, c.Err()
}

// Kind sets the Context's kind attribute.
//
// Every Context has a kind. Setting it to an empty string is equivalent to the default kind of
// "user". This value is case-sensitive. Validation rules are as follows:
//
// - It may only contain letters, numbers, and the characters ".", "_", and "-".
// - It cannot equal the literal string "kind".
// - It cannot equal "multi".
//
// If the value is invalid at the time Build() is called, you will receive an invalid Context
// whose Err() value will describe the problem.
func (b *Builder) Kind(kind Kind) *Builder {
	if kind == "" {
		b.kind = DefaultKind
	} else {
		b.kind = kind
	}
	return b
}

// Key sets the Context's key attribute.
//
// Every Context has a key, which is always a string. There are no restrictions on its value. It may
// be an empty string.
//
// The key attribute can be referenced by flag rules, flag target lists, and segments.
func (b *Builder) Key(key string) *Builder {
	b.key = key
	return b
}

// Used internally when we are deserializing an old-style user from JSON; otherwise an empty key is
// never allowed.
func (b *Builder) setAllowEmptyKey(value bool) *Builder {
	b.allowEmptyKey = value
	return b
}

// Name sets the Context's name attribute.
//
// This attribute is optional. It has the following special rules:
//
// - Unlike most other attributes, it is always a string if it is specified.
// - The LaunchDarkly dashboard treats this attribute as the preferred display name for users.
func (b *Builder) Name(name string) *Builder {
	return b.OptName(ldvalue.NewOptionalString(name))
}

// OptName sets or clears the Context's name attribute.
//
// Calling b.OptName(ldvalue.NewOptionalString("x")) is equivalent to b.Name("x"), but since it uses
// the OptionalString type, it also allows clearing a previously set name with
// b.OptName(ldvalue.OptionalString{}).
func (b *Builder) OptName(name ldvalue.OptionalString) *Builder {
	b.name = name
	return b
}

// SetBool sets an attribute to a boolean value.
//
// For rules regarding attribute names and values, see SetValue. This method is exactly
// equivalent to calling b.SetValue(attributeName, ldvalue.Bool(value)).
func (b *Builder) SetBool(attributeName string, value bool) *Builder {
	return b.SetValue(attributeName, ldvalue.Bool(value))
}

// SetFloat64 sets an attribute to a float64 numeric value. This method is exactly
// equivalent to calling b.SetValue(attributeName, ldvalue.Float64(value)).
//
// For rules regarding attribute names and values, see SetValue.
//
// Note: the LaunchDarkly model for feature flags and user attributes is based on JSON types,
// and does not distinguish between integer and floating-point types. Therefore,
// b.SetFloat64(name, float64(1.0)) is exactly equivalent to b.SetInt(name, 1).
func (b *Builder) SetFloat64(attributeName string, value float64) *Builder {
	return b.SetValue(attributeName, ldvalue.Float64(value))
}

// SetInt sets an attribute to an int numeric value.
//
// For rules regarding attribute names and values, see SetValue. This method is exactly
// equivalent to calling b.SetValue(attributeName, ldvalue.Int(value)).
//
// Note: the LaunchDarkly model for feature flags and user attributes is based on JSON types,
// and does not distinguish between integer and floating-point types. Therefore,
// b.SetFloat64(name, float64(1.0)) is exactly equivalent to b.SetInt(name, 1).
func (b *Builder) SetInt(attributeName string, value int) *Builder {
	return b.SetValue(attributeName, ldvalue.Int(value))
}

// SetString sets an attribute to a string value.
//
// For rules regarding attribute names and values, see SetValue. This method is exactly
// equivalent to calling b.SetValue(attributeName, ldvalue.String(value)).
func (b *Builder) SetString(attributeName string, value string) *Builder {
	return b.SetValue(attributeName, ldvalue.String(value))
}

// SetValue sets the value of any attribute for the Context.
//
// This method uses the ldvalue.Value type to represent a value of any JSON type: null, boolean,
// number, string, array, or object. For all attribute names that do not have special meaning
// to LaunchDarkly, you may use any of those types. Values of different JSON types are always
// treated as different values: for instance, null, false, and the empty string "" are not the
// the same, and the number 1 is not the same as the string "1".
//
// The following attribute names have special restrictions:
//
// - "kind", "key": Must be a string. Values of other types are treated as an empty string.
// See Builder.Kind() and Builder.Key().
//
// - "name", "secondary": Must be a string or null. Null is equivalent to not setting any value
// for this attribute. Values of other types are treated as null. See Builder.Name() and
// Builder.Secondary().
//
// - "transient": Must be a boolean. Values of other types are treated as false.
// See Builder.Transient().
//
// Values that are JSON arrays or objects have special behavior when referenced in flag/segment
// rules.
//
// A value of ldvalue.Null() is equivalent to removing any current non-default value of the
// attribute. Null is not a valid attribute value in the LaunchDarkly model; any expressions
// in feature flags that reference an attribute with a null value will behave as if the
// attribute did not exist.
func (b *Builder) SetValue(attributeName string, value ldvalue.Value) *Builder {
	switch attributeName {
	case AttrNameKind:
		return b.Kind(Kind(value.StringValue()))
	case AttrNameKey:
		return b.Key(value.StringValue())
	case AttrNameName:
		return b.OptName(value.AsOptionalString())
	case AttrNameSecondary:
		return b.OptSecondary(value.AsOptionalString())
	case AttrNameTransient:
		return b.Transient(value.BoolValue())
	default:
		if value.IsNull() {
			if _, found := b.attributes[attributeName]; !found {
				return b
			}
		}
		if b.attributes == nil {
			b.attributes = make(map[string]ldvalue.Value)
		} else if b.attributesCopyOnWrite {
			// See note in Build() on ___CopyOnWrite
			copied := make(map[string]ldvalue.Value, len(b.attributes)+1)
			for k, v := range b.attributes {
				copied[k] = v
			}
			b.attributes = copied
			b.attributesCopyOnWrite = false
		}
		if value.IsNull() {
			delete(b.attributes, attributeName)
		} else {
			b.attributes[attributeName] = value
		}
		return b
	}
}

// Secondary sets a secondary key attribute for the Context.
//
// This affects feature flag targeting
// (https://docs.launchdarkly.com/home/flags/targeting-users#targeting-rules-based-on-user-attributes)
// as follows: if you have chosen to bucket users by a specific attribute, the secondary key (if set)
// is used to further distinguish between users who are otherwise identical according to that attribute.
//
// Setting this value to an empty string is not the same as leaving it unset. If you need to clear this
// attribute to a "no value" state, use OptSecondary().
func (b *Builder) Secondary(value string) *Builder {
	return b.OptSecondary(ldvalue.NewOptionalString(value))
}

// OptSecondary sets a secondary key attribute for the Context.
//
// Calling b.OptSecondary(ldvalue.NewOptionalString("x")) is equivalent to b.Secondary("x"), but since it uses
// the OptionalString type, it also allows clearing a previously set name with
// b.OptSecondary(ldvalue.OptionalString{}).
func (b *Builder) OptSecondary(value ldvalue.OptionalString) *Builder {
	b.secondary = value
	return b
}

// Transient sets whether the Context is only intended for flag evaluations and should not be indexed by
// LaunchDarkly.
//
// The default value is false. False means that this Context represents an entity such as a user that you
// want to be able to see on the LaunchDarkly dashboard.
//
// Setting Transient to true excludes this Context from the database that is used by the dashboard. It does
// not exclude it from analytics event data, so it is not the same as making attributes private; all
// non-private attributes will still be included in events and data export.
func (b *Builder) Transient(value bool) *Builder {
	b.transient = value
	return b
}

// Private designates any number of Context attributes as private: that is, their values will not be sent
// to LaunchDarkly.
//
// (TKTK: possibly move some of this conceptual information to a non-platform-specific docs page and/or
// have docs team copyedit it here)
//
// Each parameter can be a simple attribute name, such as "email". Or, it can be a slash-delimited path in
// the JSON-Pointer-like syntax that LaunchDarkly uses to reference properties within a JSON object. For
// instance, if it is "/address/street", and if the value of the attribute "address" is a JSON object, then
// the "street" property within that object would be treated as private but other properties within "address"
// would not. However, numeric array index references such as "/addresses/0/street" do not work for this
// purpose and will be ignored.
//
// This action only affects analytics events that involve this particular Context. To mark some (or all)
// Context attributes as private for all users, use the overall event configuration for the SDK.
//
// The attributes "kind" and "key", and the metadata properties set by Secondary() and Transient(),
// cannot be made private.
//
// In this example, firstName is marked as private, but lastName is not:
//
//     c := ldcontext.NewBuilder("org", "my-key").
//         SetString("firstName", "Pierre").
//         SetString("lastName", "Menard").
//	       Private("firstName").
//         Build()
func (b *Builder) Private(attrRefStrings ...string) *Builder {
	refs := make([]AttrRef, 0, 20) // arbitrary capacity that's likely greater than needed, to preallocate on stack
	for _, s := range attrRefStrings {
		refs = append(refs, NewAttrRef(s))
	}
	return b.PrivateRef(refs...)
}

// PrivateRef is equivalent to Private but uses the AttrRef type. It designates any number of
// Context attributes as private: that is, their values will not be sent to LaunchDarkly.
//
// The difference between PrivateRef and Private is simply a minor optimization: Private parses
// each attribute reference when it is called (in case the reference uses a slash-delimited format),
// but PrivateRef uses existing AttrRef values. The overhead of parsing is fairly minor, but if you
// are using any slash-delimited attribute references and you would like to keep the overhead of
// constructing a user to the absolute minimum, you may wish to use PrivateRef instead of Private.
// and precompute these references.
//
// In this example, firstName is marked as private, but lastName is not:
//
//     privateStreetAttr := ldcontext.NewAttrRef("/address/street")
//     c := ldcontext.NewBuilder("org", "my-key").
//         SetString("firstName", "Pierre").
//         SetString("lastName", "Menard").
//	       PrivateRef(privateStreetAttr).
//         Build()
func (b *Builder) PrivateRef(attrRefs ...AttrRef) *Builder {
	if b.privateAttrs == nil {
		b.privateAttrs = make([]AttrRef, 0, len(attrRefs))
	} else if b.privateCopyOnWrite {
		// See note in Build() on ___CopyOnWrite
		b.privateAttrs = append([]AttrRef(nil), b.privateAttrs...)
		b.privateCopyOnWrite = false
	}
	b.privateAttrs = append(b.privateAttrs, attrRefs...)
	return b
}

// RemovePrivate removes any private attribute references previously added with AddPrivate or AddPrivateRef
// that exactly match that of any of the specified attribute references.
func (b *Builder) RemovePrivate(attrRefStrings ...string) *Builder {
	refs := make([]AttrRef, 0, 20) // arbitrary capacity that's likely greater than needed, to preallocate on stack
	for _, s := range attrRefStrings {
		refs = append(refs, NewAttrRef(s))
	}
	return b.RemovePrivateRef(refs...)
}

// RemovePrivateRef removes any private attribute references previously added with AddPrivate or
// AddPrivateRef that exactly match that of any of the specified attribute references.
func (b *Builder) RemovePrivateRef(attrRefs ...AttrRef) *Builder {
	if b.privateCopyOnWrite {
		// See note in Build() on ___CopyOnWrite
		b.privateAttrs = append([]AttrRef(nil), b.privateAttrs...)
		b.privateCopyOnWrite = false
	}
	for _, attrRefToRemove := range attrRefs {
		for i := 0; i < len(b.privateAttrs); i++ {
			if b.privateAttrs[i].String() == attrRefToRemove.String() {
				b.privateAttrs = append(b.privateAttrs[0:i], b.privateAttrs[i+1:]...)
				i--
			}
		}
	}
	return b
}

func (b *Builder) copyFrom(fromContext Context) {
	if fromContext.Multiple() {
		return
	}
	b.kind = fromContext.kind
	b.key = fromContext.key
	b.name = fromContext.name
	b.secondary = fromContext.secondary
	b.transient = fromContext.transient
	b.attributes = fromContext.attributes
	b.attributesCopyOnWrite = true
	b.privateAttrs = fromContext.privateAttrs
	b.privateCopyOnWrite = true
}

func makeFullyQualifiedKeySingleKind(kind Kind, key string, omitDefaultKind bool) string {
	// Per the users-to-contexts specification, the fully-qualified key for a single-kind context is:
	// - equal to the regular "key" property, if the kind is "user" (a.k.a. DefaultKind)
	// - or, for any other kind, it's the kind plus ":" plus the result of URL-encoding the "key"
	// property (the URL-encoding is to avoid ambiguity if the key contains colons).
	if omitDefaultKind && kind == DefaultKind {
		return key
	}
	return fmt.Sprintf("%s:%s", kind, url.PathEscape(key))
}
