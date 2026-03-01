package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEmptyResponseBody(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "get_empty_response_body_openapi.yml"))

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

	// GET with JSON response should be a function.
	_, ok := testPulumiPkg.Functions[packageName+":vms/v2:getVm"]
	assert.True(t, ok, "Expected to find function getVm in the Pulumi package spec")

	// GET with no response body (ad-hoc restart action) should also be a function.
	_, ok = testPulumiPkg.Functions[packageName+":vms/v2:getVmsRestart"]
	assert.True(t, ok, "Expected to find function getVmsRestart for ad-hoc GET action with no response body")

	// The ad-hoc action function should have no return type.
	restartFunc := testPulumiPkg.Functions[packageName+":vms/v2:getVmsRestart"]
	assert.Nil(t, restartFunc.ReturnType, "Expected ad-hoc GET action to have no return type")

	// List with JSON response should be a function.
	_, ok = testPulumiPkg.Functions[packageName+":vms/v2:listVms"]
	assert.True(t, ok, "Expected to find function listVms in the Pulumi package spec")
}
