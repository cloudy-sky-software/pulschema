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
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "simple_property_ref_openapi.yml"))

	openAPICtx := &OpenAPIContext{
		Doc: *testOpenAPIDoc,
		Pkg: &testPulumiPkg,
	}

	csharpNamespaces := map[string]string{
		"": "Provider",
	}

	_, _, err := openAPICtx.GatherResourcesFromAPI(csharpNamespaces)
	assert.Nil(t, err)

	t.Run("TestRequestSchemaWithSimpleProp", func(t *testing.T) {
		resourceSpec, ok := testPulumiPkg.Resources["fake-package:fakeresource/v2:FakeResource"]
		assert.Truef(t, ok, "Expected to find a resource called FakeResource: %v", testPulumiPkg.Resources)

		// The property simple_prop would have been converted to the SDK name in camelCase.
		assert.Equal(t, "string", resourceSpec.InputProperties["simpleProp"].Type)
	})

	t.Run("TestPlainTextResponse", func(t *testing.T) {
		funcSpec, ok := testPulumiPkg.Functions["fake-package:simpleresource/v2:getSimpleResource"]
		assert.Truef(t, ok, "Expected to find a function called getSimpleResource: %v", testPulumiPkg.Functions)

		assert.Equal(t, "string", funcSpec.ReturnType.TypeSpec.Type)
	})
}
