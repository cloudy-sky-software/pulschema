package pkg

import (
	"net/http"
	"testing"

	"github.com/cloudy-sky-software/pulschema/pkg/exclusions"
	"github.com/stretchr/testify/assert"
)

const testDebugWildcardPath = "/debug/*"

// TestExclusionIntegration tests the new exclusion feature with OpenAPIContext
func TestExclusionIntegration(t *testing.T) {
	// Test 1: Backward compatibility with ExcludedPaths
	t.Run("BackwardCompatibility", func(t *testing.T) {
		ctx := &OpenAPIContext{
			ExcludedPaths: []string{"/health", "/metrics"},
		}

		// Simulate what GatherResourcesFromAPI would do
		evaluator, err := exclusions.NewExclusionEvaluator(ctx.Exclusions, ctx.ExcludedPaths)
		assert.NoError(t, err)
		assert.NotNil(t, evaluator)

		// Test exclusions
		assert.True(t, evaluator.ShouldExclude(http.MethodGet, "/health"))
		assert.True(t, evaluator.ShouldExclude(http.MethodPost, "/health"))
		assert.True(t, evaluator.ShouldExclude(http.MethodGet, "/metrics"))
		assert.False(t, evaluator.ShouldExclude(http.MethodGet, "/api/users"))
	})

	// Test 2: New wildcard exclusions
	t.Run("WildcardExclusions", func(t *testing.T) {
		ctx := &OpenAPIContext{
			Exclusions: []exclusions.Exclusion{
				{
					PathPattern: "/internal/*",
					PatternType: exclusions.PatternTypeWildcard,
				},
			},
		}

		evaluator, err := exclusions.NewExclusionEvaluator(ctx.Exclusions, nil)
		assert.NoError(t, err)

		assert.True(t, evaluator.ShouldExclude(http.MethodGet, "/internal/debug"))
		assert.True(t, evaluator.ShouldExclude(http.MethodPost, "/internal/debug"))
		assert.False(t, evaluator.ShouldExclude(http.MethodGet, "/internal/debug/deep"))
		assert.False(t, evaluator.ShouldExclude(http.MethodGet, "/api/users"))
	})

	// Test 3: Method-specific exclusions
	t.Run("MethodSpecificExclusions", func(t *testing.T) {
		ctx := &OpenAPIContext{
			Exclusions: []exclusions.Exclusion{
				{
					Method:      http.MethodGet,
					PathPattern: testDebugWildcardPath,
					PatternType: exclusions.PatternTypeWildcard,
				},
			},
		}

		evaluator, err := exclusions.NewExclusionEvaluator(ctx.Exclusions, nil)
		assert.NoError(t, err)

		assert.True(t, evaluator.ShouldExclude(http.MethodGet, "/debug/metrics"))
		assert.False(t, evaluator.ShouldExclude(http.MethodPost, "/debug/metrics"))
		assert.False(t, evaluator.ShouldExclude(http.MethodGet, "/api/debug"))
	})

	// Test 4: Regex exclusions
	t.Run("RegexExclusions", func(t *testing.T) {
		ctx := &OpenAPIContext{
			Exclusions: []exclusions.Exclusion{
				{
					PathPattern: "^/api/v[0-9]+/test/.*",
					PatternType: exclusions.PatternTypeRegex,
				},
			},
		}

		evaluator, err := exclusions.NewExclusionEvaluator(ctx.Exclusions, nil)
		assert.NoError(t, err)

		assert.True(t, evaluator.ShouldExclude(http.MethodGet, "/api/v1/test/endpoint"))
		assert.True(t, evaluator.ShouldExclude(http.MethodPost, "/api/v2/test/resource"))
		assert.False(t, evaluator.ShouldExclude(http.MethodGet, "/api/vX/test/endpoint"))
		assert.False(t, evaluator.ShouldExclude(http.MethodGet, "/api/v1/prod/endpoint"))
	})

	// Test 5: Mixed configuration
	t.Run("MixedConfiguration", func(t *testing.T) {
		ctx := &OpenAPIContext{
			ExcludedPaths: []string{"/health"},
			Exclusions: []exclusions.Exclusion{
				{
					Method:      http.MethodGet,
					PathPattern: testDebugWildcardPath,
					PatternType: exclusions.PatternTypeWildcard,
				},
				{
					PathPattern: "/internal/**",
					PatternType: exclusions.PatternTypeWildcard,
				},
			},
		}

		evaluator, err := exclusions.NewExclusionEvaluator(ctx.Exclusions, ctx.ExcludedPaths)
		assert.NoError(t, err)

		// Legacy exclusion
		assert.True(t, evaluator.ShouldExclude(http.MethodGet, "/health"))
		assert.True(t, evaluator.ShouldExclude(http.MethodPost, "/health"))

		// Method-specific exclusion
		assert.True(t, evaluator.ShouldExclude(http.MethodGet, "/debug/metrics"))
		assert.False(t, evaluator.ShouldExclude(http.MethodPost, "/debug/metrics"))

		// Wildcard exclusion (all methods)
		assert.True(t, evaluator.ShouldExclude(http.MethodGet, "/internal/a/b/c"))
		assert.True(t, evaluator.ShouldExclude(http.MethodDelete, "/internal/x/y/z"))
	})

	// Test 6: Invalid configuration error handling
	t.Run("InvalidConfiguration", func(t *testing.T) {
		ctx := &OpenAPIContext{
			Exclusions: []exclusions.Exclusion{
				{
					PathPattern: "", // Empty pattern - should fail
				},
			},
		}

		evaluator, err := exclusions.NewExclusionEvaluator(ctx.Exclusions, nil)
		assert.Error(t, err)
		assert.Nil(t, evaluator)
	})

	// Test 7: Complex real-world scenario
	t.Run("RealWorldScenario", func(t *testing.T) {
		ctx := &OpenAPIContext{
			Exclusions: []exclusions.Exclusion{
				// Exclude all internal paths
				{
					PathPattern: "/internal/**",
					PatternType: exclusions.PatternTypeWildcard,
				},
				// Exclude GET requests to debug endpoints
				{
					Method:      http.MethodGet,
					PathPattern: testDebugWildcardPath,
					PatternType: exclusions.PatternTypeWildcard,
				},
				// Exclude versioned test endpoints using regex
				{
					PathPattern: "^/api/v[0-9]+/test/.*",
					PatternType: exclusions.PatternTypeRegex,
				},
				// Exclude specific admin operations
				{
					Method:      http.MethodPost,
					PathPattern: "/admin/users",
					PatternType: exclusions.PatternTypeExact,
				},
				{
					Method:      http.MethodDelete,
					PathPattern: "/admin/users",
					PatternType: exclusions.PatternTypeExact,
				},
			},
		}

		evaluator, err := exclusions.NewExclusionEvaluator(ctx.Exclusions, nil)
		assert.NoError(t, err)
		assert.Equal(t, 5, evaluator.Count())

		// Test each exclusion
		assert.True(t, evaluator.ShouldExclude(http.MethodGet, "/internal/secret"))
		assert.True(t, evaluator.ShouldExclude(http.MethodGet, "/debug/metrics"))
		assert.False(t, evaluator.ShouldExclude(http.MethodPost, "/debug/metrics"))
		assert.True(t, evaluator.ShouldExclude(http.MethodGet, "/api/v1/test/foo"))
		assert.False(t, evaluator.ShouldExclude(http.MethodGet, "/api/v1/prod/foo"))
		assert.True(t, evaluator.ShouldExclude(http.MethodPost, "/admin/users"))
		assert.True(t, evaluator.ShouldExclude(http.MethodDelete, "/admin/users"))
		assert.False(t, evaluator.ShouldExclude(http.MethodGet, "/admin/users"))
		assert.False(t, evaluator.ShouldExclude(http.MethodPatch, "/admin/users"))
	})
}
