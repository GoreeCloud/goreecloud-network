package runtime

import (
	"fmt"
	"strings"
	"sync"

	"github.com/netbirdio/netbird/native/conduit/control"
)

// Snapshot is the privacy-safe process-level status exposed to client surfaces.
// It deliberately contains no relay URL, peer identity, route, network, packet,
// DNS, credential, or timing detail.
type Snapshot struct {
	State            string
	Required         bool
	TransportID      string
	TransportVersion string
	Reason           string
}

type tracker struct {
	mu                sync.Mutex
	configured        bool
	request           control.ObfuscationRequest
	status            control.ObfuscationStatus
	activeConnections int
}

var processTracker tracker

func (t *tracker) ensureRequestLocked(required bool) error {
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

// Begin records that a first-party obfuscation connection is requested. An
// existing active connection remains authoritative while another connection is
// being attempted.
func Begin(required bool) error {
	t := &processTracker
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

// Negotiating records a successful connection to the dedicated transport
// endpoint. It does not report active until the relay application handshake is
// independently confirmed.
func Negotiating(protocol string) error {
	t := &processTracker
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.ensureRequestLocked(true); err != nil {
		return err
	}
	selected := control.ConduitPaddedWSSTransport()
	if !strings.EqualFold(strings.TrimSpace(protocol), selected.ID) {
		return fmt.Errorf("conduit obfuscation runtime: unexpected negotiating protocol %q", protocol)
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

// Active promotes the process state only after the actual runtime connection
// reports the reviewed transport identity and its relay handshake succeeds.
func Active(protocol string) error {
	t := &processTracker
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

// Unavailable records only an explicit absence of the dedicated reviewed
// transport endpoint. Generic network, TLS, proxy, and server failures use
// Failed instead.
func Unavailable() {
	t := &processTracker
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

// Failed records a privacy-safe transport failure without leaking endpoint or
// network details. An already-active connection remains authoritative while a
// secondary attempt fails.
func Failed() {
	t := &processTracker
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

// ReleaseActive releases exactly one runtime-confirmed connection. A failed
// last connection becomes failed; a graceful close returns to requested while
// the current process configuration still requires the transport.
func ReleaseActive(failed bool) {
	t := &processTracker
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

// ClearIfInactive removes stale process state when the running configuration no
// longer requests Conduit obfuscation. It never erases a live active connection.
func ClearIfInactive() {
	t := &processTracker
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.activeConnections == 0 {
		t.configured = false
		t.request = control.ObfuscationRequest{}
		t.status = control.ObfuscationStatus{}
	}
}

// Current returns only coarse status suitable for UI and privacy-safe
// diagnostics. An empty State means the process has no current Conduit
// obfuscation request.
func Current() Snapshot {
	t := &processTracker
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.configured {
		return Snapshot{}
	}

	snapshot := Snapshot{
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

func resetForTest() {
	t := &processTracker
	t.mu.Lock()
	defer t.mu.Unlock()
	t.configured = false
	t.request = control.ObfuscationRequest{}
	t.status = control.ObfuscationStatus{}
	t.activeConnections = 0
}
