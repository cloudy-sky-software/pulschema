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

func TestGetResourceTitleFromRequestSchema(t *testing.T) {
	s := openapi3.NewSchema()
	assert.Equal(t, "MachinesShow", getResourceTitleFromRequestSchema("machines_show", openapi3.NewSchemaRef("", s)))
}
