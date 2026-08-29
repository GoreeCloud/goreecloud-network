package control

import (
	"testing"

	"github.com/netbirdio/netbird/native/conduit/obfuscation/paddedframe"
)

func TestFirstPartyObfuscationTransportsAdvertisePaddedWSS(t *testing.T) {
	transports := FirstPartyObfuscationTransports()
	if len(transports) != 1 {
		t.Fatalf("len(transports) = %d, want 1", len(transports))
	}
	want := ObfuscationTransport{ID: paddedframe.TransportID, Version: paddedframe.Version}
	if transports[0] != want {
		t.Fatalf("transport = %#v, want %#v", transports[0], want)
	}
	if err := ValidateObfuscationTransport(transports[0]); err != nil {
		t.Fatalf("ValidateObfuscationTransport() error = %v", err)
	}
}

func TestConfirmObfuscationRuntimeProtocolActivatesPaddedWSS(t *testing.T) {
	request, _, err := NewObfuscationRequest(true, FirstPartyObfuscationTransports())
	if err != nil {
		t.Fatalf("NewObfuscationRequest() error = %v", err)
	}
	status, err := NegotiateObfuscation(request, FirstPartyObfuscationTransports())
	if err != nil {
		t.Fatalf("NegotiateObfuscation() error = %v", err)
	}
	if status.State != ObfuscationStateNegotiating {
		t.Fatalf("state = %q, want %q", status.State, ObfuscationStateNegotiating)
	}

	active, err := ConfirmObfuscationRuntimeProtocol(status, paddedframe.TransportID)
	if err != nil {
		t.Fatalf("ConfirmObfuscationRuntimeProtocol() error = %v", err)
	}
	if active.State != ObfuscationStateActive {
		t.Fatalf("state = %q, want %q", active.State, ObfuscationStateActive)
	}
	if active.SelectedTransport == nil || *active.SelectedTransport != ConduitPaddedWSSTransport() {
		t.Fatalf("selected transport = %#v, want %#v", active.SelectedTransport, ConduitPaddedWSSTransport())
	}
}

func TestConfirmObfuscationRuntimeProtocolRejectsOrdinaryRelayProtocols(t *testing.T) {
	request, _, err := NewObfuscationRequest(true, FirstPartyObfuscationTransports())
	if err != nil {
		t.Fatalf("NewObfuscationRequest() error = %v", err)
	}
	status, err := NegotiateObfuscation(request, FirstPartyObfuscationTransports())
	if err != nil {
		t.Fatalf("NegotiateObfuscation() error = %v", err)
	}

	for _, protocol := range []string{"ws", "quic", "relay", "wireguard", "force-relay"} {
		t.Run(protocol, func(t *testing.T) {
			if _, err := ConfirmObfuscationRuntimeProtocol(status, protocol); err == nil {
				t.Fatalf("ConfirmObfuscationRuntimeProtocol(%q) unexpectedly succeeded", protocol)
			}
		})
	}
}

func TestFirstPartyObfuscationTransportsReturnsFreshSlice(t *testing.T) {
	first := FirstPartyObfuscationTransports()
	first[0].ID = "mutated"
	second := FirstPartyObfuscationTransports()
	if second[0].ID != paddedframe.TransportID {
		t.Fatalf("advertised transport was mutated: %q", second[0].ID)
	}
}
