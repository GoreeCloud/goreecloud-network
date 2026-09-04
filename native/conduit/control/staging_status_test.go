package control

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildPersistedCapabilityStagingStatusRequiresExactDurableEvidence(t *testing.T) {
	requireProtectedFileStoreTestSupport(t)
	before, after := testTransitionSnapshots(t)
	acceptance := completeStagingAcceptance("control-api")
	appliedAt := time.Date(2026, 8, 28, 20, 0, 0, 0, time.UTC)
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
	transition := StoredCapabilityTransitionResult{Snapshot: after, Receipt: receipt, ReceiptPath: receiptPath}
	stagingStore, err := NewCapabilityStagingEvidenceStore(filepath.Join(t.TempDir(), "staging"))
	if err != nil {
		t.Fatal(err)
	}
	_, stagingPath, err := PersistCapabilityStagingEvidenceForTransition(
		transition,
		transitionStore,
		stagingStore,
		strings.Repeat("a", 40),
		strings.Repeat("b", 64),
		acceptance,
		appliedAt.Add(time.Minute),
	)
	if err != nil {
		t.Fatal(err)
	}

	status, err := BuildPersistedCapabilityStagingStatus(transition, transitionStore, stagingStore, stagingPath, appliedAt.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if status.Schema != CapabilityStagingStatusSchemaV1 || status.InventoryFingerprint != after.Fingerprint {
		t.Fatalf("unexpected staging status identity: %+v", status)
	}
	if !status.TransitionReceiptPersisted || status.TransitionReconciliationRequired || !status.StagingEvidencePersisted {
		t.Fatalf("unexpected staging persistence state: %+v", status)
	}
	if status.Authority != AuthorityInherited || !status.CompatibilityBridgeActive || status.ProductionCutoverAuthorized {
		t.Fatalf("staging status violated compatibility safety invariants: %+v", status)
	}
}

func TestBuildPersistedCapabilityStagingStatusRejectsUnreconciledTransition(t *testing.T) {
	requireProtectedFileStoreTestSupport(t)
	before, after := testTransitionSnapshots(t)
	acceptance := completeStagingAcceptance("control-api")
	appliedAt := time.Date(2026, 8, 28, 20, 1, 0, 0, time.UTC)
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
	if _, err := BuildPersistedCapabilityStagingStatus(transition, transitionStore, stagingStore, "unused", appliedAt.Add(time.Minute)); err == nil {
		t.Fatal("unreconciled transition unexpectedly produced staging status")
	}
}
