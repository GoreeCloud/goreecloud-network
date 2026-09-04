package control

import (
	"errors"
	"strings"

	"github.com/netbirdio/netbird/native/conduit/obfuscation/paddedframe"
)

// ConduitPaddedWSSTransport is the first-party transport identity shared by
// capability advertisement, negotiation, and runtime confirmation. Keeping the
// identity sourced from the transport package prevents control-plane claims
// from drifting away from the actual relay implementation.
func ConduitPaddedWSSTransport() ObfuscationTransport {
	return ObfuscationTransport{
		ID:      paddedframe.TransportID,
		Version: paddedframe.Version,
	}
}

// FirstPartyObfuscationTransports returns the reviewed first-party transports
// this exact source revision can advertise. Callers receive a fresh slice so
// they cannot mutate global capability state.
func FirstPartyObfuscationTransports() []ObfuscationTransport {
	return []ObfuscationTransport{ConduitPaddedWSSTransport()}
}

// ConfirmObfuscationRuntimeProtocol promotes a negotiating status to active
// only when the actual established connection reports the first-party padded
// WSS protocol identity. Ordinary ws/quic/relay/WireGuard identifiers cannot
// pass this boundary and therefore cannot be reported as active obfuscation.
func ConfirmObfuscationRuntimeProtocol(status ObfuscationStatus, protocol string) (ObfuscationStatus, error) {
	if !strings.EqualFold(strings.TrimSpace(protocol), paddedframe.TransportID) {
		return ObfuscationStatus{}, errors.New("conduit control: runtime protocol is not an accepted first-party obfuscation transport")
	}
	return ConfirmObfuscationActive(status, ConduitPaddedWSSTransport())
}
