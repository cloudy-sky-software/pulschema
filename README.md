[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/cloudy-sky-software/pulschema)

# Pulschema

Pul(umi) schema from API specs. Learn how to generate a provider using this library: https://buttondown.email/css-blog/archive/create-a-pulumi-provider-from-openapi-spec/. This library is the first part required to fully implement a native Pulumi provider purely based on OpenAPI specs. To use the Pulumi schema successfully, you
will need to construct HTTP requests based on the resource inputs given by the user. You should consider
using [`pulumi-provider-framework`](https://github.com/cloudy-sky-software/pulumi-provider-framework) for that.

Use the [template repo](https://github.com/cloudy-sky-software/pulumi-provider-template) to get started quickly.

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
-   Maps query params to a `queryParams` input on resources and functions (see below)

### Query Params

Every resource and function gets an optional `queryParams` input. Each endpoint gets its own
type with a property for every `in: query` param defined in the spec (path-level params included),
plus an `additionalParams` map of strings for arbitrary query params that aren't in the spec.

-   **Functions** (`get*`/`list*`): `queryParams` refers to a `<FuncName>QueryParams` type, for example
    `GetWidgetQueryParams`.
-   **Resources**: `queryParams` refers to a `<Resource>QueryParams` type that has one property per CRUD
    operation: `create`, `read`, `update` (PATCH), `put` (PUT) and `delete`. Each of these refers to a
    `<Resource><Op>QueryParams` type, for example `WidgetDeleteQueryParams`. `queryParams` is also an output
    property, so providers can read it from state during Read and Delete.

`queryParams` (and the operation property) is required only if the endpoint has a required query param.

Query param names are converted to camelCase in the schema. Like other properties, renamed query
params are recorded in the `sdkToApiNameMap` and `apiToSdkNameMap` metadata. Providers use them to
map the properties of `queryParams` back to the API's query param names when building the request URL.

## OpenAPI Conformance

This library does not convert OpenAPI specs without certain required modifications.
That is, you'll need to standardize your OpenAPI spec with the following rules.
This is required since the OpenaPI docs created by cloud providers aren't always perfect.

**NEW**: Refer to the [conformance repo](https://github.com/cloudy-sky-software/cloud-provider-api-conformance) for the rules related to the OpenAPI spec.

## Development

- Run `make ensure` to restore/cleanup dependencies.
- Run `make lint` to run `golangci-lint` rules.
- Run `make test` to run all the tests.

## Credits

This library would not be possible without these wonderful creations.

- https://github.com/getkin/kin-openapi - Used by the core of this library to parse and walk-through the OpenAPI doc.
- https://github.com/pulumi/pulumi-aws-native - Served as an example of a native Pulumi provider that has solved some problems.
- https://github.com/pulumi/pulumi-kubernetes - Served as a source for how to approach the conversion from OpenAPI to
  Pulumi schema.
