package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoreeCloud/goreecloud-network/internal/controlplane"
	"github.com/GoreeCloud/goreecloud-network/internal/identity"
)

type testAuthenticator struct {
	status    identity.Status
	principal identity.Principal
	err       error
}

func (a testAuthenticator) Status() identity.Status { return a.status }
func (a testAuthenticator) AuthenticateAdmin(context.Context, string) (identity.Principal, error) {
	return a.principal, a.err
}

func TestStatusEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var payload StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Product != "GoreeCloud Network" || payload.Version != Version || payload.Lifecycle != Lifecycle {
		t.Fatalf("unexpected status: %+v", payload)
	}
	if payload.Surfaces.Server != "development_identity_aware_control_plane" {
		t.Fatalf("unexpected server surface: %+v", payload.Surfaces)
	}
}

func TestOverviewReportsFailClosedIdentityBoundary(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)
	var payload controlplane.Overview
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Authentication != "identity_boundary_runtime_unconfigured" || payload.PolicyMode != "deny_by_default" {
		t.Fatalf("unexpected authority state: %+v", payload)
	}
}

func TestAuthenticationStatusDoesNotClaimRuntimeIntegration(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/authentication", nil)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)
	var payload identity.Status
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Authority != "goreecloud_identity" || payload.RuntimeConfigured || payload.ProductionAccepted {
		t.Fatalf("unexpected auth status: %+v", payload)
	}
}

func TestAdminSessionRequiresBearerToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/session", nil)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized || rec.Header().Get("WWW-Authenticate") == "" {
		t.Fatalf("expected bearer challenge, got %d %q", rec.Code, rec.Header().Get("WWW-Authenticate"))
	}
}

func TestAdminSessionFailsClosedWhenIdentityRuntimeUnconfigured(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/session", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminSessionReturnsMinimalAuthorizedPrincipal(t *testing.T) {
	auth := testAuthenticator{status: identity.Status{Authority: "goreecloud_identity", Protocol: "oauth2_token_introspection", State: "identity_introspection_configured_runtime_unaccepted", RuntimeConfigured: true, RequiredScope: DefaultAdminScope}, principal: identity.Principal{Subject: "principal-1", Username: "admin", Scopes: []string{DefaultAdminScope}}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/session", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	NewRouterWithRuntimeAndIdentity("testdata/missing-web-root", controlplane.NewState(), controlplane.VolatileStorageStatus(), auth).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var payload AdminSessionResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if !payload.Authenticated || !payload.Authorized || payload.Subject != "principal-1" || payload.Authority != "goreecloud_identity" {
		t.Fatalf("unexpected session: %+v", payload)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("admin session must be no-store")
	}
}

func TestAdminSessionRejectsInsufficientAuthority(t *testing.T) {
	auth := testAuthenticator{status: identity.Status{Authority: "goreecloud_identity", State: "identity_introspection_configured_runtime_unaccepted", RuntimeConfigured: true, RequiredScope: DefaultAdminScope}, err: identity.ErrForbidden}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/session", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	NewRouterWithRuntimeAndIdentity("testdata/missing-web-root", controlplane.NewState(), controlplane.VolatileStorageStatus(), auth).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAdminSessionRejectsInvalidToken(t *testing.T) {
	auth := testAuthenticator{status: identity.Status{Authority: "goreecloud_identity", State: "identity_introspection_configured_runtime_unaccepted", RuntimeConfigured: true, RequiredScope: DefaultAdminScope}, err: identity.ErrUnauthenticated}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/session", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	NewRouterWithRuntimeAndIdentity("testdata/missing-web-root", controlplane.NewState(), controlplane.VolatileStorageStatus(), auth).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAccessEvaluationStillDeniesByDefault(t *testing.T) {
	body := bytes.NewBufferString(`{"principalId":"principal-1","deviceId":"device-1","resourceId":"resource-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/access/evaluate", body)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)
	var payload AccessEvaluationResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Decision != "deny" || payload.ReasonCode != "NO_MATCHING_ALLOW_POLICY" || payload.Enforcement != "decision_only" {
		t.Fatalf("unexpected decision: %+v", payload)
	}
}

func TestCapabilitiesKeepMutationAndTunnelUnimplemented(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities", nil)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)
	var payload []Capability
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	states := map[string]string{}
	for _, capability := range payload {
		states[capability.ID] = capability.State
	}
	if states["identity.admin_authentication"] != "implemented_boundary_runtime_unconfigured" {
		t.Fatalf("unexpected auth capability: %q", states["identity.admin_authentication"])
	}
	if states["access.policy.management"] != "not_implemented" || states["tunnel.wireguard"] != "not_implemented" {
		t.Fatalf("mutation/tunnel truth regressed: %+v", states)
	}
}

func TestSecurityHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	NewRouter("testdata/missing-web-root").ServeHTTP(rec, req)
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" || rec.Header().Get("Content-Security-Policy") == "" {
		t.Fatal("missing security headers")
	}
}

func TestBearerTokenParser(t *testing.T) {
	if token, ok := bearerToken("Bearer abc"); !ok || token != "abc" {
		t.Fatalf("valid bearer token rejected")
	}
	for _, value := range []string{"", "Basic abc", "Bearer", "Bearer a b"} {
		if _, ok := bearerToken(value); ok {
			t.Fatalf("invalid authorization accepted: %q", value)
		}
	}
}

func TestUnavailableIdentityMapsToServiceUnavailable(t *testing.T) {
	auth := testAuthenticator{status: identity.Status{RequiredScope: DefaultAdminScope, RuntimeConfigured: true}, err: fmtError(identity.ErrUnavailable)}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/session", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	NewRouterWithRuntimeAndIdentity("testdata/missing-web-root", controlplane.NewState(), controlplane.VolatileStorageStatus(), auth).ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func fmtError(err error) error { return errors.Join(errors.New("wrapped"), err) }
