// Copyright 2022, Cloudy Sky Software.

package pkg

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/golang/glog"

	"github.com/pkg/errors"

	"github.com/pulumi/pulumi/pkg/v3/codegen"
	dotnetgen "github.com/pulumi/pulumi/pkg/v3/codegen/dotnet"
	pschema "github.com/pulumi/pulumi/pkg/v3/codegen/schema"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

const componentsSchemaRefPrefix = "#/components/schemas/"
const jsonMimeType = "application/json"
const arrayType = "array"
const parameterLocationPath = "path"

type OpenAPIContext struct {
	Doc             openapi3.T
	Pkg             *pschema.PackageSpec
	ResourceCRUDMap map[string]*CRUDOperationsMap
}

func getRootPath(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	return parts[0]
}

func getParentPath(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	lastPathPart := parts[len(parts)-1]
	if !strings.HasPrefix(lastPathPart, "{") && !strings.HasSuffix(lastPathPart, "}") {
		return path
	}

	// Skip the last path part which contains a path parameter.
	return "/" + strings.Join(parts[0:len(parts)-1], "/")
}

// GatherResourcesFromAPI gathers resources from API endpoints.
// The goal is to extract resources and map their corresponding CRUD
// operations.
//
//   - The "create" operation (denoted by a Post request) determines the schema
//     for the resource.
//   - The "read" operation (denoted by a Get request) determines the schema
//     for "invokes" or "resource get's".
//   - The "update" operation (denoted by a Patch request) determines the schema
//     for resource updates. The Patch request schema is used to determine
//     which properties can be patched when changes are detected in Diff() vs.
//     which ones will force a resource replacement.
func (o *OpenAPIContext) GatherResourcesFromAPI(csharpNamespaces map[string]string) error {
	for path, pathItem := range o.Doc.Paths {
		// Capture the iteration variable `path` because we use its pointer
		// in the crudMap.
		currentPath := path
		module := getRootPath(currentPath)

		// TODO: TEMPORARY!
		if currentPath == "/services/{serviceId}/resume" ||
			currentPath == "/services/{serviceId}/custom-domains/{id}/verify" {
			continue
		}
		//

		glog.V(3).Infof("Processing path %s\n", currentPath)

		if pathItem.Get != nil {
			parentPath := getParentPath(currentPath)
			glog.V(3).Infof("GET: Parent path for %s is %s\n", currentPath, parentPath)

			jsonReq := pathItem.Get.Responses.Get(200).Value.Content.Get(jsonMimeType)
			if jsonReq.Schema.Value == nil {
				contract.Failf("Path %s has no schema definition for status code 200", currentPath)
			}

			setReadOperationMapping := func(tok string) {
				if existing, ok := o.ResourceCRUDMap[tok]; ok {
					existing.R = &currentPath
				} else {
					o.ResourceCRUDMap[tok] = &CRUDOperationsMap{
						R: &currentPath,
					}
				}
			}

			resourceType := jsonReq.Schema.Value
			if resourceType.Type != arrayType {
				// If there is a discriminator then we should set this operation
				// as the read endpoint for each of the types in the mapping.
				if resourceType.Discriminator != nil {
					for _, ref := range resourceType.Discriminator.Mapping {
						schemaName := strings.TrimPrefix(ref, componentsSchemaRefPrefix)
						dResource := o.Doc.Components.Schemas[schemaName]
						typeToken := fmt.Sprintf("%s:%s:%s", o.Pkg.Name, module, dResource.Value.Title)
						setReadOperationMapping(typeToken)

						funcName := "get" + dResource.Value.Title
						funcTypeToken := o.Pkg.Name + ":" + module + ":" + funcName
						getterFuncSpec := o.genGetFunc(*pathItem, *dResource, module, funcName)
						o.Pkg.Functions[funcTypeToken] = getterFuncSpec
						setReadOperationMapping(funcTypeToken)
					}
				} else {
					typeToken := fmt.Sprintf("%s:%s:%s", o.Pkg.Name, module, resourceType.Title)
					setReadOperationMapping(typeToken)

					funcName := "get" + resourceType.Title
					funcTypeToken := o.Pkg.Name + ":" + module + ":" + funcName
					getterFuncSpec := o.genGetFunc(*pathItem, *jsonReq.Schema, module, funcName)
					o.Pkg.Functions[funcTypeToken] = getterFuncSpec
					setReadOperationMapping(funcTypeToken)
				}
			}

			// Add the API operation as a list* function.
			if resourceType.Type == arrayType {
				funcName := strings.TrimPrefix(jsonReq.Schema.Ref, componentsSchemaRefPrefix)
				funcTypeToken := o.Pkg.Name + ":" + module + ":" + funcName
				funcSpec := o.genListFunc(*pathItem, *jsonReq.Schema, funcName, module)
				o.Pkg.Functions[funcTypeToken] = funcSpec
				setReadOperationMapping(funcTypeToken)
			}
		}

		if pathItem.Patch != nil {
			parentPath := getParentPath(currentPath)
			glog.V(3).Infof("PATCH: Parent path for %s is %s\n", currentPath, parentPath)

			jsonReq := pathItem.Patch.RequestBody.Value.Content.Get(jsonMimeType)
			if jsonReq.Schema.Value == nil {
				contract.Failf("Path %s has no schema definition for Patch method", currentPath)
			}

			setUpdateOperationMapping := func(tok string) {
				if existing, ok := o.ResourceCRUDMap[tok]; ok {
					existing.U = &currentPath
				} else {
					o.ResourceCRUDMap[tok] = &CRUDOperationsMap{
						U: &currentPath,
					}
				}
			}

			resourceType := jsonReq.Schema.Value
			if resourceType.Discriminator != nil || len(resourceType.OneOf) > 0 || len(resourceType.AnyOf) > 0 {
				schemaNames := make([]string, 0)
				if resourceType.Discriminator != nil {
					for _, ref := range resourceType.Discriminator.Mapping {
						schemaName := strings.TrimPrefix(ref, componentsSchemaRefPrefix)
						schemaNames = append(schemaNames, schemaName)
					}
				}

				if len(resourceType.OneOf) > 0 {
					for _, ref := range resourceType.OneOf {
						schemaName := strings.TrimPrefix(ref.Ref, componentsSchemaRefPrefix)
						schemaNames = append(schemaNames, schemaName)
					}
				}

				if len(resourceType.AnyOf) > 0 {
					for _, ref := range resourceType.AnyOf {
						schemaName := strings.TrimPrefix(ref.Ref, componentsSchemaRefPrefix)
						schemaNames = append(schemaNames, schemaName)
					}
				}

				for _, n := range schemaNames {
					dResource := o.Doc.Components.Schemas[n]
					typeToken := fmt.Sprintf("%s:%s:%s", o.Pkg.Name, module, dResource.Value.Title)
					setUpdateOperationMapping(typeToken)
				}
			} else {
				typeToken := fmt.Sprintf("%s:%s:%s", o.Pkg.Name, module, resourceType.Title)
				setUpdateOperationMapping(typeToken)
			}
		}

		if pathItem.Put != nil {
			parentPath := getParentPath(currentPath)
			glog.V(3).Infof("PUT: Parent path for %s is %s\n", currentPath, parentPath)

			jsonReq := pathItem.Put.RequestBody.Value.Content.Get(jsonMimeType)
			if jsonReq.Schema.Value == nil {
				contract.Failf("Path %s has no schema definition for Put method", currentPath)
			}

			setPutOperationMapping := func(tok string) {
				if existing, ok := o.ResourceCRUDMap[tok]; ok {
					existing.P = &currentPath
				} else {
					o.ResourceCRUDMap[tok] = &CRUDOperationsMap{
						P: &currentPath,
					}
				}
			}

			resourceType := jsonReq.Schema.Value
			if resourceType.Discriminator != nil {
				for _, ref := range resourceType.Discriminator.Mapping {
					schemaName := strings.TrimPrefix(ref, componentsSchemaRefPrefix)
					dResource := o.Doc.Components.Schemas[schemaName]
					typeToken := fmt.Sprintf("%s:%s:%s", o.Pkg.Name, module, dResource.Value.Title)
					setPutOperationMapping(typeToken)
				}
			} else {
				typeToken := fmt.Sprintf("%s:%s:%s", o.Pkg.Name, module, resourceType.Title)
				setPutOperationMapping(typeToken)
			}
		}

		if pathItem.Delete != nil && pathItem.Delete.RequestBody != nil {
			parentPath := getParentPath(currentPath)
			glog.V(3).Infof("DELETE: Parent path for %s is %s\n", currentPath, parentPath)

			jsonReq := pathItem.Delete.RequestBody.Value.Content.Get(jsonMimeType)
			if jsonReq.Schema.Value == nil {
				contract.Failf("Path %s has no schema definition for Delete method", currentPath)
			}

			setDeleteOperationMapping := func(tok string) {
				if existing, ok := o.ResourceCRUDMap[tok]; ok {
					existing.D = &currentPath
				} else {
					o.ResourceCRUDMap[tok] = &CRUDOperationsMap{
						D: &currentPath,
					}
				}
			}

			resourceType := jsonReq.Schema.Value
			if resourceType.Discriminator != nil {
				for _, ref := range resourceType.Discriminator.Mapping {
					schemaName := strings.TrimPrefix(ref, componentsSchemaRefPrefix)
					dResource := o.Doc.Components.Schemas[schemaName]
					typeToken := fmt.Sprintf("%s:%s:%s", o.Pkg.Name, module, dResource.Value.Title)
					setDeleteOperationMapping(typeToken)
				}
			} else {
				typeToken := fmt.Sprintf("%s:%s:%s", o.Pkg.Name, module, resourceType.Title)
				setDeleteOperationMapping(typeToken)
			}
		}

		if pathItem.Post == nil {
			continue
		}

		jsonReq := pathItem.Post.RequestBody.Value.Content.Get(jsonMimeType)
		if jsonReq.Schema.Value == nil {
			return errors.Errorf("path %s has no api schema definition for post method", currentPath)
		}

		resourceType := jsonReq.Schema.Value
		if err := o.gatherResource(currentPath, *resourceType, pathItem.Post.Parameters, module); err != nil {
			return errors.Wrapf(err, "generating resource for api path %s", currentPath)
		}

		csharpNamespaces[module] = ToPascalCase(module)
	}

	return nil
}

