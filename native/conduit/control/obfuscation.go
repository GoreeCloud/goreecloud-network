package control

import (
	"errors"
	"strings"
)

const ObfuscationNegotiationSchemaV1 = "goreecloud-conduit-obfuscation-negotiation/v1"

// ObfuscationState reports only the coarse state needed by clients and
// operators. It intentionally does not expose peer identities, relay
// endpoints, packet data, network names, or other connection metadata.
type ObfuscationState string

const (
	ObfuscationStateRequested   ObfuscationState = "requested"
	ObfuscationStateNegotiating ObfuscationState = "negotiating"
	ObfuscationStateActive      ObfuscationState = "active"
	ObfuscationStateUnavailable ObfuscationState = "unavailable"
	ObfuscationStateFailed      ObfuscationState = "failed"
)

// ObfuscationReason is a privacy-safe machine-readable reason for an
// unavailable or failed negotiation.
type ObfuscationReason string

const (
	ObfuscationReasonNoCommonTransport ObfuscationReason = "no-common-transport"
	ObfuscationReasonTransportFailed   ObfuscationReason = "transport-failed"
)

// ObfuscationTransport identifies one reviewed transport/encapsulation
// implementation. IDs identify actual obfuscation transports; ordinary relay,
// direct-path selection, or WireGuard by itself are explicitly not transports
// for this contract.
type ObfuscationTransport struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// ObfuscationRequest contains the minimum information needed to negotiate an
// obfuscation transport. Required=true means the caller must not silently
// continue with an unobfuscated path when negotiation is unavailable or fails.
type ObfuscationRequest struct {
	Schema     string                 `json:"schema"`
	Required   bool                   `json:"required"`
	Transports []ObfuscationTransport `json:"transports"`
}

// ObfuscationStatus is the minimized negotiation result suitable for client
// state presentation and privacy-safe diagnostics.
type ObfuscationStatus struct {
	Schema           string            `json:"schema"`
	State            ObfuscationState  `json:"state"`
	Required         bool              `json:"required"`
	SelectedTransport *ObfuscationTransport `json:"selected_transport,omitempty"`
	Reason           ObfuscationReason `json:"reason,omitempty"`
}

var nonObfuscationTransportIDs = map[string]struct{}{
	"direct":      {},
	"force-relay": {},
	"p2p":         {},
	"relay":       {},
	"wireguard":   {},
}

func ValidateObfuscationTransport(transport ObfuscationTransport) error {
	id := strings.TrimSpace(transport.ID)
	version := strings.TrimSpace(transport.Version)
	if id == "" {
		return errors.New("conduit control: obfuscation transport id is required")
	}
	if version == "" {
		return errors.New("conduit control: obfuscation transport version is required")
	}
	if _, forbidden := nonObfuscationTransportIDs[strings.ToLower(id)]; forbidden {
		return errors.New("conduit control: ordinary path or tunnel primitive cannot be represented as an obfuscation transport")
	}
	return nil
}

func ValidateObfuscationRequest(request ObfuscationRequest) error {
	if request.Schema != ObfuscationNegotiationSchemaV1 {
		return errors.New("conduit control: unsupported obfuscation negotiation schema")
	}
	if len(request.Transports) == 0 {
		return errors.New("conduit control: at least one obfuscation transport is required")
	}
	seen := make(map[string]struct{}, len(request.Transports))
	for _, transport := range request.Transports {
		if err := ValidateObfuscationTransport(transport); err != nil {
			return err
		}
		key := strings.ToLower(strings.TrimSpace(transport.ID)) + "\x00" + strings.TrimSpace(transport.Version)
		if _, ok := seen[key]; ok {
			return errors.New("conduit control: duplicate obfuscation transport")
		}
		seen[key] = struct{}{}
	}
	return nil
}

func ValidateObfuscationStatus(status ObfuscationStatus) error {
	if status.Schema != ObfuscationNegotiationSchemaV1 {
		return errors.New("conduit control: unsupported obfuscation negotiation schema")
	}

	switch status.State {
	case ObfuscationStateRequested:
		if status.SelectedTransport != nil || status.Reason != "" {
			return errors.New("conduit control: requested obfuscation cannot include a selected transport or terminal reason")
		}
	case ObfuscationStateNegotiating, ObfuscationStateActive:
		if status.SelectedTransport == nil {
			return errors.New("conduit control: negotiating or active obfuscation requires a selected transport")
		}
		if err := ValidateObfuscationTransport(*status.SelectedTransport); err != nil {
			return err
		}
		if status.Reason != "" {
			return errors.New("conduit control: negotiating or active obfuscation cannot include a terminal reason")
		}
	case ObfuscationStateUnavailable:
		if status.SelectedTransport != nil || status.Reason != ObfuscationReasonNoCommonTransport {
			return errors.New("conduit control: unavailable obfuscation requires no-common-transport and no selected transport")
		}
	case ObfuscationStateFailed:
		if status.SelectedTransport != nil || status.Reason != ObfuscationReasonTransportFailed {
			return errors.New("conduit control: failed obfuscation requires transport-failed and no selected transport")
		}
	default:
		return errors.New("conduit control: invalid obfuscation state")
	}
	return nil
}

