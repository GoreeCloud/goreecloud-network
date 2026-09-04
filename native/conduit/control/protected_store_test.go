package control

import (
	"errors"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func requireProtectedFileStoreTestSupport(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("Conduit durable protected file-store tests require Unix owner-only permission semantics")
	}
}

func TestProtectedFileStoreFailsClosedOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific protected file-store boundary")
	}
	store, err := NewCapabilityInventoryStore(filepath.Join(t.TempDir(), "inventory.json"))
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := BuildCapabilityInventorySnapshot(CapabilityInventory{
		Schema: CapabilityInventorySchemaV1,
		Capabilities: []CapabilityState{
			{ID: "control-api", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
		},
	}, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if saveErr := store.Save(snapshot); !errors.Is(saveErr, ErrProtectedFileStoreUnsupported) {
		t.Fatalf("Windows inventory persistence error=%v", saveErr)
	}

	receiptStore, err := NewCapabilityTransitionReceiptStore(filepath.Join(t.TempDir(), "receipts"))
	if err != nil {
		t.Fatal(err)
	}
	if _, loadErr := receiptStore.Load(filepath.Join(t.TempDir(), "receipt.json")); !errors.Is(loadErr, ErrProtectedFileStoreUnsupported) {
		t.Fatalf("Windows receipt persistence error=%v", loadErr)
	}
}
