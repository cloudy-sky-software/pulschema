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
	assert.Equal(t, "#/types/fake-package:resource/v2:AProp", resourceSpec.InputProperties["aProp"].Ref)

	resourceSpec2, ok := testPulumiPkg.Resources["fake-package:resource/v2:SomeOtherResource"]
	assert.Truef(t, ok, "Expected to find a resource called SomeOtherResource: %v", testPulumiPkg.Resources)
	// Because AProp was already encountered and is not an enum, this resource's property
	// a_prop which is an enum should be suffixed with Enum but not yet prefixed with the
	// resource name.
	assert.Equal(t, "#/types/fake-package:resource/v2:APropEnum", resourceSpec2.InputProperties["aProp"].Ref)

	resourceSpec3, ok := testPulumiPkg.Resources["fake-package:resource/v2:LastResource"]
	assert.Truef(t, ok, "Expected to find a resource called LastResource: %v", testPulumiPkg.Resources)
	// When the next operation is encountered, the a_prop property having an inline
	// schema with different values should be renamed with the prefix of the resource,
	// as well as have the suffix Enum because, remember, we need to distinguish from
	// the regular object-type property AProp.
	assert.Equal(t, "#/types/fake-package:resource/v2:LastResourceAPropEnum", resourceSpec3.InputProperties["aProp"].Ref)
}
