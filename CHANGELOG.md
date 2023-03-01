# Change log

All notable changes to the project will be documented in this file. This project adheres to [Semantic Versioning](http://semver.org).

## [3.0.1] - 2023-03-01
### Fixed:
- Fixed unmarshaling bug in easyJSON implementation when `privateAttributes` or `redactedAttributes` were encountered in Context `_meta` attribute, but not expected.

## [3.0.0] - 2022-11-30
This major version release of `go-sdk-common` corresponds to the upcoming v6.0.0 release of the LaunchDarkly Go SDK (`go-server-sdk`), and cannot be used with earlier SDK versions.

### Added:
- The new package `ldcontext` with the types `Context` and `Kind` defines the new "context" model. "Contexts" are a replacement for the earlier concept of "users"; they can be populated with attributes in more or less the same way as before, but they also support new behaviors. More information about these features will be included in the release notes for the v6.0.0 SDK release.
- The new package `ldattr` defines the attribute reference syntax, for referencing subproperties of JSON objects in flag evaluations or private attribute configuration. Applications normally will not need to reference this package.

### Changed:
- The minimum Go version is now 1.18.
- The SDK packages now use regular Go module import paths rather than `gopkg.in` paths: `gopkg.in/launchdarkly/go-sdk-common.v2` is replaced by `github.com/launchdarkly/go-sdk-common/v3`.
- The type `lduser.User` has been redefined to be an alias for `ldcontext.Context`. This means that existing application code referencing `lduser.User` can still work as long as it is treating the user as an opaque value, and not calling methods on it that were specific to that type.
- `lduser.NewUser` and `lduser.UserBuilder` now create an instance of `Context` instead of `User`. This is as a convenience so that any code that was previously using these methods to construct a user, but did _not_ reference the `User` type directly for the result, may still be usable without changes. It is still preferable to use the new constructors and builders for `Context`.
- The `Secondary` attribute which existed in `User` does not exist in `Context` and is no longer a supported feature.
- It was previously allowable to set a user key to an empty string. In the new context model, the key is not allowed to be empty. Trying to use an empty key will cause evaluations to fail and return the default value.
- If you were using JSON serialization to produce a representation of a `User`, the new type `Context` uses a different JSON schema, so any code that reads the JSON will need to be adjusted. If you are passing the JSON to other code that uses LaunchDarkly SDKs, make sure you have updated all SDKs to versions that use the new context model. (However, _unmarshaling_ a `Context` from JSON data will still work correctly even if the JSON is in the old user format.)

### Removed:
- Removed the `Secondary` meta-attribute in `lduser.UserBuilder`.

## [2.5.1] - 2022-06-30
### Changed:
- If you create an `ldvalue.Value` with the `ldvalue.Raw(json.RawMessage)` constructor, and you pass a zero-length or nil value to the constructor, and then encode the `Value` to JSON with `json.Marshal` or an equivalent method, the JSON output will now be `null` (that is, the literal characters `null` representing a JSON null value). Previously it would have been a zero-length string, which is not valid as the JSON encoding of any value and could cause the SDK to output a malformed JSON document if the document contained such a value.

## [2.5.0] - 2021-10-14
_This release was unintended and can be ignored. It contains no code changes, only changes to the CI build._

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
