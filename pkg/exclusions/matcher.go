package exclusions

import (
	"fmt"
	"regexp"
	"strings"
)

// PatternType defines the type of pattern matching to use
type PatternType string

const (
	// PatternTypeExact matches paths exactly
	PatternTypeExact PatternType = "exact"
	// PatternTypeWildcard matches paths using wildcards (* and **)
	PatternTypeWildcard PatternType = "wildcard"
	// PatternTypeRegex matches paths using regular expressions
	PatternTypeRegex PatternType = "regex"
)

// PathMatcher is an interface for matching paths against patterns
type PathMatcher interface {
	// Matches returns true if the path matches the pattern
	Matches(path string) bool
	// Pattern returns the original pattern string
	Pattern() string
	// Type returns the pattern type
	Type() PatternType
}

// ExactMatcher matches paths exactly
type ExactMatcher struct {
	pattern string
}

// NewExactMatcher creates a new exact matcher
func NewExactMatcher(pattern string) *ExactMatcher {
	return &ExactMatcher{pattern: pattern}
}

// Matches returns true if the path exactly matches the pattern
func (m *ExactMatcher) Matches(path string) bool {
	return m.pattern == path
}

// Pattern returns the pattern string
func (m *ExactMatcher) Pattern() string {
	return m.pattern
}

// Type returns the pattern type
func (m *ExactMatcher) Type() PatternType {
	return PatternTypeExact
}

// WildcardMatcher matches paths using wildcards
// * matches any sequence of characters except /
// ** matches any sequence of characters including /
type WildcardMatcher struct {
	pattern string
	regex   *regexp.Regexp
}

// NewWildcardMatcher creates a new wildcard matcher
func NewWildcardMatcher(pattern string) (*WildcardMatcher, error) {
	// Convert wildcard pattern to regex
	regexPattern := wildcardToRegex(pattern)
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile wildcard pattern: %w", err)
	}

	return &WildcardMatcher{
		pattern: pattern,
		regex:   regex,
	}, nil
}

// wildcardToRegex converts a wildcard pattern to a regex pattern
func wildcardToRegex(pattern string) string {
	var result strings.Builder
	result.WriteString("^")

	i := 0
	for i < len(pattern) {
		if i < len(pattern)-1 && pattern[i:i+2] == "**" {
			// ** matches anything including /
			result.WriteString(".*")
			i += 2
		} else if pattern[i] == '*' {
			// * matches anything except /
			result.WriteString("[^/]*")
			i++
		} else if pattern[i] == '?' {
			// ? matches any single character except /
			result.WriteString("[^/]")
			i++
		} else {
			// Escape special regex characters
			if strings.ContainsRune(".+^$()[]{}|\\", rune(pattern[i])) {
				result.WriteRune('\\')
			}
			result.WriteByte(pattern[i])
			i++
		}
	}

	result.WriteString("$")
	return result.String()
}

// Matches returns true if the path matches the wildcard pattern
func (m *WildcardMatcher) Matches(path string) bool {
	return m.regex.MatchString(path)
}

// Pattern returns the pattern string
func (m *WildcardMatcher) Pattern() string {
	return m.pattern
}

// Type returns the pattern type
func (m *WildcardMatcher) Type() PatternType {
	return PatternTypeWildcard
}

// RegexMatcher matches paths using regular expressions
type RegexMatcher struct {
	pattern string
	regex   *regexp.Regexp
}

// NewRegexMatcher creates a new regex matcher
func NewRegexMatcher(pattern string) (*RegexMatcher, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex pattern: %w", err)
	}

	return &RegexMatcher{
		pattern: pattern,
		regex:   regex,
	}, nil
}

// Matches returns true if the path matches the regex pattern
func (m *RegexMatcher) Matches(path string) bool {
	return m.regex.MatchString(path)
}

// Pattern returns the pattern string
func (m *RegexMatcher) Pattern() string {
	return m.pattern
}

// Type returns the pattern type
func (m *RegexMatcher) Type() PatternType {
	return PatternTypeRegex
}

// NewPathMatcher creates a path matcher based on the pattern type
func NewPathMatcher(pattern string, patternType PatternType) (PathMatcher, error) {
	if pattern == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}

	switch patternType {
	case PatternTypeExact:
		return NewExactMatcher(pattern), nil
	case PatternTypeWildcard:
		return NewWildcardMatcher(pattern)
	case PatternTypeRegex:
		return NewRegexMatcher(pattern)
	default:
		return nil, fmt.Errorf("unknown pattern type: %s", patternType)
	}
}
