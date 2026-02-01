package exclusions

import (
	"fmt"
	"strings"
)

// Exclusion represents a single exclusion rule
type Exclusion struct {
	// Method is the HTTP method to match (e.g., GET, POST, PUT, DELETE).
	// If empty, the exclusion applies to all methods.
	Method string `json:"method,omitempty"`

	// PathPattern is the pattern to match against paths
	PathPattern string `json:"pathPattern"`

	// PatternType specifies how to interpret PathPattern
	// Valid values: "exact", "wildcard", "regex"
	// Default: "wildcard"
	PatternType PatternType `json:"patternType,omitempty"`
}

// EndpointMatcher matches a specific endpoint (method + path combination)
type EndpointMatcher struct {
	method      string // empty means all methods
	pathMatcher PathMatcher
}

// Matches returns true if the endpoint matches this matcher
func (m *EndpointMatcher) Matches(method, path string) bool {
	// If method is specified and doesn't match, return false
	if m.method != "" && !strings.EqualFold(m.method, method) {
		return false
	}

	// Check if path matches
	return m.pathMatcher.Matches(path)
}

// ExclusionEvaluator evaluates whether an endpoint should be excluded
type ExclusionEvaluator struct {
	matchers []EndpointMatcher
}

// NewExclusionEvaluator creates a new exclusion evaluator
func NewExclusionEvaluator(exclusions []Exclusion, legacyPaths []string) (*ExclusionEvaluator, error) {
	evaluator := &ExclusionEvaluator{
		matchers: make([]EndpointMatcher, 0, len(exclusions)+len(legacyPaths)),
	}

	// Convert legacy paths to exclusions (exact match, all methods)
	for _, path := range legacyPaths {
		if path == "" {
			continue
		}

		matcher := EndpointMatcher{
			method:      "", // all methods
			pathMatcher: NewExactMatcher(path),
		}
		evaluator.matchers = append(evaluator.matchers, matcher)
	}

	// Process new exclusions
	for i, excl := range exclusions {
		if err := validateExclusion(&excl); err != nil {
			return nil, fmt.Errorf("invalid exclusion at index %d: %w", i, err)
		}

		// Default pattern type to wildcard if not specified
		patternType := excl.PatternType
		if patternType == "" {
			patternType = PatternTypeWildcard
		}

		pathMatcher, err := NewPathMatcher(excl.PathPattern, patternType)
		if err != nil {
			return nil, fmt.Errorf("failed to create path matcher for exclusion at index %d: %w", i, err)
		}

		matcher := EndpointMatcher{
			method:      strings.ToUpper(excl.Method),
			pathMatcher: pathMatcher,
		}
		evaluator.matchers = append(evaluator.matchers, matcher)
	}

	return evaluator, nil
}

// validateExclusion validates an exclusion configuration
func validateExclusion(excl *Exclusion) error {
	if excl.PathPattern == "" {
		return fmt.Errorf("pathPattern is required")
	}

	// Validate method if specified
	if excl.Method != "" {
		method := strings.ToUpper(excl.Method)
		validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "CONNECT", "TRACE"}
		valid := false
		for _, vm := range validMethods {
			if method == vm {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid HTTP method: %s", excl.Method)
		}
	}

	// Validate pattern type if specified
	if excl.PatternType != "" {
		switch excl.PatternType {
		case PatternTypeExact, PatternTypeWildcard, PatternTypeRegex:
			// valid
		default:
			return fmt.Errorf("invalid pattern type: %s (must be exact, wildcard, or regex)", excl.PatternType)
		}
	}

	return nil
}

// ShouldExclude returns true if the given endpoint should be excluded
func (e *ExclusionEvaluator) ShouldExclude(method, path string) bool {
	method = strings.ToUpper(method)

	for _, matcher := range e.matchers {
		if matcher.Matches(method, path) {
			return true
		}
	}

	return false
}

// GetMatchingExclusions returns all matchers that match the given endpoint
// This is useful for debugging and logging purposes
func (e *ExclusionEvaluator) GetMatchingExclusions(method, path string) []string {
	method = strings.ToUpper(method)
	var matches []string

	for _, matcher := range e.matchers {
		if matcher.Matches(method, path) {
			methodStr := matcher.method
			if methodStr == "" {
				methodStr = "*"
			}
			matches = append(matches, fmt.Sprintf("%s %s (%s)",
				methodStr,
				matcher.pathMatcher.Pattern(),
				matcher.pathMatcher.Type()))
		}
	}

	return matches
}

// Count returns the total number of exclusion matchers
func (e *ExclusionEvaluator) Count() int {
	return len(e.matchers)
}
