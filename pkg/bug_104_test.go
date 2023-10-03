package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test to ensure resources with discriminated request body types
// are generated with unique resource names if there is a collision.
func TestBug104(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "bug_104_openapi.yml"))

	openAPICtx := &OpenAPIContext{
		Doc: *testOpenAPIDoc,
		Pkg: &testPulumiPkg,
	}

	csharpNamespaces := map[string]string{
		"": "Provider",
	}

	providerMetadata, _, err := openAPICtx.GatherResourcesFromAPI(csharpNamespaces)
	assert.Nil(t, err)
	assert.NotNil(t, providerMetadata)

	_, ok := testPulumiPkg.Resources["fake-package:droplets/v2:DropletAction"]
	assert.False(t, ok, "Resource DropletAction should not exist")
}
