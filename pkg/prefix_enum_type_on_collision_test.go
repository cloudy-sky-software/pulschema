package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenEnumType(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "prefix_enum_type_on_collision_openapi.yml"))

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
