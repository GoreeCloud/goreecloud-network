package controlplane

import "testing"

func TestEmptyStateIsDenyByDefault(t *testing.T) {
	state := NewState()
	overview := state.Overview(VolatileStorageStatus())
	if overview.PolicyMode != "deny_by_default" || overview.PolicyCount != 0 {
		t.Fatalf("unexpected overview: %+v", overview)
	}
	decision := state.EvaluateAccess(AccessRequest{PrincipalID: "principal-1", ResourceID: "resource-1"})
	if decision.Decision != "deny" || decision.ReasonCode != "NO_MATCHING_ALLOW_POLICY" {
		t.Fatalf("expected deny by default, got %+v", decision)
	}
}

func TestExplicitAllowPolicyCanPermitMatchingRequest(t *testing.T) {
	state := NewState()
	if err := state.PutAllowPolicy(AllowPolicy{ID: "policy-1", PrincipalID: "principal-1", DeviceID: "device-1", ResourceID: "resource-1"}); err != nil {
		t.Fatal(err)
	}
	allowed := state.EvaluateAccess(AccessRequest{PrincipalID: "principal-1", DeviceID: "device-1", ResourceID: "resource-1"})
	if allowed.Decision != "allow" || allowed.PolicyID != "policy-1" {
		t.Fatalf("expected explicit allow, got %+v", allowed)
	}
	denied := state.EvaluateAccess(AccessRequest{PrincipalID: "principal-1", DeviceID: "other-device", ResourceID: "resource-1"})
	if denied.Decision != "deny" {
		t.Fatalf("device-scoped policy must not allow a different device: %+v", denied)
	}
}

func TestInvalidContextFailsClosed(t *testing.T) {
	state := NewState()
	decision := state.EvaluateAccess(AccessRequest{PrincipalID: "principal-1"})
	if decision.Decision != "deny" || decision.ReasonCode != "INVALID_CONTEXT" {
		t.Fatalf("invalid context must fail closed: %+v", decision)
	}
}

func TestSnapshotRoundTripPreservesState(t *testing.T) {
	state := NewState()
	if err := state.PutDevice(Device{ID: "device-2", Name: "Tablet", Platform: "android", State: "approved"}); err != nil {
		t.Fatal(err)
	}
	if err := state.PutDevice(Device{ID: "device-1", Name: "Phone", Platform: "android", State: "approved"}); err != nil {
		t.Fatal(err)
	}
	if err := state.PutResource(Resource{ID: "resource-1", Name: "NAS", Type: "host"}); err != nil {
		t.Fatal(err)
	}
	if err := state.PutAllowPolicy(AllowPolicy{ID: "policy-1", PrincipalID: "principal-1", ResourceID: "resource-1"}); err != nil {
		t.Fatal(err)
	}

	snapshot := state.Snapshot()
	if len(snapshot.Devices) != 2 || snapshot.Devices[0].ID != "device-1" {
		t.Fatalf("snapshot should be stable and sorted: %+v", snapshot.Devices)
	}
	restored, err := NewStateFromSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if got := restored.Overview(VolatileStorageStatus()); got.DeviceCount != 2 || got.ResourceCount != 1 || got.PolicyCount != 1 {
		t.Fatalf("unexpected restored counts: %+v", got)
	}
}
