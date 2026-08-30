package control

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const CapabilityInventoryRecoveryPointSchemaV1 = "goreecloud-conduit-capability-inventory-recovery-point/v1"

// CapabilityInventoryRecoveryPoint is an immutable, privacy-safe recovery
// record for the Conduit migration inventory. CapabilityInventory contains no
// peer, route, packet, DNS-query, credential, device, or user state.
type CapabilityInventoryRecoveryPoint struct {
	Schema                      string                      `json:"schema"`
	CreatedAt                   string                      `json:"created_at"`
	SourceRevision              string                      `json:"source_revision"`
	SnapshotFingerprint         string                      `json:"snapshot_fingerprint"`
	Snapshot                    CapabilityInventorySnapshot `json:"snapshot"`
	ProductionCutoverAuthorized bool                        `json:"production_cutover_authorized"`
}

// CapabilityInventoryRecoveryStore persists immutable recovery points below an
// owner-protected directory. Recovery evidence is separate from the active
// CapabilityInventoryStore so replacing active state never overwrites the
// recovery record that may be needed to reverse the transition.
type CapabilityInventoryRecoveryStore struct {
	root string
}

func NewCapabilityInventoryRecoveryStore(root string) (*CapabilityInventoryRecoveryStore, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("conduit control: capability inventory recovery root is required")
	}
	return &CapabilityInventoryRecoveryStore{root: root}, nil
}

func BuildCapabilityInventoryRecoveryPoint(snapshot CapabilityInventorySnapshot, sourceRevision string, now time.Time) (CapabilityInventoryRecoveryPoint, error) {
	if err := validateCapabilityInventorySnapshot(snapshot); err != nil {
		return CapabilityInventoryRecoveryPoint{}, err
	}
	sourceRevision = strings.ToLower(strings.TrimSpace(sourceRevision))
	if !validPlatformEvidenceHex(sourceRevision, 40) {
		return CapabilityInventoryRecoveryPoint{}, errors.New("conduit control: recovery point source revision is invalid")
	}
	if now.IsZero() {
		return CapabilityInventoryRecoveryPoint{}, errors.New("conduit control: recovery point creation time is required")
	}
	return CapabilityInventoryRecoveryPoint{
		Schema:                      CapabilityInventoryRecoveryPointSchemaV1,
		CreatedAt:                   now.UTC().Format(time.RFC3339Nano),
		SourceRevision:              sourceRevision,
		SnapshotFingerprint:         snapshot.Fingerprint,
		Snapshot:                    cloneCapabilityInventorySnapshot(snapshot),
		ProductionCutoverAuthorized: false,
	}, nil
}

func (s *CapabilityInventoryRecoveryStore) Save(point CapabilityInventoryRecoveryPoint) (string, error) {
	if s == nil || strings.TrimSpace(s.root) == "" {
		return "", errors.New("conduit control: capability inventory recovery store is not initialized")
	}
	if err := requireProtectedFileStoreSupport(); err != nil {
		return "", err
	}
	if err := validateCapabilityInventoryRecoveryPoint(point); err != nil {
		return "", err
	}
	if err := os.MkdirAll(s.root, 0o700); err != nil {
		return "", fmt.Errorf("conduit control: create inventory recovery directory: %w", err)
	}
	rootInfo, err := os.Lstat(s.root)
	if err != nil {
		return "", fmt.Errorf("conduit control: inspect inventory recovery directory: %w", err)
	}
	if !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("conduit control: inventory recovery root must be a real directory")
	}
	if err := os.Chmod(s.root, 0o700); err != nil {
		return "", fmt.Errorf("conduit control: protect inventory recovery directory: %w", err)
	}
	path := filepath.Join(s.root, point.SnapshotFingerprint+".json")
	encoded, err := json.MarshalIndent(point, "", "  ")
	if err != nil {
		return "", fmt.Errorf("conduit control: encode inventory recovery point: %w", err)
	}
	encoded = append(encoded, '\n')

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", fmt.Errorf("conduit control: create immutable inventory recovery point: %w", err)
	}
	keep := false
	defer func() {
		_ = file.Close()
		if !keep {
			_ = os.Remove(path)
		}
	}()
	if _, err := file.Write(encoded); err != nil {
		return "", fmt.Errorf("conduit control: write inventory recovery point: %w", err)
	}
	if err := file.Sync(); err != nil {
		return "", fmt.Errorf("conduit control: sync inventory recovery point: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("conduit control: close inventory recovery point: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return "", fmt.Errorf("conduit control: protect inventory recovery point: %w", err)
	}
	keep = true
	return path, nil
}

