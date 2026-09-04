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

const CapabilityStagingEvidenceSchemaV1 = "goreecloud-conduit-capability-staging-evidence/v1"

// CapabilityStagingEvidence binds one isolated-validation capability to an
// exact source revision, immutable runtime artifact digest, inventory snapshot,
// and completed acceptance schema. It intentionally carries no peer, route,
// policy, credential, device, packet, DNS-query, or user data.
type CapabilityStagingEvidence struct {
	Schema                      string    `json:"schema"`
	CapabilityID                string    `json:"capability_id"`
	RecordedAt                  string    `json:"recorded_at"`
	SourceRevision              string    `json:"source_revision"`
	RuntimeArtifactSHA256       string    `json:"runtime_artifact_sha256"`
	InventoryFingerprint        string    `json:"inventory_fingerprint"`
	AcceptanceSchema            string    `json:"acceptance_schema"`
	Authority                   Authority `json:"authority"`
	CompatibilityBridgeActive   bool      `json:"compatibility_bridge_active"`
	ProductionCutoverAuthorized bool      `json:"production_cutover_authorized"`
}

type CapabilityStagingEvidenceStore struct {
	directory string
}

func NewCapabilityStagingEvidenceStore(directory string) (*CapabilityStagingEvidenceStore, error) {
	directory = strings.TrimSpace(directory)
	if directory == "" {
		return nil, errors.New("conduit control: capability staging evidence directory is required")
	}
	return &CapabilityStagingEvidenceStore{directory: directory}, nil
}

// BuildCapabilityStagingEvidence creates a minimized immutable-evidence record
// only after the target is already in isolated validation with inherited
// authority, its compatibility bridge remains active, and all source-level
// isolated acceptance gates pass. It cannot authorize production cutover.
func BuildCapabilityStagingEvidence(snapshot CapabilityInventorySnapshot, capabilityID, sourceRevision, runtimeArtifactSHA256 string, acceptance IsolatedAcceptanceEvidence, now time.Time) (CapabilityStagingEvidence, error) {
	if err := validateCapabilityInventorySnapshot(snapshot); err != nil {
		return CapabilityStagingEvidence{}, err
	}
	capabilityID = strings.TrimSpace(capabilityID)
	if capabilityID == "" {
		return CapabilityStagingEvidence{}, errors.New("conduit control: capability staging evidence id is required")
	}
	if !validSourceRevision(sourceRevision) {
		return CapabilityStagingEvidence{}, errors.New("conduit control: capability staging evidence source revision is invalid")
	}
	runtimeArtifactSHA256 = strings.ToLower(strings.TrimSpace(runtimeArtifactSHA256))
	if !validSHA256(runtimeArtifactSHA256) {
		return CapabilityStagingEvidence{}, errors.New("conduit control: capability staging evidence runtime artifact digest is invalid")
	}
	if now.IsZero() {
		return CapabilityStagingEvidence{}, errors.New("conduit control: capability staging evidence time is required")
	}

	var target *CapabilityState
	for i := range snapshot.Inventory.Capabilities {
		if snapshot.Inventory.Capabilities[i].ID == capabilityID {
			target = &snapshot.Inventory.Capabilities[i]
			break
		}
	}
	if target == nil {
		return CapabilityStagingEvidence{}, errors.New("conduit control: capability staging evidence target is missing")
	}
	if target.MigrationStage != MigrationStageIsolatedValidation || target.Authority != AuthorityInherited || !target.CompatibilityBridgeActive || target.ProductionCutoverAuthorized {
		return CapabilityStagingEvidence{}, errors.New("conduit control: capability staging evidence requires isolated validation with inherited authority and active bridge")
	}
	decision, err := EvaluateIsolatedAcceptance(*target, acceptance)
	if err != nil {
		return CapabilityStagingEvidence{}, err
	}
	if !decision.EligibleForIsolatedValidation {
		return CapabilityStagingEvidence{}, errors.New("conduit control: capability staging evidence requires complete isolated acceptance")
	}

	return CapabilityStagingEvidence{
		Schema:                      CapabilityStagingEvidenceSchemaV1,
		CapabilityID:                capabilityID,
		RecordedAt:                  now.UTC().Format(time.RFC3339Nano),
		SourceRevision:              strings.ToLower(strings.TrimSpace(sourceRevision)),
		RuntimeArtifactSHA256:       runtimeArtifactSHA256,
		InventoryFingerprint:        snapshot.Fingerprint,
		AcceptanceSchema:            IsolatedAcceptanceSchemaV1,
		Authority:                   AuthorityInherited,
		CompatibilityBridgeActive:   true,
		ProductionCutoverAuthorized: false,
	}, nil
}

