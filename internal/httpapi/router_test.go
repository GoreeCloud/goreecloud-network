package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("expected 200, got %d", rec.Code) }
	var payload StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil { t.Fatalf("decode response: %v", err) }
	if payload.Product != "GoreeCloud Network" { t.Fatalf("unexpected product: %q", payload.Product) }
	if payload.Version != Version || payload.Lifecycle != Lifecycle { t.Fatalf("unexpected version/lifecycle: %+v", payload) }
	if payload.Surfaces.Android != "source_bootstrap" || payload.Surfaces.GoogleTV != "source_bootstrap" || payload.Surfaces.IOS != "source_bootstrap" { t.Fatalf("client source states must remain truthful: %+v", payload.Surfaces) }
}

func TestCapabilitiesDoNotClaimTunnelImplementation(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities", nil)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)
	var payload []Capability
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil { t.Fatalf("decode response: %v", err) }
	for _, capability := range payload {
		if capability.ID == "tunnel.wireguard" && capability.State != "not_implemented" { t.Fatalf("tunnel must not be represented as implemented: %+v", capability) }
	}
}

func TestSecurityHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" { t.Fatalf("missing nosniff header: %q", got) }
	if got := rec.Header().Get("Content-Security-Policy"); got == "" { t.Fatal("missing content security policy") }
}
