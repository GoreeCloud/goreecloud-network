package identity

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthenticatorDisabledWithoutRuntimeConfiguration(t *testing.T) {
	auth, err := NewAuthenticator(Config{RequiredScope: "goreecloud.network.admin"})
	if err != nil {
		t.Fatal(err)
	}
	status := auth.Status()
	if status.State != "identity_boundary_runtime_unconfigured" || status.RuntimeConfigured || status.ProductionAccepted {
		t.Fatalf("unexpected disabled status: %+v", status)
	}
	if _, err := auth.AuthenticateAdmin(context.Background(), "token"); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected fail-closed not configured result, got %v", err)
	}
}

func TestAuthenticatorRejectsPartialConfiguration(t *testing.T) {
	_, err := NewAuthenticator(Config{IntrospectionURL: "https://identity.example/introspect", ClientID: "network", RequiredScope: "goreecloud.network.admin"})
	if err == nil {
		t.Fatal("expected partial configuration error")
	}
}

func TestAuthenticatorRejectsNonTLSRemoteEndpoint(t *testing.T) {
	_, err := NewAuthenticator(Config{IntrospectionURL: "http://identity.example/introspect", ClientID: "network", ClientSecret: "secret", RequiredScope: "goreecloud.network.admin"})
	if err == nil {
		t.Fatal("expected non-TLS remote endpoint rejection")
	}
}

func TestAuthenticateAdminAcceptsActiveScopedIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		clientID, secret, ok := r.BasicAuth()
		if !ok || clientID != "network-client" || secret != "test-secret" {
			t.Fatalf("unexpected client authentication")
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("token") != "valid-token" || r.Form.Get("token_type_hint") != "access_token" {
			t.Fatalf("unexpected form: %v", r.Form)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"active":             true,
			"sub":                "identity-user-1",
			"preferred_username": "admin",
			"scope":              "openid goreecloud.network.admin profile",
			"iss":                "https://identity.goreecloud.com/application/o/network/",
			"aud":                []string{"goreecloud-network"},
		})
	}))
	defer server.Close()

	auth, err := NewAuthenticator(Config{IntrospectionURL: server.URL, ClientID: "network-client", ClientSecret: "test-secret", ExpectedIssuer: "https://identity.goreecloud.com/application/o/network/", ExpectedAudience: "goreecloud-network", RequiredScope: "goreecloud.network.admin"})
	if err != nil {
		t.Fatal(err)
	}
	principal, err := auth.AuthenticateAdmin(context.Background(), "valid-token")
	if err != nil {
		t.Fatal(err)
	}
	if principal.Subject != "identity-user-1" || principal.Username != "admin" || !contains(principal.Scopes, "goreecloud.network.admin") {
		t.Fatalf("unexpected principal: %+v", principal)
	}
}

func TestAuthenticateAdminRejectsInactiveAndInsufficientTokens(t *testing.T) {
	responses := map[string]map[string]any{
		"inactive":  {"active": false},
		"user-only": {"active": true, "sub": "identity-user-1", "scope": "openid profile"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_ = json.NewEncoder(w).Encode(responses[r.Form.Get("token")])
	}))
	defer server.Close()
	auth, err := NewAuthenticator(Config{IntrospectionURL: server.URL, ClientID: "network", ClientSecret: "secret", RequiredScope: "goreecloud.network.admin"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := auth.AuthenticateAdmin(context.Background(), "inactive"); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
	if _, err := auth.AuthenticateAdmin(context.Background(), "user-only"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestAuthenticationStatusDoesNotExposeClientSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	auth, err := NewAuthenticator(Config{IntrospectionURL: server.URL, ClientID: "network", ClientSecret: "top-secret", RequiredScope: "goreecloud.network.admin"})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(auth.Status())
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) == "" || strings.Contains(string(encoded), "top-secret") {
		t.Fatalf("status leaked secret: %s", encoded)
	}
}
