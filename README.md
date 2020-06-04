# LaunchDarkly Go SDK Core Packages

[![Circle CI](https://circleci.com/gh/launchdarkly/go-sdk-common.svg?style=svg)](https://circleci.com/gh/launchdarkly/go-sdk-common)

## This is a prerelease branch

The `v2` branch currently contains prerelease code to support development of Go SDK 5.0.0. For the source code of the latest release that is used by Go SDK 4.x, see the [`v1` branch](https://github.com/launchdarkly/go-sdk-common/tree/v1).

## Overview

This repository contains packages that are shared between the [LaunchDarkly Go SDK](https://github.com/launchdarkly/go-server-sdk) and other LaunchDarkly Go components.

Applications using the LaunchDarkly Go SDK will usually not need to reference these packages directly. If you do (for instance, if you are using the SDK's JSONVariation method, which returns the type `Value` from the `ldvalue` package), you should import the same major version of the package that is imported by the SDK. See the SDK documentation for more details.

Note that the base import path is `gopkg.in/launchdarkly/go-sdk-common.v2` (to ensure that you receive the latest release of major version 2.x), not `github.com/launchdarkly/go-sdk-common`. Also, unlike `go-server-sdk` this does not have `server` in the name, because nothing in this repository is specific to the LaunchDarkly server-side model; it could be used in a client-side context.

## Supported Go versions

This version of the project has been tested with Go 1.13 through 1.14.

## Learn more

Check out our [documentation](http://docs.launchdarkly.com) for in-depth instructions on configuring and using LaunchDarkly. You can also head straight to the [complete reference guide for the Go SDK](http://docs.launchdarkly.com/docs/go-sdk-reference), or the [generated API documentation](https://godoc.org/gopkg.in/launchdarkly/go-sdk-common.v2) for this project.

## Contributing

We encourage pull requests and other contributions from the community. Check out our [contributing guidelines](CONTRIBUTING.md) for instructions on how to contribute to this SDK.

## About LaunchDarkly

* LaunchDarkly is a continuous delivery platform that provides feature flags as a service and allows developers to iterate quickly and safely. We allow you to easily flag your features and manage them from the LaunchDarkly dashboard.  With LaunchDarkly, you can:
    * Roll out a new feature to a subset of your users (like a group of users who opt-in to a beta tester group), gathering feedback and bug reports from real-world use cases.
    * Gradually roll out a feature to an increasing percentage of users, and track the effect that the feature has on key metrics (for instance, how likely is a user to complete a purchase if they have feature A versus feature B?).
    * Turn off a feature that you realize is causing performance problems in production, without needing to re-deploy, or even restart the application with a changed configuration file.
    * Grant access to certain features based on user attributes, like payment plan (eg: users on the ‘gold’ plan get access to more features than users in the ‘silver’ plan). Disable parts of your application to facilitate maintenance, without taking everything offline.
* LaunchDarkly provides feature flag SDKs for a wide variety of languages and technologies. Check out [our documentation](https://docs.launchdarkly.com/docs) for a complete list.
* Explore LaunchDarkly
    * [launchdarkly.com](https://www.launchdarkly.com/ "LaunchDarkly Main Website") for more information
    * [docs.launchdarkly.com](https://docs.launchdarkly.com/  "LaunchDarkly Documentation") for our documentation and SDK reference guides
    * [apidocs.launchdarkly.com](https://apidocs.launchdarkly.com/  "LaunchDarkly API Documentation") for our API documentation
    * [blog.launchdarkly.com](https://blog.launchdarkly.com/  "LaunchDarkly Blog Documentation") for the latest product updates
    * [Feature Flagging Guide](https://github.com/launchdarkly/featureflags/  "Feature Flagging Guide") for best practices and strategies
