// Package status defines the privacy-minimized GoreeCloud infrastructure status boundary.
package status

import "time"

// SchemaVersion is the current GoreeCloud Infrastructure Status contract version.
const SchemaVersion = 1

// Producer identifies the GoreeCloud service adapter and current runtime authority.
type Producer struct {
	ServiceID        string `json:"service_id"`
	AdapterID        string `json:"adapter_id"`
	RuntimeAuthority string `json:"runtime_authority"`
}

// Privacy declares the sensitive-data classes excluded from a status document.
type Privacy struct {
	ContainsCredentials         bool `json:"contains_credentials"`
	ContainsPersonalData        bool `json:"contains_personal_data"`
	ContainsRawLogs             bool `json:"contains_raw_logs"`
	ContainsNetworkIdentifiers  bool `json:"contains_network_identifiers"`
	ContainsQueryData           bool `json:"contains_query_data"`
	ContainsCertificateMaterial bool `json:"contains_certificate_material"`
}

// Acceptance preserves runtime-acceptance and production-approval boundaries.
type Acceptance struct {
	RuntimeAcceptanceRequired bool `json:"runtime_acceptance_required"`
	ProductionApproved        bool `json:"production_approved"`
}

// Capability reports one approved coarse-grained Network capability state.
type Capability struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

// Snapshot is a complete privacy-minimized Infrastructure Status v1 document.
type Snapshot struct {
	SchemaVersion int          `json:"schema_version"`
	Producer      Producer     `json:"producer"`
	GeneratedAt   string       `json:"generated_at"`
	State         string       `json:"state"`
	Privacy       Privacy      `json:"privacy"`
	Acceptance    Acceptance   `json:"acceptance"`
	Capabilities  []Capability `json:"capabilities"`
}

// DevelopmentSnapshot is conservative until GoreeCloud Network runtime state is
// supplied by an accepted local adapter rather than inferred by Manager.
func DevelopmentSnapshot(now time.Time) Snapshot {
	return Snapshot{
		SchemaVersion: SchemaVersion,
		Producer: Producer{
			ServiceID:        "goreecloud-network",
			AdapterID:        "goreecloud-network/status-v1",
			RuntimeAuthority: "GoreeCloud/NetBirdDataPlane",
		},
		GeneratedAt: now.UTC().Format(time.RFC3339),
		State:       "development",
		Privacy:     Privacy{},
		Acceptance: Acceptance{
			RuntimeAcceptanceRequired: true,
			ProductionApproved:        false,
		},
		Capabilities: []Capability{
			{ID: "private-connectivity", State: "pending"},
			{ID: "peer-coordination", State: "pending"},
			{ID: "access-policy", State: "pending"},
			{ID: "network-dns", State: "pending"},
		},
	}
}
