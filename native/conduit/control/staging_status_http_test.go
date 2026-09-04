package control

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type capabilityStagingStatusProviderFunc func(context.Context) (CapabilityStagingStatus, error)

func (f capabilityStagingStatusProviderFunc) CapabilityStagingStatus(ctx context.Context) (CapabilityStagingStatus, error) {
	return f(ctx)
}

func TestCapabilityStagingStatusHandlerReturnsMinimizedEvidence(t *testing.T) {
	status := validCapabilityStagingStatusForHTTPTest()
	handler := CapabilityStagingStatusHandler{
		Provider: capabilityStagingStatusProviderFunc(func(context.Context) (CapabilityStagingStatus, error) {
			return status, nil
		}),
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/conduit/staging-status", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
	var decoded CapabilityStagingStatus
	if err := json.Unmarshal(recorder.Body.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != status {
		t.Fatalf("decoded status = %+v, want %+v", decoded, status)
	}
	for _, forbidden := range []string{
		"capability_id",
		"source_revision",
		"runtime_artifact_sha256",
		"peer",
		"route",
		"policy",
		"credential",
		"packet",
		"dns_query",
	} {
		if strings.Contains(recorder.Body.String(), forbidden) {
			t.Fatalf("response unexpectedly exposes %q: %s", forbidden, recorder.Body.String())
		}
	}
}

func TestCapabilityStagingStatusHandlerFailsClosed(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		provider CapabilityStagingStatusProvider
		wantCode int
	}{
		{
			name:     "method",
			method:   http.MethodPost,
			provider: capabilityStagingStatusProviderFunc(func(context.Context) (CapabilityStagingStatus, error) { return validCapabilityStagingStatusForHTTPTest(), nil }),
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "missing provider",
			method:   http.MethodGet,
			provider: nil,
			wantCode: http.StatusServiceUnavailable,
		},
		{
			name:   "provider error",
			method: http.MethodGet,
			provider: capabilityStagingStatusProviderFunc(func(context.Context) (CapabilityStagingStatus, error) {
				return CapabilityStagingStatus{}, errors.New("unavailable")
			}),
			wantCode: http.StatusServiceUnavailable,
		},
		{
			name:   "unsafe status",
			method: http.MethodGet,
			provider: capabilityStagingStatusProviderFunc(func(context.Context) (CapabilityStagingStatus, error) {
				status := validCapabilityStagingStatusForHTTPTest()
				status.ProductionCutoverAuthorized = true
				return status, nil
			}),
			wantCode: http.StatusServiceUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler := CapabilityStagingStatusHandler{Provider: tc.provider}
			handler.ServeHTTP(recorder, httptest.NewRequest(tc.method, "/conduit/staging-status", nil))
			if recorder.Code != tc.wantCode {
				t.Fatalf("status code = %d, want %d: %s", recorder.Code, tc.wantCode, recorder.Body.String())
			}
		})
	}
}

func TestValidateCapabilityStagingStatusRejectsIncompleteEvidence(t *testing.T) {
	status := validCapabilityStagingStatusForHTTPTest()
	status.TransitionReceiptPersisted = false
	if err := ValidateCapabilityStagingStatus(status); err == nil {
		t.Fatal("incomplete durable evidence unexpectedly accepted")
	}
}

func validCapabilityStagingStatusForHTTPTest() CapabilityStagingStatus {
	return CapabilityStagingStatus{
		Schema:                           CapabilityStagingStatusSchemaV1,
		GeneratedAt:                      time.Date(2026, 8, 29, 4, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
		InventoryFingerprint:             strings.Repeat("a", 64),
		TransitionReceiptPersisted:       true,
		TransitionReconciliationRequired: false,
		StagingEvidencePersisted:         true,
		AcceptanceSchema:                 IsolatedAcceptanceSchemaV1,
		Authority:                        AuthorityInherited,
		CompatibilityBridgeActive:        true,
		ProductionCutoverAuthorized:      false,
	}
}
