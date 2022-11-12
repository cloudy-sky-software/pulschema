# Pulschema

Pul(umi) schema from API specs.

## What Is This?

This module is a library that can convert an OpenAPI spec to a Pulumi schema spec.
From there, using Pulumi's codegen tools, one can generate the necessary language
SDKs for a provider.

## Features

-   Handles discriminated types
-   Handles AllOf, OneOf, AnyOf
-   Creates a metadata map for resource type tokens that map to CRUD operations
-   Generates schema for Pulumi functions, aka invokes, from `GET` methods
-   Maps path params as required inputs in the resource schema for easier mapping of inputs
    to HTTP requests

## Prerequisites

This library does not convert OpenAPI specs without certain required modifications.
That is, you'll need to standardize your OpenAPI spec with the following rules.
This is required since the OpenaPI docs created by cloud providers aren't always perfect.

### Resource names

Set the `operationId` of an endpoint path's request method or the `title` property of the
request body schema type.

### Path params

If a resource is a child resource linked by the ID of a parent resource, then ensure that the
path param that represents the parent resource's ID is not called just `id`. The `id` path param
should always only be used by the primary resource that the endpoint serves.

For example, in the endpoint `/servers/{server_id}/volumes/{id}`, `servers` is the parent resource
for `volumes`. The `{server_id}` path param is automatically added as a required input when generating
the resource schema for a `volume` resource. And this endpoint in particular is to "get" a single volume
belonging to a server. Therefore, the volume is the primary resource in this case and the server is the
parent resource.

### Resource operations

To map an endpoint to a "creatable" resource, the endpoint path must have a `POST` request method under it.
Such resources can have `PATCH`, `GET` and `DELETE` endpoints that takes an `id` path param as the last path
param, since that uniquely identifies a specific resource.

### Enums

Make sure that properties are pointing to the correct enum type ref. Avoid inline enums if you can.

### Input/output properties

`pulschema` attempts to gather input and output properties for a resource based on the `readOnly` property.
So be sure to set properties as `readOnly: true` if it should not be supplied as an input during resource
creation.

## Credits

This library would not be possible without these wonderful creations.

-   https://github.com/getkin/kin-openapi - Used by the core of this library to parse and walk-through the OpenAPI doc.
-   https://github.com/pulumi/pulumi-aws-native - Served as an example of a native Pulumi provider that has solved some problems.
-   https://github.com/pulumi/pulumi-kubernetes - Served as a source for how to approach the conversion from OpenAPI to
    Pulumi schema.
