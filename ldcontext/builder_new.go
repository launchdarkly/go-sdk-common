package ldcontext

import (
	"sort"

	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/lderrors"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
)

// ContextBuilder is a mutable object that uses the builder pattern to specify properties for a Context.
//
// Obtain an instance of ContextBuilder by calling [NewContextBuilder]. Then, call [ContextBuilder.Kind]
// to start defining a context with a specific Kind. This returns a [KindBuilder] that can set properties
// for that Kind. KindBuilders support chaining: call Kind again to start defining a different Kind. This
// builds a multi-context. Finally, call Build on any KindBuilder to produce the immutable [Context].
//
//	context := ldcontext.NewContextBuilder().
//		Kind("user", "user-key").
//		Name("my-name").
//		SetString("country", "us").
//		Kind("org", "org-key").
//		SetString("type", "enterprise").
//		Build()
//
// A ContextBuilder and its KindBuilders should not be accessed by multiple goroutines at once.
// Once you have called [ContextBuilder.Build], the resulting Context is immutable and safe to
// use from multiple goroutines.
//
// # Context attributes
//
// There are several built-in attribute names with special meaning in LaunchDarkly, and
// restrictions on the type of their value. These have their own builder methods: see
// [KindBuilder.Key], [KindBuilder.Name], [KindBuilder.Anonymous], and [KindBuilder.SetValue].
//
// You may also set any number of other attributes with whatever names are useful for your
// application (subject to validation constraints; see [KindBuilder.SetValue] for rules regarding
// attribute names). These attributes can have any data type that is supported in JSON:
// boolean, number, string, array, or object.
//
// # Setting attributes with simple value types
//
// For convenience, there are setter methods for simple types:
//
//	context := ldcontext.NewContextBuilder().
//		Kind("user", "user-key").
//		SetBool("a", true).    // this attribute has a boolean value
//		SetString("b", "xyz"). // this attribute has a string value
//		SetInt("c", 3).        // this attribute has an integer numeric value
//		SetFloat64("d", 4.5).  // this attribute has a floating-point numeric value
//		Build()
//
// # Setting attributes with complex value types
//
// JSON arrays and objects are represented by the [ldvalue.Value] type. The [KindBuilder.SetValue]
// method takes a value of this type.
//
// The [ldvalue] package provides several ways to construct values of each type. Here are some examples;
// for more information, see [ldvalue.Value].
//
//	context := ldcontext.NewContextBuilder().
//		Kind("user", "user-key").
//		SetValue("arrayAttr1",
//			ldvalue.ArrayOf(ldvalue.String("a"), ldvalue.String("b"))).
//		SetValue("arrayAttr2",
//			ldvalue.CopyArbitraryValue([]string{"a", "b"})).
//		SetValue("objectAttr1",
//			ldvalue.ObjectBuild().SetString("color", "green").Build()).
//		SetValue("objectAttr2",
//			ldvalue.FromJSONMarshal(MyStructType{Color: "green"})).
//		Build()
//
// Arrays and objects have special meanings in LaunchDarkly flag evaluation:
//   - An array of values means "try to match any of these values to the targeting rule."
//   - An object allows you to match a property within the object to the targeting rule. For instance,
//     in the example above, a targeting rule could reference /objectAttr1/color to match the value
//     "green". Nested property references like /objectAttr1/address/street are allowed if a property
//     contains another JSON object.
//
// # Private attributes
//
// You may designate certain attributes, or values within them, as "private", meaning that their
// values are not included in analytics data sent to LaunchDarkly. See [KindBuilder.Private].
//
//	context := ldcontext.NewContextBuilder().
//		Kind("user", "user-key").
//		SetString("email", "test@example.com").
//		Private("email").
//		Build()
type ContextBuilder struct {
	singleBuilders map[Kind]*Builder
}

// KindBuilder is a mutable object that uses the builder pattern to specify properties for a context
// of a specific Kind within a multi-context.
//
// See [ContextBuilder] for more about how to use this type.
type KindBuilder struct {
	*ContextBuilder
	singleBuilder *Builder
}

// NewContextBuilder creates a ContextBuilder for building a single or multi-context.
//
// See [ContextBuilder] for more information.
func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{singleBuilders: make(map[Kind]*Builder)}
}

