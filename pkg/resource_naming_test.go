package pkg

import (
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestGetResourceTitleFromOperationID_PUToperationIdWithSetting(t *testing.T) {
	assert.Equal(t, "TableSetting", getResourceTitleFromOperationID("tableSetting", http.MethodPut, false))
	assert.Equal(t, "TableSetting", getResourceTitleFromOperationID("table_setting", http.MethodPut, false))
	assert.Equal(t, "TheTableSetting", getResourceTitleFromOperationID("setTheTableSetting", http.MethodPut, false))
	assert.Equal(t, "StandardOperationIdName", getResourceTitleFromOperationID("standardOperationIdName", http.MethodPut, false))
	assert.Equal(t, "TheThing", getResourceTitleFromOperationID("updateTheThing", http.MethodPut, false))
	assert.Equal(t, "Machines", getResourceTitleFromOperationID("machines_show", "GET", false))
}

func TestGetResourceTitleFromOperationID_PUToperationIdWithSetup(t *testing.T) {
	assert.Equal(t, "SetupKeys", getResourceTitleFromOperationID("setupKeys", http.MethodPut, false))
	assert.Equal(t, "SetupKeys", getResourceTitleFromOperationID("setup_keys", http.MethodPut, false))
	assert.Equal(t, "SetupKeys", getResourceTitleFromOperationID("setSetupKeys", http.MethodPut, false))
}

func TestGetResourceTitleFromOperationID_POSToperationIdWithSetup(t *testing.T) {
	assert.Equal(t, "SetupKeys", getResourceTitleFromOperationID("createSetupKeys", http.MethodPost, false))
	assert.Equal(t, "SetupKeys", getResourceTitleFromOperationID("setupKeys", http.MethodPost, false))
}

func TestGetResourceTitleFromOperationID_POSToperationIdWithPosture(t *testing.T) {
	assert.Equal(t, "Posture", getResourceTitleFromOperationID("createPosture", http.MethodPost, false))
	assert.Equal(t, "PostureRules", getResourceTitleFromOperationID("postureRules", http.MethodPost, false))
	assert.Equal(t, "PostureRules", getResourceTitleFromOperationID("posture_rules", http.MethodPost, false))
}

func TestGetResourceTitleFromOperationID_PUToperationIdWithPosture(t *testing.T) {
	assert.Equal(t, "PostureRules", getResourceTitleFromOperationID("updatePostureRules", http.MethodPut, false))
	assert.Equal(t, "PostureRules", getResourceTitleFromOperationID("setPostureRules", http.MethodPut, false))
}

func TestGetResourceTitleFromRequestSchema(t *testing.T) {
	s := openapi3.NewSchema()
	assert.Equal(t, "MachinesShow", getResourceTitleFromRequestSchema("machines_show", openapi3.NewSchemaRef("", s)))
}
