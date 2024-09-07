package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBug131(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "bug_131_openapi.yml"))

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

	t.Run("Renamed", func(t *testing.T) {
		funcSpec, ok := testPulumiPkg.Functions[packageName+":actions/v2:listActions"]
		assert.True(t, ok, "Expected to find function listActions in the Pulumi package spec")

		assert.NotEmpty(t, funcSpec.ReturnType.TypeSpec.Ref)
		assert.Equal(t, funcSpec.ReturnType.TypeSpec.Ref, "#/types/fake-package:actions/v2:ListActionsItems")
	})

	// Test that a response type without an allOf definition in its
	// response type is named with the suffix Properties.
	t.Run("PropertiesResponseType", func(t *testing.T) {
		funcSpec, ok := testPulumiPkg.Functions[packageName+":actions2/v2:listActions2"]
		assert.True(t, ok, "Expected to find function listActions2 in the Pulumi package spec")

		assert.NotEmpty(t, funcSpec.ReturnType.TypeSpec.Ref)
		assert.Equal(t, funcSpec.ReturnType.TypeSpec.Ref, "#/types/fake-package:actions2/v2:ListActions2Properties")
	})
}