// genListFunc returns a function spec for a GET API endpoint that returns a list of objects.
// The item type can have a discriminator in the schema. This method will return a type
// that will refer to an output type that uses the discriminator properties to correctly
// type the output result.
func (o *OpenAPIContext) genListFunc(pathItem openapi3.PathItem, returnTypeSchema openapi3.SchemaRef, funcName, module string) pschema.FunctionSpec {
	parentName := ToPascalCase(funcName)
	funcPkgCtx := &resourceContext{
		mod:               module,
		pkg:               o.Pkg,
		openapiComponents: o.Doc.Components,
		visitedTypes:      codegen.NewStringSet(),
	}

	requiredInputs := codegen.NewStringSet()
	inputProps := make(map[string]pschema.PropertySpec)
	for _, param := range pathItem.Get.Parameters {
		if param.Value.In != parameterLocationPath {
			continue
		}

		paramName := param.Value.Name
		inputProps[paramName] = pschema.PropertySpec{
			Description: param.Value.Description,
			TypeSpec:    pschema.TypeSpec{Type: "string"},
		}
		requiredInputs.Add(paramName)
	}

	outputPropType, _ := funcPkgCtx.propertyTypeSpec(parentName, returnTypeSchema)

	return pschema.FunctionSpec{
		Description: pathItem.Description,
		Inputs: &pschema.ObjectTypeSpec{
			Properties: inputProps,
			Required:   requiredInputs.SortedValues(),
		},
		Outputs: &pschema.ObjectTypeSpec{
			Properties: map[string]pschema.PropertySpec{
				"items": {
					TypeSpec: *outputPropType,
				},
			},
			Required: []string{"items"},
		},
	}
}

