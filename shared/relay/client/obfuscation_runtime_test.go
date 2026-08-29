package client

import (
	"testing"

	"github.com/netbirdio/netbird/native/conduit/control"
)

func TestConduitObfuscationRuntimeLifecycle(t *testing.T) {
	resetConduitObfuscationRuntimeForTest()
	t.Cleanup(resetConduitObfuscationRuntimeForTest)

	if got := ConduitObfuscationRuntimeSnapshot().State; got != "" {
		t.Fatalf("initial state = %q, want empty", got)
	}
	if err := beginConduitObfuscationAttempt(true); err != nil {
		t.Fatalf("begin attempt: %v", err)
	}
	if got := ConduitObfuscationRuntimeSnapshot().State; got != string(control.ObfuscationStateRequested) {
		t.Fatalf("requested state = %q", got)
	}

	transport := control.ConduitPaddedWSSTransport()
	if err := markConduitObfuscationNegotiating(transport.ID); err != nil {
		t.Fatalf("mark negotiating: %v", err)
	}
	negotiating := ConduitObfuscationRuntimeSnapshot()
	if negotiating.State != string(control.ObfuscationStateNegotiating) {
		t.Fatalf("negotiating state = %q", negotiating.State)
	}
	if negotiating.TransportID != transport.ID || negotiating.TransportVersion != transport.Version {
		t.Fatalf("negotiating transport = %q/%q, want %q/%q", negotiating.TransportID, negotiating.TransportVersion, transport.ID, transport.Version)
	}

	if err := markConduitObfuscationActive(transport.ID); err != nil {
		t.Fatalf("mark active: %v", err)
	}
	active := ConduitObfuscationRuntimeSnapshot()
	if active.State != string(control.ObfuscationStateActive) {
		t.Fatalf("active state = %q", active.State)
	}

	// A second relay attempt must not downgrade an already active process-level
	// state while at least one padded connection is still alive.
	if err := beginConduitObfuscationAttempt(true); err != nil {
		t.Fatalf("begin second attempt: %v", err)
	}
	markConduitObfuscationFailed()
	if got := ConduitObfuscationRuntimeSnapshot().State; got != string(control.ObfuscationStateActive) {
		t.Fatalf("state after secondary failure = %q, want active", got)
	}

	releaseConduitObfuscationActiveConnection(true)
	failed := ConduitObfuscationRuntimeSnapshot()
	if failed.State != string(control.ObfuscationStateFailed) || failed.Reason != string(control.ObfuscationReasonTransportFailed) {
		t.Fatalf("terminal failure = state %q reason %q", failed.State, failed.Reason)
	}

	if err := beginConduitObfuscationAttempt(true); err != nil {
		t.Fatalf("retry attempt: %v", err)
	}
	if got := ConduitObfuscationRuntimeSnapshot().State; got != string(control.ObfuscationStateRequested) {
		t.Fatalf("retry state = %q, want requested", got)
	}
}

func TestConduitObfuscationRuntimeUnavailableAndClear(t *testing.T) {
	resetConduitObfuscationRuntimeForTest()
	t.Cleanup(resetConduitObfuscationRuntimeForTest)

	if err := beginConduitObfuscationAttempt(true); err != nil {
		t.Fatalf("begin attempt: %v", err)
	}
	markConduitObfuscationUnavailable()
	unavailable := ConduitObfuscationRuntimeSnapshot()
	if unavailable.State != string(control.ObfuscationStateUnavailable) || unavailable.Reason != string(control.ObfuscationReasonNoCommonTransport) {
		t.Fatalf("unavailable = state %q reason %q", unavailable.State, unavailable.Reason)
	}

	clearConduitObfuscationRuntimeIfInactive()
	if got := ConduitObfuscationRuntimeSnapshot().State; got != "" {
		t.Fatalf("cleared state = %q, want empty", got)
	}
}
