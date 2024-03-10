package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSdkName(t *testing.T) {
	assert.Equal(t, "stringProp", ToSdkName("string_prop"))
	assert.Equal(t, "stringProp", ToSdkName("string.prop"))
	assert.Equal(t, "stringPropProp", ToSdkName("string.prop.prop"))
}

func TestStartsWithNumber(t *testing.T) {
	assert.True(t, startsWithNumber("1_var"))
	assert.True(t, startsWithNumber("1var"))
	assert.False(t, startsWithNumber("var"))
}
