package control

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestBuildCapabilityStagingEvidenceBindsIsolatedCapability(t *testing.T) {
	_, after := testTransitionSnapshots(t)
	acceptance := completeStagingAcceptance("control-api")
	record, err := BuildCapabilityStagingEvidence(
		after,
		"control-api",
		"7490de58f8b6c8ed1f80bdc2cb6ef20623e823e6",
		strings.Repeat("a", 64),
		acceptance,
		time.Date(2026, 8, 27, 16, 45, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if record.CapabilityID != "control-api" || record.InventoryFingerprint != after.Fingerprint {
		t.Fatalf("unexpected staging evidence binding: %+v", record)
	}
	if record.Authority != AuthorityInherited || !record.CompatibilityBridgeActive || record.ProductionCutoverAuthorized {
		t.Fatal("staging evidence violated authority, bridge, or cutover invariants")
	}
}

func TestBuildCapabilityStagingEvidenceRejectsIncompleteAcceptance(t *testing.T) {
	_, after := testTransitionSnapshots(t)
	acceptance := completeStagingAcceptance("control-api")
	acceptance.SecurityPrivacyValidated = false
	if _, err := BuildCapabilityStagingEvidence(
		after,
		"control-api",
		"7490de58f8b6c8ed1f80bdc2cb6ef20623e823e6",
		strings.Repeat("b", 64),
		acceptance,
		time.Now().UTC(),
	); err == nil {
		t.Fatal("incomplete staging acceptance unexpectedly produced evidence")
	}
}

func TestCapabilityStagingEvidenceStoreIsImmutableAndOwnerOnly(t *testing.T) {
	requireProtectedFileStoreTestSupport(t)
	_, after := testTransitionSnapshots(t)
	record, err := BuildCapabilityStagingEvidence(
		after,
		"control-api",
		"7490de58f8b6c8ed1f80bdc2cb6ef20623e823e6",
		strings.Repeat("c", 64),
		completeStagingAcceptance("control-api"),
		time.Date(2026, 8, 27, 16, 46, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewCapabilityStagingEvidenceStore(filepath.Join(t.TempDir(), "staging"))
	if err != nil {
		t.Fatal(err)
	}
	path, err := store.Save(record)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(record); err == nil {
		t.Fatal("immutable staging evidence was unexpectedly replaceable")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("staging evidence mode=%#o", info.Mode().Perm())
	}
	loaded, err := store.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded, record) {
		t.Fatalf("loaded staging evidence mismatch: %+v", loaded)
	}
}

func completeStagingAcceptance(capabilityID string) IsolatedAcceptanceEvidence {
	return IsolatedAcceptanceEvidence{
		Schema:                         IsolatedAcceptanceSchemaV1,
		CapabilityID:                   capabilityID,
		ExactSourceRevision:            true,
		ImmutableRuntimeArtifact:       true,
		StateMigrationValidated:        true,
		BackupRestoreProven:            true,
		RollbackRehearsed:              true,
		ClientNetworkingValidated:      true,
		SecurityPrivacyValidated:       true,
		ManagerIntegrationValidated:    true,
		PrivacyShieldValidated:         true,
		WardveilSecurityValidated:      true,
		EverkeepValidated:              true,
		GlazeUIStableValidated:         true,
		MeshCoordinationValidated:      true,
		IdentityIntegrationValidated:   true,
		GovernanceIntegrationValidated: true,
	}
}
