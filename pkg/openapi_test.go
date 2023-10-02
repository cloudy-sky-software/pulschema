package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureIdHierarchyInRequestPath(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "openapi.yml"))

	path := "/v2/uptime/checks/{check_id}/alerts/{alert_id}"
	pathItem := testOpenAPIDoc.Paths.Find(path)
	transformedPath := ensureIDHierarchyInRequestPath(path, pathItem)
	assert.Equal(t, "/v2/uptime/checks/{check_id}/alerts/{id}", transformedPath)

	checkIDParam := pathItem.Get.Parameters.GetByInAndName("path", "check_id")
	assert.NotNil(t, checkIDParam)

	idParam := pathItem.Get.Parameters.GetByInAndName("path", "id")
	assert.NotNil(t, idParam)
}
