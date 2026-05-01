package exclusions

import (
	"net/http"
	"testing"
)

const (
	testAPIPostsPath     = "/api/posts"
	testAPIUsersWildPath = "/api/users/*"
	testInvalidValue     = "invalid"
)

func TestNewExclusionEvaluator(t *testing.T) {
	tests := []struct {
		name        string
		exclusions  []Exclusion
		legacyPaths []string
		wantErr     bool
		wantCount   int
	}{
		{
			name:        "empty exclusions",
			exclusions:  []Exclusion{},
			legacyPaths: []string{},
			wantCount:   0,
		},
		{
			name:        "only legacy paths",
			legacyPaths: []string{"/api/users", testAPIPostsPath},
			wantCount:   2,
		},
		{
			name: "only new exclusions",
			exclusions: []Exclusion{
				{PathPattern: testAPIUsersWildPath, PatternType: PatternTypeWildcard},
				{Method: http.MethodGet, PathPattern: testAPIPostsPath, PatternType: PatternTypeExact},
			},
			wantCount: 2,
		},
		{
			name: "mixed legacy and new",
			exclusions: []Exclusion{
				{PathPattern: testAPIUsersWildPath},
			},
			legacyPaths: []string{testAPIPostsPath},
			wantCount:   2,
		},
		{
			name: "invalid exclusion - empty pattern",
			exclusions: []Exclusion{
				{PathPattern: ""},
			},
			wantErr: true,
		},
		{
			name: "invalid exclusion - invalid method",
			exclusions: []Exclusion{
				{Method: "INVALID", PathPattern: "/api/users"},
			},
			wantErr: true,
		},
		{
			name: "invalid exclusion - invalid pattern type",
			exclusions: []Exclusion{
				{PathPattern: "/api/users", PatternType: testInvalidValue},
			},
			wantErr: true,
		},
		{
			name: "invalid exclusion - invalid regex",
			exclusions: []Exclusion{
				{PathPattern: "[invalid", PatternType: PatternTypeRegex},
			},
			wantErr: true,
		},
		{
			name: "valid methods",
			exclusions: []Exclusion{
				{Method: http.MethodGet, PathPattern: "/api/users"},
				{Method: "post", PathPattern: testAPIPostsPath}, // lowercase should work
				{Method: http.MethodPut, PathPattern: "/api/items"},
				{Method: http.MethodPatch, PathPattern: "/api/items"},
				{Method: http.MethodDelete, PathPattern: "/api/items"},
			},
			wantCount: 5,
		},
		{
			name:        "ignore empty legacy paths",
			legacyPaths: []string{"", "/api/users", ""},
			wantCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator, err := NewExclusionEvaluator(tt.exclusions, tt.legacyPaths)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewExclusionEvaluator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if evaluator.Count() != tt.wantCount {
				t.Errorf("NewExclusionEvaluator() count = %v, want %v", evaluator.Count(), tt.wantCount)
			}
		})
	}
}

