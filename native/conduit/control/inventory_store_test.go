package control

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCapabilityInventoryStoreRoundTripsValidatedSnapshot(t *testing.T) {
	requireProtectedFileStoreTestSupport(t)
	path := filepath.Join(t.TempDir(), "state", "inventory.json")
	store, err := NewCapabilityInventoryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	inventory := CapabilityInventory{
		Schema: CapabilityInventorySchemaV1,
		Capabilities: []CapabilityState{
			{ID: "control-api", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
		},
	}
	snapshot, err := BuildCapabilityInventorySnapshot(inventory, time.Date(2026, 8, 27, 8, 45, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(snapshot); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Fingerprint != snapshot.Fingerprint || loaded.Inventory.Capabilities[0] != inventory.Capabilities[0] {
		t.Fatalf("unexpected loaded snapshot: %+v", loaded)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("snapshot mode=%#o", info.Mode().Perm())
	}
}

func TestAdvanceStoredCapabilityToIsolatedValidationRequiresCurrentFingerprint(t *testing.T) {
	requireProtectedFileStoreTestSupport(t)
	path := filepath.Join(t.TempDir(), "inventory.json")
	store, err := NewCapabilityInventoryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	inventory := CapabilityInventory{
		Schema: CapabilityInventorySchemaV1,
		Capabilities: []CapabilityState{
			{ID: "control-api", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
			{ID: "relay", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
		},
	}
	initial, err := BuildCapabilityInventorySnapshot(inventory, time.Date(2026, 8, 27, 8, 45, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(initial); err != nil {
		t.Fatal(err)
	}
	if _, err := AdvanceStoredCapabilityToIsolatedValidation(store, "stale", "control-api", completeTransitionEvidence("control-api"), time.Date(2026, 8, 27, 8, 46, 0, 0, time.UTC)); err == nil {
		t.Fatal("stale snapshot fingerprint unexpectedly advanced stored capability")
	}

	next, err := AdvanceStoredCapabilityToIsolatedValidation(store, initial.Fingerprint, "control-api", completeTransitionEvidence("control-api"), time.Date(2026, 8, 27, 8, 46, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if next.Fingerprint == initial.Fingerprint {
		t.Fatal("stored transition did not change snapshot fingerprint")
	}
	if next.Inventory.Capabilities[0].MigrationStage != MigrationStageIsolatedValidation {
		t.Fatalf("target stage=%q", next.Inventory.Capabilities[0].MigrationStage)
	}
	if next.Inventory.Capabilities[0].Authority != AuthorityInherited || !next.Inventory.Capabilities[0].CompatibilityBridgeActive || next.Inventory.Capabilities[0].ProductionCutoverAuthorized {
		t.Fatal("stored transition changed authority, bridge, or cutover authorization")
	}
	if next.Inventory.Capabilities[1] != inventory.Capabilities[1] {
		t.Fatal("stored transition changed non-target capability")
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Fingerprint != next.Fingerprint {
		t.Fatal("stored transition was not durably persisted")
	}
}

func TestCapabilityInventoryStoreRejectsCutoverAuthorization(t *testing.T) {
	store, err := NewCapabilityInventoryStore(filepath.Join(t.TempDir(), "inventory.json"))
	if err != nil {
		t.Fatal(err)
	}
	inventory := CapabilityInventory{
		Schema: CapabilityInventorySchemaV1,
		Capabilities: []CapabilityState{
			{ID: "control-api", MigrationStage: "isolated-validation", Authority: AuthorityInherited, CompatibilityBridgeActive: true, ProductionCutoverAuthorized: true},
		},
	}
	if _, err := BuildCapabilityInventorySnapshot(inventory, time.Now().UTC()); err == nil {
		t.Fatal("production cutover authorization unexpectedly entered a source inventory snapshot")
	}
	if _, err := store.Load(); err == nil {
		t.Fatal("empty inventory store unexpectedly loaded")
	}
}