// NewObfuscationRequest creates a validated request and the corresponding
// requested state. The caller is still responsible for obtaining an endpoint
// capability advertisement before negotiation.
func NewObfuscationRequest(required bool, transports []ObfuscationTransport) (ObfuscationRequest, ObfuscationStatus, error) {
	request := ObfuscationRequest{
		Schema:     ObfuscationNegotiationSchemaV1,
		Required:   required,
		Transports: append([]ObfuscationTransport(nil), transports...),
	}
	if err := ValidateObfuscationRequest(request); err != nil {
		return ObfuscationRequest{}, ObfuscationStatus{}, err
	}
	status := ObfuscationStatus{
		Schema:   ObfuscationNegotiationSchemaV1,
		State:    ObfuscationStateRequested,
		Required: required,
	}
	return request, status, nil
}

// NegotiateObfuscation selects the first exact client-preference match from an
// endpoint's advertised reviewed transports. A successful capability match is
// only negotiating; it does not become active until the selected transport is
// actually established and ConfirmObfuscationActive is called.
func NegotiateObfuscation(request ObfuscationRequest, endpointTransports []ObfuscationTransport) (ObfuscationStatus, error) {
	if err := ValidateObfuscationRequest(request); err != nil {
		return ObfuscationStatus{}, err
	}
	for _, transport := range endpointTransports {
		if err := ValidateObfuscationTransport(transport); err != nil {
			return ObfuscationStatus{}, err
		}
	}

	for _, clientTransport := range request.Transports {
		for _, endpointTransport := range endpointTransports {
			if strings.EqualFold(strings.TrimSpace(clientTransport.ID), strings.TrimSpace(endpointTransport.ID)) && strings.TrimSpace(clientTransport.Version) == strings.TrimSpace(endpointTransport.Version) {
				selected := clientTransport
				status := ObfuscationStatus{
					Schema:            ObfuscationNegotiationSchemaV1,
					State:             ObfuscationStateNegotiating,
					Required:          request.Required,
					SelectedTransport: &selected,
				}
				return status, ValidateObfuscationStatus(status)
			}
		}
	}

	status := ObfuscationStatus{
		Schema:   ObfuscationNegotiationSchemaV1,
		State:    ObfuscationStateUnavailable,
		Required: request.Required,
		Reason:   ObfuscationReasonNoCommonTransport,
	}
	return status, ValidateObfuscationStatus(status)
}

// ConfirmObfuscationActive is the only helper in this contract that promotes a
// negotiation to active. The runtime-established transport must exactly match
// the negotiated selection.
func ConfirmObfuscationActive(status ObfuscationStatus, established ObfuscationTransport) (ObfuscationStatus, error) {
	if err := ValidateObfuscationStatus(status); err != nil {
		return ObfuscationStatus{}, err
	}
	if status.State != ObfuscationStateNegotiating || status.SelectedTransport == nil {
		return ObfuscationStatus{}, errors.New("conduit control: obfuscation can become active only from negotiating")
	}
	if err := ValidateObfuscationTransport(established); err != nil {
		return ObfuscationStatus{}, err
	}
	if !strings.EqualFold(strings.TrimSpace(status.SelectedTransport.ID), strings.TrimSpace(established.ID)) || strings.TrimSpace(status.SelectedTransport.Version) != strings.TrimSpace(established.Version) {
		return ObfuscationStatus{}, errors.New("conduit control: established obfuscation transport does not match negotiated selection")
	}
	active := status
	active.State = ObfuscationStateActive
	return active, ValidateObfuscationStatus(active)
}

// FailObfuscation returns a terminal privacy-safe failure state and removes any
// selected transport from the status rather than retaining path detail.
func FailObfuscation(status ObfuscationStatus) (ObfuscationStatus, error) {
	if err := ValidateObfuscationStatus(status); err != nil {
		return ObfuscationStatus{}, err
	}
	if status.State != ObfuscationStateNegotiating && status.State != ObfuscationStateActive {
		return ObfuscationStatus{}, errors.New("conduit control: only negotiating or active obfuscation can transition to failed")
	}
	failed := ObfuscationStatus{
		Schema:   ObfuscationNegotiationSchemaV1,
		State:    ObfuscationStateFailed,
		Required: status.Required,
		Reason:   ObfuscationReasonTransportFailed,
	}
	return failed, ValidateObfuscationStatus(failed)
}

// AllowsUnobfuscatedFallback makes downgrade policy explicit. Required mode
// never permits an unobfuscated fallback, including after unavailable or failed
// negotiation.
func AllowsUnobfuscatedFallback(status ObfuscationStatus) bool {
	if ValidateObfuscationStatus(status) != nil || status.Required {
		return false
	}
	return status.State == ObfuscationStateUnavailable || status.State == ObfuscationStateFailed
}
