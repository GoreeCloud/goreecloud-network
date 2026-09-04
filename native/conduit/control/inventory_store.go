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

const CapabilityInventorySnapshotSchemaV1 = "goreecloud-conduit-capability-inventory-snapshot/v1"

type CapabilityInventorySnapshot struct {
	Schema      string              `json:"schema"`
	UpdatedAt   string              `json:"updated_at"`
	Fingerprint string              `json:"fingerprint"`
	Inventory   CapabilityInventory `json:"inventory"`
}

type CapabilityInventoryStore struct {
	path string
}

func NewCapabilityInventoryStore(path string) (*CapabilityInventoryStore, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("conduit control: capability inventory store path is required")
	}
	return &CapabilityInventoryStore{path: path}, nil
}

func BuildCapabilityInventorySnapshot(inventory CapabilityInventory, now time.Time) (CapabilityInventorySnapshot, error) {
	if err := ValidateCapabilityInventory(inventory); err != nil {
		return CapabilityInventorySnapshot{}, err
	}
	if now.IsZero() {
		return CapabilityInventorySnapshot{}, errors.New("conduit control: capability inventory snapshot time is required")
	}
	fingerprint, err := capabilityInventoryFingerprint(inventory)
	if err != nil {
		return CapabilityInventorySnapshot{}, err
	}
	return CapabilityInventorySnapshot{
		Schema:      CapabilityInventorySnapshotSchemaV1,
		UpdatedAt:   now.UTC().Format(time.RFC3339Nano),
		Fingerprint: fingerprint,
		Inventory: CapabilityInventory{
			Schema:       inventory.Schema,
			Capabilities: append([]CapabilityState(nil), inventory.Capabilities...),
		},
	}, nil
}

func (s *CapabilityInventoryStore) Load() (CapabilityInventorySnapshot, error) {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return CapabilityInventorySnapshot{}, errors.New("conduit control: capability inventory store is not initialized")
	}
	if supportErr := requireProtectedFileStoreSupport(); supportErr != nil {
		return CapabilityInventorySnapshot{}, supportErr
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return CapabilityInventorySnapshot{}, err
	}
	var snapshot CapabilityInventorySnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return CapabilityInventorySnapshot{}, fmt.Errorf("conduit control: decode capability inventory snapshot: %w", err)
	}
	if err := validateCapabilityInventorySnapshot(snapshot); err != nil {
		return CapabilityInventorySnapshot{}, err
	}
	return snapshot, nil
}

func (s *CapabilityInventoryStore) Save(snapshot CapabilityInventorySnapshot) error {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return errors.New("conduit control: capability inventory store is not initialized")
	}
	if supportErr := requireProtectedFileStoreSupport(); supportErr != nil {
		return supportErr
	}
	if err := validateCapabilityInventorySnapshot(snapshot); err != nil {
		return err
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("conduit control: encode capability inventory snapshot: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("conduit control: create capability inventory directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".capability-inventory-*")
	if err != nil {
		return fmt.Errorf("conduit control: create temporary capability inventory snapshot: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("conduit control: protect temporary capability inventory snapshot: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("conduit control: write capability inventory snapshot: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("conduit control: sync capability inventory snapshot: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("conduit control: close capability inventory snapshot: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("conduit control: replace capability inventory snapshot: %w", err)
	}
	if err := os.Chmod(s.path, 0o600); err != nil {
		return fmt.Errorf("conduit control: protect capability inventory snapshot: %w", err)
	}
	return nil
}

// AdvanceStoredCapabilityToIsolatedValidation performs a compare-and-swap style
// source-state transition. The expected snapshot fingerprint prevents stale
// migration tooling from overwriting a newer capability inventory.
func AdvanceStoredCapabilityToIsolatedValidation(store *CapabilityInventoryStore, expectedFingerprint, capabilityID string, evidence IsolatedAcceptanceEvidence, now time.Time) (CapabilityInventorySnapshot, error) {
	if store == nil {
		return CapabilityInventorySnapshot{}, errors.New("conduit control: capability inventory store is required")
	}
	current, err := store.Load()
	if err != nil {
		return CapabilityInventorySnapshot{}, err
	}
	expectedFingerprint = strings.ToLower(strings.TrimSpace(expectedFingerprint))
	if expectedFingerprint == "" || expectedFingerprint != current.Fingerprint {
		return CapabilityInventorySnapshot{}, errors.New("conduit control: capability inventory snapshot fingerprint mismatch")
	}
	nextInventory, err := AdvanceInventoryCapabilityToIsolatedValidation(current.Inventory, capabilityID, evidence)
	if err != nil {
		return CapabilityInventorySnapshot{}, err
	}
	next, err := BuildCapabilityInventorySnapshot(nextInventory, now)
	if err != nil {
		return CapabilityInventorySnapshot{}, err
	}
	if next.Fingerprint == current.Fingerprint {
		return CapabilityInventorySnapshot{}, errors.New("conduit control: capability inventory transition produced no state change")
	}
	if err := store.Save(next); err != nil {
		return CapabilityInventorySnapshot{}, err
	}
	return next, nil
}

func validateCapabilityInventorySnapshot(snapshot CapabilityInventorySnapshot) error {
	if snapshot.Schema != CapabilityInventorySnapshotSchemaV1 {
		return errors.New("conduit control: unsupported capability inventory snapshot schema")
	}
	if _, err := time.Parse(time.RFC3339Nano, snapshot.UpdatedAt); err != nil {
		return errors.New("conduit control: capability inventory snapshot updated_at is invalid")
	}
	if err := ValidateCapabilityInventory(snapshot.Inventory); err != nil {
		return err
	}
	fingerprint, err := capabilityInventoryFingerprint(snapshot.Inventory)
	if err != nil {
		return err
	}
	if strings.ToLower(strings.TrimSpace(snapshot.Fingerprint)) != fingerprint {
		return errors.New("conduit control: capability inventory snapshot fingerprint mismatch")
	}
	return nil
}

func capabilityInventoryFingerprint(inventory CapabilityInventory) (string, error) {
	if err := ValidateCapabilityInventory(inventory); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(inventory)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}
