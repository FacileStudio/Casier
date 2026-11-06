package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/tronc/apiref"
)

// A hand-written registry rots the moment a route is added without it, which is
// how an API ends up with documentation that lies. This turns that drift into a
// failing build. If it fails, fix the registry in the module's
// documentation.go - never the test.
//
// It needs no database: buildTestRouter mounts every route against a nil handle.
//
// /spaces is ignored on purpose. It is not an API route: it is the 301 left by
// the 2026-08-04 rename to /projects, serving bookmarks and nothing else, and
// documenting a redirect as an endpoint would invite a client to call it.
func TestEveryRouteIsDocumented(t *testing.T) {
	if missing := apiref.Undocumented(testRouter, refConfig(), "/spaces"); len(missing) > 0 {
		t.Errorf("routes missing from the API registry:\n  %s", strings.Join(missing, "\n  "))
	}
}

func TestRegistryIsComplete(t *testing.T) {
	if issues := apiref.Incomplete(
		refConfig(),
		"/auth/logout",
		"/registre/deletions",
		"/projects/{slug}",
		"/projects/{slug}/members",
		"/projects/{slug}/invites/{inviteId}",
		"/projects/{slug}/tokens/{tokenId}",
		"/projects/{slug}/webhooks/{webhookId}",
		"/projects/{slug}/environments/{envSlug}",
		"/projects/{slug}/environments/{envSlug}/secrets/{key}",
		"/projects/{slug}/environments/{envSlug}/secrets/{key}/restore",
		"/projects/{slug}/environments/{envSlug}/secrets/{key}/tags",
		"/projects/{slug}/environments/{envSlug}/secrets/{key}/acls",
		"/projects/{slug}/environments/{envSlug}/secrets/{key}/acls/{aclId}",
		"/invites/{token}/accept",
	); len(issues) > 0 {
		t.Errorf("incomplete documentation routes: %v", issues)
	}
}

func TestReferenceIsServed(t *testing.T) {
	page := httptest.NewRecorder()
	testRouter.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if page.Code != http.StatusOK {
		t.Fatalf("GET /docs = %d, want 200", page.Code)
	}
	if !strings.Contains(page.Body.String(), `data-url="/docs/openapi.json"`) {
		t.Errorf("the reference page does not point at the spec: %s", page.Body.String())
	}

	spec := httptest.NewRecorder()
	testRouter.ServeHTTP(spec, httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil))
	if spec.Code != http.StatusOK {
		t.Fatalf("GET /docs/openapi.json = %d, want 200", spec.Code)
	}
	var document struct {
		OpenAPI string `json:"openapi"`
		Info    struct {
			Title string `json:"title"`
		} `json:"info"`
		Paths map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(spec.Body.Bytes(), &document); err != nil {
		t.Fatalf("the spec is not JSON: %v", err)
	}
	if !strings.HasPrefix(document.OpenAPI, "3.1") {
		t.Errorf("openapi = %q, want 3.1.x", document.OpenAPI)
	}
	if document.Info.Title != refConfig().Title {
		t.Errorf("info.title = %q, want %q", document.Info.Title, refConfig().Title)
	}
	if len(document.Paths) == 0 {
		t.Error("the spec documents no paths")
	}
}
