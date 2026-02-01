package pkg

import (
	"testing"

	"github.com/cloudy-sky-software/pulschema/pkg/exclusions"
	"github.com/stretchr/testify/assert"
)

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
		assert.True(t, evaluator.ShouldExclude("GET", "/health"))
		assert.True(t, evaluator.ShouldExclude("POST", "/health"))
		assert.True(t, evaluator.ShouldExclude("GET", "/metrics"))
		assert.False(t, evaluator.ShouldExclude("GET", "/api/users"))
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

		assert.True(t, evaluator.ShouldExclude("GET", "/internal/debug"))
		assert.True(t, evaluator.ShouldExclude("POST", "/internal/debug"))
		assert.False(t, evaluator.ShouldExclude("GET", "/internal/debug/deep"))
		assert.False(t, evaluator.ShouldExclude("GET", "/api/users"))
	})

	// Test 3: Method-specific exclusions
	t.Run("MethodSpecificExclusions", func(t *testing.T) {
		ctx := &OpenAPIContext{
			Exclusions: []exclusions.Exclusion{
				{
					Method:      "GET",
					PathPattern: "/debug/*",
					PatternType: exclusions.PatternTypeWildcard,
				},
			},
		}

		evaluator, err := exclusions.NewExclusionEvaluator(ctx.Exclusions, nil)
		assert.NoError(t, err)

		assert.True(t, evaluator.ShouldExclude("GET", "/debug/metrics"))
		assert.False(t, evaluator.ShouldExclude("POST", "/debug/metrics"))
		assert.False(t, evaluator.ShouldExclude("GET", "/api/debug"))
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

		assert.True(t, evaluator.ShouldExclude("GET", "/api/v1/test/endpoint"))
		assert.True(t, evaluator.ShouldExclude("POST", "/api/v2/test/resource"))
		assert.False(t, evaluator.ShouldExclude("GET", "/api/vX/test/endpoint"))
		assert.False(t, evaluator.ShouldExclude("GET", "/api/v1/prod/endpoint"))
	})

	// Test 5: Mixed configuration
	t.Run("MixedConfiguration", func(t *testing.T) {
		ctx := &OpenAPIContext{
			ExcludedPaths: []string{"/health"},
			Exclusions: []exclusions.Exclusion{
				{
					Method:      "GET",
					PathPattern: "/debug/*",
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
		assert.True(t, evaluator.ShouldExclude("GET", "/health"))
		assert.True(t, evaluator.ShouldExclude("POST", "/health"))

		// Method-specific exclusion
		assert.True(t, evaluator.ShouldExclude("GET", "/debug/metrics"))
		assert.False(t, evaluator.ShouldExclude("POST", "/debug/metrics"))

		// Wildcard exclusion (all methods)
		assert.True(t, evaluator.ShouldExclude("GET", "/internal/a/b/c"))
		assert.True(t, evaluator.ShouldExclude("DELETE", "/internal/x/y/z"))
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
					Method:      "GET",
					PathPattern: "/debug/*",
					PatternType: exclusions.PatternTypeWildcard,
				},
				// Exclude versioned test endpoints using regex
				{
					PathPattern: "^/api/v[0-9]+/test/.*",
					PatternType: exclusions.PatternTypeRegex,
				},
				// Exclude specific admin operations
				{
					Method:      "POST",
					PathPattern: "/admin/users",
					PatternType: exclusions.PatternTypeExact,
				},
				{
					Method:      "DELETE",
					PathPattern: "/admin/users",
					PatternType: exclusions.PatternTypeExact,
				},
			},
		}

		evaluator, err := exclusions.NewExclusionEvaluator(ctx.Exclusions, nil)
		assert.NoError(t, err)
		assert.Equal(t, 5, evaluator.Count())

		// Test each exclusion
		assert.True(t, evaluator.ShouldExclude("GET", "/internal/secret"))
		assert.True(t, evaluator.ShouldExclude("GET", "/debug/metrics"))
		assert.False(t, evaluator.ShouldExclude("POST", "/debug/metrics"))
		assert.True(t, evaluator.ShouldExclude("GET", "/api/v1/test/foo"))
		assert.False(t, evaluator.ShouldExclude("GET", "/api/v1/prod/foo"))
		assert.True(t, evaluator.ShouldExclude("POST", "/admin/users"))
		assert.True(t, evaluator.ShouldExclude("DELETE", "/admin/users"))
		assert.False(t, evaluator.ShouldExclude("GET", "/admin/users"))
		assert.False(t, evaluator.ShouldExclude("PATCH", "/admin/users"))
	})
}
