package android

import (
	"testing"

	"github.com/netbirdio/netbird/native/conduit/obfuscation/paddedframe"
	relayclient "github.com/netbirdio/netbird/shared/relay/client"
)

func TestConduitTransportBindingIdentity(t *testing.T) {
	if EnvKeyNBRelayTransport != relayclient.EnvRelayTransport {
		t.Fatalf("EnvKeyNBRelayTransport = %q, want %q", EnvKeyNBRelayTransport, relayclient.EnvRelayTransport)
	}
	if ConduitPaddedWSSTransportID != paddedframe.TransportID {
		t.Fatalf("ConduitPaddedWSSTransportID = %q, want %q", ConduitPaddedWSSTransportID, paddedframe.TransportID)
	}
	if ConduitPaddedWSSTransportVersion != paddedframe.Version {
		t.Fatalf("ConduitPaddedWSSTransportVersion = %q, want %q", ConduitPaddedWSSTransportVersion, paddedframe.Version)
	}
}
