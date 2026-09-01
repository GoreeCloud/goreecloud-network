package status

import "time"

const (
	stateReady       = "ready"
	statePartial     = "partial"
	stateUnavailable = "unavailable"

	capabilityVerified  = "verified"
	capabilityAttention = "attention"
)

// RuntimeEvidence contains only coarse service health. It must never be
// populated with peers, routes, users, tokens, setup keys, IP addresses,
// public keys, ACL rules, DNS labels, activity history, or raw logs.
type RuntimeEvidence struct {
	PrivateConnectivityReady bool
	PeerCoordinationReady     bool
	AccessPolicyReady         bool
	NetworkDNSReady           bool
}

// SnapshotFromEvidence converts bounded NetBird-derived runtime evidence into
// the GoreeCloud-owned status contract. Runtime evidence never grants
// production approval by itself.
func SnapshotFromEvidence(now time.Time, evidence RuntimeEvidence) Snapshot {
	snapshot := DevelopmentSnapshot(now)
	readyCount := 0
	for _, ready := range []bool{
		evidence.PrivateConnectivityReady,
		evidence.PeerCoordinationReady,
		evidence.AccessPolicyReady,
		evidence.NetworkDNSReady,
	} {
		if ready {
			readyCount++
		}
	}

	switch readyCount {
	case 0:
		snapshot.State = stateUnavailable
	case 4:
		snapshot.State = stateReady
	default:
		snapshot.State = statePartial
	}

	snapshot.Capabilities = []Capability{
		{ID: "private-connectivity", State: evidenceState(evidence.PrivateConnectivityReady)},
		{ID: "peer-coordination", State: evidenceState(evidence.PeerCoordinationReady)},
		{ID: "access-policy", State: evidenceState(evidence.AccessPolicyReady)},
		{ID: "network-dns", State: evidenceState(evidence.NetworkDNSReady)},
	}
	return snapshot
}

func evidenceState(ready bool) string {
	if ready {
		return capabilityVerified
	}
	return capabilityAttention
}
