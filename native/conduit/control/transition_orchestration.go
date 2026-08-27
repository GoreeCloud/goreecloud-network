package control

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type StoredCapabilityTransitionResult struct {
	Snapshot               CapabilityInventorySnapshot
	Receipt                CapabilityTransitionReceipt
	ReceiptPath            string
	ReconciliationRequired bool
	IdempotentReplay       bool
}

// ApplyStoredCapabilityTransitionWithReceipt coordinates the durable inventory
// transition and its immutable audit receipt without pretending the two files
// form one filesystem transaction. If snapshot persistence succeeds but receipt
// persistence fails, the returned reconciliation material is deterministic and
// can be repaired safely. A later retry recognizes an exact existing receipt.
func ApplyStoredCapabilityTransitionWithReceipt(inventoryStore *CapabilityInventoryStore, receiptStore *CapabilityTransitionReceiptStore, expectedFingerprint, capabilityID string, evidence IsolatedAcceptanceEvidence, now time.Time) (StoredCapabilityTransitionResult, error) {
	if inventoryStore == nil || receiptStore == nil {
		return StoredCapabilityTransitionResult{}, errors.New("conduit control: inventory and receipt stores are required")
	}
	if now.IsZero() {
		return StoredCapabilityTransitionResult{}, errors.New("conduit control: stored transition time is required")
	}
	expectedFingerprint = strings.ToLower(strings.TrimSpace(expectedFingerprint))
	capabilityID = strings.TrimSpace(capabilityID)
	if expectedFingerprint == "" || capabilityID == "" {
		return StoredCapabilityTransitionResult{}, errors.New("conduit control: expected fingerprint and capability id are required")
	}

	current, err := inventoryStore.Load()
	if err != nil {
		return StoredCapabilityTransitionResult{}, err
	}
	if current.Fingerprint != expectedFingerprint {
		return replayStoredCapabilityTransition(receiptStore, current, expectedFingerprint, capabilityID, evidence)
	}

	nextInventory, err := AdvanceInventoryCapabilityToIsolatedValidation(current.Inventory, capabilityID, evidence)
	if err != nil {
		return StoredCapabilityTransitionResult{}, err
	}
	next, err := BuildCapabilityInventorySnapshot(nextInventory, now)
	if err != nil {
		return StoredCapabilityTransitionResult{}, err
	}
	if next.Fingerprint == current.Fingerprint {
		return StoredCapabilityTransitionResult{}, errors.New("conduit control: capability transition produced no state change")
	}
	receipt, err := BuildCapabilityTransitionReceipt(current, next, capabilityID, evidence.Schema, now)
	if err != nil {
		return StoredCapabilityTransitionResult{}, err
	}
	result := StoredCapabilityTransitionResult{Snapshot: next, Receipt: receipt, ReceiptPath: transitionReceiptPath(receiptStore, receipt)}

	if err := inventoryStore.Save(next); err != nil {
		return StoredCapabilityTransitionResult{}, err
	}
	path, err := receiptStore.Save(receipt)
	if err != nil {
		result.ReconciliationRequired = true
		return result, fmt.Errorf("conduit control: capability snapshot advanced but transition receipt requires reconciliation: %w", err)
	}
	result.ReceiptPath = path
	return result, nil
}

func replayStoredCapabilityTransition(receiptStore *CapabilityTransitionReceiptStore, current CapabilityInventorySnapshot, expectedFingerprint, capabilityID string, evidence IsolatedAcceptanceEvidence) (StoredCapabilityTransitionResult, error) {
	var target *CapabilityState
	for i := range current.Inventory.Capabilities {
		if current.Inventory.Capabilities[i].ID == capabilityID {
			target = &current.Inventory.Capabilities[i]
			break
		}
	}
	if target == nil || target.MigrationStage != MigrationStageIsolatedValidation || target.Authority != AuthorityInherited || !target.CompatibilityBridgeActive || target.ProductionCutoverAuthorized {
		return StoredCapabilityTransitionResult{}, errors.New("conduit control: capability inventory snapshot fingerprint mismatch")
	}
	decision, err := EvaluateIsolatedAcceptance(*target, evidence)
	if err != nil {
		return StoredCapabilityTransitionResult{}, err
	}
	if !decision.EligibleForIsolatedValidation {
		return StoredCapabilityTransitionResult{}, errors.New("conduit control: idempotent transition replay requires complete isolated acceptance evidence")
	}
	probe := CapabilityTransitionReceipt{
		Schema:                      CapabilityTransitionReceiptSchemaV1,
		CapabilityID:                capabilityID,
		FromSnapshotFingerprint:     expectedFingerprint,
		ToSnapshotFingerprint:       current.Fingerprint,
		FromStage:                   "implementation",
		ToStage:                     MigrationStageIsolatedValidation,
		Authority:                   AuthorityInherited,
		CompatibilityBridgeActive:   true,
		EvidenceSchema:              evidence.Schema,
		ProductionCutoverAuthorized: false,
	}
	path := transitionReceiptPath(receiptStore, probe)
	existing, err := receiptStore.Load(path)
	if err != nil {
		return StoredCapabilityTransitionResult{}, errors.New("conduit control: capability inventory changed and no exact transition receipt proves an idempotent replay")
	}
	if existing.CapabilityID != capabilityID || existing.FromSnapshotFingerprint != expectedFingerprint || existing.ToSnapshotFingerprint != current.Fingerprint || existing.EvidenceSchema != evidence.Schema || existing.FromStage != "implementation" || existing.ToStage != MigrationStageIsolatedValidation || existing.Authority != AuthorityInherited || !existing.CompatibilityBridgeActive || existing.ProductionCutoverAuthorized {
		return StoredCapabilityTransitionResult{}, errors.New("conduit control: existing transition receipt does not match requested replay")
	}
	return StoredCapabilityTransitionResult{Snapshot: current, Receipt: existing, ReceiptPath: path, IdempotentReplay: true}, nil
}

func transitionReceiptPath(store *CapabilityTransitionReceiptStore, receipt CapabilityTransitionReceipt) string {
	fingerprint := sha256.Sum256([]byte(receipt.CapabilityID + "\x00" + receipt.FromSnapshotFingerprint + "\x00" + receipt.ToSnapshotFingerprint))
	return filepath.Join(store.directory, "transition-"+hex.EncodeToString(fingerprint[:8])+".json")
}
