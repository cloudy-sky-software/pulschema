package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSecretProperty tests that properties with the `x-pulumi-secret` extension
// mark the Pulumi property as a secret in the generated SDK.
func TestSecretProperty(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "secret_property_openapi.yml"))

	openAPICtx := &OpenAPIContext{
		Doc: *testOpenAPIDoc,
		Pkg: &testPulumiPkg,
	}

	csharpNamespaces := map[string]string{
		"": "Provider",
	}

	_, _, err := openAPICtx.GatherResourcesFromAPI(csharpNamespaces)
	assert.Nil(t, err)

	// Verify that aren't mapped from OneOf (aggregated properties) support secret properties
	resourceSpec, _ := testPulumiPkg.Resources["fake-package:fakeresource/v2:FakeResource"]
	assert.Truef(t, resourceSpec.InputProperties["refProp"].Secret, "Expected refProp to be a secret property, but it was not: %v", resourceSpec.InputProperties["refProp"])
	assert.Truef(t, resourceSpec.InputProperties["directStringProp"].Secret, "Expected directStringProp to be a secret property, but it was not: %v", resourceSpec.InputProperties["directStringProp"])
	assert.Falsef(t, resourceSpec.InputProperties["notSecretProp"].Secret, "Expected notSecretProp to not be a secret property, but it was: %v", resourceSpec.InputProperties["notSecretProp"])
	assert.Falsef(t, resourceSpec.InputProperties["explicitNotSecretProp"].Secret, "Expected explicitNotSecretProp to not be a secret property, but it was: %v", resourceSpec.InputProperties["explicitNotSecretProp"])
	assert.Truef(t, resourceSpec.Properties["aProp"].Secret, "Expected output property aProp to be a secret property, but it was not: %v", resourceSpec.Properties["aProp"])
	assert.Falsef(t, resourceSpec.Properties["notSecretProp"].Secret, "Expected output property notSecretProp to not be a secret property, but it was: %v", resourceSpec.Properties["notSecretProp"])

	// Verify that are mapped from OneOf (aggregated properties) support secret properties
	oneOfResourceSpec, _ := testPulumiPkg.Resources["fake-package:oneofresource/v2:OneofResource"]
	assert.Truef(t, oneOfResourceSpec.InputProperties["aProp"].Secret, "Expected aProp to be a secret property, but it was not: %v", oneOfResourceSpec.InputProperties["aProp"])
	assert.Truef(t, oneOfResourceSpec.InputProperties["refProp"].Secret, "Expected refProp to be a secret property, but it was not: %v", oneOfResourceSpec.InputProperties["refProp"])
	assert.Truef(t, oneOfResourceSpec.InputProperties["directStringProp"].Secret, "Expected directStringProp to be a secret property, but it was not: %v", oneOfResourceSpec.InputProperties["directStringProp"])
	assert.Falsef(t, oneOfResourceSpec.InputProperties["notSecretProp"].Secret, "Expected notSecretProp to not be a secret property, but it was: %v", oneOfResourceSpec.InputProperties["notSecretProp"])
	assert.Truef(t, oneOfResourceSpec.Properties["aProp"].Secret, "Expected output property aProp to be a secret property, but it was not: %v", oneOfResourceSpec.Properties["aProp"])
	assert.Falsef(t, oneOfResourceSpec.Properties["notSecretProp"].Secret, "Expected output property notSecretProp to not be a secret property, but it was: %v", oneOfResourceSpec.Properties["notSecretProp"])

}
