package status

import (
	"testing"
	"time"
)

func TestSnapshotFromEvidenceReadyStillRequiresAcceptance(t *testing.T) {
	snapshot := SnapshotFromEvidence(
		time.Date(2026, 9, 1, 17, 0, 0, 0, time.UTC),
		RuntimeEvidence{
			PrivateConnectivityReady: true,
			PeerCoordinationReady:     true,
			AccessPolicyReady:         true,
			NetworkDNSReady:           true,
		},
	)
	if snapshot.State != "ready" {
		t.Fatalf("expected ready state, got %q", snapshot.State)
	}
	if snapshot.Acceptance.ProductionApproved || !snapshot.Acceptance.RuntimeAcceptanceRequired {
		t.Fatal("runtime evidence must not bypass GoreeCloud acceptance")
	}
	for _, capability := range snapshot.Capabilities {
		if capability.State != "verified" {
			t.Fatalf("expected %s capability verified, got %q", capability.ID, capability.State)
		}
	}
}

func TestSnapshotFromEvidencePartial(t *testing.T) {
	snapshot := SnapshotFromEvidence(
		time.Date(2026, 9, 1, 17, 0, 0, 0, time.UTC),
		RuntimeEvidence{PeerCoordinationReady: true, AccessPolicyReady: true},
	)
	if snapshot.State != "partial" {
		t.Fatalf("expected partial state, got %q", snapshot.State)
	}
	if got := capabilityState(snapshot, "private-connectivity"); got != "attention" {
		t.Fatalf("expected private connectivity attention, got %q", got)
	}
}

func TestSnapshotFromEvidenceNoReadyCapabilitiesUnavailable(t *testing.T) {
	snapshot := SnapshotFromEvidence(
		time.Date(2026, 9, 1, 17, 0, 0, 0, time.UTC),
		RuntimeEvidence{},
	)
	if snapshot.State != "unavailable" {
		t.Fatalf("expected unavailable state, got %q", snapshot.State)
	}
	for _, capability := range snapshot.Capabilities {
		if capability.State != "attention" {
			t.Fatalf("expected %s attention, got %q", capability.ID, capability.State)
		}
	}
}

func capabilityState(snapshot Snapshot, id string) string {
	for _, capability := range snapshot.Capabilities {
		if capability.ID == id {
			return capability.State
		}
	}
	return ""
}
