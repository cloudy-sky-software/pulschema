package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCommonPathParams tests if path params that are
// common to all operations are considered.
func TestCommonPathParams(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "common_path_params_openapi.yml"))

	openAPICtx := &OpenAPIContext{
		Doc: *testOpenAPIDoc,
		Pkg: &testPulumiPkg,
	}

	csharpNamespaces := map[string]string{
		"": "Provider",
	}

	_, _, err := openAPICtx.GatherResourcesFromAPI(csharpNamespaces)
	assert.Nil(t, err)

	subResource, ok := testPulumiPkg.Resources["fake-package:fakeresource/v2:SubResource"]
	assert.Truef(t, ok, "Expected to find a resource called SubResource: %v", testPulumiPkg.Resources)

	// Ensure that the input properties for the resource contains
	// the expected id property.
	assert.Contains(t, subResource.InputProperties, "id")

	// Ensure that the "get" func also contains the id
	// as an input properties.
	getFunc, ok := testPulumiPkg.Functions["fake-package:fakeresource/v2:listSubResources"]
	assert.Truef(t, ok, "Expected to find a list func listSubResources: %v", testPulumiPkg.Functions)
	assert.NotNil(t, getFunc.Inputs)
	assert.Contains(t, getFunc.Inputs.Properties, "id")
}
