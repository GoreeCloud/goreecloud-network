package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoreeCloud/goreecloud-network/internal/controlplane"
)

func TestStatusEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	rec := httptest.NewRecorder()

	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Product != "GoreeCloud Network" {
		t.Fatalf("unexpected product: %q", payload.Product)
	}
	if payload.Version != Version || payload.Lifecycle != Lifecycle {
		t.Fatalf("unexpected version/lifecycle: %+v", payload)
	}
	if payload.Surfaces.Server != "development_control_plane" || payload.Surfaces.Android != "development_client" || payload.Surfaces.GoogleTV != "development_client" || payload.Surfaces.IOS != "development_client" {
		t.Fatalf("surface states must remain truthful: %+v", payload.Surfaces)
	}
}

func TestOverviewStartsEmptyAndTruthful(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)

	var payload controlplane.Overview
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.DeviceCount != 0 || payload.ResourceCount != 0 || payload.PolicyCount != 0 {
		t.Fatalf("default server must not invent inventory: %+v", payload)
	}
	if payload.PolicyMode != "deny_by_default" || payload.Persistence != "volatile_development_memory" || payload.Authentication != "not_implemented" {
		t.Fatalf("unexpected authority state: %+v", payload)
	}
}

func TestDeviceInventoryStartsEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices", nil)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)

	var payload []controlplane.Device
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload) != 0 {
		t.Fatalf("default inventory must be empty: %+v", payload)
	}
}

func TestAccessEvaluationDeniesByDefault(t *testing.T) {
	body := bytes.NewBufferString(`{"principalId":"principal-1","deviceId":"device-1","resourceId":"resource-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/access/evaluate", body)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)

	var payload AccessEvaluationResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Decision != "deny" || payload.ReasonCode != "NO_MATCHING_ALLOW_POLICY" || payload.Enforcement != "decision_only" {
		t.Fatalf("expected decision-only deny by default, got %+v", payload)
	}
}

func TestAccessEvaluationUsesExplicitAllowPolicy(t *testing.T) {
	state := controlplane.NewState()
	if err := state.PutAllowPolicy(controlplane.AllowPolicy{ID: "policy-1", PrincipalID: "principal-1", DeviceID: "device-1", ResourceID: "resource-1"}); err != nil {
		t.Fatal(err)
	}
	body := bytes.NewBufferString(`{"principalId":"principal-1","deviceId":"device-1","resourceId":"resource-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/access/evaluate", body)
	rec := httptest.NewRecorder()
	NewRouterWithState("testdata/missing-web-root", state).ServeHTTP(rec, req)

	var payload AccessEvaluationResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Decision != "allow" || payload.PolicyID != "policy-1" {
		t.Fatalf("expected explicit allow policy, got %+v", payload)
	}
}

func TestCapabilitiesDoNotClaimTunnelImplementation(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities", nil)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)

	var payload []Capability
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	for _, capability := range payload {
		if capability.ID == "tunnel.wireguard" && capability.State != "not_implemented" {
			t.Fatalf("tunnel must not be represented as implemented: %+v", capability)
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("missing nosniff header: %q", got)
	}
	if got := rec.Header().Get("Content-Security-Policy"); got == "" {
		t.Fatal("missing content security policy")
	}
}
