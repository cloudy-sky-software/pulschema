package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGenEnumType tests that schemas that have conflicting
// enums are legal. This is because they can have the
// same enum names but with different values as inline
// enums instead of refs.
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

	resourceSpec, ok := testPulumiPkg.Resources["fake-package:resource/v2:SomeResource"]
	assert.Truef(t, ok, "Expected to find a resource called SomeResource: %v", testPulumiPkg.Resources)

	// Due to the ordering of the paths, the type that has an enum prop point to a ref
	// would be encountered first and therefore would not have any collision.
	assert.Equal(t, "#/types/fake-package:resource/v2:EnumProp", resourceSpec.InputProperties["enumProp"].Ref)

	resourceSpec2, ok := testPulumiPkg.Resources["fake-package:resource/v2:SomeOtherResource"]
	assert.Truef(t, ok, "Expected to find a resource called SomeOtherResource: %v", testPulumiPkg.Resources)

	// When the next operation is encountered, the enum_prop property having an inline
	// schema with different values should be renamed with the prefix of the resource.
	assert.Equal(t, "#/types/fake-package:resource/v2:SomeOtherResourceEnumProp", resourceSpec2.InputProperties["enumProp"].Ref)
}