// genGetFunc returns a function spec for a GET API endpoint that returns a single object.
// The single object can have a discriminator in the schema. This method will return a type
// that will refer to an output type that uses the discriminator properties to correctly
// type the output result.
func (o *OpenAPIContext) genGetFunc(pathItem openapi3.PathItem, returnTypeSchema openapi3.SchemaRef, module, funcName string) pschema.FunctionSpec {
	parentName := ToPascalCase(funcName)
	funcPkgCtx := &resourceContext{
		mod:               module,
		pkg:               o.Pkg,
		openapiComponents: o.Doc.Components,
		visitedTypes:      codegen.NewStringSet(),
	}

	requiredInputs := codegen.NewStringSet()
	inputProps := make(map[string]pschema.PropertySpec)
	for _, param := range pathItem.Get.Parameters {
		if param.Value.In != parameterLocationPath {
			continue
		}

		paramName := param.Value.Name
		inputProps[paramName] = pschema.PropertySpec{
			Description: param.Value.Description,
			TypeSpec:    pschema.TypeSpec{Type: "string"},
		}
		requiredInputs.Add(paramName)
	}

	outputPropType, _ := funcPkgCtx.propertyTypeSpec(parentName, returnTypeSchema)

	return pschema.FunctionSpec{
		Description: pathItem.Description,
		Inputs: &pschema.ObjectTypeSpec{
			Properties: inputProps,
			Required:   requiredInputs.SortedValues(),
		},
		Outputs: &pschema.ObjectTypeSpec{
			Properties: map[string]pschema.PropertySpec{
				"items": {
					TypeSpec: *outputPropType,
				},
			},
			Required: []string{"items"},
		},
	}
}

