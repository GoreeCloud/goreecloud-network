package control

import (
	"errors"
	"strings"
	"time"
)

// PersistCapabilityStagingEvidenceForTransition writes capability staging
// evidence only after the corresponding immutable transition receipt is proven
// durable and exact. It refuses unresolved snapshot/receipt reconciliation and
// never changes capability authority, compatibility-bridge state, or cutover
// authorization.
func PersistCapabilityStagingEvidenceForTransition(
	transition StoredCapabilityTransitionResult,
	transitionStore *CapabilityTransitionReceiptStore,
	stagingStore *CapabilityStagingEvidenceStore,
	sourceRevision string,
	runtimeArtifactSHA256 string,
	acceptance IsolatedAcceptanceEvidence,
	now time.Time,
) (CapabilityStagingEvidence, string, error) {
	if transitionStore == nil || stagingStore == nil {
		return CapabilityStagingEvidence{}, "", errors.New("conduit control: transition and staging evidence stores are required")
	}
	if transition.ReconciliationRequired {
		return CapabilityStagingEvidence{}, "", errors.New("conduit control: capability staging evidence requires reconciled transition persistence")
	}
	if strings.TrimSpace(transition.ReceiptPath) == "" {
		return CapabilityStagingEvidence{}, "", errors.New("conduit control: capability staging evidence requires a persisted transition receipt path")
	}
	if err := validateCapabilityInventorySnapshot(transition.Snapshot); err != nil {
		return CapabilityStagingEvidence{}, "", err
	}
	if err := validateCapabilityTransitionReceipt(transition.Receipt); err != nil {
		return CapabilityStagingEvidence{}, "", err
	}
	if transition.Receipt.ToSnapshotFingerprint != transition.Snapshot.Fingerprint {
		return CapabilityStagingEvidence{}, "", errors.New("conduit control: transition receipt does not bind the staging snapshot")
	}
	if transition.Receipt.EvidenceSchema != acceptance.Schema {
		return CapabilityStagingEvidence{}, "", errors.New("conduit control: transition receipt evidence schema does not match staging acceptance")
	}
	if now.IsZero() {
		return CapabilityStagingEvidence{}, "", errors.New("conduit control: capability staging evidence time is required")
	}
	appliedAt, err := time.Parse(time.RFC3339Nano, transition.Receipt.AppliedAt)
	if err != nil {
		return CapabilityStagingEvidence{}, "", errors.New("conduit control: transition receipt applied_at is invalid")
	}
	if now.UTC().Before(appliedAt) {
		return CapabilityStagingEvidence{}, "", errors.New("conduit control: capability staging evidence cannot precede its transition receipt")
	}

	persistedReceipt, err := transitionStore.Load(transition.ReceiptPath)
	if err != nil {
		return CapabilityStagingEvidence{}, "", errors.New("conduit control: capability staging evidence requires a readable persisted transition receipt")
	}
	if persistedReceipt != transition.Receipt {
		return CapabilityStagingEvidence{}, "", errors.New("conduit control: persisted transition receipt does not match transition result")
	}

	evidence, err := BuildCapabilityStagingEvidence(
		transition.Snapshot,
		transition.Receipt.CapabilityID,
		sourceRevision,
		runtimeArtifactSHA256,
		acceptance,
		now,
	)
	if err != nil {
		return CapabilityStagingEvidence{}, "", err
	}
	path, err := stagingStore.Save(evidence)
	if err != nil {
		return CapabilityStagingEvidence{}, "", err
	}
	return evidence, path, nil
}
