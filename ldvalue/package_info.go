// Package ldvalue provides the LaunchDarkly SDK's general value type, [Value]. LaunchDarkly
// supports the standard JSON data types of null, boolean, number, string, array, and object (map), for
// any feature flag variation or context attribute. The [Value] type can contain any of these.
//
// This package also provides several helper types:
//   - [OptionalBool], [OptionalInt], and [OptionalString], which are safer alternatives to using
//     pointers for values.
//   - [ValueArray] and [ValueMap], which provide immutable representations of JSON arrays and objects.
//
// All value types in this package support several kinds of marshaling/unmarshaling, as follows:
//
// # JSON conversion with MarshalJSON and UnmarshalJSON
//
// All value types in this package have MarshalJSON() and UnmarshalJSON() methods, so they can be used
// with the Marshal and Unmarshal functions in the [encoding/json] package. The result of JSON
// conversions depends on the type; see MarshalJSON() and UnmarshalJSON() for each type.
//
// They also have a convenience method, JSONString(), that is equivalent to calling
// [encoding/json.Marshal]() and then casting to a string.
//
// # String conversion with String method
//
// All value types in this package have a String() method, conforming to the [fmt.Stringer] interface.
// This is a human-readable string representation whose format depends on the type; see String() for
// each type.
//
// # Text conversion with TextMarshaler and TextUnmarshaler methods
//
// All value types in this package have MarshalText() and UnmarshalText() methods allowing them to be
// used with any packages that support the [encoding.TextMarshaler] and [encoding.TextUnmarshaler]
// interfaces, such as gcfg. The format of this representation depends on the type, see
// MarshalText() and UnmarshalText() for each type.
package ldvalue
