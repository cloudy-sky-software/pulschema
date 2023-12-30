# Pulschema

Pul(umi) schema from API specs.

## What Is This?

This module is a library that can convert an OpenAPI spec to a Pulumi schema spec.
From there, using Pulumi's codegen tools, one can generate the necessary language
SDKs for a provider.

## Features

-   Handles discriminated types
-   Handles `AllOf`, `OneOf`, `AnyOf`
-   Creates a metadata map for resource type tokens that map to CRUD operations
-   Generates schema for Pulumi functions, aka invokes, from `GET` methods
-   Maps path params as required inputs in the resource schema for easier mapping of inputs
    to HTTP requests

## OpenAPI Conformance

This library does not convert OpenAPI specs without certain required modifications.
That is, you'll need to standardize your OpenAPI spec with the following rules.
This is required since the OpenaPI docs created by cloud providers aren't always perfect.

**NEW**: Refer to the [conformance repo](https://github.com/cloudy-sky-software/cloud-provider-api-spec) for the rules related to the OpenAPI spec.

## Credits

This library would not be possible without these wonderful creations.

-   https://github.com/getkin/kin-openapi - Used by the core of this library to parse and walk-through the OpenAPI doc.
-   https://github.com/pulumi/pulumi-aws-native - Served as an example of a native Pulumi provider that has solved some problems.
-   https://github.com/pulumi/pulumi-kubernetes - Served as a source for how to approach the conversion from OpenAPI to
    Pulumi schema.
