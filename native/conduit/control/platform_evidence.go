package control

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const PlatformAcceptanceEvidenceSchemaV1 = "goreecloud-conduit-platform-acceptance-evidence/v1"

const (
	PlatformPrivacyShield    = "privacy-shield"
	PlatformWardveilSecurity = "wardveil-security"
	PlatformEverkeep         = "everkeep"
	PlatformMesh             = "goreecloud-mesh"
	PlatformIdentity         = "goreecloud-identity"
	PlatformGovernance       = "goreecloud-governance"
)

// PlatformAcceptanceRecord is a deliberately minimized reference to evidence
// produced by another GoreeCloud platform system. The Conduit control plane
// consumes only the platform identity, immutable evidence digest, and accepted
// result; raw platform state does not cross this boundary.
type PlatformAcceptanceRecord struct {
	Platform       string `json:"platform"`
	EvidenceSHA256 string `json:"evidence_sha256"`
	Accepted       bool   `json:"accepted"`
}

// PlatformAcceptanceEvidence binds the required platform-system evidence to
// the exact Conduit source revision being evaluated. It intentionally contains
// no peers, routes, policies, credentials, device identifiers, packet data,
// DNS queries, user information, or unrestricted diagnostics.
type PlatformAcceptanceEvidence struct {
	Schema                       string                   `json:"schema"`
	SourceRevision               string                   `json:"source_revision"`
	PrivacyShield                PlatformAcceptanceRecord `json:"privacy_shield"`
	WardveilSecurity             PlatformAcceptanceRecord `json:"wardveil_security"`
	Everkeep                     PlatformAcceptanceRecord `json:"everkeep"`
	Mesh                         PlatformAcceptanceRecord `json:"mesh"`
	Identity                     PlatformAcceptanceRecord `json:"identity"`
	Governance                   PlatformAcceptanceRecord `json:"governance"`
	ProductionCutoverAuthorized  bool                     `json:"production_cutover_authorized"`
}

// ApplyPlatformAcceptanceEvidence validates a complete minimized platform
// evidence bundle and reflects its accepted results into the existing isolated
// acceptance contract. This function cannot authorize production cutover.
func ApplyPlatformAcceptanceEvidence(
	evidence IsolatedAcceptanceEvidence,
	bundle PlatformAcceptanceEvidence,
) (IsolatedAcceptanceEvidence, error) {
	if err := ValidatePlatformAcceptanceEvidence(bundle); err != nil {
		return evidence, err
	}

	evidence.PrivacyShieldValidated = true
	evidence.WardveilSecurityValidated = true
	evidence.EverkeepValidated = true
	evidence.MeshCoordinationValidated = true
	evidence.IdentityIntegrationValidated = true
	evidence.GovernanceIntegrationValidated = true
	return evidence, nil
}

// ValidatePlatformAcceptanceEvidence fails closed unless every required
// platform record is explicitly accepted and bound to an immutable SHA-256
// evidence identity. A valid source-level bundle still never authorizes a
// production cutover.
func ValidatePlatformAcceptanceEvidence(bundle PlatformAcceptanceEvidence) error {
	if bundle.Schema != PlatformAcceptanceEvidenceSchemaV1 {
		return errors.New("conduit control: unsupported platform acceptance evidence schema")
	}
	if !validPlatformEvidenceHex(bundle.SourceRevision, 40) {
		return errors.New("conduit control: platform acceptance source revision is invalid")
	}
	if bundle.ProductionCutoverAuthorized {
		return errors.New("conduit control: platform acceptance evidence cannot authorize production cutover")
	}

	records := []struct {
		expected string
		record   PlatformAcceptanceRecord
	}{
		{PlatformPrivacyShield, bundle.PrivacyShield},
		{PlatformWardveilSecurity, bundle.WardveilSecurity},
		{PlatformEverkeep, bundle.Everkeep},
		{PlatformMesh, bundle.Mesh},
		{PlatformIdentity, bundle.Identity},
		{PlatformGovernance, bundle.Governance},
	}
	for _, item := range records {
		if strings.TrimSpace(item.record.Platform) != item.expected {
			return fmt.Errorf("conduit control: platform acceptance record mismatch for %s", item.expected)
		}
		if !item.record.Accepted {
			return fmt.Errorf("conduit control: platform acceptance evidence is not accepted for %s", item.expected)
		}
		if !validPlatformEvidenceHex(item.record.EvidenceSHA256, sha256.Size*2) {
			return fmt.Errorf("conduit control: platform acceptance evidence digest is invalid for %s", item.expected)
		}
	}
	return nil
}

func validPlatformEvidenceHex(value string, length int) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) != length {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