// gatherResource generates a resource spec from a POST API endpoint schema and
// adds it to the Pulumi schema spec.
func (o *OpenAPIContext) gatherResource(
	apiPath string,
	resourceType openapi3.Schema,
	pathParams openapi3.Parameters,
	module string) error {
	var resourceTypeToken *string
	var err error

	if resourceType.Discriminator != nil {
		for _, mappingRef := range resourceType.Discriminator.Mapping {
			schemaName := strings.TrimPrefix(mappingRef, componentsSchemaRefPrefix)
			typeSchema, ok := o.Doc.Components.Schemas[schemaName]
			if !ok {
				return errors.Errorf("%s not found in api schemas for discriminated type in path %s", schemaName, apiPath)
			}

			resourceTypeToken, err = o.gatherResourceProperties(*typeSchema.Value, apiPath, module)
		}
	} else {
		resourceTypeToken, err = o.gatherResourceProperties(resourceType, apiPath, module)
	}

	if err != nil {
		return errors.Wrapf(err, "gathering resource from api path %s", apiPath)
	}

	resourceSpec := o.Pkg.Resources[*resourceTypeToken]
	requiredInputs := codegen.NewStringSet(resourceSpec.RequiredInputs...)

	// If this endpoint path has path parameters,
	// then those should be required inputs too.
	for _, param := range pathParams {
		if param.Value.In != parameterLocationPath {
			continue
		}

		paramName := param.Value.Name
		resourceSpec.InputProperties[paramName] = pschema.PropertySpec{
			Description: param.Value.Description,
			TypeSpec:    pschema.TypeSpec{Type: "string"},
		}
		requiredInputs.Add(paramName)
	}

	o.Pkg.Resources[*resourceTypeToken] = resourceSpec

	return nil
}

