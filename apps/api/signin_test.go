package main

import (
	"net/http"
	"testing"
)

// TestPasswordSignInIsGone pins the launch rule: Registre is the only way for
// a person in, so no register, login or reset route answers at all.
func TestPasswordSignInIsGone(t *testing.T) {
	routes := []struct{ method, path string }{
		{http.MethodPost, "/api/auth/register"},
		{http.MethodPost, "/api/auth/login"},
		{http.MethodPost, "/api/auth/forgot-password"},
		{http.MethodPost, "/api/auth/reset-password"},
	}
	for i, route := range routes {
		status, _ := call(route.method, route.path, `{"email":"a@b.test","password":"a-long-enough-password"}`, "", i+1)
		if status != http.StatusNotFound && status != http.StatusMethodNotAllowed {
			t.Errorf("%s %s answered %d, want it unrouted", route.method, route.path, status)
		}
	}
}

// TestAPITokensStillAuthenticate guards the machine path that removing the
// password routes must never touch: a casier_ token reaches the API on its own,
// with no session and no password anywhere.
func TestAPITokensStillAuthenticate(t *testing.T) {
	requireDatabase(t)
	if _, err := seed(testDB, testKeyring); err != nil {
		t.Fatalf("seed: %v", err)
	}
	for i, caller := range liveTokens {
		status, _ := call(http.MethodGet, "/api/auth/me", "", bearers[caller], 100+i)
		if status != http.StatusOK {
			t.Errorf("%s: GET /api/auth/me answered %d, want 200", caller, status)
		}
	}
	for i, caller := range []string{tExpired, tRevoked, badToken} {
		status, _ := call(http.MethodGet, "/api/auth/me", "", bearers[caller], 200+i)
		if status != http.StatusUnauthorized {
			t.Errorf("%s: GET /api/auth/me answered %d, want 401", caller, status)
		}
	}
}
