package status

import (
	"testing"
	"time"
)

func TestDevelopmentSnapshotDoesNotExposeNetworkIdentity(t *testing.T) {
	snapshot := DevelopmentSnapshot(time.Date(2026, 9, 1, 16, 0, 0, 0, time.UTC))
	if snapshot.SchemaVersion != 1 || snapshot.Producer.ServiceID != "goreecloud-network" {
		t.Fatal("unexpected GoreeCloud Network contract identity")
	}
	if snapshot.State != "development" || snapshot.Acceptance.ProductionApproved {
		t.Fatal("development adapter must not claim production readiness")
	}
	if !snapshot.Acceptance.RuntimeAcceptanceRequired {
		t.Fatal("runtime acceptance must remain explicit")
	}
	if snapshot.Privacy.ContainsCredentials || snapshot.Privacy.ContainsPersonalData || snapshot.Privacy.ContainsRawLogs || snapshot.Privacy.ContainsNetworkIdentifiers || snapshot.Privacy.ContainsQueryData || snapshot.Privacy.ContainsCertificateMaterial {
		t.Fatal("network status must exclude peer, route, credential, and other sensitive data")
	}
}
