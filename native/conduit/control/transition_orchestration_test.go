package control

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func storedTransitionFixture(t *testing.T) (*CapabilityInventoryStore, CapabilityInventorySnapshot) {
	t.Helper()
	store, err := NewCapabilityInventoryStore(filepath.Join(t.TempDir(), "inventory.json"))
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := BuildCapabilityInventorySnapshot(CapabilityInventory{
		Schema: CapabilityInventorySchemaV1,
		Capabilities: []CapabilityState{
			{ID: "control-api", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
			{ID: "relay", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
		},
	}, time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(snapshot); err != nil {
		t.Fatal(err)
	}
	return store, snapshot
}

func TestApplyStoredCapabilityTransitionWithReceiptSupportsExactReplay(t *testing.T) {
	inventoryStore, before := storedTransitionFixture(t)
	receiptStore, err := NewCapabilityTransitionReceiptStore(filepath.Join(t.TempDir(), "receipts"))
	if err != nil {
		t.Fatal(err)
	}
	evidence := completeTransitionEvidence("control-api")
	result, err := ApplyStoredCapabilityTransitionWithReceipt(inventoryStore, receiptStore, before.Fingerprint, "control-api", evidence, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if result.ReconciliationRequired || result.IdempotentReplay {
		t.Fatalf("unexpected first transition result: %+v", result)
	}
	if _, err := os.Stat(result.ReceiptPath); err != nil {
		t.Fatal(err)
	}
	if result.Receipt.FromSnapshotFingerprint != before.Fingerprint || result.Receipt.ToSnapshotFingerprint != result.Snapshot.Fingerprint {
		t.Fatal("receipt does not bind before and after snapshot fingerprints")
	}

	replay, err := ApplyStoredCapabilityTransitionWithReceipt(inventoryStore, receiptStore, before.Fingerprint, "control-api", evidence, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if !replay.IdempotentReplay || replay.ReconciliationRequired || replay.ReceiptPath != result.ReceiptPath {
		t.Fatalf("exact retry was not recognized as idempotent: %+v", replay)
	}
}

func TestApplyStoredCapabilityTransitionReturnsReconciliationMaterial(t *testing.T) {
	inventoryStore, before := storedTransitionFixture(t)
	receiptRoot := filepath.Join(t.TempDir(), "receipt-root")
	if err := os.WriteFile(receiptRoot, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	receiptStore, err := NewCapabilityTransitionReceiptStore(receiptRoot)
	if err != nil {
		t.Fatal(err)
	}
	result, err := ApplyStoredCapabilityTransitionWithReceipt(inventoryStore, receiptStore, before.Fingerprint, "control-api", completeTransitionEvidence("control-api"), time.Now())
	if err == nil {
		t.Fatal("receipt persistence failure unexpectedly succeeded")
	}
	if !result.ReconciliationRequired || result.Snapshot.Fingerprint == before.Fingerprint || result.Receipt.FromSnapshotFingerprint != before.Fingerprint {
		t.Fatalf("missing deterministic reconciliation material: %+v", result)
	}
	stored, loadErr := inventoryStore.Load()
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	if stored.Fingerprint != result.Snapshot.Fingerprint {
		t.Fatal("snapshot did not persist before receipt reconciliation failure")
	}
}