// NewContextBuilderFromContext creates a ContextBuilder whose properties are the same as an existing
// Context. You may then change the ContextBuilder's state in any way and call [ContextBuilder.Build]
// to create a new independent [Context].
func NewContextBuilderFromContext(fromContext Context) *ContextBuilder {
	cb := &ContextBuilder{singleBuilders: make(map[Kind]*Builder)}
	for _, c := range fromContext.GetAllIndividualContexts(nil) {
		b := &Builder{}
		b.copyFrom(fromContext)
		cb.singleBuilders[c.Kind()] = b
	}
	return cb
}

// Add adds one or more contexts to a ContextBuilder. Only one context of each kind is allowed.
// If multiple contexts are added with the same kind, the last one fully replaces the previous ones.
func (cb *ContextBuilder) Add(contexts ...Context) *ContextBuilder {
	for _, c := range contexts {
		if c.Multiple() {
			for _, ic := range c.multiContexts {
				cb.Add(ic)
			}
		} else {
			b := &Builder{}
			b.copyFrom(c)
			cb.singleBuilders[c.Kind()] = b
		}
	}
	return cb
}

// Kind starts building a new context with the specified kind and key in the current
// ContextBuilder, returning a [KindBuilder] that can set properties for that inner context.
// If a context with that kind has already been defined, its key is set to the new value,
// all other properties are preserved, and the returned KindBuilder can be used to replace or
// add additional properties.
//
// Every [Context] has a kind. Setting it to an empty string is equivalent to the default kind of
// "user". This value is case-sensitive. Validation rules are as follows:
//
//   - It may only contain letters, numbers, and the characters ".", "_", and "-".
//   - It cannot equal the literal string "kind".
//   - It cannot equal the literal string "multi" ([MultiKind]).
//
// The key parameter is the value of the context's "key" attribute. It is always a string.
// There are no restrictions on its value except that it cannot be empty.
//
// If either value is invalid at the time [ContextBuilder.Build] is called, you will receive an
// invalid Context whose [Context.Err] value will describe the problem.
func (cb *ContextBuilder) Kind(kind Kind, key string) *KindBuilder {
	if kind == "" {
		kind = DefaultKind
	}
	singleBuilder, ok := cb.singleBuilders[kind]
	if !ok {
		b := &Builder{}
		singleBuilder = b.Key(key).Kind(kind)
		cb.singleBuilders[kind] = singleBuilder
	}
	singleBuilder.Key(key)
	return &KindBuilder{ContextBuilder: cb, singleBuilder: singleBuilder}
}

// Build creates a Context from the current ContextBuilder and KindBuilder properties.
//
// The [Context] is immutable and will not be affected by any subsequent actions on the ContextBuilder.
//
// It is possible for a ContextBuilder to represent an invalid state. Instead of returning two
// values (Context, error), Build always returns a Context and you can call [Context.Err]
// to see if it has an error. See [Context.Err] for more information about invalid Context
// conditions. Using a single-return-value syntax is more convenient for application code, since
// in normal usage an application will never build an invalid Context. If you pass an invalid
// Context to an SDK method, the SDK will detect this and will generally log a description
// of the error.
//
// You may call [ContextBuilder.TryBuild] instead of Build if you prefer to use two-value return
// semantics, but the validation behavior is the same for both.
func (cb *ContextBuilder) Build() Context {
	if len(cb.singleBuilders) == 0 {
		return Context{defined: true, err: lderrors.ErrContextKindMultiWithNoKinds{}}
	}

	if len(cb.singleBuilders) == 1 {
		// If only one context kind was added, the result is just the same as building that one context
		for _, singleBuilder := range cb.singleBuilders {
			return singleBuilder.Build()
		}
		panic("impossible")
	}

	// Sort the list by kind - this makes our output deterministic and will also be important when we
	// compute a fully qualified key.
	kinds := make([]string, 0, len(cb.singleBuilders))
	for kind := range cb.singleBuilders {
		kinds = append(kinds, string(kind))
	}
	sort.Strings(kinds)

	ret := Context{
		defined:       true,
		kind:          MultiKind,
		multiContexts: make([]Context, 0, len(cb.singleBuilders)),
	}

	var individualErrors map[string]error
	for _, kind := range kinds {
		singleBuilder := cb.singleBuilders[Kind(kind)]
		ctx, err := singleBuilder.TryBuild()
		if err != nil {
			if individualErrors == nil {
				individualErrors = make(map[string]error)
			}
			individualErrors[kind] = err
			continue
		}
		ret.multiContexts = append(ret.multiContexts, ctx)
	}
	if len(individualErrors) != 0 {
		ret.err = lderrors.ErrContextPerKindErrors{Errors: individualErrors}
		return ret
	}

	// Fully-qualified key for multi-context is defined as "kind1:key1:kind2:key2" etc., where kinds are in
	// alphabetical order (we have already sorted them above) and keys are URL-encoded. In this case we
	// do _not_ omit a default kind of "user".
	for _, c := range ret.multiContexts {
		if ret.fullyQualifiedKey != "" {
			ret.fullyQualifiedKey += ":"
		}
		ret.fullyQualifiedKey += makeFullyQualifiedKeySingleKind(c.kind, c.key, false)
	}

	return ret
}