// gatherResourceProperties generates a resource spec's input and output properties
// based on its API schema. Returns the Pulumi type token for the newly-added resource.
func (o *OpenAPIContext) gatherResourceProperties(resourceAPISchema openapi3.Schema, apiPath, module string) (*string, error) {
	pkgCtx := &resourceContext{
		mod:               module,
		pkg:               o.Pkg,
		resourceName:      resourceAPISchema.Title,
		openapiComponents: o.Doc.Components,
		visitedTypes:      codegen.NewStringSet(),
	}

	inputProperties := make(map[string]pschema.PropertySpec)
	properties := make(map[string]pschema.PropertySpec)
	requiredInputs := codegen.NewStringSet()
	for propName, prop := range resourceAPISchema.Properties {
		propSpec := pkgCtx.genPropertySpec(ToPascalCase(propName), *prop)
		if !prop.Value.ReadOnly {
			inputProperties[propName] = propSpec
		}
		properties[propName] = propSpec
	}

	// Create a set of required inputs for this resource.
	// Filter out required props that are marked as read-only.
	for _, requiredProp := range resourceAPISchema.Required {
		propSchema := resourceAPISchema.Properties[requiredProp]
		if propSchema.Value.ReadOnly {
			continue
		}

		requiredInputs.Add(requiredProp)
	}

	if len(resourceAPISchema.AllOf) > 0 {
		parentName := ToPascalCase(resourceAPISchema.Title)
		var types []pschema.TypeSpec
		for _, schemaRef := range resourceAPISchema.AllOf {
			typ, err := pkgCtx.propertyTypeSpec(parentName, *schemaRef)
			if err != nil {
				return nil, errors.Wrapf(err, "generating property type spec from allOf schema for %s", resourceAPISchema.Title)
			}
			types = append(types, *typ)
		}

		// Now that all of the types have been added to schema's Types,
		// gather all of their properties and smash them together into
		// a new type and get rid of those top-level ones.
		typeToken := fmt.Sprintf("%s:%s:%s", o.Pkg.Name, module, parentName)
		for _, t := range types {
			refTypeTok := strings.TrimPrefix(t.Ref, "#/types/")
			refType := pkgCtx.pkg.Types[refTypeTok]

			for name, propSpec := range refType.Properties {
				inputProperties[name] = propSpec
				properties[name] = propSpec
			}

			for _, r := range refType.Required {
				if requiredInputs.Has(r) {
					continue
				}
				requiredInputs.Add(r)
			}

			pkgCtx.visitedTypes.Delete(refTypeTok)
			delete(pkgCtx.pkg.Types, refTypeTok)
		}

		if existing, ok := o.ResourceCRUDMap[typeToken]; ok {
			existing.C = &apiPath
		} else {
			o.ResourceCRUDMap[typeToken] = &CRUDOperationsMap{
				C: &apiPath,
			}
		}

		o.Pkg.Resources[typeToken] = pschema.ResourceSpec{
			ObjectTypeSpec: pschema.ObjectTypeSpec{
				Description: resourceAPISchema.Description,
				Type:        "object",
				Properties:  properties,
				Required:    resourceAPISchema.Required,
			},
			InputProperties: inputProperties,
			RequiredInputs:  requiredInputs.SortedValues(),
		}

		return &typeToken, nil
	}

	typeToken := fmt.Sprintf("%s:%s:%s", o.Pkg.Name, module, resourceAPISchema.Title)
	if existing, ok := o.ResourceCRUDMap[typeToken]; ok {
		existing.C = &apiPath
	} else {
		o.ResourceCRUDMap[typeToken] = &CRUDOperationsMap{
			C: &apiPath,
		}
	}

	o.Pkg.Resources[typeToken] = pschema.ResourceSpec{
		ObjectTypeSpec: pschema.ObjectTypeSpec{
			Description: resourceAPISchema.Description,
			Type:        "object",
			Properties:  properties,
			Required:    resourceAPISchema.Required,
		},
		InputProperties: inputProperties,
		RequiredInputs:  requiredInputs.SortedValues(),
	}

	return &typeToken, nil
}