func TestExclusionEvaluator_ShouldExclude(t *testing.T) {
	tests := []struct {
		name        string
		exclusions  []Exclusion
		legacyPaths []string
		testCases   []struct {
			method string
			path   string
			want   bool
		}
	}{
		{
			name:        "legacy exact match",
			legacyPaths: []string{"/api/users"},
			testCases: []struct {
				method string
				path   string
				want   bool
			}{
				{http.MethodGet, "/api/users", true},
				{http.MethodPost, "/api/users", true},
				{http.MethodGet, "/api/users/123", false},
				{http.MethodGet, testAPIPostsPath, false},
			},
		},
		{
			name: "wildcard pattern - all methods",
			exclusions: []Exclusion{
				{PathPattern: testAPIUsersWildPath},
			},
			testCases: []struct {
				method string
				path   string
				want   bool
			}{
				{http.MethodGet, "/api/users/123", true},
				{http.MethodPost, "/api/users/123", true},
				{http.MethodDelete, "/api/users/123", true},
				{http.MethodGet, "/api/users", false},
				{http.MethodGet, "/api/users/123/posts", false},
			},
		},
		{
			name: "wildcard pattern - specific method",
			exclusions: []Exclusion{
				{Method: http.MethodGet, PathPattern: testAPIUsersWildPath},
			},
			testCases: []struct {
				method string
				path   string
				want   bool
			}{
				{http.MethodGet, "/api/users/123", true},
				{http.MethodPost, "/api/users/123", false},
				{http.MethodDelete, "/api/users/123", false},
			},
		},
		{
			name: "double wildcard",
			exclusions: []Exclusion{
				{PathPattern: "/internal/**"},
			},
			testCases: []struct {
				method string
				path   string
				want   bool
			}{
				{http.MethodGet, "/internal/debug", true},
				{http.MethodGet, "/internal/debug/metrics", true},
				{http.MethodGet, "/internal/debug/metrics/cpu", true},
				{http.MethodGet, "/api/internal/debug", false},
			},
		},
		{
			name: "regex pattern",
			exclusions: []Exclusion{
				{PathPattern: "^/api/v[0-9]+/internal/.*", PatternType: PatternTypeRegex},
			},
			testCases: []struct {
				method string
				path   string
				want   bool
			}{
				{http.MethodGet, "/api/v1/internal/debug", true},
				{http.MethodGet, "/api/v2/internal/metrics", true},
				{http.MethodGet, "/api/vX/internal/debug", false},
				{http.MethodGet, "/api/v1/public/users", false},
			},
		},
		{
			name: "multiple exclusions",
			exclusions: []Exclusion{
				{Method: http.MethodGet, PathPattern: testAPIUsersWildPath},
				{Method: http.MethodPost, PathPattern: "/api/posts/*"},
				{PathPattern: "/internal/**"},
			},
			testCases: []struct {
				method string
				path   string
				want   bool
			}{
				{http.MethodGet, "/api/users/123", true},
				{http.MethodPost, "/api/users/123", false},
				{http.MethodPost, "/api/posts/456", true},
				{http.MethodGet, "/api/posts/456", false},
				{http.MethodDelete, "/internal/debug", true},
				{http.MethodGet, "/internal/metrics/cpu", true},
			},
		},
		{
			name: "case insensitive method matching",
			exclusions: []Exclusion{
				{Method: "get", PathPattern: "/api/users"},
			},
			testCases: []struct {
				method string
				path   string
				want   bool
			}{
				{http.MethodGet, "/api/users", true},
				{"get", "/api/users", true},
				{"Get", "/api/users", true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator, err := NewExclusionEvaluator(tt.exclusions, tt.legacyPaths)
			if err != nil {
				t.Fatalf("NewExclusionEvaluator() error = %v", err)
			}

			for _, tc := range tt.testCases {
				got := evaluator.ShouldExclude(tc.method, tc.path)
				if got != tc.want {
					t.Errorf("ShouldExclude(%s, %s) = %v, want %v",
						tc.method, tc.path, got, tc.want)
				}
			}
		})
	}
}

func TestExclusionEvaluator_GetMatchingExclusions(t *testing.T) {
	exclusions := []Exclusion{
		{Method: http.MethodGet, PathPattern: testAPIUsersWildPath, PatternType: PatternTypeWildcard},
		{Method: http.MethodGet, PathPattern: "/api/v1/users/*", PatternType: PatternTypeWildcard},
		{PathPattern: "/internal/**", PatternType: PatternTypeWildcard},
		{PathPattern: "^/api/v[0-9]+/.*", PatternType: PatternTypeRegex},
	}

	evaluator, err := NewExclusionEvaluator(exclusions, nil)
	if err != nil {
		t.Fatalf("NewExclusionEvaluator() error = %v", err)
	}

	tests := []struct {
		name      string
		method    string
		path      string
		wantCount int
	}{
		{
			name:      "single match",
			method:    http.MethodGet,
			path:      "/api/users/123",
			wantCount: 1,
		},
		{
			name:      "multiple matches",
			method:    http.MethodGet,
			path:      "/api/v1/users/123",
			wantCount: 2, // matches both GET /api/users/* and regex
		},
		{
			name:      "no matches",
			method:    http.MethodPost,
			path:      testAPIPostsPath,
			wantCount: 0,
		},
		{
			name:      "internal path matches all methods",
			method:    http.MethodDelete,
			path:      "/internal/debug",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := evaluator.GetMatchingExclusions(tt.method, tt.path)
			if len(matches) != tt.wantCount {
				t.Errorf("GetMatchingExclusions(%s, %s) returned %d matches, want %d. Matches: %v",
					tt.method, tt.path, len(matches), tt.wantCount, matches)
			}
		})
	}
}

func TestEndpointMatcher_Matches(t *testing.T) {
	tests := []struct {
		name           string
		matcherMethod  string
		matcherPattern string
		testMethod     string
		testPath       string
		want           bool
	}{
		{
			name:           "exact method and path match",
			matcherMethod:  http.MethodGet,
			matcherPattern: "/api/users",
			testMethod:     http.MethodGet,
			testPath:       "/api/users",
			want:           true,
		},
		{
			name:           "method mismatch",
			matcherMethod:  http.MethodGet,
			matcherPattern: "/api/users",
			testMethod:     http.MethodPost,
			testPath:       "/api/users",
			want:           false,
		},
		{
			name:           "path mismatch",
			matcherMethod:  http.MethodGet,
			matcherPattern: "/api/users",
			testMethod:     http.MethodGet,
			testPath:       testAPIPostsPath,
			want:           false,
		},
		{
			name:           "all methods wildcard",
			matcherMethod:  "",
			matcherPattern: "/api/users",
			testMethod:     http.MethodDelete,
			testPath:       "/api/users",
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := EndpointMatcher{
				method:      tt.matcherMethod,
				pathMatcher: NewExactMatcher(tt.matcherPattern),
			}

			if got := matcher.Matches(tt.testMethod, tt.testPath); got != tt.want {
				t.Errorf("EndpointMatcher.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateExclusion(t *testing.T) {
	tests := []struct {
		name      string
		exclusion Exclusion
		wantErr   bool
	}{
		{
			name: "valid exclusion",
			exclusion: Exclusion{
				Method:      http.MethodGet,
				PathPattern: testAPIUsersWildPath,
				PatternType: PatternTypeWildcard,
			},
			wantErr: false,
		},
		{
			name: "empty pattern",
			exclusion: Exclusion{
				PathPattern: "",
			},
			wantErr: true,
		},
		{
			name: "invalid method",
			exclusion: Exclusion{
				Method:      "INVALID",
				PathPattern: "/api/users",
			},
			wantErr: true,
		},
		{
			name: "invalid pattern type",
			exclusion: Exclusion{
				PathPattern: "/api/users",
				PatternType: testInvalidValue,
			},
			wantErr: true,
		},
		{
			name: "valid - no method",
			exclusion: Exclusion{
				PathPattern: testAPIUsersWildPath,
			},
			wantErr: false,
		},
		{
			name: "valid - no pattern type (defaults to wildcard)",
			exclusion: Exclusion{
				PathPattern: testAPIUsersWildPath,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExclusion(&tt.exclusion)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateExclusion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
