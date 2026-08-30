package control

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCapabilityInventoryRecoveryStoreRoundTrip(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("owner-only file-mode contract is intentionally unsupported on Windows")
	}
	now := time.Date(2026, 8, 30, 3, 30, 0, 0, time.UTC)
	snapshot := testRecoveryInventorySnapshot(t, "implementation", now)
	point, err := BuildCapabilityInventoryRecoveryPoint(snapshot, strings.Repeat("a", 40), now)
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewCapabilityInventoryRecoveryStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path, err := store.Save(point)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("recovery point mode = %o, want 600", info.Mode().Perm())
	}
	loaded, err := store.Load(snapshot.Fingerprint)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SnapshotFingerprint != snapshot.Fingerprint || loaded.SourceRevision != strings.Repeat("a", 40) {
		t.Fatalf("unexpected loaded recovery point: %+v", loaded)
	}
	if loaded.ProductionCutoverAuthorized {
		t.Fatal("recovery point unexpectedly authorized production cutover")
	}
	if _, err := store.Save(point); err == nil {
		t.Fatal("immutable recovery point was unexpectedly overwritten")
	}
}

func TestRestoreCapabilityInventoryRecoveryPointCompareAndSwap(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("owner-only file-mode contract is intentionally unsupported on Windows")
	}
	now := time.Date(2026, 8, 30, 3, 30, 0, 0, time.UTC)
	original := testRecoveryInventorySnapshot(t, "implementation", now)
	point, err := BuildCapabilityInventoryRecoveryPoint(original, strings.Repeat("b", 40), now)
	if err != nil {
		t.Fatal(err)
	}
	current := testRecoveryInventorySnapshot(t, "isolated-validation", now.Add(time.Minute))
	activeStore, err := NewCapabilityInventoryStore(filepath.Join(t.TempDir(), "inventory.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := activeStore.Save(current); err != nil {
		t.Fatal(err)
	}

	if _, err := RestoreCapabilityInventoryRecoveryPoint(activeStore, point, strings.Repeat("0", 64), now.Add(2*time.Minute)); err == nil {
		t.Fatal("stale expected fingerprint unexpectedly restored recovery point")
	}
	restored, err := RestoreCapabilityInventoryRecoveryPoint(activeStore, point, current.Fingerprint, now.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if restored.Fingerprint != original.Fingerprint {
		t.Fatalf("restored fingerprint = %s, want %s", restored.Fingerprint, original.Fingerprint)
	}
	loaded, err := activeStore.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Fingerprint != original.Fingerprint || loaded.Inventory.Capabilities[0].MigrationStage != "implementation" {
		t.Fatalf("active store did not restore recovery inventory: %+v", loaded)
	}
}

func TestCapabilityInventoryRecoveryPointRejectsCutoverAndTampering(t *testing.T) {
	now := time.Date(2026, 8, 30, 3, 30, 0, 0, time.UTC)
	snapshot := testRecoveryInventorySnapshot(t, "implementation", now)
	point, err := BuildCapabilityInventoryRecoveryPoint(snapshot, strings.Repeat("c", 40), now)
	if err != nil {
		t.Fatal(err)
	}
	point.ProductionCutoverAuthorized = true
	if err := validateCapabilityInventoryRecoveryPoint(point); err == nil {
		t.Fatal("cutover-authorizing recovery point unexpectedly validated")
	}

	point.ProductionCutoverAuthorized = false
	point.SnapshotFingerprint = strings.Repeat("d", 64)
	if err := validateCapabilityInventoryRecoveryPoint(point); err == nil {
		t.Fatal("tampered recovery fingerprint unexpectedly validated")
	}
}

func testRecoveryInventorySnapshot(t *testing.T, stage string, now time.Time) CapabilityInventorySnapshot {
	t.Helper()
	snapshot, err := BuildCapabilityInventorySnapshot(CapabilityInventory{
		Schema: CapabilityInventorySchemaV1,
		Capabilities: []CapabilityState{
			{
				ID:                          "relay-transport",
				MigrationStage:              stage,
				Authority:                   AuthorityInherited,
				CompatibilityBridgeActive:   true,
				ProductionCutoverAuthorized: false,
			},
		},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}
