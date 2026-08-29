package control

import (
	"errors"
	"strings"
	"time"
)

const CapabilityStagingStatusSchemaV1 = "goreecloud-conduit-capability-staging-status/v1"

// CapabilityStagingStatus is minimized durable-evidence status for central
// GoreeCloud consumers. It intentionally omits capability IDs, source
// revisions, runtime-artifact digests, peers, routes, policies, credentials,
// packet data, DNS queries, and other operational detail.
type CapabilityStagingStatus struct {
	Schema                           string    `json:"schema"`
	GeneratedAt                      string    `json:"generated_at"`
	InventoryFingerprint             string    `json:"inventory_fingerprint"`
	TransitionReceiptPersisted       bool      `json:"transition_receipt_persisted"`
	TransitionReconciliationRequired bool      `json:"transition_reconciliation_required"`
	StagingEvidencePersisted         bool      `json:"staging_evidence_persisted"`
	AcceptanceSchema                 string    `json:"acceptance_schema"`
	Authority                        Authority `json:"authority"`
	CompatibilityBridgeActive        bool      `json:"compatibility_bridge_active"`
	ProductionCutoverAuthorized      bool      `json:"production_cutover_authorized"`
}

// BuildPersistedCapabilityStagingStatus proves that the exact transition
// receipt and staging evidence are both durable before producing minimized
// status. It refuses unresolved reconciliation and never changes authority,
// compatibility-bridge state, or production cutover authorization.
func BuildPersistedCapabilityStagingStatus(
	transition StoredCapabilityTransitionResult,
	transitionStore *CapabilityTransitionReceiptStore,
	stagingStore *CapabilityStagingEvidenceStore,
	stagingPath string,
	now time.Time,
) (CapabilityStagingStatus, error) {
	if transitionStore == nil || stagingStore == nil {
		return CapabilityStagingStatus{}, errors.New("conduit control: transition and staging evidence stores are required")
	}
	if now.IsZero() {
		return CapabilityStagingStatus{}, errors.New("conduit control: capability staging status generation time is required")
	}
	if transition.ReconciliationRequired {
		return CapabilityStagingStatus{}, errors.New("conduit control: capability staging status requires reconciled transition persistence")
	}
	if strings.TrimSpace(transition.ReceiptPath) == "" || strings.TrimSpace(stagingPath) == "" {
		return CapabilityStagingStatus{}, errors.New("conduit control: persisted transition and staging evidence paths are required")
	}
	if err := validateCapabilityInventorySnapshot(transition.Snapshot); err != nil {
		return CapabilityStagingStatus{}, err
	}
	if err := validateCapabilityTransitionReceipt(transition.Receipt); err != nil {
		return CapabilityStagingStatus{}, err
	}
	if transition.Receipt.ToSnapshotFingerprint != transition.Snapshot.Fingerprint {
		return CapabilityStagingStatus{}, errors.New("conduit control: transition receipt does not bind the status snapshot")
	}

	persistedReceipt, err := transitionStore.Load(transition.ReceiptPath)
	if err != nil {
		return CapabilityStagingStatus{}, errors.New("conduit control: capability staging status requires a readable persisted transition receipt")
	}
	if persistedReceipt != transition.Receipt {
		return CapabilityStagingStatus{}, errors.New("conduit control: persisted transition receipt does not match transition result")
	}
	staging, err := stagingStore.Load(stagingPath)
	if err != nil {
		return CapabilityStagingStatus{}, errors.New("conduit control: capability staging status requires readable persisted staging evidence")
	}
	if staging.CapabilityID != transition.Receipt.CapabilityID ||
		staging.InventoryFingerprint != transition.Snapshot.Fingerprint ||
		staging.AcceptanceSchema != transition.Receipt.EvidenceSchema {
		return CapabilityStagingStatus{}, errors.New("conduit control: persisted staging evidence does not match transition identity")
	}
	if staging.Authority != transition.Receipt.Authority ||
		staging.CompatibilityBridgeActive != transition.Receipt.CompatibilityBridgeActive ||
		staging.ProductionCutoverAuthorized != transition.Receipt.ProductionCutoverAuthorized {
		return CapabilityStagingStatus{}, errors.New("conduit control: persisted staging evidence does not match transition safety state")
	}
	recordedAt, err := time.Parse(time.RFC3339Nano, staging.RecordedAt)
	if err != nil {
		return CapabilityStagingStatus{}, errors.New("conduit control: persisted staging evidence recorded_at is invalid")
	}
	appliedAt, err := time.Parse(time.RFC3339Nano, transition.Receipt.AppliedAt)
	if err != nil {
		return CapabilityStagingStatus{}, errors.New("conduit control: transition receipt applied_at is invalid")
	}
	if recordedAt.Before(appliedAt) {
		return CapabilityStagingStatus{}, errors.New("conduit control: staging evidence predates its transition receipt")
	}
	if staging.Authority != AuthorityInherited || !staging.CompatibilityBridgeActive || staging.ProductionCutoverAuthorized {
		return CapabilityStagingStatus{}, errors.New("conduit control: persisted staging evidence violates compatibility safety invariants")
	}

	return CapabilityStagingStatus{
		Schema:                           CapabilityStagingStatusSchemaV1,
		GeneratedAt:                      now.UTC().Format(time.RFC3339Nano),
		InventoryFingerprint:             transition.Snapshot.Fingerprint,
		TransitionReceiptPersisted:       true,
		TransitionReconciliationRequired: false,
		StagingEvidencePersisted:         true,
		AcceptanceSchema:                 staging.AcceptanceSchema,
		Authority:                        AuthorityInherited,
		CompatibilityBridgeActive:        true,
		ProductionCutoverAuthorized:      false,
	}, nil
}
