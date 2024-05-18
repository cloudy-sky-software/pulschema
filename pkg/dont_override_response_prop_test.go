package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDontOverrideResponseProp tests if object-type
// props in response schema are not overridden by
// the same prop from the request body schema.
func TestDontOverrideResponseProp(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "dont_override_response_props_openapi.yml"))

	openAPICtx := &OpenAPIContext{
		Doc: *testOpenAPIDoc,
		Pkg: &testPulumiPkg,
	}

	csharpNamespaces := map[string]string{
		"": "Provider",
	}

	_, _, err := openAPICtx.GatherResourcesFromAPI(csharpNamespaces)
	assert.Nil(t, err)

	resourceSpec, ok := testPulumiPkg.Resources["fake-package:fakeresource/v2:FakeResource"]
	assert.Truef(t, ok, "Expected to find a resource called FakeResource: %v", testPulumiPkg.Resources)

	// The output property `objectProp` for the resource
	// should use the response type schema `parent_object_2`.
	assert.Contains(t, resourceSpec.Properties, "objectProp")
	objectProp := resourceSpec.Properties["objectProp"]
	assert.Equal(t, "#/types/fake-package:fakeresource/v2:ParentObject2", objectProp.TypeSpec.Ref)
}