// genPropertySpec returns a property spec from an schema ref.
// The type spec of the returned property spec can be any of the
// supported types in Pulumi, including ref's to other types
// within the schema. In the case of ref's to other types, those
// other types are automatically added to the Pulumi schema spec's
// `Types` property.
func (ctx *resourceContext) genPropertySpec(propName string, p openapi3.SchemaRef) pschema.PropertySpec {
	propertySpec := pschema.PropertySpec{
		Description: p.Value.Description,
	}
	if p.Value.Default != nil {
		propertySpec.Default = p.Value.Default
	}
	languageName := strings.ToUpper(propName[:1]) + propName[1:]
	if languageName == ctx.resourceName {
		// .NET does not allow properties to be the same as the enclosing class - so special case these
		propertySpec.Language = map[string]pschema.RawMessage{
			"csharp": rawMessage(dotnetgen.CSharpPropertyInfo{
				Name: languageName + "Value",
			}),
		}
	}
	// JSONSchema type includes `$ref` and `$schema` properties, and $ is an invalid character in
	// the generated names. Replace them with `Ref` and `Schema`.
	if strings.HasPrefix(propName, "$") {
		propertySpec.Language = map[string]pschema.RawMessage{
			"csharp": rawMessage(map[string]interface{}{
				"name": strings.ToUpper(propName[1:2]) + propName[2:],
			}),
		}
	}

	typeSpec, err := ctx.propertyTypeSpec(propName, p)
	if err != nil {
		contract.Failf("Failed to generate type spec (resource: %s, prop %s): %v", ctx.resourceName, propName, err)
	}

	propertySpec.TypeSpec = *typeSpec

	return propertySpec
}

