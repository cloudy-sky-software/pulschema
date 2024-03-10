package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSimplePropertyRef tests if schema references to
// simple property types are inlined instead of creating
// a schema type.
func TestSimplePropertyRef(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "simple_property_ref.yml"))

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

	// The property simple_prop would have been converted to the SDK name in camelCase.
	assert.Equal(t, "string", resourceSpec.InputProperties["simpleProp"].Type)
}
