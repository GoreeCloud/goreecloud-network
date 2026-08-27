package control

import "testing"

func TestIsolatedAcceptanceFailsClosedWithMissingGates(t *testing.T) {
	capability := CapabilityState{ID: "control-api", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true}
	evidence := completeIsolatedEvidence()
	evidence.RollbackRehearsed = false
	evidence.BackupRestoreProven = false

	decision, err := EvaluateIsolatedAcceptance(capability, evidence)
	if err != nil {
		t.Fatal(err)
	}
	if decision.EligibleForIsolatedValidation {
		t.Fatal("incomplete evidence unexpectedly became isolated-validation eligible")
	}
	if len(decision.MissingGates) != 2 {
		t.Fatalf("missing gates = %v", decision.MissingGates)
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
		Schema:                    IsolatedAcceptanceSchemaV1,
		CapabilityID:              "control-api",
		ExactSourceRevision:       true,
		ImmutableRuntimeArtifact:  true,
		StateMigrationValidated:   true,
		BackupRestoreProven:       true,
		RollbackRehearsed:         true,
		ClientNetworkingValidated: true,
		SecurityPrivacyValidated:  true,
	}
}
