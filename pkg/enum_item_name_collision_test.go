package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEnumItemNameCollision tests that enums with values
// that have the same name as the enclosing enum type they
// are declared in are properly renamed.
func TestEnumItemNameCollision(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "enum_item_name_collision.yml"))

	openAPICtx := &OpenAPIContext{
		Doc: *testOpenAPIDoc,
		Pkg: &testPulumiPkg,
	}

	csharpNamespaces := map[string]string{
		"": "Provider",
	}

	_, _, err := openAPICtx.GatherResourcesFromAPI(csharpNamespaces)
	assert.Nil(t, err)

	resourceSpec, ok := testPulumiPkg.Resources["fake-package:resource/v2:SomeResource"]
	assert.Truef(t, ok, "Expected to find a resource called SomeResource: %v", testPulumiPkg.Resources)
	assert.Equal(t, "#/types/fake-package:resource/v2:MyEnum", resourceSpec.InputProperties["myEnum"].Ref)

	assert.Equal(t, "MyEnum_", testPulumiPkg.Types["fake-package:resource/v2:MyEnum"].Enum[0].Name)
}
