package control

import (
	"errors"
	"strings"
)

const IsolatedAcceptanceSchemaV1 = "goreecloud-conduit-isolated-acceptance/v1"

// IsolatedAcceptanceEvidence records only bounded pass/fail evidence needed to
// decide whether a Conduit capability may advance into isolated validation.
// It intentionally carries no peers, routes, policies, credentials, devices,
// packet data, DNS queries, or user information.
type IsolatedAcceptanceEvidence struct {
	Schema                         string `json:"schema"`
	CapabilityID                   string `json:"capability_id"`
	ExactSourceRevision            bool   `json:"exact_source_revision"`
	ImmutableRuntimeArtifact       bool   `json:"immutable_runtime_artifact"`
	StateMigrationValidated        bool   `json:"state_migration_validated"`
	BackupRestoreProven            bool   `json:"backup_restore_proven"`
	RollbackRehearsed              bool   `json:"rollback_rehearsed"`
	ClientNetworkingValidated      bool   `json:"client_networking_validated"`
	SecurityPrivacyValidated       bool   `json:"security_privacy_validated"`
	PrivacyShieldValidated         bool   `json:"privacy_shield_validated"`
	WardveilSecurityValidated      bool   `json:"wardveil_security_validated"`
	EverkeepValidated              bool   `json:"everkeep_validated"`
	MeshCoordinationValidated      bool   `json:"mesh_coordination_validated"`
	IdentityIntegrationValidated   bool   `json:"identity_integration_validated"`
	GovernanceIntegrationValidated bool   `json:"governance_integration_validated"`
}

// IsolatedAcceptanceDecision is a source-level readiness result. It cannot
// authorize production cutover or transfer authoritative state.
type IsolatedAcceptanceDecision struct {
	EligibleForIsolatedValidation bool     `json:"eligible_for_isolated_validation"`
	MissingGates                  []string `json:"missing_gates,omitempty"`
	ProductionCutoverAuthorized   bool     `json:"production_cutover_authorized"`
}

func EvaluateIsolatedAcceptance(capability CapabilityState, evidence IsolatedAcceptanceEvidence) (IsolatedAcceptanceDecision, error) {
	decision := IsolatedAcceptanceDecision{ProductionCutoverAuthorized: false}
	if strings.TrimSpace(capability.ID) == "" {
		return decision, errors.New("conduit control: capability id is required")
	}
	if capability.ProductionCutoverAuthorized {
		return decision, errors.New("conduit control: source capability state cannot authorize production cutover")
	}
	if capability.Authority == AuthorityNative {
		return decision, errors.New("conduit control: isolated acceptance must precede native authority")
	}
	if evidence.Schema != IsolatedAcceptanceSchemaV1 {
		return decision, errors.New("conduit control: unsupported isolated acceptance schema")
	}
	if evidence.CapabilityID != capability.ID {
		return decision, errors.New("conduit control: isolated acceptance evidence capability mismatch")
	}

	gates := []struct {
		name string
		pass bool
	}{
		{"exact_source_revision", evidence.ExactSourceRevision},
		{"immutable_runtime_artifact", evidence.ImmutableRuntimeArtifact},
		{"state_migration_validated", evidence.StateMigrationValidated},
		{"backup_restore_proven", evidence.BackupRestoreProven},
		{"rollback_rehearsed", evidence.RollbackRehearsed},
		{"client_networking_validated", evidence.ClientNetworkingValidated},
		{"security_privacy_validated", evidence.SecurityPrivacyValidated},
		{"privacy_shield_validated", evidence.PrivacyShieldValidated},
		{"wardveil_security_validated", evidence.WardveilSecurityValidated},
		{"everkeep_validated", evidence.EverkeepValidated},
		{"mesh_coordination_validated", evidence.MeshCoordinationValidated},
		{"identity_integration_validated", evidence.IdentityIntegrationValidated},
		{"governance_integration_validated", evidence.GovernanceIntegrationValidated},
	}
	for _, gate := range gates {
		if !gate.pass {
			decision.MissingGates = append(decision.MissingGates, gate.name)
		}
	}
	decision.EligibleForIsolatedValidation = len(decision.MissingGates) == 0
	return decision, nil
}
