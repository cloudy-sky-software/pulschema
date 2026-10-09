// Copyright 2022, Cloudy Sky Software.

package pkg

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/golang/glog"

	"github.com/pkg/errors"

	"github.com/pulumi/pulumi/pkg/v3/codegen"
	pschema "github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

// Names of the properties of a resource's queryParams type,
// one for each of the resource's CRUD operations.
const (
	queryParamsOpCreate = "create"
	queryParamsOpRead   = "read"
	queryParamsOpUpdate = "update"
	queryParamsOpPut    = "put"
	queryParamsOpDelete = "delete"
)

// resourceOpQueryParams is the query params type
// for a single CRUD operation of a resource.
type resourceOpQueryParams struct {
	typeSpec pschema.TypeSpec
	required bool
}

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
			addNameOverride(paramName, sdkName, o.queryParamNameMap)
		}

		propSpec := pschema.PropertySpec{
			Description: param.Value.Description,
			TypeSpec:    pschema.TypeSpec{Type: typeString},
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

	if _, ok := properties[additionalQueryParamsPropName]; ok {
		glog.Warningf("Query params type %s has a query param called %s. "+
			"Arbitrary query params cannot be specified for this endpoint.", typeName, additionalQueryParamsPropName)
	} else {
		properties[additionalQueryParamsPropName] = pschema.PropertySpec{
			Description: "Additional query params to send with the request that are not defined in the API spec.",
			TypeSpec: pschema.TypeSpec{
				Type:                 typeObject,
				AdditionalProperties: &pschema.TypeSpec{Type: typeString},
			},
		}
	}

	tok := fmt.Sprintf("%s:%s:%s", o.Pkg.Name, module, typeName)
	o.Pkg.Types[tok] = pschema.ComplexTypeSpec{
		ObjectTypeSpec: pschema.ObjectTypeSpec{
			Description: "Query params for the API request.",
			Type:        typeObject,
			Properties:  properties,
			Required:    required.SortedValues(),
		},
	}

	return &pschema.TypeSpec{Ref: typesSchemaRefPrefix + tok}, len(required) > 0, nil
}

// addResourceOpQueryParams generates the query params type for a
// CRUD operation of a resource and records it, so that it can be
// added to the resource's queryParams type.
//
// The endpoints for a resource can be processed in any order. So
// if the resource hasn't been gathered yet, the queryParams property
// is added to it once its create endpoint is processed.
func (o *OpenAPIContext) addResourceOpQueryParams(tok, opName string, params openapi3.Parameters) error {
	tokParts := strings.Split(tok, ":")
	module, resourceName := tokParts[1], tokParts[2]

	typeName := resourceName + ToPascalCase(opName) + "QueryParams"
	typeSpec, hasRequired, err := o.genQueryParamsType(module, typeName, params)
	if err != nil {
		return errors.Wrapf(err, "generating %s query params type for resource %s", opName, tok)
	}

	if _, ok := o.resourceQueryParams[tok]; !ok {
		o.resourceQueryParams[tok] = make(map[string]resourceOpQueryParams)
	}
	o.resourceQueryParams[tok][opName] = resourceOpQueryParams{
		typeSpec: *typeSpec,
		required: hasRequired,
	}

	o.setResourceQueryParamsProp(tok)
	return nil
}

// setResourceQueryParamsProp (re)generates the queryParams type for
// a resource from the operation query params types recorded so far
// and sets the queryParams property on the resource. It does nothing
// if the resource hasn't been gathered yet.
func (o *OpenAPIContext) setResourceQueryParamsProp(tok string) {
	// Only consider resources that were gathered from the
	// current spec, i.e. those with a create endpoint.
	crud, ok := o.resourceCRUDMap[tok]
	if !ok || crud.C == nil {
		return
	}
	resourceSpec, ok := o.Pkg.Resources[tok]
	if !ok {
		return
	}

	tokParts := strings.Split(tok, ":")
	module, resourceName := tokParts[1], tokParts[2]
	typeTok := fmt.Sprintf("%s:%s:%s", o.Pkg.Name, module, resourceName+"QueryParams")
	typeRef := typesSchemaRefPrefix + typeTok

	existingInput, hasInput := resourceSpec.InputProperties[queryParamsPropName]
	existingOutput, hasOutput := resourceSpec.Properties[queryParamsPropName]
	if (hasInput && existingInput.Ref != typeRef) || (hasOutput && existingOutput.Ref != typeRef) {
		glog.Warningf("Resource %s already has a property called %s. Skipping query params for it.", tok, queryParamsPropName)
		return
	}

	properties := make(map[string]pschema.PropertySpec)
	required := codegen.NewStringSet()
	for opName, opQueryParams := range o.resourceQueryParams[tok] {
		properties[opName] = pschema.PropertySpec{
			Description: fmt.Sprintf("Query params for the %s operation.", opName),
			TypeSpec:    opQueryParams.typeSpec,
		}
		if opQueryParams.required {
			required.Add(opName)
		}
	}

	o.Pkg.Types[typeTok] = pschema.ComplexTypeSpec{
		ObjectTypeSpec: pschema.ObjectTypeSpec{
			Description: "Query params for each of the operations of the resource.",
			Type:        typeObject,
			Properties:  properties,
			Required:    required.SortedValues(),
		},
	}

	queryParamsProp := pschema.PropertySpec{
		Description: "Query params to send with the API requests for this resource.",
		TypeSpec:    pschema.TypeSpec{Ref: typeRef},
	}

	resourceSpec.InputProperties[queryParamsPropName] = queryParamsProp
	if resourceSpec.Properties == nil {
		resourceSpec.Properties = make(map[string]pschema.PropertySpec)
	}
	resourceSpec.Properties[queryParamsPropName] = queryParamsProp

	if len(required) > 0 {
		requiredInputs := codegen.NewStringSet(resourceSpec.RequiredInputs...)
		requiredInputs.Add(queryParamsPropName)
		resourceSpec.RequiredInputs = requiredInputs.SortedValues()
	}

	o.Pkg.Resources[tok] = resourceSpec
}