// TryBuild is an alternative to Build that returns any validation errors as a second value.
//
// As described in [ContextBuilder.Build], there are several ways the state of a [Context] could
// be invalid. Since in normal usage it is possible to be confident that these will not occur,
// the Build method is designed for convenient use within expressions by returning a single
// Context value, and any validation problems are contained within that value where they can be
// detected by calling the context's [Context.Err] method. But, if you prefer to use the
// two-value pattern that is common in Go, you can call TryBuild instead:
//
//	c, err := ldcontext.NewContextBuilder().
//		Kind("user", "user-key").
//		Name("my-name").
//		TryBuild()
//	if err != nil {
//		// do whatever is appropriate if building the context failed
//	}
//
// The two return values are the same as to 1. the Context that would be returned by Build(),
// and 2. the result of calling [Context.Err] on that Context. So, the above example is exactly
// equivalent to:
//
//	c := ldcontext.NewContextBuilder().
//		Kind("user", "user-key").
//		Name("my-name").
//		Build()
//	if c.Err() != nil {
//		// do whatever is appropriate if building the context failed
//	}
//
// Note that unlike some Go methods where the first return value is normally an
// uninitialized zero value if the error is non-nil, the Context returned by TryBuild in case
// of an error is not completely uninitialized: it does contain the error information as well,
// so that if it is mistakenly passed to an SDK method, the SDK can tell what the error was.
func (cb *ContextBuilder) TryBuild() (Context, error) {
	c := cb.Build()
	return c, c.Err()
}

// Key sets the Context's key attribute.
//
// Every [Context] has a key, which is always a string. There are no restrictions on its value except
// that it cannot be empty.
//
// The key attribute can be referenced by flag rules, flag target lists, and segments.
//
// If the key is empty at the time [ContextBuilder.Build] is called, you will receive an invalid Context
// whose [Context.Err] value will describe the problem.
func (kb *KindBuilder) Key(key string) *KindBuilder {
	kb.singleBuilder.Key(key)
	return kb
}

// Name sets the Context's name attribute.
//
// This attribute is optional. It has the following special rules:
//   - Unlike most other attributes, it is always a string if it is specified.
//   - The LaunchDarkly dashboard treats this attribute as the preferred display name for contexts.
func (kb *KindBuilder) Name(name string) *KindBuilder {
	kb.singleBuilder.Name(name)
	return kb
}

// OptName sets or clears the Context's name attribute.
//
// Calling b.OptName(ldvalue.NewOptionalString("x")) is equivalent to b.Name("x"), but since it uses
// the OptionalString type, it also allows clearing a previously set name with
// b.OptName(ldvalue.OptionalString{}).
func (kb *KindBuilder) OptName(name ldvalue.OptionalString) *KindBuilder {
	kb.singleBuilder.OptName(name)
	return kb
}

// SetBool sets an attribute to a boolean value.
//
// For rules regarding attribute names and values, see [KindBuilder.SetValue]. This method is exactly
// equivalent to calling b.SetValue(attributeName, ldvalue.Bool(value)).
func (kb *KindBuilder) SetBool(attributeName string, value bool) *KindBuilder {
	kb.singleBuilder.SetBool(attributeName, value)
	return kb
}

