package control

import "testing"

func TestCapabilityInventoryRejectsUnsafeAuthorityClaims(t *testing.T) {
	t.Parallel()

	base := CapabilityInventory{
		Schema: CapabilityInventorySchemaV1,
		Capabilities: []CapabilityState{{
			ID:                        "control",
			MigrationStage:            "implementation",
			Authority:                 AuthorityInherited,
			CompatibilityBridgeActive: true,
		}},
	}
	if err := ValidateCapabilityInventory(base); err != nil {
		t.Fatalf("valid inventory rejected: %v", err)
	}

	cutover := base
	cutover.Capabilities = append([]CapabilityState(nil), base.Capabilities...)
	cutover.Capabilities[0].ProductionCutoverAuthorized = true
	if err := ValidateCapabilityInventory(cutover); err == nil {
		t.Fatal("source inventory authorized production cutover")
	}

	nativeWithBridge := base
	nativeWithBridge.Capabilities = append([]CapabilityState(nil), base.Capabilities...)
	nativeWithBridge.Capabilities[0].Authority = AuthorityNative
	if err := ValidateCapabilityInventory(nativeWithBridge); err == nil {
		t.Fatal("native authority retained inherited compatibility bridge")
	}
}

func TestCapabilityInventoryRejectsDuplicateIDs(t *testing.T) {
	t.Parallel()

	inventory := CapabilityInventory{
		Schema: CapabilityInventorySchemaV1,
		Capabilities: []CapabilityState{
			{ID: "api", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
			{ID: "api", MigrationStage: "contract", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
		},
	}
	if err := ValidateCapabilityInventory(inventory); err == nil {
		t.Fatal("duplicate capability IDs were accepted")
	}
}

func TestSummarizeCapabilitiesIsAggregateOnly(t *testing.T) {
	t.Parallel()

	inventory := CapabilityInventory{
		Schema: CapabilityInventorySchemaV1,
		Capabilities: []CapabilityState{
			{ID: "control", MigrationStage: "implementation", Authority: AuthorityInherited, CompatibilityBridgeActive: true},
			{ID: "api", MigrationStage: "implementation", Authority: AuthorityTransitional, CompatibilityBridgeActive: true},
			{ID: "diagnostics", MigrationStage: "implementation", Authority: AuthorityNative},
		},
	}
	summary, err := SummarizeCapabilities(inventory)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Total != 3 || summary.Inherited != 1 || summary.Transitional != 1 || summary.Native != 1 || summary.Bridged != 2 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}
