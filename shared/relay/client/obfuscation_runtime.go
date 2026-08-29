package client

import (
	"sync"

	"github.com/netbirdio/netbird/native/conduit/control"
)

// ObfuscationRuntimeSnapshot is the privacy-safe process-level status exposed
// to mobile clients. It deliberately contains no relay URL, peer identity,
// route, network, packet, DNS, credential, or timing detail.
type ObfuscationRuntimeSnapshot struct {
	State            string
	Required         bool
	TransportID      string
	TransportVersion string
	Reason           string
}

type conduitObfuscationRuntimeTracker struct {
	mu                sync.Mutex
	configured        bool
	request           control.ObfuscationRequest
	status            control.ObfuscationStatus
	activeConnections int
}

var conduitObfuscationRuntime conduitObfuscationRuntimeTracker

func (t *conduitObfuscationRuntimeTracker) ensureRequestLocked(required bool) error {
	if t.configured && t.request.Required == required {
		return nil
	}
	request, status, err := control.NewObfuscationRequest(required, control.FirstPartyObfuscationTransports())
	if err != nil {
		return err
	}
	t.configured = true
	t.request = request
	if t.activeConnections == 0 {
		t.status = status
	}
	return nil
}

func beginConduitObfuscationAttempt(required bool) error {
	t := &conduitObfuscationRuntime
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.ensureRequestLocked(required); err != nil {
		return err
	}
	if t.activeConnections == 0 {
		_, requested, err := control.NewObfuscationRequest(required, control.FirstPartyObfuscationTransports())
		if err != nil {
			return err
		}
		t.status = requested
	}
	return nil
}

func markConduitObfuscationNegotiating(protocol string) error {
	t := &conduitObfuscationRuntime
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.ensureRequestLocked(true); err != nil {
		return err
	}
	selected := control.ConduitPaddedWSSTransport()
	if protocol != selected.ID {
		return control.ValidateObfuscationStatus(control.ObfuscationStatus{})
	}
	status, err := control.NegotiateObfuscation(t.request, []control.ObfuscationTransport{selected})
	if err != nil {
		return err
	}
	if t.activeConnections == 0 {
		t.status = status
	}
	return nil
}

func markConduitObfuscationActive(protocol string) error {
	t := &conduitObfuscationRuntime
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.ensureRequestLocked(true); err != nil {
		return err
	}
	status, err := control.NegotiateObfuscation(t.request, control.FirstPartyObfuscationTransports())
	if err != nil {
		return err
	}
	status, err = control.ConfirmObfuscationRuntimeProtocol(status, protocol)
	if err != nil {
		return err
	}
	t.activeConnections++
	t.status = status
	return nil
}

func markConduitObfuscationUnavailable() {
	t := &conduitObfuscationRuntime
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.ensureRequestLocked(true); err != nil || t.activeConnections > 0 {
		return
	}
	t.status = control.ObfuscationStatus{
		Schema:   control.ObfuscationNegotiationSchemaV1,
		State:    control.ObfuscationStateUnavailable,
		Required: true,
		Reason:   control.ObfuscationReasonNoCommonTransport,
	}
}

func markConduitObfuscationFailed() {
	t := &conduitObfuscationRuntime
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.ensureRequestLocked(true); err != nil || t.activeConnections > 0 {
		return
	}
	t.status = control.ObfuscationStatus{
		Schema:   control.ObfuscationNegotiationSchemaV1,
		State:    control.ObfuscationStateFailed,
		Required: true,
		Reason:   control.ObfuscationReasonTransportFailed,
	}
}

func releaseConduitObfuscationActiveConnection(failed bool) {
	t := &conduitObfuscationRuntime
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.activeConnections > 0 {
		t.activeConnections--
	}
	if t.activeConnections > 0 || !t.configured {
		return
	}
	if failed {
		t.status = control.ObfuscationStatus{
			Schema:   control.ObfuscationNegotiationSchemaV1,
			State:    control.ObfuscationStateFailed,
			Required: t.request.Required,
			Reason:   control.ObfuscationReasonTransportFailed,
		}
		return
	}
	_, requested, err := control.NewObfuscationRequest(t.request.Required, control.FirstPartyObfuscationTransports())
	if err == nil {
		t.status = requested
	}
}

func clearConduitObfuscationRuntimeIfInactive() {
	t := &conduitObfuscationRuntime
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.activeConnections == 0 {
		t.configured = false
		t.request = control.ObfuscationRequest{}
		t.status = control.ObfuscationStatus{}
	}
}

// ConduitObfuscationRuntimeSnapshot returns only coarse status suitable for UI
// and privacy-safe diagnostics. An empty State means the Conduit obfuscation
// transport is not currently requested by the running relay client.
func ConduitObfuscationRuntimeSnapshot() ObfuscationRuntimeSnapshot {
	t := &conduitObfuscationRuntime
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.configured {
		return ObfuscationRuntimeSnapshot{}
	}

	snapshot := ObfuscationRuntimeSnapshot{
		State:    string(t.status.State),
		Required: t.status.Required,
		Reason:   string(t.status.Reason),
	}
	if t.status.SelectedTransport != nil {
		snapshot.TransportID = t.status.SelectedTransport.ID
		snapshot.TransportVersion = t.status.SelectedTransport.Version
	}
	return snapshot
}

func resetConduitObfuscationRuntimeForTest() {
	t := &conduitObfuscationRuntime
	t.mu.Lock()
	defer t.mu.Unlock()
	t.configured = false
	t.request = control.ObfuscationRequest{}
	t.status = control.ObfuscationStatus{}
	t.activeConnections = 0
}