// SetFloat64 sets an attribute to a float64 numeric value.
//
// For rules regarding attribute names and values, see [KindBuilder.SetValue]. This method is exactly
// equivalent to calling b.SetValue(attributeName, ldvalue.Float64(value)).
//
// Note: the LaunchDarkly model for feature flags and user attributes is based on JSON types,
// and JSON does not distinguish between integer and floating-point types. Therefore,
// b.SetFloat64(name, float64(1.0)) is exactly equivalent to b.SetInt(name, 1).
func (kb *KindBuilder) SetFloat64(attributeName string, value float64) *KindBuilder {
	kb.singleBuilder.SetFloat64(attributeName, value)
	return kb
}

// SetInt sets an attribute to an int numeric value.
//
// For rules regarding attribute names and values, see [KindBuilder.SetValue]. This method is exactly
// equivalent to calling b.SetValue(attributeName, ldvalue.Int(value)).
//
// Note: the LaunchDarkly model for feature flags and user attributes is based on JSON types,
// and JSON does not distinguish between integer and floating-point types. Therefore,
// b.SetFloat64(name, float64(1.0)) is exactly equivalent to b.SetInt(name, 1).
func (kb *KindBuilder) SetInt(attributeName string, value int) *KindBuilder {
	kb.singleBuilder.SetInt(attributeName, value)
	return kb
}

// SetString sets an attribute to a string value.
//
// For rules regarding attribute names and values, see [KindBuilder.SetValue]. This method is exactly
// equivalent to calling b.SetValue(attributeName, ldvalue.String(value)).
func (kb *KindBuilder) SetString(attributeName string, value string) *KindBuilder {
	kb.singleBuilder.SetString(attributeName, value)
	return kb
}

// SetValue sets the value of any attribute for the Context.
//
// This method uses the [ldvalue.Value] type to represent a value of any JSON type: boolean,
// number, string, array, or object. The [ldvalue] package provides several ways to construct
// values of each type.
//
// The return value is always the same [KindBuilder], for convenience (to allow method chaining).
//
// # Allowable attribute names
//
// The attribute names "kind", "key", "name", and "anonymous" have special meaning in
// LaunchDarkly. You may not set "kind" with SetValue; that must be specified with [ContextBuilder.Kind].
//
// You may set "key", "name", and "anonymous" with SetValue, as an alternative to using the
// methods [KindBuilder.Key], [KindBuilder.Name], and [KindBuilder.Anonymous]. However,
// there are restrictions on the value type: "key" must be a string, "name" must
// be a string or null, and "anonymous" must be a boolean. Any value of an unsupported type
// is ignored (leaving the attribute unchanged).
func (kb *KindBuilder) SetValue(attributeName string, value ldvalue.Value) *KindBuilder {
	_ = kb.TrySetValue(attributeName, value)
	return kb
}

// TrySetValue sets the value of any attribute for the Context.
//
// This is the same as [KindBuilder.SetValue], except that it returns true for success, or false if the
// parameters violated one of the restrictions described for SetValue (for instance,
// attempting to set "key" to a value that was not a string).
func (kb *KindBuilder) TrySetValue(attributeName string, value ldvalue.Value) bool {
	// We mostly defer to kb.singlebuilder.TrySetValue, but we have an extra restriction
	// that you cannot set "kind" with SetValue.
	switch attributeName {
	case ldattr.KindAttr:
		return false
	default:
		return kb.singleBuilder.TrySetValue(attributeName, value)
	}
}

// Anonymous sets whether the Context is only intended for flag evaluations and should not be indexed by
// LaunchDarkly.
//
// The default value is false. False means that this [Context] represents an entity such as a user that you
// want to be able to see on the LaunchDarkly dashboard.
//
// Setting Anonymous to true excludes this Context from the database that is used by the dashboard. It does
// not exclude it from analytics event data, so it is not the same as making attributes private; all
// non-private attributes will still be included in events and data export. There is no limitation on what
// other attributes may be included (so, for instance, Anonymous does not mean there is no [KindBuilder.Name]).
//
// This value is also addressable in evaluations as the attribute name "anonymous". It is always treated as
// a boolean true or false in evaluations; it cannot be null/undefined.
func (kb *KindBuilder) Anonymous(value bool) *KindBuilder {
	kb.singleBuilder.Anonymous(value)
	return kb
}

