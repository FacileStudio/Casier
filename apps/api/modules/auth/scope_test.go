package auth

import "testing"

func TestScopeFromTokenRow(t *testing.T) {
	cases := []struct {
		name            string
		row             apiTokenScopeRow
		wantWrite       bool
		allowedEnvs     []string
		deniedEnvs      []string
		wantProjectID   int64
		wantTokenID     int64
		otherProjectIDs []int64
	}{
		{
			name:            "read only token cannot write",
			row:             apiTokenScopeRow{TokenID: 1, ProjectID: 7, Environments: []string{"dev"}, Permissions: "read"},
			wantWrite:       false,
			allowedEnvs:     []string{"dev"},
			deniedEnvs:      []string{"prod"},
			wantProjectID:   7,
			wantTokenID:     1,
			otherProjectIDs: []int64{8},
		},
		{
			name:          "read write token can write",
			row:           apiTokenScopeRow{TokenID: 2, ProjectID: 7, Environments: []string{"dev", "staging"}, Permissions: "read,write"},
			wantWrite:     true,
			allowedEnvs:   []string{"dev", "staging"},
			deniedEnvs:    []string{"prod"},
			wantProjectID: 7,
			wantTokenID:   2,
		},
		{
			name:          "a token marked for every environment reaches all of them",
			row:           apiTokenScopeRow{TokenID: 3, ProjectID: 9, AllEnvironments: true, Permissions: "read,write"},
			wantWrite:     true,
			allowedEnvs:   []string{"dev", "staging", "prod", "anything"},
			wantProjectID: 9,
			wantTokenID:   3,
		},
		{
			name:          "a scoped token whose grants are all gone reaches nothing",
			row:           apiTokenScopeRow{TokenID: 4, ProjectID: 9, Permissions: "read"},
			wantWrite:     false,
			deniedEnvs:    []string{"dev", "prod"},
			wantProjectID: 9,
			wantTokenID:   4,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			scope := scopeFromTokenRow(testCase.row)
			if scope == nil {
				t.Fatal("expected a non-nil scope for an API token")
			}
			if scope.TokenID != testCase.wantTokenID {
				t.Errorf("TokenID = %d, want %d", scope.TokenID, testCase.wantTokenID)
			}
			if scope.ProjectID != testCase.wantProjectID {
				t.Errorf("ProjectID = %d, want %d", scope.ProjectID, testCase.wantProjectID)
			}
			if got := scope.AllowsWrite(); got != testCase.wantWrite {
				t.Errorf("AllowsWrite() = %t, want %t", got, testCase.wantWrite)
			}
			if !scope.AllowsProject(testCase.wantProjectID) {
				t.Errorf("AllowsProject(%d) = false, want true", testCase.wantProjectID)
			}
			for _, projectID := range testCase.otherProjectIDs {
				if scope.AllowsProject(projectID) {
					t.Errorf("AllowsProject(%d) = true, want false", projectID)
				}
			}
			for _, slug := range testCase.allowedEnvs {
				if !scope.AllowsEnvironment(slug) {
					t.Errorf("AllowsEnvironment(%q) = false, want true", slug)
				}
			}
			for _, slug := range testCase.deniedEnvs {
				if scope.AllowsEnvironment(slug) {
					t.Errorf("AllowsEnvironment(%q) = true, want false", slug)
				}
			}
		})
	}
}

func TestSessionDataHasNoScope(t *testing.T) {
	data := &Data{Email: "user@example.com"}
	if data.GetScope() != nil {
		t.Error("session identity must carry a nil scope")
	}
	var missing *Data
	if missing.GetScope() != nil {
		t.Error("nil Data must yield a nil scope")
	}
}
