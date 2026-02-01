package exclusions

import (
	"testing"
)

func TestExactMatcher(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{
			name:    "exact match",
			pattern: "/api/users",
			path:    "/api/users",
			want:    true,
		},
		{
			name:    "no match - different path",
			pattern: "/api/users",
			path:    "/api/posts",
			want:    false,
		},
		{
			name:    "no match - subset",
			pattern: "/api/users",
			path:    "/api/users/123",
			want:    false,
		},
		{
			name:    "no match - superset",
			pattern: "/api/users/123",
			path:    "/api/users",
			want:    false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewExactMatcher(tt.pattern)
			if got := matcher.Matches(tt.path); got != tt.want {
				t.Errorf("ExactMatcher.Matches() = %v, want %v", got, tt.want)
			}
			if matcher.Pattern() != tt.pattern {
				t.Errorf("ExactMatcher.Pattern() = %v, want %v", matcher.Pattern(), tt.pattern)
			}
			if matcher.Type() != PatternTypeExact {
				t.Errorf("ExactMatcher.Type() = %v, want %v", matcher.Type(), PatternTypeExact)
			}
		})
	}
}

func TestWildcardMatcher(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
		wantErr bool
	}{
		{
			name:    "single wildcard - match",
			pattern: "/api/*/users",
			path:    "/api/v1/users",
			want:    true,
		},
		{
			name:    "single wildcard - no match across slash",
			pattern: "/api/*/users",
			path:    "/api/v1/test/users",
			want:    false,
		},
		{
			name:    "double wildcard - match across slashes",
			pattern: "/api/**/users",
			path:    "/api/v1/test/users",
			want:    true,
		},
		{
			name:    "double wildcard - match single level",
			pattern: "/api/**/users",
			path:    "/api/v1/users",
			want:    true,
		},
		{
			name:    "trailing wildcard - match",
			pattern: "/api/users/*",
			path:    "/api/users/123",
			want:    true,
		},
		{
			name:    "trailing wildcard - match multiple levels with **",
			pattern: "/api/users/**",
			path:    "/api/users/123/posts/456",
			want:    true,
		},
		{
			name:    "trailing wildcard - no match different base",
			pattern: "/api/users/*",
			path:    "/api/posts/123",
			want:    false,
		},
		{
			name:    "multiple wildcards",
			pattern: "/api/*/users/*/posts",
			path:    "/api/v1/users/123/posts",
			want:    true,
		},
		{
			name:    "question mark wildcard - single char",
			pattern: "/api/v?/users",
			path:    "/api/v1/users",
			want:    true,
		},
		{
			name:    "question mark wildcard - no match multiple chars",
			pattern: "/api/v?/users",
			path:    "/api/v12/users",
			want:    false,
		},
		{
			name:    "wildcard at start",
			pattern: "*/users",
			path:    "/api/users",
			want:    true,
		},
		{
			name:    "complex pattern",
			pattern: "/api/*/users/**/posts/*.json",
			path:    "/api/v1/users/123/456/posts/789.json",
			want:    true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := NewWildcardMatcher(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWildcardMatcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			
			if got := matcher.Matches(tt.path); got != tt.want {
				t.Errorf("WildcardMatcher.Matches() = %v, want %v (pattern: %s, path: %s)", 
					got, tt.want, tt.pattern, tt.path)
			}
			if matcher.Pattern() != tt.pattern {
				t.Errorf("WildcardMatcher.Pattern() = %v, want %v", matcher.Pattern(), tt.pattern)
			}
			if matcher.Type() != PatternTypeWildcard {
				t.Errorf("WildcardMatcher.Type() = %v, want %v", matcher.Type(), PatternTypeWildcard)
			}
		})
	}
}

func TestRegexMatcher(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
		wantErr bool
	}{
		{
			name:    "simple regex",
			pattern: "^/api/v[0-9]+/users$",
			path:    "/api/v1/users",
			want:    true,
		},
		{
			name:    "simple regex - no match",
			pattern: "^/api/v[0-9]+/users$",
			path:    "/api/v1/posts",
			want:    false,
		},
		{
			name:    "complex regex",
			pattern: "^/api/v[0-9]+/(users|posts)/[0-9]+$",
			path:    "/api/v1/users/123",
			want:    true,
		},
		{
			name:    "complex regex - alternative match",
			pattern: "^/api/v[0-9]+/(users|posts)/[0-9]+$",
			path:    "/api/v2/posts/456",
			want:    true,
		},
		{
			name:    "any internal path",
			pattern: ".*/internal/.*",
			path:    "/api/v1/internal/debug",
			want:    true,
		},
		{
			name:    "invalid regex",
			pattern: "[invalid",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := NewRegexMatcher(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRegexMatcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			
			if got := matcher.Matches(tt.path); got != tt.want {
				t.Errorf("RegexMatcher.Matches() = %v, want %v", got, tt.want)
			}
			if matcher.Pattern() != tt.pattern {
				t.Errorf("RegexMatcher.Pattern() = %v, want %v", matcher.Pattern(), tt.pattern)
			}
			if matcher.Type() != PatternTypeRegex {
				t.Errorf("RegexMatcher.Type() = %v, want %v", matcher.Type(), PatternTypeRegex)
			}
		})
	}
}

func TestNewPathMatcher(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		patternType PatternType
		wantType    PatternType
		wantErr     bool
	}{
		{
			name:        "exact matcher",
			pattern:     "/api/users",
			patternType: PatternTypeExact,
			wantType:    PatternTypeExact,
		},
		{
			name:        "wildcard matcher",
			pattern:     "/api/users/*",
			patternType: PatternTypeWildcard,
			wantType:    PatternTypeWildcard,
		},
		{
			name:        "regex matcher",
			pattern:     "^/api/.*",
			patternType: PatternTypeRegex,
			wantType:    PatternTypeRegex,
		},
		{
			name:        "empty pattern",
			pattern:     "",
			patternType: PatternTypeExact,
			wantErr:     true,
		},
		{
			name:        "invalid pattern type",
			pattern:     "/api/users",
			patternType: "invalid",
			wantErr:     true,
		},
		{
			name:        "invalid regex",
			pattern:     "[invalid",
			patternType: PatternTypeRegex,
			wantErr:     true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := NewPathMatcher(tt.pattern, tt.patternType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPathMatcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			
			if matcher.Type() != tt.wantType {
				t.Errorf("NewPathMatcher() type = %v, want %v", matcher.Type(), tt.wantType)
			}
		})
	}
}

func TestWildcardToRegex(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		testPath string
		want     bool
	}{
		{
			name:     "special chars are escaped",
			pattern:  "/api/users.json",
			testPath: "/api/users.json",
			want:     true,
		},
		{
			name:     "dot not treated as wildcard",
			pattern:  "/api/users.json",
			testPath: "/api/userszjson",
			want:     false,
		},
		{
			name:     "plus is escaped",
			pattern:  "/api/users+posts",
			testPath: "/api/users+posts",
			want:     true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := NewWildcardMatcher(tt.pattern)
			if err != nil {
				t.Fatalf("NewWildcardMatcher() error = %v", err)
			}
			
			if got := matcher.Matches(tt.testPath); got != tt.want {
				t.Errorf("wildcardToRegex() matches = %v, want %v", got, tt.want)
			}
		})
	}
}
