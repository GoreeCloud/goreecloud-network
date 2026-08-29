package control

import "testing"

func TestObfuscationNegotiationRequiresRuntimeConfirmation(t *testing.T) {
	client := []ObfuscationTransport{
		{ID: "reviewed-transport", Version: "1"},
		{ID: "alternate-transport", Version: "1"},
	}
	request, requested, err := NewObfuscationRequest(true, client)
	if err != nil {
		t.Fatalf("NewObfuscationRequest() error = %v", err)
	}
	if requested.State != ObfuscationStateRequested {
		t.Fatalf("requested state = %q, want %q", requested.State, ObfuscationStateRequested)
	}

	status, err := NegotiateObfuscation(request, []ObfuscationTransport{{ID: "reviewed-transport", Version: "1"}})
	if err != nil {
		t.Fatalf("NegotiateObfuscation() error = %v", err)
	}
	if status.State != ObfuscationStateNegotiating {
		t.Fatalf("negotiated state = %q, want %q", status.State, ObfuscationStateNegotiating)
	}
	if status.SelectedTransport == nil || status.SelectedTransport.ID != "reviewed-transport" {
		t.Fatalf("selected transport = %#v", status.SelectedTransport)
	}

	active, err := ConfirmObfuscationActive(status, ObfuscationTransport{ID: "reviewed-transport", Version: "1"})
	if err != nil {
		t.Fatalf("ConfirmObfuscationActive() error = %v", err)
	}
	if active.State != ObfuscationStateActive {
		t.Fatalf("active state = %q, want %q", active.State, ObfuscationStateActive)
	}
}

func TestRequiredObfuscationNeverAllowsSilentFallback(t *testing.T) {
	request, _, err := NewObfuscationRequest(true, []ObfuscationTransport{{ID: "reviewed-transport", Version: "1"}})
	if err != nil {
		t.Fatalf("NewObfuscationRequest() error = %v", err)
	}
	status, err := NegotiateObfuscation(request, []ObfuscationTransport{{ID: "different-transport", Version: "1"}})
	if err != nil {
		t.Fatalf("NegotiateObfuscation() error = %v", err)
	}
	if status.State != ObfuscationStateUnavailable {
		t.Fatalf("state = %q, want %q", status.State, ObfuscationStateUnavailable)
	}
	if status.Reason != ObfuscationReasonNoCommonTransport {
		t.Fatalf("reason = %q, want %q", status.Reason, ObfuscationReasonNoCommonTransport)
	}
	if AllowsUnobfuscatedFallback(status) {
		t.Fatal("required obfuscation unexpectedly allowed unobfuscated fallback")
	}
}

func TestPreferredObfuscationFallbackIsExplicit(t *testing.T) {
	request, _, err := NewObfuscationRequest(false, []ObfuscationTransport{{ID: "reviewed-transport", Version: "1"}})
	if err != nil {
		t.Fatalf("NewObfuscationRequest() error = %v", err)
	}
	status, err := NegotiateObfuscation(request, nil)
	if err != nil {
		t.Fatalf("NegotiateObfuscation() error = %v", err)
	}
	if !AllowsUnobfuscatedFallback(status) {
		t.Fatal("preferred obfuscation should allow an explicit fallback after unavailable negotiation")
	}
}

func TestOrdinaryRelayAndWireGuardCannotMasqueradeAsObfuscation(t *testing.T) {
	for _, id := range []string{"force-relay", "relay", "direct", "p2p", "wireguard"} {
		t.Run(id, func(t *testing.T) {
			_, _, err := NewObfuscationRequest(true, []ObfuscationTransport{{ID: id, Version: "1"}})
			if err == nil {
				t.Fatalf("NewObfuscationRequest(%q) succeeded; ordinary path/tunnel primitive must not be obfuscation", id)
			}
		})
	}
}

func TestObfuscationActivationRejectsDifferentRuntimeTransport(t *testing.T) {
	request, _, err := NewObfuscationRequest(true, []ObfuscationTransport{{ID: "reviewed-transport", Version: "1"}})
	if err != nil {
		t.Fatalf("NewObfuscationRequest() error = %v", err)
	}
	status, err := NegotiateObfuscation(request, []ObfuscationTransport{{ID: "reviewed-transport", Version: "1"}})
	if err != nil {
		t.Fatalf("NegotiateObfuscation() error = %v", err)
	}
	if _, err := ConfirmObfuscationActive(status, ObfuscationTransport{ID: "different-transport", Version: "1"}); err == nil {
		t.Fatal("ConfirmObfuscationActive() succeeded with a different runtime transport")
	}
}

func TestFailedObfuscationRedactsSelectedTransportAndPreservesRequiredPolicy(t *testing.T) {
	request, _, err := NewObfuscationRequest(true, []ObfuscationTransport{{ID: "reviewed-transport", Version: "1"}})
	if err != nil {
		t.Fatalf("NewObfuscationRequest() error = %v", err)
	}
	status, err := NegotiateObfuscation(request, []ObfuscationTransport{{ID: "reviewed-transport", Version: "1"}})
	if err != nil {
		t.Fatalf("NegotiateObfuscation() error = %v", err)
	}
	failed, err := FailObfuscation(status)
	if err != nil {
		t.Fatalf("FailObfuscation() error = %v", err)
	}
	if failed.State != ObfuscationStateFailed || failed.Reason != ObfuscationReasonTransportFailed {
		t.Fatalf("failed status = %#v", failed)
	}
	if failed.SelectedTransport != nil {
		t.Fatalf("failed status retained selected transport: %#v", failed.SelectedTransport)
	}
	if AllowsUnobfuscatedFallback(failed) {
		t.Fatal("required failed obfuscation unexpectedly allowed fallback")
	}
}

func TestActiveObfuscationCannotBeClaimedWithoutSelectedTransport(t *testing.T) {
	status := ObfuscationStatus{
		Schema:   ObfuscationNegotiationSchemaV1,
		State:    ObfuscationStateActive,
		Required: true,
	}
	if err := ValidateObfuscationStatus(status); err == nil {
		t.Fatal("ValidateObfuscationStatus() accepted active without a selected transport")
	}
}
