package control

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const CapabilityTransitionReceiptSchemaV1 = "goreecloud-conduit-capability-transition-receipt/v1"

type CapabilityTransitionReceipt struct {
	Schema                      string    `json:"schema"`
	CapabilityID                string    `json:"capability_id"`
	AppliedAt                   string    `json:"applied_at"`
	FromSnapshotFingerprint     string    `json:"from_snapshot_fingerprint"`
	ToSnapshotFingerprint       string    `json:"to_snapshot_fingerprint"`
	FromStage                   string    `json:"from_stage"`
	ToStage                     string    `json:"to_stage"`
	Authority                   Authority `json:"authority"`
	CompatibilityBridgeActive   bool      `json:"compatibility_bridge_active"`
	EvidenceSchema              string    `json:"evidence_schema"`
	ProductionCutoverAuthorized bool      `json:"production_cutover_authorized"`
}

type CapabilityTransitionReceiptStore struct {
	directory string
}

func NewCapabilityTransitionReceiptStore(directory string) (*CapabilityTransitionReceiptStore, error) {
	directory = strings.TrimSpace(directory)
	if directory == "" {
		return nil, errors.New("conduit control: capability transition receipt directory is required")
	}
	return &CapabilityTransitionReceiptStore{directory: directory}, nil
}

// BuildCapabilityTransitionReceipt proves one bounded source-state transition.
// It rejects any before/after pair where a non-target capability changed or the
// target crossed an authority, bridge, or production-cutover boundary.
func BuildCapabilityTransitionReceipt(before, after CapabilityInventorySnapshot, capabilityID, evidenceSchema string, now time.Time) (CapabilityTransitionReceipt, error) {
	if err := validateCapabilityInventorySnapshot(before); err != nil {
		return CapabilityTransitionReceipt{}, err
	}
	if err := validateCapabilityInventorySnapshot(after); err != nil {
		return CapabilityTransitionReceipt{}, err
	}
	capabilityID = strings.TrimSpace(capabilityID)
	if capabilityID == "" {
		return CapabilityTransitionReceipt{}, errors.New("conduit control: capability id is required")
	}
	if strings.TrimSpace(evidenceSchema) != IsolatedAcceptanceSchemaV1 {
		return CapabilityTransitionReceipt{}, errors.New("conduit control: transition receipt requires isolated acceptance evidence schema")
	}
	if now.IsZero() {
		return CapabilityTransitionReceipt{}, errors.New("conduit control: capability transition receipt time is required")
	}
	if before.Fingerprint == after.Fingerprint {
		return CapabilityTransitionReceipt{}, errors.New("conduit control: transition receipt requires a changed snapshot fingerprint")
	}
	if len(before.Inventory.Capabilities) != len(after.Inventory.Capabilities) {
		return CapabilityTransitionReceipt{}, errors.New("conduit control: transition receipt capability inventory size changed")
	}

	beforeByID := make(map[string]CapabilityState, len(before.Inventory.Capabilities))
	for _, capability := range before.Inventory.Capabilities {
		beforeByID[capability.ID] = capability
	}
	afterByID := make(map[string]CapabilityState, len(after.Inventory.Capabilities))
	for _, capability := range after.Inventory.Capabilities {
		afterByID[capability.ID] = capability
	}
	beforeTarget, ok := beforeByID[capabilityID]
	if !ok {
		return CapabilityTransitionReceipt{}, errors.New("conduit control: transition receipt target missing from before snapshot")
	}
	afterTarget, ok := afterByID[capabilityID]
	if !ok {
		return CapabilityTransitionReceipt{}, errors.New("conduit control: transition receipt target missing from after snapshot")
	}
	for id, beforeCapability := range beforeByID {
		afterCapability, exists := afterByID[id]
		if !exists {
			return CapabilityTransitionReceipt{}, errors.New("conduit control: transition receipt capability set changed")
		}
		if id == capabilityID {
			continue
		}
		if beforeCapability != afterCapability {
			return CapabilityTransitionReceipt{}, errors.New("conduit control: transition receipt detected a non-target capability change")
		}
	}

	if beforeTarget.MigrationStage != "implementation" || afterTarget.MigrationStage != MigrationStageIsolatedValidation {
		return CapabilityTransitionReceipt{}, errors.New("conduit control: transition receipt requires implementation to isolated-validation")
	}
	if beforeTarget.Authority != AuthorityInherited || afterTarget.Authority != AuthorityInherited {
		return CapabilityTransitionReceipt{}, errors.New("conduit control: transition receipt cannot cross inherited authority")
	}
	if !beforeTarget.CompatibilityBridgeActive || !afterTarget.CompatibilityBridgeActive {
		return CapabilityTransitionReceipt{}, errors.New("conduit control: transition receipt requires the compatibility bridge to remain active")
	}
	if beforeTarget.ProductionCutoverAuthorized || afterTarget.ProductionCutoverAuthorized {
		return CapabilityTransitionReceipt{}, errors.New("conduit control: transition receipt cannot authorize production cutover")
	}

	return CapabilityTransitionReceipt{
		Schema:                      CapabilityTransitionReceiptSchemaV1,
		CapabilityID:                capabilityID,
		AppliedAt:                   now.UTC().Format(time.RFC3339Nano),
		FromSnapshotFingerprint:     before.Fingerprint,
		ToSnapshotFingerprint:       after.Fingerprint,
		FromStage:                   beforeTarget.MigrationStage,
		ToStage:                     afterTarget.MigrationStage,
		Authority:                   afterTarget.Authority,
		CompatibilityBridgeActive:   afterTarget.CompatibilityBridgeActive,
		EvidenceSchema:              IsolatedAcceptanceSchemaV1,
		ProductionCutoverAuthorized: false,
	}, nil
}

