package control

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testTransitionSnapshots(t *testing.T) (CapabilityInventorySnapshot, CapabilityInventorySnapshot) {
	t.Helper()
	beforeInventory := CapabilityInventory{
		Schema: CapabilityInventorySchemaV1,
		Capabilities: []CapabilityState{
			{ID: "control-api", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
			{ID: "relay", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
		},
	}
	before, err := BuildCapabilityInventorySnapshot(beforeInventory, time.Date(2026, 8, 27, 10, 15, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	afterInventory, err := AdvanceInventoryCapabilityToIsolatedValidation(beforeInventory, "control-api", completeTransitionEvidence("control-api"))
	if err != nil {
		t.Fatal(err)
	}
	after, err := BuildCapabilityInventorySnapshot(afterInventory, time.Date(2026, 8, 27, 10, 16, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return before, after
}

func TestBuildCapabilityTransitionReceiptBindsExactSafeTransition(t *testing.T) {
	before, after := testTransitionSnapshots(t)
	receipt, err := BuildCapabilityTransitionReceipt(before, after, "control-api", IsolatedAcceptanceSchemaV1, time.Date(2026, 8, 27, 10, 17, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Schema != CapabilityTransitionReceiptSchemaV1 || receipt.FromSnapshotFingerprint != before.Fingerprint || receipt.ToSnapshotFingerprint != after.Fingerprint {
		t.Fatalf("unexpected transition receipt: %+v", receipt)
	}
	if receipt.Authority != AuthorityInherited || !receipt.CompatibilityBridgeActive || receipt.ProductionCutoverAuthorized {
		t.Fatal("transition receipt violated authority, bridge, or cutover safety invariants")
	}
}

func TestBuildCapabilityTransitionReceiptRejectsNonTargetChange(t *testing.T) {
	before, after := testTransitionSnapshots(t)
	after.Inventory.Capabilities[1].MigrationStage = MigrationStageIsolatedValidation
	fingerprint, err := capabilityInventoryFingerprint(after.Inventory)
	if err != nil {
		t.Fatal(err)
	}
	after.Fingerprint = fingerprint
	if _, err := BuildCapabilityTransitionReceipt(before, after, "control-api", IsolatedAcceptanceSchemaV1, time.Now().UTC()); err == nil {
		t.Fatal("transition receipt unexpectedly accepted a non-target capability change")
	}
}

func TestCapabilityTransitionReceiptStoreIsImmutableAndOwnerOnly(t *testing.T) {
	requireProtectedFileStoreTestSupport(t)
	before, after := testTransitionSnapshots(t)
	receipt, err := BuildCapabilityTransitionReceipt(before, after, "control-api", IsolatedAcceptanceSchemaV1, time.Date(2026, 8, 27, 10, 17, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewCapabilityTransitionReceiptStore(filepath.Join(t.TempDir(), "receipts"))
	if err != nil {
		t.Fatal(err)
	}
	path, err := store.Save(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(receipt); err == nil {
		t.Fatal("immutable transition receipt was unexpectedly replaceable")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("transition receipt mode=%#o", info.Mode().Perm())
	}
	loaded, err := store.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded != receipt {
		t.Fatalf("loaded transition receipt mismatch: %+v", loaded)
	}
}

func TestCapabilityTransitionReceiptStoreRejectsOutsidePath(t *testing.T) {
	store, err := NewCapabilityTransitionReceiptStore(filepath.Join(t.TempDir(), "receipts"))
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "receipt.json")
	if err := os.WriteFile(outside, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(outside); err == nil {
		t.Fatal("transition receipt store unexpectedly loaded a path outside its boundary")
	}
}
