package exclusions

import (
	"net/http"
	"testing"
)

const (
	pathAPIUsers       = "/api/users"
	pathAPIPosts       = "/api/posts"
	pathAPIUsersWild   = "/api/users/*"
	pathAPIUsers123    = "/api/users/123"
	pathAPIItems       = "/api/items"
	pathAPIV1UsersWild = "/api/v1/users/*"
	pathInternalGlob   = "/internal/**"
	pathInternalDebug  = "/internal/debug"
	invalidStr         = "invalid"
	invalidBracketStr  = "[invalid"
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
			legacyPaths: []string{pathAPIUsers, pathAPIPosts},
			wantCount:   2,
		},
		{
			name: "only new exclusions",
			exclusions: []Exclusion{
				{PathPattern: pathAPIUsersWild, PatternType: PatternTypeWildcard},
				{Method: http.MethodGet, PathPattern: pathAPIPosts, PatternType: PatternTypeExact},
			},
			wantCount: 2,
		},
		{
			name: "mixed legacy and new",
			exclusions: []Exclusion{
				{PathPattern: pathAPIUsersWild},
			},
			legacyPaths: []string{pathAPIPosts},
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
				{Method: "INVALID", PathPattern: pathAPIUsers},
			},
			wantErr: true,
		},
		{
			name: "invalid exclusion - invalid pattern type",
			exclusions: []Exclusion{
				{PathPattern: pathAPIUsers, PatternType: PatternType(invalidStr)},
			},
			wantErr: true,
		},
		{
			name: "invalid exclusion - invalid regex",
			exclusions: []Exclusion{
				{PathPattern: invalidBracketStr, PatternType: PatternTypeRegex},
			},
			wantErr: true,
		},
		{
			name: "valid methods",
			exclusions: []Exclusion{
				{Method: http.MethodGet, PathPattern: pathAPIUsers},
				{Method: "post", PathPattern: pathAPIPosts}, // lowercase should work
				{Method: http.MethodPut, PathPattern: pathAPIItems},
				{Method: http.MethodPatch, PathPattern: pathAPIItems},
				{Method: http.MethodDelete, PathPattern: pathAPIItems},
			},
			wantCount: 5,
		},
		{
			name:        "ignore empty legacy paths",
			legacyPaths: []string{"", pathAPIUsers, ""},
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
			legacyPaths: []string{pathAPIUsers},
			testCases: []struct {
				method string
				path   string
				want   bool
			}{
				{http.MethodGet, pathAPIUsers, true},
				{http.MethodPost, pathAPIUsers, true},
				{http.MethodGet, pathAPIUsers123, false},
				{http.MethodGet, pathAPIPosts, false},
			},
		},
		{
			name: "wildcard pattern - all methods",
			exclusions: []Exclusion{
				{PathPattern: pathAPIUsersWild},
			},
			testCases: []struct {
				method string
				path   string
				want   bool
			}{
				{http.MethodGet, pathAPIUsers123, true},
				{http.MethodPost, pathAPIUsers123, true},
				{http.MethodDelete, pathAPIUsers123, true},
				{http.MethodGet, pathAPIUsers, false},
				{http.MethodGet, "/api/users/123/posts", false},
			},
		},
		{
			name: "wildcard pattern - specific method",
			exclusions: []Exclusion{
				{Method: http.MethodGet, PathPattern: pathAPIUsersWild},
			},
			testCases: []struct {
				method string
				path   string
				want   bool
			}{
				{http.MethodGet, pathAPIUsers123, true},
				{http.MethodPost, pathAPIUsers123, false},
				{http.MethodDelete, pathAPIUsers123, false},
			},
		},
		{
			name: "double wildcard",
			exclusions: []Exclusion{
				{PathPattern: pathInternalGlob},
			},
			testCases: []struct {
				method string
				path   string
				want   bool
			}{
				{http.MethodGet, pathInternalDebug, true},
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
				{Method: http.MethodGet, PathPattern: pathAPIUsersWild},
				{Method: http.MethodPost, PathPattern: "/api/posts/*"},
				{PathPattern: pathInternalGlob},
			},
			testCases: []struct {
				method string
				path   string
				want   bool
			}{
				{http.MethodGet, pathAPIUsers123, true},
				{http.MethodPost, pathAPIUsers123, false},
				{http.MethodPost, "/api/posts/456", true},
				{http.MethodGet, "/api/posts/456", false},
				{http.MethodDelete, pathInternalDebug, true},
				{http.MethodGet, "/internal/metrics/cpu", true},
			},
		},
		{
			name: "case insensitive method matching",
			exclusions: []Exclusion{
				{Method: "get", PathPattern: pathAPIUsers},
			},
			testCases: []struct {
				method string
				path   string
				want   bool
			}{
				{http.MethodGet, pathAPIUsers, true},
				{"get", pathAPIUsers, true},
				{"Get", pathAPIUsers, true},
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
		{Method: http.MethodGet, PathPattern: pathAPIUsersWild, PatternType: PatternTypeWildcard},
		{Method: http.MethodGet, PathPattern: pathAPIV1UsersWild, PatternType: PatternTypeWildcard},
		{PathPattern: pathInternalGlob, PatternType: PatternTypeWildcard},
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
			path:      pathAPIUsers123,
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
			path:      pathAPIPosts,
			wantCount: 0,
		},
		{
			name:      "internal path matches all methods",
			method:    http.MethodDelete,
			path:      pathInternalDebug,
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
			matcherPattern: pathAPIUsers,
			testMethod:     http.MethodGet,
			testPath:       pathAPIUsers,
			want:           true,
		},
		{
			name:           "method mismatch",
			matcherMethod:  http.MethodGet,
			matcherPattern: pathAPIUsers,
			testMethod:     http.MethodPost,
			testPath:       pathAPIUsers,
			want:           false,
		},
		{
			name:           "path mismatch",
			matcherMethod:  http.MethodGet,
			matcherPattern: pathAPIUsers,
			testMethod:     http.MethodGet,
			testPath:       pathAPIPosts,
			want:           false,
		},
		{
			name:           "all methods wildcard",
			matcherMethod:  "",
			matcherPattern: pathAPIUsers,
			testMethod:     http.MethodDelete,
			testPath:       pathAPIUsers,
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
				PathPattern: pathAPIUsersWild,
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
				PathPattern: pathAPIUsers,
			},
			wantErr: true,
		},
		{
			name: "invalid pattern type",
			exclusion: Exclusion{
				PathPattern: pathAPIUsers,
				PatternType: PatternType(invalidStr),
			},
			wantErr: true,
		},
		{
			name: "valid - no method",
			exclusion: Exclusion{
				PathPattern: pathAPIUsersWild,
			},
			wantErr: false,
		},
		{
			name: "valid - no pattern type (defaults to wildcard)",
			exclusion: Exclusion{
				PathPattern: pathAPIUsersWild,
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
