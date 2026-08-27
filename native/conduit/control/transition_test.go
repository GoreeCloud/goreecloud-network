package control

import "testing"

func completeIsolatedEvidence(id string) IsolatedAcceptanceEvidence {
	return IsolatedAcceptanceEvidence{
		Schema:                    IsolatedAcceptanceSchemaV1,
		CapabilityID:              id,
		ExactSourceRevision:       true,
		ImmutableRuntimeArtifact:  true,
		StateMigrationValidated:   true,
		BackupRestoreProven:       true,
		RollbackRehearsed:         true,
		ClientNetworkingValidated: true,
		SecurityPrivacyValidated:  true,
	}
}

func TestAdvanceToIsolatedValidationPreservesAuthorityAndBridge(t *testing.T) {
	capability := CapabilityState{
		ID:                        "control-api",
		MigrationStage:            "implementation",
		Authority:                 AuthorityInherited,
		CompatibilityBridgeActive: true,
	}
	next, err := AdvanceToIsolatedValidation(capability, completeIsolatedEvidence(capability.ID))
	if err != nil {
		t.Fatal(err)
	}
	if next.MigrationStage != MigrationStageIsolatedValidation {
		t.Fatalf("stage=%q", next.MigrationStage)
	}
	if next.Authority != AuthorityInherited || !next.CompatibilityBridgeActive {
		t.Fatal("isolated validation unexpectedly transferred authority or removed bridge")
	}
	if next.ProductionCutoverAuthorized {
		t.Fatal("isolated validation unexpectedly authorized production cutover")
	}
}

func TestAdvanceToIsolatedValidationRejectsIncompleteEvidence(t *testing.T) {
	capability := CapabilityState{ID: "control-api", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true}
	evidence := completeIsolatedEvidence(capability.ID)
	evidence.RollbackRehearsed = false
	if _, err := AdvanceToIsolatedValidation(capability, evidence); err == nil {
		t.Fatal("incomplete evidence unexpectedly advanced migration stage")
	}
}

func TestAdvanceToIsolatedValidationRejectsWrongStartingStage(t *testing.T) {
	capability := CapabilityState{ID: "control-api", MigrationStage: "contract", Authority: AuthorityInherited, CompatibilityBridgeActive: true}
	if _, err := AdvanceToIsolatedValidation(capability, completeIsolatedEvidence(capability.ID)); err == nil {
		t.Fatal("non-implementation stage unexpectedly advanced")
	}
}