// Save writes one immutable owner-only receipt. The deterministic filename and
// O_EXCL semantics prevent an accepted transition receipt from being replaced.
func (s *CapabilityTransitionReceiptStore) Save(receipt CapabilityTransitionReceipt) (string, error) {
	if s == nil || strings.TrimSpace(s.directory) == "" {
		return "", errors.New("conduit control: capability transition receipt store is not initialized")
	}
	if err := validateCapabilityTransitionReceipt(receipt); err != nil {
		return "", err
	}
	if err := os.MkdirAll(s.directory, 0o700); err != nil {
		return "", fmt.Errorf("conduit control: create transition receipt directory: %w", err)
	}
	if err := os.Chmod(s.directory, 0o700); err != nil {
		return "", fmt.Errorf("conduit control: protect transition receipt directory: %w", err)
	}
	fingerprint := sha256.Sum256([]byte(receipt.CapabilityID + "\x00" + receipt.FromSnapshotFingerprint + "\x00" + receipt.ToSnapshotFingerprint))
	path := filepath.Join(s.directory, "transition-"+hex.EncodeToString(fingerprint[:8])+".json")
	encoded, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return "", fmt.Errorf("conduit control: encode transition receipt: %w", err)
	}
	encoded = append(encoded, '\n')
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return "", errors.New("conduit control: capability transition receipt already exists")
		}
		return "", fmt.Errorf("conduit control: create transition receipt: %w", err)
	}
	if _, err := file.Write(encoded); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("conduit control: write transition receipt: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("conduit control: sync transition receipt: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("conduit control: close transition receipt: %w", err)
	}
	return path, nil
}

func (s *CapabilityTransitionReceiptStore) Load(path string) (CapabilityTransitionReceipt, error) {
	if s == nil || strings.TrimSpace(s.directory) == "" {
		return CapabilityTransitionReceipt{}, errors.New("conduit control: capability transition receipt store is not initialized")
	}
	cleanDirectory := filepath.Clean(s.directory)
	cleanPath := filepath.Clean(path)
	relative, err := filepath.Rel(cleanDirectory, cleanPath)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return CapabilityTransitionReceipt{}, errors.New("conduit control: transition receipt path is outside the receipt store")
	}
	info, err := os.Lstat(cleanPath)
	if err != nil {
		return CapabilityTransitionReceipt{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return CapabilityTransitionReceipt{}, errors.New("conduit control: transition receipt file is not protected")
	}
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return CapabilityTransitionReceipt{}, err
	}
	var receipt CapabilityTransitionReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return CapabilityTransitionReceipt{}, fmt.Errorf("conduit control: decode transition receipt: %w", err)
	}
	if err := validateCapabilityTransitionReceipt(receipt); err != nil {
		return CapabilityTransitionReceipt{}, err
	}
	return receipt, nil
}

func validateCapabilityTransitionReceipt(receipt CapabilityTransitionReceipt) error {
	if receipt.Schema != CapabilityTransitionReceiptSchemaV1 {
		return errors.New("conduit control: unsupported capability transition receipt schema")
	}
	if strings.TrimSpace(receipt.CapabilityID) == "" {
		return errors.New("conduit control: capability transition receipt id is required")
	}
	if _, err := time.Parse(time.RFC3339Nano, receipt.AppliedAt); err != nil {
		return errors.New("conduit control: capability transition receipt applied_at is invalid")
	}
	for _, fingerprint := range []string{receipt.FromSnapshotFingerprint, receipt.ToSnapshotFingerprint} {
		decoded, err := hex.DecodeString(strings.TrimSpace(fingerprint))
		if err != nil || len(decoded) != sha256.Size {
			return errors.New("conduit control: capability transition receipt fingerprint is invalid")
		}
	}
	if receipt.FromSnapshotFingerprint == receipt.ToSnapshotFingerprint {
		return errors.New("conduit control: capability transition receipt fingerprints must differ")
	}
	if receipt.FromStage != "implementation" || receipt.ToStage != MigrationStageIsolatedValidation {
		return errors.New("conduit control: capability transition receipt stage boundary is invalid")
	}
	if receipt.Authority != AuthorityInherited || !receipt.CompatibilityBridgeActive || receipt.ProductionCutoverAuthorized {
		return errors.New("conduit control: capability transition receipt safety invariants are invalid")
	}
	if receipt.EvidenceSchema != IsolatedAcceptanceSchemaV1 {
		return errors.New("conduit control: capability transition receipt evidence schema is invalid")
	}
	return nil
}
