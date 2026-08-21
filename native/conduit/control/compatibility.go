package control

import (
	"context"
	"errors"
)

// CompatibilityBridge converts an inherited provider into the first-party
// Conduit status contract without transferring state authority.
type CompatibilityBridge struct {
	Inherited Provider
}

func (b CompatibilityBridge) Status(ctx context.Context) (Status, error) {
	if b.Inherited == nil {
		return Status{}, errors.New("conduit control: inherited compatibility provider is required")
	}
	status, err := b.Inherited.Status(ctx)
	if err != nil {
		return Status{}, err
	}
	status.Schema = SchemaV1
	status.Authority = AuthorityInherited
	status.CompatibilityBridgeActive = true
	status.ProductionCutoverAuthorized = false
	if err := ValidateStatus(status); err != nil {
		return Status{}, err
	}
	return status, nil
}
