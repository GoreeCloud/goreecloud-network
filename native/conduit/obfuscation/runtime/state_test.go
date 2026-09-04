package runtime

import (
	"testing"

	"github.com/netbirdio/netbird/native/conduit/control"
)

func TestLifecycle(t *testing.T) {
	resetForTest()
	t.Cleanup(resetForTest)

	if got := Current().State; got != "" {
		t.Fatalf("initial state = %q, want empty", got)
	}
	if err := Begin(true); err != nil {
		t.Fatalf("begin: %v", err)
	}
	if got := Current().State; got != string(control.ObfuscationStateRequested) {
		t.Fatalf("requested state = %q", got)
	}

	transport := control.ConduitPaddedWSSTransport()
	if err := Negotiating(transport.ID); err != nil {
		t.Fatalf("negotiating: %v", err)
	}
	negotiating := Current()
	if negotiating.State != string(control.ObfuscationStateNegotiating) {
		t.Fatalf("negotiating state = %q", negotiating.State)
	}
	if negotiating.TransportID != transport.ID || negotiating.TransportVersion != transport.Version {
		t.Fatalf("negotiating transport = %q/%q, want %q/%q", negotiating.TransportID, negotiating.TransportVersion, transport.ID, transport.Version)
	}

	if err := Active(transport.ID); err != nil {
		t.Fatalf("active: %v", err)
	}
	if got := Current().State; got != string(control.ObfuscationStateActive) {
		t.Fatalf("active state = %q", got)
	}

	// A secondary failed attempt must not downgrade an already active process.
	if err := Begin(true); err != nil {
		t.Fatalf("second begin: %v", err)
	}
	Failed()
	if got := Current().State; got != string(control.ObfuscationStateActive) {
		t.Fatalf("state after secondary failure = %q, want active", got)
	}

	ReleaseActive(true)
	failed := Current()
	if failed.State != string(control.ObfuscationStateFailed) || failed.Reason != string(control.ObfuscationReasonTransportFailed) {
		t.Fatalf("failed = state %q reason %q", failed.State, failed.Reason)
	}

	if err := Begin(true); err != nil {
		t.Fatalf("retry begin: %v", err)
	}
	if got := Current().State; got != string(control.ObfuscationStateRequested) {
		t.Fatalf("retry state = %q, want requested", got)
	}
}

func TestUnavailableAndClear(t *testing.T) {
	resetForTest()
	t.Cleanup(resetForTest)

	if err := Begin(true); err != nil {
		t.Fatalf("begin: %v", err)
	}
	Unavailable()
	unavailable := Current()
	if unavailable.State != string(control.ObfuscationStateUnavailable) || unavailable.Reason != string(control.ObfuscationReasonNoCommonTransport) {
		t.Fatalf("unavailable = state %q reason %q", unavailable.State, unavailable.Reason)
	}

	ClearIfInactive()
	if got := Current().State; got != "" {
		t.Fatalf("cleared state = %q, want empty", got)
	}
}

func TestUnexpectedProtocolRejected(t *testing.T) {
	resetForTest()
	t.Cleanup(resetForTest)

	if err := Begin(true); err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := Negotiating("ws"); err == nil {
		t.Fatal("Negotiating(ws) unexpectedly succeeded")
	}
	if err := Active("quic"); err == nil {
		t.Fatal("Active(quic) unexpectedly succeeded")
	}
	if got := Current().State; got == string(control.ObfuscationStateActive) {
		t.Fatal("ordinary relay protocol became active obfuscation")
	}
}