// Private designates any number of Context attributes, or properties within them, as private: that is,
// their values will not be sent to LaunchDarkly in analytics data.
//
// This action only affects analytics events that involve this particular [Context]. To mark some (or all)
// Context attributes as private for all context, use the overall event configuration for the SDK.
//
// In this example, firstName is marked as private, but lastName is not:
//
//	c := ldcontext.NewContextBuilder().
//		Kind("user", "my-key").
//		SetString("firstName", "Pierre").
//		SetString("lastName", "Menard").
//		Private("firstName").
//		Build()
//
// The attributes "kind", "key", and "anonymous" cannot be made private.
//
// This is a metadata property, rather than an attribute that can be addressed in evaluations: that is,
// a rule clause that references the attribute name "private" will not use this value, but instead will
// use whatever value (if any) you have set for that name with a method such as [KindBuilder.SetString].
//
// # Designating an entire attribute as private
//
// If the parameter is an attribute name such as "email" that does not start with a '/' character, the
// entire attribute is private.
//
// # Designating a property within a JSON object as private
//
// If the parameter starts with a '/' character, it is interpreted as a slash-delimited path to a
// property within a JSON object. The first path component is an attribute name, and each following
// component is a property name.
//
// For instance, suppose that the attribute "address" had the following JSON object value:
// {"street": {"line1": "abc", "line2": "def"}, "city": "ghi"}
//
//   - Calling either Private("address") or Private("/address") would cause the entire "address"
//     attribute to be private.
//   - Calling Private("/address/street") would cause the "street" property to be private, so that
//     only {"city": "ghi"} is included in analytics.
//   - Calling Private("/address/street/line2") would cause only "line2" within "street" to be private,
//     so that {"street": {"line1": "abc"}, "city": "ghi"} is included in analytics.
//
// This syntax deliberately resembles JSON Pointer, but other JSON Pointer features such as array
// indexing are not supported for Private.
//
// If an attribute's actual name starts with a '/' character, you must use the same escaping syntax as
// JSON Pointer: replace "~" with "~0", and "/" with "~1".
func (kb *KindBuilder) Private(attrRefStrings ...string) *KindBuilder {
	kb.singleBuilder.Private(attrRefStrings...)
	return kb
}

// PrivateRef is equivalent to Private, but uses the ldattr.Ref type. It designates any number of
// Context attributes, or properties within them, as private: that is, their values will not be
// sent to LaunchDarkly.
//
// Application code is unlikely to need to use the ldattr.Ref type directly; however, in cases where
// you are constructing Contexts constructed repeatedly with the same set of private attributes, if
// you are also using complex private attribute path references such as "/address/street", converting
// this to an [ldattr.Ref] once and reusing it in many PrivateRef calls is slightly more efficient than
// calling [KindBuilder.Private] (since it does not need to parse the path repeatedly).
func (kb *KindBuilder) PrivateRef(attrRefs ...ldattr.Ref) *KindBuilder {
	kb.singleBuilder.PrivateRef(attrRefs...)
	return kb
}

// RemovePrivate removes any private attribute references previously added with [KindBuilder.Private]
// or [KindBuilder.PrivateRef] that exactly match any of the specified attribute references.
func (kb *KindBuilder) RemovePrivate(attrRefStrings ...string) *KindBuilder {
	kb.singleBuilder.RemovePrivate(attrRefStrings...)
	return kb
}

// RemovePrivateRef removes any private attribute references previously added with [KindBuilder.Private]
// or [KindBuilder.PrivateRef] that exactly match that of any of the specified attribute references.
//
// Application code is unlikely to need to use the [ldattr.Ref] type directly, and can use
// RemovePrivate with a string parameter to accomplish the same thing. This method is mainly for
// use by internal LaunchDarkly SDK and service code which uses ldattr.Ref.
func (kb *KindBuilder) RemovePrivateRef(attrRefs ...ldattr.Ref) *KindBuilder {
	kb.singleBuilder.RemovePrivateRef(attrRefs...)
	return kb
}
