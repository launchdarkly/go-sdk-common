# Change log

All notable changes to the project will be documented in this file. This project adheres to [Semantic Versioning](http://semver.org).

## [2.5.0] - 2021-10-14
### Added:
- Convenience methods for working with JSON object and array values: `LdValue.Dictionary`, `LdValue.List`, `LdValue.ObjectBuilder.Set`, and `LdValue.ObjectBuilder.Copy`.

## [2.4.0] - 2021-07-19
### Added:
- In `ldreason`, added new optional status information related to the new big segments feature.

## [2.3.0] - 2021-06-17
### Added:
- The SDK now supports the ability to control the proportion of traffic allocation to an experiment. This works in conjunction with a new platform feature now available to early access customers.

## [2.2.3] - 2021-06-03
### Fixed:
- Updated `go-jsonstream` to [v1.0.1](https://github.com/launchdarkly/go-jsonstream/releases/tag/1.0.1) to incorporate a bugfix in JSON number parsing.

## [2.2.2] - 2021-01-15
### Changed:
- Greatly improved the efficiency of deserializing `lduser` and `ldvalue` types from JSON when the `launchdarkly_easyjson` build tag is enabled, by using the EasyJSON API more directly than before. When the build tag is not enabled, these changes have no effect.

## [2.2.1] - 2021-01-04
### Fixed:
- Parsing a `User` from JSON failed if there was a `privateAttributeNames` property whose value was `null`. This has been fixed so that it behaves the same as if the property had a value of `[]` or if it was not present at all.

## [2.2.0] - 2020-12-17
### Added:
- All types that can be converted to or from JSON now have `WriteToJSONWriter` and `ReadFromJSONReader` methods that use the new [`go-jsonstream`](https://github.com/launchdarkly/go-jsonstream) API for greater efficiency, although `json.Marshal` and `json.Unmarshal` still also work.

### Deprecated:
- The `jsonstream` subpackage in `go-sdk-common` is now deprecated in favor of [`go-jsonstream`](https://github.com/launchdarkly/go-jsonstream). The Go SDK no longer uses `jsonstream`, but it is retained here for backward compatibility with any other code that may have been using it. Some types still have `WriteToJSONBuffer` methods for using `jsonstream`; these are also deprecated.


## [2.1.0] - 2020-12-14
### Added:
- `IsDefined()` method for `ldvalue.Value`, `ldreason.EvaluationReason`, and `ldtime.UnixMillisecondTime`.
- `ValueArray` and `ValueMap` types in `ldvalue`, for representing immutable JSON array/object data in contexts where only an array or an object is allowed, as opposed to the more general `ldvalue.Value`. This is mainly used by LaunchDarkly internal components but may be useful elsewhere.

### Changed:
- In `lduser.NewUserBuilderFromUser()`, if the original user had custom attributes and/or private attributes, the map that holds that data now has copy-on-write behavior: that is, the builder will only allocate a new map if you actually make changes to those attributes.

## [2.0.1] - 2020-10-08
### Fixed:
- Trying to unmarshal a JSON `null` value into `lduser.User` now returns an error of type `*json.UnmarshalTypeError`, rather than misleadingly returning `lduser.ErrMissingKey()`.
- Trying to unmarshal a JSON value of the wrong type into `ldvalue.OptionalBool`, `ldvalue.OptionalInt`, or `ldvalue.OptionalString` now returns an error of type `*json.UnmarshalTypeError`, rather than a generic error string from `errors.New()`.
- The error `lduser.ErrMissingKey()` had an empty error string.

## [2.0.0] - 2020-09-18
Initial release of the newer types that will be used in Go SDK 5.0 and above.

### Added:
- Package `ldlog`, which was formerly a subpackage of `go-server-sdk`.
- Package `ldlogtest`, containing test helpers for use with `ldlog`.
- Package `ldreason`, containing `EvaluationReason` and related types that were formerly in `go-server-sdk`.
- Package `ldtime`, containing `UnixMillisecondTime`.
- Package `lduser`, containing `User` and related types that were formerly in `go-server-sdk`.
- Package `jsonstream`, a fast JSON encoding tool that is used internally by the SDK.
- `ldvalue.OptionalString` now implements `encoding.TextMarshaler` and `encoding.TextUnmarshaler`. This is not used by the Go SDK, but can be helpful when using `OptionalString` in other contexts.
- `ldvalue.OptionalBool` and `ldvalue.OptionalInt` are analogous to `ldvalue.OptionalString`, representing values that may be undefined without using pointers. These are used in the Go SDK.

### Changed:
- The minimum Go version is now 1.14.
- The `User` type is now opaque and immutable; there is no direct access to its fields.
- The `User` type no longer uses pointers or `interface{}` internally, decreasing the need for heap allocations.
- Reading a `User` from JSON with `json.Unmarshal` now returns an error if the `key` property is missing or null.
- `EvaluationDetail.VariationIndex` is now an `OptionalInt` rather than an `int`.
- `EvaluationReason` is now a struct.
- This project is now a Go module, although it can still be used from non-module code.

### Removed:
- In `ldvalue`, there are no longer methods for wrapping an existing `interface{}` value in a `Value`.
- All deprecated members of types that were moved here from `go-server-sdk` have been removed.

## [1.0.0] - 2020-02-03
Initial release. This will be used in versions 4.16.0 and above of the LaunchDarkly Server-Side SDK for Go.
