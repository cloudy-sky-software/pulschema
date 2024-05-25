package pkg

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMultiplePathsUsingSameRef tests that a type used in an allOf
// definition as well as a regular reference don't cause collisions.
// In other words, the random order in which request paths in an API
// spec are sometimes processed should not add/remove such a type
// based on which request is processed first.
func TestMultiplePathsUsingSameRef(t *testing.T) {
	mustReadTestOpenAPIDoc(t, filepath.Join("testdata", "multiple_paths_using_same_refs_openapi.yml"))

	openAPICtx := &OpenAPIContext{
		Doc: *testOpenAPIDoc,
		Pkg: &testPulumiPkg,
	}

	csharpNamespaces := map[string]string{
		"": "Provider",
	}

	count := 0

	// Due to the non-deterministic nature of iterating the request
	// paths in the OpenAPI spec, we should execute this test a
	// few times to guarantee we are safe from the issue we are
	// testing for.
	for {
		if count >= 10 {
			break
		}
		_, _, err := openAPICtx.GatherResourcesFromAPI(csharpNamespaces)
		assert.Nil(t, err)

		assert.Contains(t, testPulumiPkg.Types, "fake-package:resources:Meta")

		count++
	}
}
