package control

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInventoryStatusHandlerExposesAggregatePersistedStateOnly(t *testing.T) {
	requireProtectedFileStoreTestSupport(t)
	store, err := NewCapabilityInventoryStore(filepath.Join(t.TempDir(), "inventory.json"))
	if err != nil {
		t.Fatal(err)
	}
	inventory := CapabilityInventory{
		Schema: CapabilityInventorySchemaV1,
		Capabilities: []CapabilityState{
			{ID: "control-api", MigrationStage: MigrationStageIsolatedValidation, Authority: AuthorityInherited, CompatibilityBridgeActive: true},
			{ID: "relay", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
		},
	}
	snapshot, err := BuildCapabilityInventorySnapshot(inventory, time.Date(2026, 8, 27, 9, 15, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(snapshot); err != nil {
		t.Fatal(err)
	}
	handler := InventoryStatusHandler{Store: store, Now: func() time.Time { return time.Date(2026, 8, 27, 9, 16, 0, 0, time.UTC) }}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/conduit/capabilities/status", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "control-api") || strings.Contains(recorder.Body.String(), "relay") {
		t.Fatal("aggregate capability inventory status leaked individual capability identifiers")
	}
	var status CapabilityInventoryStatus
	if err := json.Unmarshal(recorder.Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.Schema != CapabilityInventoryStatusSchemaV1 || status.SnapshotFingerprint != snapshot.Fingerprint {
		t.Fatalf("unexpected aggregate status: %+v", status)
	}
	if status.Summary.Total != 2 || status.Summary.Inherited != 2 || status.Summary.Bridged != 2 {
		t.Fatalf("unexpected aggregate summary: %+v", status.Summary)
	}
	if status.ProductionCutoverAuthorized {
		t.Fatal("aggregate source status unexpectedly authorized production cutover")
	}
	if recorder.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("aggregate migration status must not be cached")
	}
}

func TestInventoryStatusHandlerIsReadOnlyAndFailsClosed(t *testing.T) {
	handler := InventoryStatusHandler{}
	post := httptest.NewRecorder()
	handler.ServeHTTP(post, httptest.NewRequest(http.MethodPost, "/conduit/capabilities/status", nil))
	if post.Code != http.StatusMethodNotAllowed || post.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("unexpected POST response: status=%d allow=%q", post.Code, post.Header().Get("Allow"))
	}
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/conduit/capabilities/status", nil))
	if get.Code != http.StatusServiceUnavailable {
		t.Fatalf("missing store status=%d", get.Code)
	}
}
