package control

import "errors"

const CapabilityInventorySchemaV1 = "goreecloud-conduit-capability-inventory/v1"

// CapabilityState is a privacy-safe migration record for one first-party
// Conduit product capability. It contains no peer, route, policy, credential,
// device, packet, DNS-query, or user data.
type CapabilityState struct {
	ID                          string    `json:"id"`
	MigrationStage              string    `json:"migration_stage"`
	Authority                   Authority `json:"authority"`
	CompatibilityBridgeActive   bool      `json:"compatibility_bridge_active"`
	ProductionCutoverAuthorized bool      `json:"production_cutover_authorized"`
}

// CapabilityInventory is a first-party control-plane read model that lets
// Manager, Wardveil Security, Privacy Shield, and migration tooling inspect
// native replacement progress without reading inherited operational state.
type CapabilityInventory struct {
	Schema       string            `json:"schema"`
	Capabilities []CapabilityState `json:"capabilities"`
}

func ValidateCapabilityInventory(inventory CapabilityInventory) error {
	if inventory.Schema != CapabilityInventorySchemaV1 {
		return errors.New("conduit control: unsupported capability inventory schema")
	}
	seen := make(map[string]struct{}, len(inventory.Capabilities))
	for _, capability := range inventory.Capabilities {
		if capability.ID == "" {
			return errors.New("conduit control: capability id is required")
		}
		if _, ok := seen[capability.ID]; ok {
			return errors.New("conduit control: duplicate capability id")
		}
		seen[capability.ID] = struct{}{}
		if capability.MigrationStage == "" {
			return errors.New("conduit control: capability migration_stage is required")
		}
		switch capability.Authority {
		case AuthorityInherited, AuthorityTransitional, AuthorityNative:
		default:
			return errors.New("conduit control: invalid capability authority")
		}
		if capability.ProductionCutoverAuthorized {
			return errors.New("conduit control: source capability inventory cannot authorize production cutover")
		}
		if capability.Authority == AuthorityNative && capability.CompatibilityBridgeActive {
			return errors.New("conduit control: native capability cannot retain an inherited compatibility bridge")
		}
	}
	return nil
}

// CapabilitySummary provides aggregate migration evidence without exposing
// individual capability identifiers to higher-level status consumers.
type CapabilitySummary struct {
	Total        int `json:"total"`
	Inherited   int `json:"inherited"`
	Transitional int `json:"transitional"`
	Native      int `json:"native"`
	Bridged     int `json:"bridged"`
}

func SummarizeCapabilities(inventory CapabilityInventory) (CapabilitySummary, error) {
	if err := ValidateCapabilityInventory(inventory); err != nil {
		return CapabilitySummary{}, err
	}
	var summary CapabilitySummary
	for _, capability := range inventory.Capabilities {
		summary.Total++
		switch capability.Authority {
		case AuthorityInherited:
			summary.Inherited++
		case AuthorityTransitional:
			summary.Transitional++
		case AuthorityNative:
			summary.Native++
		}
		if capability.CompatibilityBridgeActive {
			summary.Bridged++
		}
	}
	return summary, nil
}
