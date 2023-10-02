package pkg

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"

	pschema "github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

const packageName = "fake-package"

var testOpenAPIDocBytes []byte
var testOpenAPIDoc *openapi3.T

var testPulumiPkg = pschema.PackageSpec{
	Name:        packageName,
	Description: "A Pulumi package for creating and managing FakeCloud resources.",
	DisplayName: "FakePackage",
	License:     "Apache-2.0",
	Keywords: []string{
		"pulumi",
		packageName,
		"category/cloud",
		"kind/native",
	},
	Homepage:   "https://cloudysky.software",
	Publisher:  "Cloudy Sky Software",
	Repository: "https://github.com/cloudy-sky-software/pulumi-fakecloud",

	Config: pschema.ConfigSpec{
		Variables: map[string]pschema.PropertySpec{
			"apiKey": {
				Description: "The API key",
				TypeSpec:    pschema.TypeSpec{Type: "string"},
				Language: map[string]pschema.RawMessage{
					"csharp": rawMessage(map[string]interface{}{
						"name": "ApiKey",
					}),
				},
				Secret: true,
			},
		},
	},

	Provider: pschema.ResourceSpec{
		ObjectTypeSpec: pschema.ObjectTypeSpec{
			Description: "The provider type for the FakeCloud package.",
			Type:        "object",
		},
		InputProperties: map[string]pschema.PropertySpec{
			"apiKey": {
				DefaultInfo: &pschema.DefaultSpec{
					Environment: []string{
						"FAKECLOUD_APIKEY",
					},
				},
				Description: "The FakeCloud API key.",
				TypeSpec:    pschema.TypeSpec{Type: "string"},
				Language: map[string]pschema.RawMessage{
					"csharp": rawMessage(map[string]interface{}{
						"name": "ApiKey",
					}),
				},
				Secret: true,
			},
		},
	},

	PluginDownloadURL: "github://api.github.com/cloudy-sky-software/pulumi-fake-cloud",
	Types:             map[string]pschema.ComplexTypeSpec{},
	Resources:         map[string]pschema.ResourceSpec{},
	Functions:         map[string]pschema.FunctionSpec{},
	Language:          map[string]pschema.RawMessage{},
}

func mustReadTestOpenAPIDoc(t *testing.T, path string) {
	t.Helper()

	assert.NotEmpty(t, path, "Path to an OpenAPI spec file is required")

	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Failed reading openapi.yml:")
		panic(err)
	}

	testOpenAPIDocBytes = b

	doc, err := openapi3.NewLoader().LoadFromData(testOpenAPIDocBytes)
	if err != nil {
		fmt.Println("Failed to load openapi.yml:")
		panic(err)
	}

	ctx := context.Background()

	if err := doc.Validate(ctx, openapi3.DisableExamplesValidation()); err != nil {
		fmt.Println("OpenAPI spec failed validation:")
		panic(err)
	}

	testOpenAPIDoc = doc
}

func TestMain(m *testing.M) {
	m.Run()
}