// propertyTypeSpec converts an API schema to a Pulumi property type spec.
func (ctx *resourceContext) propertyTypeSpec(parentName string, propSchema openapi3.SchemaRef) (*pschema.TypeSpec, error) {
	// References to other type definitions as long as the type is not an array.
	// Arrays will be handled later in this method.
	if propSchema.Ref != "" && propSchema.Value.Type != arrayType {
		schemaName := strings.TrimPrefix(propSchema.Ref, componentsSchemaRefPrefix)
		typName := ToPascalCase(schemaName)
		tok := fmt.Sprintf("%s:%s:%s", ctx.pkg.Name, ctx.mod, typName)

		typeSchema, ok := ctx.openapiComponents.Schemas[schemaName]
		if !ok {
			return nil, errors.Errorf("definition %s not found in resource %s", schemaName, parentName)
		}

		if !ctx.visitedTypes.Has(tok) {
			ctx.visitedTypes.Add(tok)
			specs, requiredSpecs, err := ctx.genProperties(typName, *typeSchema.Value)
			if err != nil {
				return nil, err
			}

			ctx.pkg.Types[tok] = pschema.ComplexTypeSpec{
				ObjectTypeSpec: pschema.ObjectTypeSpec{
					Description: typeSchema.Value.Description,
					Type:        "object",
					Properties:  specs,
					Required:    requiredSpecs.SortedValues(),
				},
			}
		}

		referencedTypeName := fmt.Sprintf("#/types/%s", tok)
		return &pschema.TypeSpec{Ref: referencedTypeName}, nil
	}

	// Inline properties.
	if len(propSchema.Value.Properties) > 0 {
		typName := parentName + "Properties"
		tok := fmt.Sprintf("%s:%s:%s", ctx.pkg.Name, ctx.mod, typName)
		specs, requiredSpecs, err := ctx.genProperties(typName, *propSchema.Value)
		if err != nil {
			return nil, err
		}

		ctx.pkg.Types[tok] = pschema.ComplexTypeSpec{
			ObjectTypeSpec: pschema.ObjectTypeSpec{
				Description: propSchema.Value.Description,
				Type:        "object",
				Properties:  specs,
				Required:    requiredSpecs.SortedValues(),
			},
		}
		referencedTypeName := fmt.Sprintf("#/types/%s", tok)
		return &pschema.TypeSpec{Ref: referencedTypeName}, nil
	}

	// Union types.
	if len(propSchema.Value.OneOf) > 0 {
		var types []pschema.TypeSpec
		for _, schemaRef := range propSchema.Value.OneOf {
			typ, err := ctx.propertyTypeSpec(parentName, *schemaRef)
			if err != nil {
				return nil, err
			}
			types = append(types, *typ)
		}

		var discriminator *pschema.DiscriminatorSpec
		if propSchema.Value.Discriminator != nil {
			discriminator = &pschema.DiscriminatorSpec{
				PropertyName: ToSdkName(propSchema.Value.Discriminator.PropertyName),
			}

			mapping := make(map[string]string)
			for discriminatorProperyValue, apiSchemaRef := range propSchema.Value.Discriminator.Mapping {
				resourceTypeName := strings.TrimPrefix(apiSchemaRef, "#/components/schemas/")
				resourceTypeName = ToPascalCase(resourceTypeName)
				for _, typeSpec := range types {
					if !strings.Contains(typeSpec.Ref, resourceTypeName) {
						continue
					}
					mapping[discriminatorProperyValue] = typeSpec.Ref
				}
			}
			discriminator.Mapping = mapping
		}

		return &pschema.TypeSpec{
			OneOf:         types,
			Discriminator: discriminator,
		}, nil
	}

	if len(propSchema.Value.AllOf) > 0 {
		properties, requiredSpecs, err := ctx.genPropertiesFromAllOf(parentName, propSchema.Value.AllOf)
		if err != nil {
			return nil, errors.Wrap(err, "generating properties from allOf schema definition")
		}

		tok := fmt.Sprintf("%s:%s:%s", ctx.pkg.Name, ctx.mod, ToPascalCase(parentName))
		ctx.pkg.Types[tok] = pschema.ComplexTypeSpec{
			ObjectTypeSpec: pschema.ObjectTypeSpec{
				Description: propSchema.Value.Description,
				Type:        "object",
				Properties:  properties,
				Required:    requiredSpecs.SortedValues(),
			},
		}

		return &pschema.TypeSpec{
			Ref: fmt.Sprintf("#/types/%s", tok),
		}, nil
	}

	if len(propSchema.Value.Enum) > 0 {
		enum, err := ctx.genEnumType(parentName, *propSchema.Value)
		if err != nil {
			return nil, err
		}
		if enum != nil {
			return enum, nil
		}
	}

	// All other types.
	switch propSchema.Value.Type {
	case openapi3.TypeInteger:
		return &pschema.TypeSpec{Type: "integer"}, nil
	case openapi3.TypeString:
		return &pschema.TypeSpec{Type: "string"}, nil
	case openapi3.TypeBoolean:
		return &pschema.TypeSpec{Type: "boolean"}, nil
	case openapi3.TypeNumber:
		return &pschema.TypeSpec{Type: "number"}, nil
	case openapi3.TypeObject:
		return &pschema.TypeSpec{Ref: "pulumi.json#/Any"}, nil
	case openapi3.TypeArray:
		elementType, err := ctx.propertyTypeSpec(parentName+"Item", *propSchema.Value.Items)
		if err != nil {
			return nil, err
		}
		return &pschema.TypeSpec{
			Type:  arrayType,
			Items: elementType,
		}, nil
	}

	return nil, errors.Errorf("failed to generate property types for %+v", propSchema)
}

