package control

import (
	"errors"
	"strings"
)

// AdvanceInventoryCapabilityToIsolatedValidation applies an isolated-validation
// transition to one capability in the privacy-safe source inventory. It returns
// a copied inventory and never mutates inherited operational state, transfers
// authority, removes the compatibility bridge, or authorizes production cutover.
func AdvanceInventoryCapabilityToIsolatedValidation(inventory CapabilityInventory, capabilityID string, evidence IsolatedAcceptanceEvidence) (CapabilityInventory, error) {
	if err := ValidateCapabilityInventory(inventory); err != nil {
		return CapabilityInventory{}, err
	}
	capabilityID = strings.TrimSpace(capabilityID)
	if capabilityID == "" {
		return CapabilityInventory{}, errors.New("conduit control: capability id is required")
	}

	next := CapabilityInventory{
		Schema:       inventory.Schema,
		Capabilities: append([]CapabilityState(nil), inventory.Capabilities...),
	}
	found := false
	for i := range next.Capabilities {
		if next.Capabilities[i].ID != capabilityID {
			continue
		}
		advanced, err := AdvanceToIsolatedValidation(next.Capabilities[i], evidence)
		if err != nil {
			return CapabilityInventory{}, err
		}
		next.Capabilities[i] = advanced
		found = true
		break
	}
	if !found {
		return CapabilityInventory{}, errors.New("conduit control: capability not found in inventory")
	}
	if err := ValidateCapabilityInventory(next); err != nil {
		return CapabilityInventory{}, err
	}
	return next, nil
}