func (s *CapabilityInventoryRecoveryStore) Load(snapshotFingerprint string) (CapabilityInventoryRecoveryPoint, error) {
	if s == nil || strings.TrimSpace(s.root) == "" {
		return CapabilityInventoryRecoveryPoint{}, errors.New("conduit control: capability inventory recovery store is not initialized")
	}
	if err := requireProtectedFileStoreSupport(); err != nil {
		return CapabilityInventoryRecoveryPoint{}, err
	}
	rootInfo, err := os.Lstat(s.root)
	if err != nil {
		return CapabilityInventoryRecoveryPoint{}, err
	}
	if !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return CapabilityInventoryRecoveryPoint{}, errors.New("conduit control: inventory recovery root must be a real directory")
	}
	if rootInfo.Mode().Perm()&0o077 != 0 {
		return CapabilityInventoryRecoveryPoint{}, errors.New("conduit control: inventory recovery directory permissions are too broad")
	}
	snapshotFingerprint = strings.ToLower(strings.TrimSpace(snapshotFingerprint))
	if !validPlatformEvidenceHex(snapshotFingerprint, 64) {
		return CapabilityInventoryRecoveryPoint{}, errors.New("conduit control: recovery snapshot fingerprint is invalid")
	}
	path := filepath.Join(s.root, snapshotFingerprint+".json")
	info, err := os.Lstat(path)
	if err != nil {
		return CapabilityInventoryRecoveryPoint{}, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return CapabilityInventoryRecoveryPoint{}, errors.New("conduit control: recovery point must be a regular non-symlink file")
	}
	if info.Mode().Perm()&0o077 != 0 {
		return CapabilityInventoryRecoveryPoint{}, errors.New("conduit control: recovery point permissions are too broad")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return CapabilityInventoryRecoveryPoint{}, err
	}
	var point CapabilityInventoryRecoveryPoint
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&point); err != nil {
		return CapabilityInventoryRecoveryPoint{}, fmt.Errorf("conduit control: decode inventory recovery point: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return CapabilityInventoryRecoveryPoint{}, errors.New("conduit control: inventory recovery point contains trailing JSON data")
		}
		return CapabilityInventoryRecoveryPoint{}, fmt.Errorf("conduit control: decode trailing inventory recovery data: %w", err)
	}
	if err := validateCapabilityInventoryRecoveryPoint(point); err != nil {
		return CapabilityInventoryRecoveryPoint{}, err
	}
	if point.SnapshotFingerprint != snapshotFingerprint {
		return CapabilityInventoryRecoveryPoint{}, errors.New("conduit control: recovery point filename fingerprint mismatch")
	}
	return point, nil
}

// RestoreCapabilityInventoryRecoveryPoint performs an explicit compare-and-swap
// restore of the privacy-safe migration inventory. The caller must supply the
// exact currently active fingerprint so stale rollback tooling cannot overwrite
// a newer transition. This function cannot authorize production cutover.
func RestoreCapabilityInventoryRecoveryPoint(
	activeStore *CapabilityInventoryStore,
	point CapabilityInventoryRecoveryPoint,
	expectedCurrentFingerprint string,
	now time.Time,
) (CapabilityInventorySnapshot, error) {
	if activeStore == nil {
		return CapabilityInventorySnapshot{}, errors.New("conduit control: active capability inventory store is required")
	}
	if err := validateCapabilityInventoryRecoveryPoint(point); err != nil {
		return CapabilityInventorySnapshot{}, err
	}
	if now.IsZero() {
		return CapabilityInventorySnapshot{}, errors.New("conduit control: recovery restore time is required")
	}
	current, err := activeStore.Load()
	if err != nil {
		return CapabilityInventorySnapshot{}, err
	}
	expectedCurrentFingerprint = strings.ToLower(strings.TrimSpace(expectedCurrentFingerprint))
	if !validPlatformEvidenceHex(expectedCurrentFingerprint, 64) || expectedCurrentFingerprint != current.Fingerprint {
		return CapabilityInventorySnapshot{}, errors.New("conduit control: active inventory changed after recovery was planned")
	}
	if current.Fingerprint == point.SnapshotFingerprint {
		return CapabilityInventorySnapshot{}, errors.New("conduit control: recovery point already matches active inventory")
	}
	restored, err := BuildCapabilityInventorySnapshot(point.Snapshot.Inventory, now)
	if err != nil {
		return CapabilityInventorySnapshot{}, err
	}
	if restored.Fingerprint != point.SnapshotFingerprint {
		return CapabilityInventorySnapshot{}, errors.New("conduit control: recovered inventory fingerprint mismatch")
	}
	if err := activeStore.Save(restored); err != nil {
		return CapabilityInventorySnapshot{}, err
	}
	return restored, nil
}

func validateCapabilityInventoryRecoveryPoint(point CapabilityInventoryRecoveryPoint) error {
	if point.Schema != CapabilityInventoryRecoveryPointSchemaV1 {
		return errors.New("conduit control: unsupported capability inventory recovery-point schema")
	}
	if _, err := time.Parse(time.RFC3339Nano, point.CreatedAt); err != nil {
		return errors.New("conduit control: recovery point created_at is invalid")
	}
	if !validPlatformEvidenceHex(point.SourceRevision, 40) {
		return errors.New("conduit control: recovery point source revision is invalid")
	}
	if point.ProductionCutoverAuthorized {
		return errors.New("conduit control: recovery point cannot authorize production cutover")
	}
	if err := validateCapabilityInventorySnapshot(point.Snapshot); err != nil {
		return err
	}
	if point.SnapshotFingerprint != point.Snapshot.Fingerprint || !validPlatformEvidenceHex(point.SnapshotFingerprint, 64) {
		return errors.New("conduit control: recovery point snapshot fingerprint mismatch")
	}
	return nil
}

func cloneCapabilityInventorySnapshot(snapshot CapabilityInventorySnapshot) CapabilityInventorySnapshot {
	clone := snapshot
	clone.Inventory = CapabilityInventory{
		Schema:       snapshot.Inventory.Schema,
		Capabilities: append([]CapabilityState(nil), snapshot.Inventory.Capabilities...),
	}
	return clone
}
