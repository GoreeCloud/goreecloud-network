package control

import "testing"

func TestIsolatedAcceptanceFailsClosedWithMissingGates(t *testing.T) {
	capability := CapabilityState{ID: "control-api", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true}
	evidence := completeIsolatedEvidence()
	evidence.RollbackRehearsed = false
	evidence.BackupRestoreProven = false
	evidence.MeshCoordinationValidated = false
	evidence.IdentityIntegrationValidated = false
	evidence.GovernanceIntegrationValidated = false

	decision, err := EvaluateIsolatedAcceptance(capability, evidence)
	if err != nil {
		t.Fatal(err)
	}
	if decision.EligibleForIsolatedValidation {
		t.Fatal("incomplete evidence unexpectedly became isolated-validation eligible")
	}
	want := []string{
		"backup_restore_proven",
		"rollback_rehearsed",
		"mesh_coordination_validated",
		"identity_integration_validated",
		"governance_integration_validated",
	}
	if len(decision.MissingGates) != len(want) {
		t.Fatalf("missing gates = %v, want %v", decision.MissingGates, want)
	}
	for i := range want {
		if decision.MissingGates[i] != want[i] {
			t.Fatalf("missing gates = %v, want %v", decision.MissingGates, want)
		}
	}
	if decision.ProductionCutoverAuthorized {
		t.Fatal("source acceptance decision authorized production cutover")
	}
}

func TestIsolatedAcceptanceCanBecomeEligibleWithoutCutoverAuthority(t *testing.T) {
	capability := CapabilityState{ID: "control-api", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true}
	decision, err := EvaluateIsolatedAcceptance(capability, completeIsolatedEvidence())
	if err != nil {
		t.Fatal(err)
	}
	if !decision.EligibleForIsolatedValidation {
		t.Fatalf("complete evidence not eligible: %v", decision.MissingGates)
	}
	if decision.ProductionCutoverAuthorized {
		t.Fatal("isolated validation eligibility authorized production cutover")
	}
}

func TestIsolatedAcceptanceRejectsCapabilityMismatch(t *testing.T) {
	capability := CapabilityState{ID: "control-api", MigrationStage: "implementation", Authority: AuthorityInherited}
	evidence := completeIsolatedEvidence()
	evidence.CapabilityID = "routing"
	if _, err := EvaluateIsolatedAcceptance(capability, evidence); err == nil {
		t.Fatal("mismatched capability evidence unexpectedly accepted")
	}
}

func TestIsolatedAcceptanceRejectsNativeAuthority(t *testing.T) {
	capability := CapabilityState{ID: "control-api", MigrationStage: "native-accepted", Authority: AuthorityNative}
	if _, err := EvaluateIsolatedAcceptance(capability, completeIsolatedEvidence()); err == nil {
		t.Fatal("native authority unexpectedly passed pre-native isolated acceptance")
	}
}

func completeIsolatedEvidence() IsolatedAcceptanceEvidence {
	return IsolatedAcceptanceEvidence{
		Schema:                         IsolatedAcceptanceSchemaV1,
		CapabilityID:                   "control-api",
		ExactSourceRevision:            true,
		ImmutableRuntimeArtifact:       true,
		StateMigrationValidated:        true,
		BackupRestoreProven:            true,
		RollbackRehearsed:              true,
		ClientNetworkingValidated:      true,
		SecurityPrivacyValidated:       true,
		PrivacyShieldValidated:         true,
		WardveilSecurityValidated:      true,
		EverkeepValidated:              true,
		MeshCoordinationValidated:      true,
		IdentityIntegrationValidated:   true,
		GovernanceIntegrationValidated: true,
	}
}
