# Change log

All notable changes to the project will be documented in this file. This project adheres to [Semantic Versioning](http://semver.org).

## [2.0.0-beta.1] - 2020-07-01

Initial beta release of the 2.x major version that will be used with versions 5.0.0 and above of the LaunchDarkly Server-Side SDK for Go.

The main difference from 1.x is that it adds the `lduser` package, which provides the `User` and `UserBuilder` types that were formerly in the main package of the Go SDK repository. The user types have been broken out because they represent the standard user property schema used by the whole LaunchDarkly system, and can therefore be used outside of the SDK. Similarly, the new `ldreason` package contains `EvaluationDetail` and `EvaluationReason`, which are used in the SDK but are also part of the standard LaunchDarkly event schema.

## [1.0.0] - 2020-02-03
Initial release. This will be used in versions 4.16.0 and above of the LaunchDarkly Server-Side SDK for Go.
