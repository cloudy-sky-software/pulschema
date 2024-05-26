package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResourceInputsWithOneOfNoDiscriminator tests if
// resource request bodies with oneOf definitions that
// don't define a discriminator are treated as allOf
// with none of the inputs being required.
func TestResourceInputsWithOneOfNoDiscriminator(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "resource_inputs_oneof_no_discriminator_openapi.yml"))

	openAPICtx := &OpenAPIContext{
		Doc: *testOpenAPIDoc,
		Pkg: &testPulumiPkg,
	}

	csharpNamespaces := map[string]string{
		"": "Provider",
	}

	_, _, err := openAPICtx.GatherResourcesFromAPI(csharpNamespaces)
	assert.Nil(t, err)

	resourceSpec, ok := testPulumiPkg.Resources["fake-package:resource/v2:Resource"]
	assert.Truef(t, ok, "Expected to find a resource called Resource: %v", testPulumiPkg.Resources)

	// The resource should have all props from all types that was in the oneOf
	// definition.
	assert.Equal(t, "string", resourceSpec.InputProperties["simpleProp"].Type)
	assert.Equal(t, "string", resourceSpec.InputProperties["anotherProp"].Type)
	// None of them should be required because the API does not
	// describe a discriminator. It's up to the user to know
	// which set of inputs are required based on the "type"
	// of request they would like to send to the API.
	assert.Empty(t, resourceSpec.Required)
}
