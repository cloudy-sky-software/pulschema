package pkg

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestGetResourceTitleFromOperationID_PUToperationIdWithSetting(t *testing.T) {
	assert.Equal(t, "TableSetting", getResourceTitleFromOperationID("tableSetting", "PUT", false))
	assert.Equal(t, "TableSetting", getResourceTitleFromOperationID("table_setting", "PUT", false))
	assert.Equal(t, "TheTableSetting", getResourceTitleFromOperationID("setTheTableSetting", "PUT", false))
	assert.Equal(t, "StandardOperationIdName", getResourceTitleFromOperationID("standardOperationIdName", "PUT", false))
	assert.Equal(t, "TheThing", getResourceTitleFromOperationID("updateTheThing", "PUT", false))
	assert.Equal(t, "Machines", getResourceTitleFromOperationID("machines_show", "GET", false))
}

func TestGetResourceTitleFromOperationID_PUToperationIdWithSetup(t *testing.T) {
	assert.Equal(t, "SetupKeys", getResourceTitleFromOperationID("setupKeys", "PUT", false))
	assert.Equal(t, "SetupKeys", getResourceTitleFromOperationID("setup_keys", "PUT", false))
	assert.Equal(t, "SetupKeys", getResourceTitleFromOperationID("setSetupKeys", "PUT", false))
}

func TestGetResourceTitleFromOperationID_POSToperationIdWithSetup(t *testing.T) {
	assert.Equal(t, "SetupKeys", getResourceTitleFromOperationID("createSetupKeys", "POST", false))
	assert.Equal(t, "SetupKeys", getResourceTitleFromOperationID("setupKeys", "POST", false))
}

func TestGetResourceTitleFromOperationID_POSToperationIdWithPosture(t *testing.T) {
	assert.Equal(t, "Posture", getResourceTitleFromOperationID("createPosture", "POST", false))
	assert.Equal(t, "PostureRules", getResourceTitleFromOperationID("postureRules", "POST", false))
	assert.Equal(t, "PostureRules", getResourceTitleFromOperationID("posture_rules", "POST", false))
}

func TestGetResourceTitleFromOperationID_PUToperationIdWithPosture(t *testing.T) {
	assert.Equal(t, "PostureRules", getResourceTitleFromOperationID("updatePostureRules", "PUT", false))
	assert.Equal(t, "PostureRules", getResourceTitleFromOperationID("setPostureRules", "PUT", false))
}

func TestGetResourceTitleFromRequestSchema(t *testing.T) {
	s := openapi3.NewSchema()
	assert.Equal(t, "MachinesShow", getResourceTitleFromRequestSchema("machines_show", openapi3.NewSchemaRef("", s)))
}
