// Copyright 2022, Cloudy Sky Software.

package pkg

import (
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/golang/glog"

	"github.com/pkg/errors"

	"github.com/pulumi/pulumi/pkg/v3/codegen"
	pschema "github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

const (
	// QueryParamsPropName is the name of the input property of
	// resources and functions that holds the query params to
	// send with the API requests.
	QueryParamsPropName = "queryParams"
	// AdditionalQueryParamsPropName is the name of the property
	// of each query params type that holds arbitrary query params
	// not defined in the API spec.
	AdditionalQueryParamsPropName = "additionalParams"
)

// Names of the properties of a resource's queryParams type,
// one for each of the resource's CRUD operations.
const (
	QueryParamsOpCreate = "create"
	QueryParamsOpRead   = "read"
	QueryParamsOpUpdate = "update"
	QueryParamsOpPut    = "put"
	QueryParamsOpDelete = "delete"
)

// mergeParameters returns a new slice containing the path-level
// common parameters and the operation-level parameters. An
// operation-level parameter overrides a path-level parameter
// with the same name and location.
//
// In OpenAPI spec, a single path (eg. /v1/example) can
// have multiple operation types such as GET, POST, PUT,
// DELETE.
func mergeParameters(common, op openapi3.Parameters) openapi3.Parameters {
	type paramKey struct {
		name string
		in   string
	}

	opParams := make(map[paramKey]struct{}, len(op))
	for _, p := range op {
		if p == nil || p.Value == nil {
			continue
		}
		opParams[paramKey{name: p.Value.Name, in: p.Value.In}] = struct{}{}
	}

	merged := make(openapi3.Parameters, 0, len(common)+len(op))
	for _, p := range common {
		if p == nil || p.Value == nil {
			continue
		}
		if _, ok := opParams[paramKey{name: p.Value.Name, in: p.Value.In}]; ok {
			continue
		}
		merged = append(merged, p)
	}

	for _, p := range op {
		if p == nil || p.Value == nil {
			continue
		}
		merged = append(merged, p)
	}

	return merged
}

// genQueryParamsType generates an object type named typeName for the
// query params in params. In addition to a property for each query
// param, the type always has an additionalParams property that can
// hold arbitrary query params not defined in the API spec.
// Returns a ref to the generated type and whether any of the query
// params are required.
func (o *OpenAPIContext) genQueryParamsType(module, typeName string, params openapi3.Parameters) (*pschema.TypeSpec, bool, error) {
	ctx := &resourceContext{
		mod:               module,
		pkg:               o.Pkg,
		resourceName:      typeName,
		openapiComponents: *o.Doc.Components,
		visitedTypes:      o.visitedTypes,
		sdkToAPINameMap:   o.sdkToAPINameMap,
		apiToSDKNameMap:   o.apiToSDKNameMap,
		pathParamMap:      o.pathParamNameMap,
	}

	properties := make(map[string]pschema.PropertySpec)
	required := codegen.NewStringSet()

	for _, param := range params {
		if param == nil || param.Value == nil || param.Value.In != parameterLocationQuery {
			continue
		}

		paramName := param.Value.Name
		sdkName := ToSdkName(paramName)

		if sdkName != paramName {
			addNameOverride(sdkName, paramName, o.sdkToAPINameMap)
			addNameOverride(paramName, sdkName, o.apiToSDKNameMap)
		}

		propSpec := pschema.PropertySpec{
			Description: param.Value.Description,
			TypeSpec:    pschema.TypeSpec{Type: openapi3.TypeString},
		}

		// Params can be content-encoded instead of having a schema,
		// in which case they are treated as plain strings.
		if param.Value.Schema != nil && param.Value.Schema.Value != nil {
			schema := param.Value.Schema.Value
			typeSpec, _, err := ctx.propertyTypeSpec(typeName+ToPascalCase(paramName), *param.Value.Schema)
			if err != nil {
				return nil, false, errors.Wrapf(err, "generating type spec for query param %s", paramName)
			}
			propSpec.TypeSpec = *typeSpec

			if propSpec.Description == "" {
				propSpec.Description = schema.Description
			}
			if schema.Default != nil && !schema.Type.Is(openapi3.TypeArray) {
				propSpec.Default = schema.Default
			}
		}

		properties[sdkName] = propSpec
		if param.Value.Required {
			required.Add(sdkName)
		}
	}

	if _, ok := properties[AdditionalQueryParamsPropName]; ok {
		glog.Warningf("Query params type %s has a query param called %s. "+
			"Arbitrary query params cannot be specified for this endpoint.", typeName, AdditionalQueryParamsPropName)
	} else {
		properties[AdditionalQueryParamsPropName] = pschema.PropertySpec{
			Description: "Additional query params to send with the request that are not defined in the API spec.",
			TypeSpec: pschema.TypeSpec{
				Type:                 openapi3.TypeObject,
				AdditionalProperties: &pschema.TypeSpec{Type: openapi3.TypeString},
			},
		}
	}

	tok := fmt.Sprintf("%s:%s:%s", o.Pkg.Name, module, typeName)
	o.Pkg.Types[tok] = pschema.ComplexTypeSpec{
		ObjectTypeSpec: pschema.ObjectTypeSpec{
			Description: "Query params for the API request.",
			Type:        openapi3.TypeObject,
			Properties:  properties,
			Required:    required.SortedValues(),
		},
	}

	return &pschema.TypeSpec{Ref: typesSchemaRefPrefix + tok}, len(required) > 0, nil
}

// addQueryParamsToResources adds a queryParams input (and output)
// property to each resource gathered from the API spec. The
// queryParams type of a resource has a property for each of its
// CRUD operations, each of which refers to a type with the query
// params of that operation's endpoint.
//
// This must be called after all paths have been processed since
// the endpoints for a resource can be processed in any order.
func (o *OpenAPIContext) addQueryParamsToResources() error {
	toks := make([]string, 0, len(o.resourceCRUDMap))
	for tok := range o.resourceCRUDMap {
		toks = append(toks, tok)
	}
	sort.Strings(toks)

	for _, tok := range toks {
		resourceSpec, ok := o.Pkg.Resources[tok]
		if !ok {
			continue
		}

		if _, ok := resourceSpec.InputProperties[QueryParamsPropName]; ok {
			glog.Warningf("Resource %s already has an input property called %s. Skipping query params for it.", tok, QueryParamsPropName)
			continue
		}
		if _, ok := resourceSpec.Properties[QueryParamsPropName]; ok {
			glog.Warningf("Resource %s already has an output property called %s. Skipping query params for it.", tok, QueryParamsPropName)
			continue
		}

		crud := o.resourceCRUDMap[tok]
		tokParts := strings.Split(tok, ":")
		module, resourceName := tokParts[1], tokParts[2]

		ops := []struct {
			name string
			path *string
			op   func(*openapi3.PathItem) *openapi3.Operation
		}{
			{QueryParamsOpCreate, crud.C, func(p *openapi3.PathItem) *openapi3.Operation {
				// Resources can be created with a PUT request
				// when there is no POST endpoint.
				if p.Post != nil {
					return p.Post
				}
				return p.Put
			}},
			{QueryParamsOpRead, crud.R, func(p *openapi3.PathItem) *openapi3.Operation { return p.Get }},
			{QueryParamsOpUpdate, crud.U, func(p *openapi3.PathItem) *openapi3.Operation { return p.Patch }},
			{QueryParamsOpPut, crud.P, func(p *openapi3.PathItem) *openapi3.Operation { return p.Put }},
			{QueryParamsOpDelete, crud.D, func(p *openapi3.PathItem) *openapi3.Operation { return p.Delete }},
		}

		properties := make(map[string]pschema.PropertySpec)
		required := codegen.NewStringSet()
		for _, op := range ops {
			if op.path == nil {
				continue
			}
			pathItem := o.Doc.Paths.Find(*op.path)
			if pathItem == nil {
				return errors.Errorf("path item for path %s not found", *op.path)
			}
			operation := op.op(pathItem)
			if operation == nil {
				continue
			}

			typeName := resourceName + ToPascalCase(op.name) + "QueryParams"
			typeSpec, hasRequired, err := o.genQueryParamsType(module, typeName, mergeParameters(pathItem.Parameters, operation.Parameters))
			if err != nil {
				return errors.Wrapf(err, "generating %s query params type for resource %s", op.name, tok)
			}

			properties[op.name] = pschema.PropertySpec{
				Description: fmt.Sprintf("Query params for the %s operation.", op.name),
				TypeSpec:    *typeSpec,
			}
			if hasRequired {
				required.Add(op.name)
			}
		}

		if len(properties) == 0 {
			continue
		}

		typeTok := fmt.Sprintf("%s:%s:%s", o.Pkg.Name, module, resourceName+"QueryParams")
		o.Pkg.Types[typeTok] = pschema.ComplexTypeSpec{
			ObjectTypeSpec: pschema.ObjectTypeSpec{
				Description: "Query params for each of the operations of the resource.",
				Type:        openapi3.TypeObject,
				Properties:  properties,
				Required:    required.SortedValues(),
			},
		}

		queryParamsProp := pschema.PropertySpec{
			Description: "Query params to send with the API requests for this resource.",
			TypeSpec:    pschema.TypeSpec{Ref: typesSchemaRefPrefix + typeTok},
		}

		if resourceSpec.InputProperties == nil {
			resourceSpec.InputProperties = make(map[string]pschema.PropertySpec)
		}
		resourceSpec.InputProperties[QueryParamsPropName] = queryParamsProp
		// Providers can read the query params from the
		// state during Read and Delete operations.
		if resourceSpec.Properties == nil {
			resourceSpec.Properties = make(map[string]pschema.PropertySpec)
		}
		resourceSpec.Properties[QueryParamsPropName] = queryParamsProp

		if len(required) > 0 {
			requiredInputs := codegen.NewStringSet(resourceSpec.RequiredInputs...)
			requiredInputs.Add(QueryParamsPropName)
			resourceSpec.RequiredInputs = requiredInputs.SortedValues()
		}

		o.Pkg.Resources[tok] = resourceSpec
	}

	return nil
}
