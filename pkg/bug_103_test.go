package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBug103(t *testing.T) {
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

	// projects_list_resources
	_, ok = testPulumiPkg.Functions[packageName+":projects/v2:listProjectsResources"]
	assert.True(t, ok, "Expected to find function listProjectsResources in the Pulumi package spec")

	// projects_list_resources_default
	_, ok = testPulumiPkg.Functions[packageName+":projects/v2:listProjectsResourcesDefault"]
	assert.True(t, ok, "Expected to find function listProjectsResourcesDefault in the Pulumi package spec")
}
