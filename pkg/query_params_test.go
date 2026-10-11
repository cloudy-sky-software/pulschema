package pkg

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"

	pschema "github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

// noopLoader is a schema loader for packages that
// don't reference any external packages.
type noopLoader struct {
	pschema.Loader
}

func (noopLoader) LoadPackageV2(_ context.Context, descriptor *pschema.PackageDescriptor) (*pschema.Package, error) {
	return nil, fmt.Errorf("unexpected load of package %s", descriptor.Name)
}

// TestQueryParams tests that query params are added as
// a queryParams input to resources and functions.
func TestQueryParams(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "query_params_openapi.yml"))

	// Use a fresh package spec so that the result can be bound
	// without interference from types added by other tests.
	pkgSpec := testPulumiPkg
	pkgSpec.Types = map[string]pschema.ComplexTypeSpec{}
	pkgSpec.Resources = map[string]pschema.ResourceSpec{}
	pkgSpec.Functions = map[string]pschema.FunctionSpec{}

	openAPICtx := &OpenAPIContext{
		Doc: *testOpenAPIDoc,
		Pkg: &pkgSpec,
	}

	csharpNamespaces := map[string]string{
		"": providerNamespace,
	}

	metadata, _, err := openAPICtx.GatherResourcesFromAPI(csharpNamespaces)
	assert.Nil(t, err)

	typePrefix := "fake-package:things/v1:"
	refPrefix := "#/types/" + typePrefix

	thing, ok := pkgSpec.Resources[typePrefix+"Thing"]
	assert.Truef(t, ok, "Expected to find a resource called Thing: %v", pkgSpec.Resources)

	// The resource should have queryParams as both an input and an output.
	assert.Contains(t, thing.InputProperties, QueryParamsPropName)
	assert.Contains(t, thing.Properties, QueryParamsPropName)
	assert.Equal(t, refPrefix+"ThingQueryParams", thing.InputProperties[QueryParamsPropName].Ref)
	assert.Equal(t, refPrefix+"ThingQueryParams", thing.Properties[QueryParamsPropName].Ref)
	// The delete and update endpoints have required query params.
	assert.Contains(t, thing.RequiredInputs, QueryParamsPropName)

	thingQueryParams, ok := pkgSpec.Types[typePrefix+"ThingQueryParams"]
	assert.True(t, ok)
	for _, op := range []string{QueryParamsOpCreate, QueryParamsOpRead, QueryParamsOpUpdate, QueryParamsOpDelete} {
		assert.Contains(t, thingQueryParams.Properties, op)
	}
	assert.NotContains(t, thingQueryParams.Properties, QueryParamsOpPut)
	assert.Equal(t, refPrefix+"ThingCreateQueryParams", thingQueryParams.Properties[QueryParamsOpCreate].Ref)
	assert.Equal(t, refPrefix+"ThingReadQueryParams", thingQueryParams.Properties[QueryParamsOpRead].Ref)
	assert.Equal(t, refPrefix+"ThingUpdateQueryParams", thingQueryParams.Properties[QueryParamsOpUpdate].Ref)
	assert.Equal(t, refPrefix+"ThingDeleteQueryParams", thingQueryParams.Properties[QueryParamsOpDelete].Ref)
	assert.ElementsMatch(t, []string{QueryParamsOpDelete, QueryParamsOpUpdate}, thingQueryParams.Required)

	createQueryParams := pkgSpec.Types[typePrefix+"ThingCreateQueryParams"]
	assert.Equal(t, "boolean", createQueryParams.Properties["dryRun"].Type)
	assert.Empty(t, createQueryParams.Required)

	deleteQueryParams := pkgSpec.Types[typePrefix+"ThingDeleteQueryParams"]
	assert.Equal(t, "boolean", deleteQueryParams.Properties["force"].Type)
	assert.Equal(t, []string{"force"}, deleteQueryParams.Required)
	additionalParams, ok := deleteQueryParams.Properties[AdditionalQueryParamsPropName]
	assert.True(t, ok)
	assert.Equal(t, openapi3.TypeObject, additionalParams.Type)
	assert.Equal(t, openapi3.TypeString, additionalParams.AdditionalProperties.Type)

	// The path-level common query param should be in each of the
	// operation types for the /{id} path.
	readQueryParams := pkgSpec.Types[typePrefix+"ThingReadQueryParams"]
	assert.Contains(t, readQueryParams.Properties, "expand")
	assert.Contains(t, readQueryParams.Properties, "region")
	assert.Equal(t, "The region of the thing.", readQueryParams.Properties["region"].Description)
	assert.Contains(t, deleteQueryParams.Properties, "region")

	// The operation-level param overrides the path-level param.
	updateQueryParams := pkgSpec.Types[typePrefix+"ThingUpdateQueryParams"]
	assert.Equal(t, "Overridden region.", updateQueryParams.Properties["region"].Description)
	assert.Equal(t, []string{"region"}, updateQueryParams.Required)

	// The GET and DELETE endpoints of a resource created via PUT
	// are processed before the resource is gathered. Their query
	// params should still be added to the resource.
	gizmoPrefix := "fake-package:gizmos/v1:"
	gizmo, ok := pkgSpec.Resources[gizmoPrefix+"Gizmo"]
	assert.Truef(t, ok, "Expected to find a resource called Gizmo: %v", pkgSpec.Resources)
	assert.Equal(t, "#/types/"+gizmoPrefix+"GizmoQueryParams", gizmo.InputProperties[QueryParamsPropName].Ref)
	assert.Contains(t, gizmo.RequiredInputs, QueryParamsPropName)

	gizmoQueryParams := pkgSpec.Types[gizmoPrefix+"GizmoQueryParams"]
	for _, op := range []string{QueryParamsOpCreate, QueryParamsOpRead, QueryParamsOpPut, QueryParamsOpDelete} {
		assert.Contains(t, gizmoQueryParams.Properties, op)
	}
	assert.Equal(t, []string{QueryParamsOpDelete}, gizmoQueryParams.Required)
	assert.Contains(t, pkgSpec.Types[gizmoPrefix+"GizmoCreateQueryParams"].Properties, "idempotencyKey")
	assert.Contains(t, pkgSpec.Types[gizmoPrefix+"GizmoReadQueryParams"].Properties, "includeParts")
	assert.Equal(t, []string{"purge"}, pkgSpec.Types[gizmoPrefix+"GizmoDeleteQueryParams"].Required)

	// Functions.
	getFunc, ok := pkgSpec.Functions[typePrefix+"getThing"]
	assert.Truef(t, ok, "Expected to find a get func getThing: %v", pkgSpec.Functions)
	assert.Equal(t, refPrefix+"GetThingQueryParams", getFunc.Inputs.Properties[QueryParamsPropName].Ref)
	assert.NotContains(t, getFunc.Inputs.Required, QueryParamsPropName)

	listFunc, ok := pkgSpec.Functions[typePrefix+"listThings"]
	assert.Truef(t, ok, "Expected to find a list func listThings: %v", pkgSpec.Functions)
	assert.Equal(t, refPrefix+"ListThingsQueryParams", listFunc.Inputs.Properties[QueryParamsPropName].Ref)
	assert.NotContains(t, listFunc.Inputs.Required, QueryParamsPropName)

	listQueryParams := pkgSpec.Types[typePrefix+"ListThingsQueryParams"]
	assert.Equal(t, "integer", listQueryParams.Properties["pageSize"].Type)
	assert.Equal(t, refPrefix+"ListThingsQueryParamsSort", listQueryParams.Properties["sort"].Ref)
	sortEnum, ok := pkgSpec.Types[typePrefix+"ListThingsQueryParamsSort"]
	assert.True(t, ok)
	assert.Len(t, sortEnum.Enum, 2)

	// Name maps.
	assert.Equal(t, "page_size", metadata.SDKToAPINameMap["pageSize"])

	// The resulting schema should bind without errors.
	_, err = pschema.ImportSpec(pkgSpec, nil, noopLoader{}, pschema.ValidationOptions{})
	assert.Nil(t, err)
}