func (s *CapabilityStagingEvidenceStore) Save(evidence CapabilityStagingEvidence) (string, error) {
	if s == nil || strings.TrimSpace(s.directory) == "" {
		return "", errors.New("conduit control: capability staging evidence store is not initialized")
	}
	if supportErr := requireProtectedFileStoreSupport(); supportErr != nil {
		return "", supportErr
	}
	if err := validateCapabilityStagingEvidence(evidence); err != nil {
		return "", err
	}
	if err := os.MkdirAll(s.directory, 0o700); err != nil {
		return "", fmt.Errorf("conduit control: create capability staging evidence directory: %w", err)
	}
	if err := os.Chmod(s.directory, 0o700); err != nil {
		return "", fmt.Errorf("conduit control: protect capability staging evidence directory: %w", err)
	}
	identity := sha256.Sum256([]byte(evidence.CapabilityID + "\x00" + evidence.InventoryFingerprint + "\x00" + evidence.RuntimeArtifactSHA256))
	path := filepath.Join(s.directory, "staging-"+hex.EncodeToString(identity[:8])+".json")
	encoded, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		return "", fmt.Errorf("conduit control: encode capability staging evidence: %w", err)
	}
	encoded = append(encoded, '\n')
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return "", errors.New("conduit control: capability staging evidence already exists")
		}
		return "", fmt.Errorf("conduit control: create capability staging evidence: %w", err)
	}
	if _, err := file.Write(encoded); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("conduit control: write capability staging evidence: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("conduit control: sync capability staging evidence: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("conduit control: close capability staging evidence: %w", err)
	}
	return path, nil
}

func (s *CapabilityStagingEvidenceStore) Load(path string) (CapabilityStagingEvidence, error) {
	if s == nil || strings.TrimSpace(s.directory) == "" {
		return CapabilityStagingEvidence{}, errors.New("conduit control: capability staging evidence store is not initialized")
	}
	if supportErr := requireProtectedFileStoreSupport(); supportErr != nil {
		return CapabilityStagingEvidence{}, supportErr
	}
	cleanDirectory := filepath.Clean(s.directory)
	cleanPath := filepath.Clean(path)
	relative, err := filepath.Rel(cleanDirectory, cleanPath)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return CapabilityStagingEvidence{}, errors.New("conduit control: capability staging evidence path is outside the evidence store")
	}
	info, err := os.Lstat(cleanPath)
	if err != nil {
		return CapabilityStagingEvidence{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return CapabilityStagingEvidence{}, errors.New("conduit control: capability staging evidence file is not protected")
	}
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return CapabilityStagingEvidence{}, err
	}
	var evidence CapabilityStagingEvidence
	if err := json.Unmarshal(data, &evidence); err != nil {
		return CapabilityStagingEvidence{}, fmt.Errorf("conduit control: decode capability staging evidence: %w", err)
	}
	if err := validateCapabilityStagingEvidence(evidence); err != nil {
		return CapabilityStagingEvidence{}, err
	}
	return evidence, nil
}

func validateCapabilityStagingEvidence(evidence CapabilityStagingEvidence) error {
	if evidence.Schema != CapabilityStagingEvidenceSchemaV1 {
		return errors.New("conduit control: unsupported capability staging evidence schema")
	}
	if strings.TrimSpace(evidence.CapabilityID) == "" {
		return errors.New("conduit control: capability staging evidence id is required")
	}
	if _, err := time.Parse(time.RFC3339Nano, evidence.RecordedAt); err != nil {
		return errors.New("conduit control: capability staging evidence recorded_at is invalid")
	}
	if !validSourceRevision(evidence.SourceRevision) {
		return errors.New("conduit control: capability staging evidence source revision is invalid")
	}
	if !validSHA256(evidence.RuntimeArtifactSHA256) || !validSHA256(evidence.InventoryFingerprint) {
		return errors.New("conduit control: capability staging evidence fingerprint is invalid")
	}
	if evidence.AcceptanceSchema != IsolatedAcceptanceSchemaV1 {
		return errors.New("conduit control: capability staging evidence acceptance schema is invalid")
	}
	if evidence.Authority != AuthorityInherited || !evidence.CompatibilityBridgeActive || evidence.ProductionCutoverAuthorized {
		return errors.New("conduit control: capability staging evidence safety invariants are invalid")
	}
	return nil
}

func validSourceRevision(value string) bool {
	decoded, err := hex.DecodeString(strings.TrimSpace(value))
	return err == nil && (len(decoded) == 20 || len(decoded) == sha256.Size)
}

func validSHA256(value string) bool {
	decoded, err := hex.DecodeString(strings.TrimSpace(value))
	return err == nil && len(decoded) == sha256.Size
}
