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
	if !updated.ManagerIntegrationValidated ||
		!updated.PrivacyShieldValidated ||
		!updated.WardveilSecurityValidated ||
		!updated.EverkeepValidated ||
		!updated.GlazeUIStableValidated ||
		!updated.MeshCoordinationValidated ||
		!updated.IdentityIntegrationValidated ||
		!updated.GovernanceIntegrationValidated {
		t.Fatalf("required platform gates were not applied: %+v", updated)
	}
}

func TestValidatePlatformAcceptanceEvidenceRequiresEveryPlatform(t *testing.T) {
	bundle := completePlatformAcceptanceEvidence()
	bundle.Manager.Accepted = false
	if err := ValidatePlatformAcceptanceEvidence(bundle); err == nil || !strings.Contains(err.Error(), PlatformManager) {
		t.Fatalf("Manager rejection error = %v", err)
	}

	bundle = completePlatformAcceptanceEvidence()
	bundle.GlazeUI.Accepted = false
	if err := ValidatePlatformAcceptanceEvidence(bundle); err == nil || !strings.Contains(err.Error(), PlatformGlazeUI) {
		t.Fatalf("Glaze UI rejection error = %v", err)
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
		Manager:        platformAcceptanceRecord(PlatformManager, "1"),
		PrivacyShield:  platformAcceptanceRecord(PlatformPrivacyShield, "2"),
		WardveilSecurity: platformAcceptanceRecord(PlatformWardveilSecurity, "3"),
		Everkeep:       platformAcceptanceRecord(PlatformEverkeep, "4"),
		GlazeUI:        platformAcceptanceRecord(PlatformGlazeUI, "5"),
		Mesh:           platformAcceptanceRecord(PlatformMesh, "6"),
		Identity:       platformAcceptanceRecord(PlatformIdentity, "7"),
		Governance:     platformAcceptanceRecord(PlatformGovernance, "8"),
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
