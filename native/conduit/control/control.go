// Package control defines the first-party GoreeCloud Conduit control-plane
// contracts used during the controlled fork-to-native transition.
package control

import (
	"context"
	"errors"
	"time"
)

const SchemaV1 = "goreecloud-conduit-control-status/v1"

// Authority describes which implementation currently owns authoritative state.
type Authority string

const (
	AuthorityInherited    Authority = "inherited"
	AuthorityTransitional Authority = "transitional"
	AuthorityNative       Authority = "native"
)

// Status is the privacy-safe, read-only Conduit control-plane status contract.
// It intentionally contains no peer inventory, policy contents, routes,
// credentials, tokens, keys, packet data, DNS queries, or raw diagnostics.
type Status struct {
	Schema                      string    `json:"schema"`
	GeneratedAt                 time.Time `json:"generated_at"`
	Authority                   Authority `json:"authority"`
	MigrationStage              string    `json:"migration_stage"`
	CompatibilityBridgeActive   bool      `json:"compatibility_bridge_active"`
	ProductionCutoverAuthorized bool      `json:"production_cutover_authorized"`
}

// Provider supplies the current underlying control-plane state needed to build
// the Conduit status contract. During the migration phase this may be backed by
// inherited implementation code; callers must not infer native authority from
// use of this interface.
type Provider interface {
	Status(context.Context) (Status, error)
}

// ValidateStatus enforces fail-closed invariants for the first native contract.
func ValidateStatus(status Status) error {
	if status.Schema != SchemaV1 {
		return errors.New("conduit control: unsupported status schema")
	}
	if status.GeneratedAt.IsZero() {
		return errors.New("conduit control: generated_at is required")
	}
	switch status.Authority {
	case AuthorityInherited, AuthorityTransitional, AuthorityNative:
	default:
		return errors.New("conduit control: invalid authority")
	}
	if status.MigrationStage == "" {
		return errors.New("conduit control: migration_stage is required")
	}
	if status.ProductionCutoverAuthorized {
		return errors.New("conduit control: source contract cannot authorize production cutover")
	}
	if status.Authority == AuthorityNative && status.CompatibilityBridgeActive {
		return errors.New("conduit control: native authority cannot retain an active inherited bridge")
	}
	return nil
}
