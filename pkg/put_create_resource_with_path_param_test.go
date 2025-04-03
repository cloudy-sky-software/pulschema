package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPutCreateResourceWithPathParam tests creating a resource
// with a PUT and a path param
func TestPutCreateResourceWithPathParam(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "put_create_resource_with_path_param.yml"))

	openAPICtx := &OpenAPIContext{
		Doc: *testOpenAPIDoc,
		Pkg: &testPulumiPkg,
	}

	csharpNamespaces := map[string]string{
		"": "Provider",
	}

	metadata, _, err := openAPICtx.GatherResourcesFromAPI(csharpNamespaces)
	assert.Nil(t, err)

	t.Run("FakeResource", func(t *testing.T) {
		fakeResource, ok := testPulumiPkg.Resources["fake-package:fakeresource/v2:FakeResource"]
		assert.Truef(t, ok, "Expected to find a resource called FakeResource: %v", testPulumiPkg.Resources)

		crudMap, ok := metadata.ResourceCRUDMap["fake-package:fakeresource/v2:FakeResource"]
		assert.Truef(t, ok, "Expected to find a CRUD map entry for FakeResource")
		// According the test API spec for FakeResource, it can be created
		// using a PUT endpoint and there is no POST endpoint for it.
		// So let's verify that in the metadata as well.
		assert.NotNil(t, crudMap.P, "PUT endpoint for FakeResource must be non-nil")
		assert.Equal(t, *crudMap.P, *crudMap.C, "POST endpoint for FakeResource must be the same as PUT endpoint")

		// Ensure that the input properties for the resource contains
		// the expected id property.
		assert.Contains(t, fakeResource.InputProperties, "someId")

		// Ensure that the "get" func also contains the id
		// as an input properties.
		getFunc, ok := testPulumiPkg.Functions["fake-package:fakeresource/v2:listFakeResources"]
		assert.Truef(t, ok, "Expected to find a list func listFakeResources: %v", testPulumiPkg.Functions)
		assert.NotNil(t, getFunc.Inputs)
		assert.Contains(t, getFunc.Inputs.Properties, "someId")
	})

	t.Run("DifferentResource", func(t *testing.T) {
		_, ok := testPulumiPkg.Resources["fake-package:differentresource/v2:DifferentResource"]
		assert.Truef(t, ok, "Expected to find a resource called DifferentResource: %v", testPulumiPkg.Resources)

		crudMap, ok := metadata.ResourceCRUDMap["fake-package:differentresource/v2:DifferentResource"]
		assert.Truef(t, ok, "Expected to find a CRUD map entry for DifferentResource")
		// According the test API spec for DifferentResource, it can be created
		// using the POST endpoint and updated via a PUT endpoint.
		assert.NotNil(t, crudMap.P, "PUT endpoint for DifferentResource must be non-nil")
		// The PUT and POST endpoints must be different here.
		assert.NotEqual(t, *crudMap.P, *crudMap.C, "POST endpoint for DifferentResource must NOT be the same as PUT endpoint")
	})
}
