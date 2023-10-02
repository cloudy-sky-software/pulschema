package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGatherResources(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "bug_103_openapi.yml"))

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

	// apps_get_logs_active_deployment_aggregate
	_, ok := testPulumiPkg.Functions[packageName+":apps/v2:getAppsLogsActiveDeploymentAggregate"]
	assert.True(t, ok, "Expected to find function getAppsLogsActiveDeploymentAggregate in the Pulumi package spec")

	// apps_get_logs_aggregate
	_, ok = testPulumiPkg.Functions[packageName+":apps/v2:getAppsLogsAggregate"]
	assert.True(t, ok, "Expected to find function getAppsLogsAggregate in the Pulumi package spec")
}
