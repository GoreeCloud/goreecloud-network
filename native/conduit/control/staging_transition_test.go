package control

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPersistCapabilityStagingEvidenceForTransitionRequiresPersistedExactReceipt(t *testing.T) {
	requireProtectedFileStoreTestSupport(t)
	before, after := testTransitionSnapshots(t)
	acceptance := completeStagingAcceptance("control-api")
	appliedAt := time.Date(2026, 8, 27, 17, 0, 0, 0, time.UTC)
	receipt, err := BuildCapabilityTransitionReceipt(before, after, "control-api", acceptance.Schema, appliedAt)
	if err != nil {
		t.Fatal(err)
	}
	transitionStore, err := NewCapabilityTransitionReceiptStore(filepath.Join(t.TempDir(), "transitions"))
	if err != nil {
		t.Fatal(err)
	}
	receiptPath, err := transitionStore.Save(receipt)
	if err != nil {
		t.Fatal(err)
	}
	stagingStore, err := NewCapabilityStagingEvidenceStore(filepath.Join(t.TempDir(), "staging"))
	if err != nil {
		t.Fatal(err)
	}
	transition := StoredCapabilityTransitionResult{
		Snapshot:    after,
		Receipt:     receipt,
		ReceiptPath: receiptPath,
	}

	evidence, path, err := PersistCapabilityStagingEvidenceForTransition(
		transition,
		transitionStore,
		stagingStore,
		"183df2ca0353e035ccf6dea54dbd268a125dad6c",
		strings.Repeat("d", 64),
		acceptance,
		appliedAt.Add(time.Minute),
	)
	if err != nil {
		t.Fatal(err)
	}
	if path == "" || evidence.InventoryFingerprint != after.Fingerprint || evidence.CapabilityID != receipt.CapabilityID {
		t.Fatalf("unexpected transition-bound staging evidence: %+v path=%q", evidence, path)
	}
	if evidence.Authority != AuthorityInherited || !evidence.CompatibilityBridgeActive || evidence.ProductionCutoverAuthorized {
		t.Fatal("transition-bound staging evidence violated safety invariants")
	}
}

func TestPersistCapabilityStagingEvidenceForTransitionRejectsReconciliation(t *testing.T) {
	requireProtectedFileStoreTestSupport(t)
	before, after := testTransitionSnapshots(t)
	acceptance := completeStagingAcceptance("control-api")
	appliedAt := time.Date(2026, 8, 27, 17, 1, 0, 0, time.UTC)
	receipt, err := BuildCapabilityTransitionReceipt(before, after, "control-api", acceptance.Schema, appliedAt)
	if err != nil {
		t.Fatal(err)
	}
	transitionStore, err := NewCapabilityTransitionReceiptStore(filepath.Join(t.TempDir(), "transitions"))
	if err != nil {
		t.Fatal(err)
	}
	stagingStore, err := NewCapabilityStagingEvidenceStore(filepath.Join(t.TempDir(), "staging"))
	if err != nil {
		t.Fatal(err)
	}
	transition := StoredCapabilityTransitionResult{
		Snapshot:               after,
		Receipt:                receipt,
		ReceiptPath:            transitionReceiptPath(transitionStore, receipt),
		ReconciliationRequired: true,
	}

	if _, _, err := PersistCapabilityStagingEvidenceForTransition(
		transition,
		transitionStore,
		stagingStore,
		"183df2ca0353e035ccf6dea54dbd268a125dad6c",
		strings.Repeat("e", 64),
		acceptance,
		appliedAt.Add(time.Minute),
	); err == nil {
		t.Fatal("unreconciled transition unexpectedly produced staging evidence")
	}
}

func TestPersistCapabilityStagingEvidenceForTransitionRejectsReceiptMismatch(t *testing.T) {
	requireProtectedFileStoreTestSupport(t)
	before, after := testTransitionSnapshots(t)
	acceptance := completeStagingAcceptance("control-api")
	appliedAt := time.Date(2026, 8, 27, 17, 2, 0, 0, time.UTC)
	receipt, err := BuildCapabilityTransitionReceipt(before, after, "control-api", acceptance.Schema, appliedAt)
	if err != nil {
		t.Fatal(err)
	}
	transitionStore, err := NewCapabilityTransitionReceiptStore(filepath.Join(t.TempDir(), "transitions"))
	if err != nil {
		t.Fatal(err)
	}
	receiptPath, err := transitionStore.Save(receipt)
	if err != nil {
		t.Fatal(err)
	}
	stagingStore, err := NewCapabilityStagingEvidenceStore(filepath.Join(t.TempDir(), "staging"))
	if err != nil {
		t.Fatal(err)
	}
	tampered := receipt
	tampered.AppliedAt = appliedAt.Add(time.Second).Format(time.RFC3339Nano)
	transition := StoredCapabilityTransitionResult{Snapshot: after, Receipt: tampered, ReceiptPath: receiptPath}

	if _, _, err := PersistCapabilityStagingEvidenceForTransition(
		transition,
		transitionStore,
		stagingStore,
		"183df2ca0353e035ccf6dea54dbd268a125dad6c",
		strings.Repeat("f", 64),
		acceptance,
		appliedAt.Add(time.Minute),
	); err == nil {
		t.Fatal("mismatched persisted transition receipt unexpectedly produced staging evidence")
	}
}
