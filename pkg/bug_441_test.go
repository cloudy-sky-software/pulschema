package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBug441(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "bug_441_openapi.yml"))

	openAPICtx := &OpenAPIContext{
		Doc: *testOpenAPIDoc,
		Pkg: &testPulumiPkg,
	}

	csharpNamespaces := map[string]string{
		"": providerNamespace,
	}

	providerMetadata, _, err := openAPICtx.GatherResourcesFromAPI(csharpNamespaces)
	assert.Nil(t, err)
	assert.NotNil(t, providerMetadata)

	t.Run("AdditionalProperties", func(t *testing.T) {
		machineConfig, ok := testPulumiPkg.Types[packageName+":machines:MachineConfig"]
		assert.True(t, ok, "Expected to find type MachineConfig in the Pulumi package spec")

		assert.NotEmpty(t, machineConfig.Properties)
		assert.Contains(t, machineConfig.Properties, "checks")
		checksProperty := machineConfig.Properties["checks"]
		t.Logf("%v", checksProperty)
		assert.NotNil(t, checksProperty.AdditionalProperties)
	})
}
