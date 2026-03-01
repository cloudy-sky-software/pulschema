package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetEmptyRequestBody tests that GET endpoints with no request body
// are handled correctly. GET endpoints should not require a request body,
// and both request and response bodies should be optional.
func TestGetEmptyRequestBody(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "get_empty_request_body_openapi.yml"))

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

	t.Run("GetWithJsonResponse", func(t *testing.T) {
		// The GET /v2/dns/settings endpoint should produce a get function
		// even though it has no request body.
		funcSpec, ok := testPulumiPkg.Functions[packageName+":dns/v2:getDnsSetting"]
		assert.Truef(t, ok, "Expected to find function getDnsSetting: %v", testPulumiPkg.Functions)
		assert.NotNil(t, funcSpec.ReturnType, "Expected getDnsSetting to have a return type")
	})

	t.Run("GetWithNoResponseBody", func(t *testing.T) {
		// The GET /v2/vms/{vm_id}/restart endpoint should produce a get function
		// with no return type since it only has a 204 No Content response.
		funcSpec, ok := testPulumiPkg.Functions[packageName+":vms/v2:getRestartVm"]
		assert.Truef(t, ok, "Expected to find function getRestartVm: %v", testPulumiPkg.Functions)
		// No return type since the endpoint has no response body.
		assert.Nil(t, funcSpec.ReturnType, "Expected getRestartVm to have no return type")
	})

	t.Run("ListWithJsonResponse", func(t *testing.T) {
		// The GET /v2/vms endpoint with array response should produce a list function.
		funcSpec, ok := testPulumiPkg.Functions[packageName+":vms/v2:listVms"]
		assert.Truef(t, ok, "Expected to find function listVms: %v", testPulumiPkg.Functions)
		assert.NotNil(t, funcSpec.ReturnType, "Expected listVms to have a return type")
	})
}