// genProperties returns a map of the property names and their corresponding
// property type spec and the required properties as a sorted set.
func (ctx *resourceContext) genProperties(parentName string, typeSchema openapi3.Schema) (map[string]pschema.PropertySpec, codegen.StringSet, error) {
	specs := map[string]pschema.PropertySpec{}
	requiredSpecs := codegen.NewStringSet()
	for _, name := range codegen.SortedKeys(typeSchema.Properties) {
		value := typeSchema.Properties[name]
		sdkName := ToSdkName(name)

		typeSpec, err := ctx.propertyTypeSpec(parentName+ToPascalCase(name), *value)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "property %s", name)
		}
		propertySpec := pschema.PropertySpec{
			Description: value.Value.Description,
			TypeSpec:    *typeSpec,
		}
		if value.Value.Default != nil {
			propertySpec.Default = value.Value.Default
		}

		specs[sdkName] = propertySpec
	}
	for _, name := range typeSchema.Required {
		sdkName := ToSdkName(name)
		if _, has := specs[sdkName]; has {
			requiredSpecs.Add(sdkName)
		}
	}

	if len(typeSchema.AllOf) > 0 {
		return ctx.genPropertiesFromAllOf(parentName, typeSchema.AllOf)

	}
	return specs, requiredSpecs, nil
}

// genPropertiesFromAllOf returns a map of property names and their corresponding
// property type spec gathered from a type's allOf schema.
func (ctx *resourceContext) genPropertiesFromAllOf(parentName string, allOf openapi3.SchemaRefs) (map[string]pschema.PropertySpec, codegen.StringSet, error) {
	var types []pschema.TypeSpec
	for _, schemaRef := range allOf {
		typ, err := ctx.propertyTypeSpec(parentName, *schemaRef)
		if err != nil {
			return nil, nil, err
		}
		types = append(types, *typ)
	}

	// Now that all of the types have been added to schema's Types,
	// gather all of their properties and smash them together into
	// a new type.
	properties := make(map[string]pschema.PropertySpec)
	requiredSpecs := codegen.NewStringSet()
	for _, t := range types {
		refTypeTok := strings.TrimPrefix(t.Ref, "#/types/")
		refType := ctx.pkg.Types[refTypeTok]

		for name, propSpec := range refType.Properties {
			properties[name] = propSpec
		}

		for _, r := range refType.Required {
			if requiredSpecs.Has(r) {
				continue
			}
			requiredSpecs.Add(r)
		}

		ctx.visitedTypes.Delete(refTypeTok)
		delete(ctx.pkg.Types, refTypeTok)
	}

	return properties, requiredSpecs, nil
}

// genEnumType generates the enum type for a given schema.
func (ctx *resourceContext) genEnumType(enumName string, propSchema openapi3.Schema) (*pschema.TypeSpec, error) {
	if len(propSchema.Type) == 0 {
		return nil, nil
	}
	if propSchema.Type != openapi3.TypeString {
		return nil, nil
	}

	typName := ToPascalCase(enumName)
	tok := fmt.Sprintf("%s:%s:%s", ctx.pkg.Name, ctx.mod, typName)

	enumSpec := &pschema.ComplexTypeSpec{
		Enum: []pschema.EnumValueSpec{},
		ObjectTypeSpec: pschema.ObjectTypeSpec{
			Description: propSchema.Description,
			Type:        "string",
		},
	}

	values := codegen.NewStringSet()
	for _, val := range propSchema.Enum {
		str := ToPascalCase(val.(string))
		if values.Has(str) {
			continue
		}
		values.Add(str)
		enumVal := pschema.EnumValueSpec{
			Value: val,
			Name:  str,
		}
		enumSpec.Enum = append(enumSpec.Enum, enumVal)
	}

	// Make sure that the type name we composed doesn't clash with another type
	// already defined in the schema earlier. The same enum does show up in multiple
	// places of specs, so we want to error only if they a) have the same name
	// b) the list of values does not match.
	if other, ok := ctx.pkg.Types[tok]; ok {
		same := len(enumSpec.Enum) == len(other.Enum)
		for _, val := range other.Enum {
			same = same && values.Has(val.Name)
		}
		if !same {
			return nil, errors.Errorf("duplicate enum %q: %+v vs. %+v", tok, enumSpec.Enum, other.Enum)
		}
	}
	ctx.pkg.Types[tok] = *enumSpec

	referencedTypeName := fmt.Sprintf("#/types/%s", tok)
	return &pschema.TypeSpec{
		Ref: referencedTypeName,
	}, nil
}
