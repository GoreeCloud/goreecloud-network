package control

import (
	"strings"
	"testing"
)

func TestApplyPlatformAcceptanceEvidenceSetsRequiredGates(t *testing.T) {
	evidence := IsolatedAcceptanceEvidence{
		Schema:       IsolatedAcceptanceSchemaV1,
		CapabilityID: "conduit-control",
	}
	updated, err := ApplyPlatformAcceptanceEvidence(evidence, completePlatformAcceptanceEvidence())
	if err != nil {
		t.Fatal(err)
	}
	if !updated.PrivacyShieldValidated ||
		!updated.WardveilSecurityValidated ||
		!updated.EverkeepValidated ||
		!updated.MeshCoordinationValidated ||
		!updated.IdentityIntegrationValidated ||
		!updated.GovernanceIntegrationValidated {
		t.Fatalf("required platform gates were not applied: %+v", updated)
	}
}

func TestValidatePlatformAcceptanceEvidenceRequiresEveryPlatform(t *testing.T) {
	bundle := completePlatformAcceptanceEvidence()
	bundle.Everkeep.Accepted = false
	if err := ValidatePlatformAcceptanceEvidence(bundle); err == nil || !strings.Contains(err.Error(), PlatformEverkeep) {
		t.Fatalf("Everkeep rejection error = %v", err)
	}
}

func TestValidatePlatformAcceptanceEvidenceRejectsWrongPlatformIdentity(t *testing.T) {
	bundle := completePlatformAcceptanceEvidence()
	bundle.Mesh.Platform = PlatformIdentity
	if err := ValidatePlatformAcceptanceEvidence(bundle); err == nil || !strings.Contains(err.Error(), PlatformMesh) {
		t.Fatalf("Mesh identity mismatch error = %v", err)
	}
}

func TestValidatePlatformAcceptanceEvidenceRejectsInvalidDigest(t *testing.T) {
	bundle := completePlatformAcceptanceEvidence()
	bundle.PrivacyShield.EvidenceSHA256 = strings.Repeat("z", 64)
	if err := ValidatePlatformAcceptanceEvidence(bundle); err == nil || !strings.Contains(err.Error(), PlatformPrivacyShield) {
		t.Fatalf("invalid digest error = %v", err)
	}
}

func TestValidatePlatformAcceptanceEvidenceRejectsCutoverClaim(t *testing.T) {
	bundle := completePlatformAcceptanceEvidence()
	bundle.ProductionCutoverAuthorized = true
	if err := ValidatePlatformAcceptanceEvidence(bundle); err == nil {
		t.Fatal("platform evidence unexpectedly authorized production cutover")
	}
}

func TestValidatePlatformAcceptanceEvidenceRejectsWrongSourceIdentity(t *testing.T) {
	bundle := completePlatformAcceptanceEvidence()
	bundle.SourceRevision = strings.Repeat("a", 39)
	if err := ValidatePlatformAcceptanceEvidence(bundle); err == nil {
		t.Fatal("invalid source revision unexpectedly accepted")
	}
}

func completePlatformAcceptanceEvidence() PlatformAcceptanceEvidence {
	return PlatformAcceptanceEvidence{
		Schema:         PlatformAcceptanceEvidenceSchemaV1,
		SourceRevision: strings.Repeat("a", 40),
		PrivacyShield:  platformAcceptanceRecord(PlatformPrivacyShield, "1"),
		WardveilSecurity: platformAcceptanceRecord(PlatformWardveilSecurity, "2"),
		Everkeep:       platformAcceptanceRecord(PlatformEverkeep, "3"),
		Mesh:           platformAcceptanceRecord(PlatformMesh, "4"),
		Identity:       platformAcceptanceRecord(PlatformIdentity, "5"),
		Governance:     platformAcceptanceRecord(PlatformGovernance, "6"),
		ProductionCutoverAuthorized: false,
	}
}

func platformAcceptanceRecord(platform string, hexDigit string) PlatformAcceptanceRecord {
	return PlatformAcceptanceRecord{
		Platform:       platform,
		EvidenceSHA256: strings.Repeat(hexDigit, 64),
		Accepted:       true,
	}
}
