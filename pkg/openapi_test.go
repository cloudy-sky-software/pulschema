package pkg

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

var testOpenAPIDocBytes []byte
var testOpenAPIDoc *openapi3.T

func mustReadTestOpenAPIDoc(t *testing.T) {
	t.Helper()

	b, err := os.ReadFile(filepath.Join("testdata", "openapi.yml"))
	if err != nil {
		t.Fatalf("Failed reading openapi.yml: %v", err)
	}

	testOpenAPIDocBytes = b

	doc, err := openapi3.NewLoader().LoadFromData(testOpenAPIDocBytes)
	if err != nil {
		t.Fatalf("Failed to load openapi.yml: %v", err)
	}

	ctx := context.Background()

	if err := doc.Validate(ctx, openapi3.DisableExamplesValidation()); err != nil {
		t.Fatalf("OpenAPI spec failed validation: %v", err)
	}

	testOpenAPIDoc = doc
}

func TestEnsureIdHierarchyInRequestPath(t *testing.T) {
	mustReadTestOpenAPIDoc(t)

	path := "/v2/uptime/checks/{check_id}/alerts/{alert_id}"
	pathItem := testOpenAPIDoc.Paths.Find(path)
	transformedPath := ensureIDHierarchyInRequestPath(path, pathItem)
	assert.Equal(t, "/v2/uptime/checks/{check_id}/alerts/{id}", transformedPath)

	checkIDParam := pathItem.Get.Parameters.GetByInAndName("path", "check_id")
	assert.NotNil(t, checkIDParam)

	idParam := pathItem.Get.Parameters.GetByInAndName("path", "id")
	assert.NotNil(t, idParam)
}
