package control

import "testing"

func TestAdvanceInventoryCapabilityToIsolatedValidationUpdatesOnlyTarget(t *testing.T) {
	inventory := CapabilityInventory{
		Schema: CapabilityInventorySchemaV1,
		Capabilities: []CapabilityState{
			{ID: "control-api", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
			{ID: "relay", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
		},
	}
	original := append([]CapabilityState(nil), inventory.Capabilities...)
	next, err := AdvanceInventoryCapabilityToIsolatedValidation(inventory, "control-api", completeTransitionEvidence("control-api"))
	if err != nil {
		t.Fatal(err)
	}
	if next.Capabilities[0].MigrationStage != MigrationStageIsolatedValidation {
		t.Fatalf("target stage=%q", next.Capabilities[0].MigrationStage)
	}
	if next.Capabilities[0].Authority != AuthorityInherited || !next.Capabilities[0].CompatibilityBridgeActive || next.Capabilities[0].ProductionCutoverAuthorized {
		t.Fatal("inventory transition changed target authority, bridge, or cutover authorization")
	}
	if next.Capabilities[1] != original[1] {
		t.Fatal("inventory transition changed a non-target capability")
	}
	if inventory.Capabilities[0] != original[0] {
		t.Fatal("inventory transition mutated the source inventory")
	}
}

func TestAdvanceInventoryCapabilityToIsolatedValidationRejectsMissingCapability(t *testing.T) {
	inventory := CapabilityInventory{
		Schema: CapabilityInventorySchemaV1,
		Capabilities: []CapabilityState{
			{ID: "control-api", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
		},
	}
	if _, err := AdvanceInventoryCapabilityToIsolatedValidation(inventory, "missing", completeTransitionEvidence("missing")); err == nil {
		t.Fatal("missing capability unexpectedly advanced")
	}
}

func TestAdvanceInventoryCapabilityToIsolatedValidationFailsClosedOnEvidence(t *testing.T) {
	inventory := CapabilityInventory{
		Schema: CapabilityInventorySchemaV1,
		Capabilities: []CapabilityState{
			{ID: "control-api", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
		},
	}
	evidence := completeTransitionEvidence("control-api")
	evidence.SecurityPrivacyValidated = false
	if _, err := AdvanceInventoryCapabilityToIsolatedValidation(inventory, "control-api", evidence); err == nil {
		t.Fatal("incomplete isolated acceptance evidence unexpectedly advanced inventory")
	}
}
