package control

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type providerFunc func(context.Context) (Status, error)

func (f providerFunc) Status(ctx context.Context) (Status, error) { return f(ctx) }

func TestCompatibilityBridgeForcesInheritedAuthority(t *testing.T) {
	bridge := CompatibilityBridge{Inherited: providerFunc(func(context.Context) (Status, error) {
		return Status{
			GeneratedAt:                 time.Unix(1, 0).UTC(),
			Authority:                   AuthorityNative,
			MigrationStage:              "implementation",
			CompatibilityBridgeActive:   false,
			ProductionCutoverAuthorized: true,
		}, nil
	})}

	status, err := bridge.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Authority != AuthorityInherited {
		t.Fatalf("authority = %q, want %q", status.Authority, AuthorityInherited)
	}
	if !status.CompatibilityBridgeActive {
		t.Fatal("compatibility bridge must remain active")
	}
	if status.ProductionCutoverAuthorized {
		t.Fatal("source contract must not authorize production cutover")
	}
}

func TestStatusHandlerReadOnlyAndFailClosed(t *testing.T) {
	provider := providerFunc(func(context.Context) (Status, error) {
		return Status{
			Schema:                      SchemaV1,
			GeneratedAt:                 time.Unix(1, 0).UTC(),
			Authority:                   AuthorityInherited,
			MigrationStage:              "implementation",
			CompatibilityBridgeActive:   true,
			ProductionCutoverAuthorized: false,
		}, nil
	})
	handler := StatusHandler{Provider: provider}

	get := httptest.NewRequest(http.MethodGet, "/v1/conduit/status", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, get)
	if response.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", response.Code, http.StatusOK)
	}
	var status Status
	if err := json.NewDecoder(response.Body).Decode(&status); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if status.Schema != SchemaV1 {
		t.Fatalf("schema = %q, want %q", status.Schema, SchemaV1)
	}
	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}

	post := httptest.NewRequest(http.MethodPost, "/v1/conduit/status", nil)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, post)
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

func TestValidateStatusRejectsSourceCutoverAuthorization(t *testing.T) {
	err := ValidateStatus(Status{
		Schema:                      SchemaV1,
		GeneratedAt:                 time.Unix(1, 0).UTC(),
		Authority:                   AuthorityInherited,
		MigrationStage:              "implementation",
		CompatibilityBridgeActive:   true,
		ProductionCutoverAuthorized: true,
	})
	if err == nil {
		t.Fatal("expected production cutover authorization to be rejected")
	}
}
