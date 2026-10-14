package pkg

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

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

	machineConfig, ok := testPulumiPkg.Types[packageName+":machines:MachineConfig"]
	assert.True(t, ok, "Expected to find type MachineConfig in the Pulumi package spec")

	assert.NotEmpty(t, machineConfig.Properties)

	t.Run("additionalProperties with ref", func(t *testing.T) {
		assert.Contains(t, machineConfig.Properties, "checks")

		checksProperty := machineConfig.Properties["checks"]
		assert.NotNil(t, checksProperty.AdditionalProperties)
		// The additional properties ref name should
		// be prefixed with the parent schema name.
		//
		// In this case, the property `checks` is
		// under the `MachineConfig` schema. Therefore,
		// the ref type name for `checks` should be
		// MachineConfig + Checks + Properties.
		assert.Equal(t, fmt.Sprintf("#/types/%s:machines:MachineConfigChecksProperties", packageName), checksProperty.AdditionalProperties.Ref)
	})

	t.Run("additionalProperties with array", func(t *testing.T) {
		assert.Contains(t, machineConfig.Properties, "groups")

		groupsProperty := machineConfig.Properties["groups"]
		assert.NotNil(t, groupsProperty.AdditionalProperties)
		assert.Equal(t, openapi3.TypeArray, groupsProperty.AdditionalProperties.Type)
		assert.NotNil(t, groupsProperty.AdditionalProperties.Items)
		assert.Equal(t, openapi3.TypeString, groupsProperty.AdditionalProperties.Items.Type)
	})

	t.Run("additionalProperties with primitive type", func(t *testing.T) {
		assert.Contains(t, machineConfig.Properties, "autoApprovers")

		autoApproversProperty := machineConfig.Properties["autoApprovers"]
		assert.NotNil(t, autoApproversProperty.AdditionalProperties)
		assert.Equal(t, openapi3.TypeString, autoApproversProperty.AdditionalProperties.Type)
	})
}
