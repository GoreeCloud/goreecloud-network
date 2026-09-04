package android

import (
	"os"

	"github.com/netbirdio/netbird/native/conduit/control"
	"github.com/netbirdio/netbird/native/conduit/obfuscation/paddedframe"
	obfuscationruntime "github.com/netbirdio/netbird/native/conduit/obfuscation/runtime"
)

var (
	// Export the canonical state names through gomobile so Java never has to
	// duplicate the wire/runtime contract strings.
	ConduitObfuscationStateRequested   = string(control.ObfuscationStateRequested)
	ConduitObfuscationStateNegotiating = string(control.ObfuscationStateNegotiating)
	ConduitObfuscationStateActive      = string(control.ObfuscationStateActive)
	ConduitObfuscationStateUnavailable = string(control.ObfuscationStateUnavailable)
	ConduitObfuscationStateFailed      = string(control.ObfuscationStateFailed)
)

// GetConduitObfuscationRuntimeState returns only the coarse privacy-safe state
// of the currently requested Conduit transport. If the running process no
// longer requests conduit-padded-wss, stale inactive state is cleared and no
// obfuscation state is exposed.
func GetConduitObfuscationRuntimeState() string {
	if os.Getenv(EnvKeyNBRelayTransport) != paddedframe.TransportID {
		obfuscationruntime.ClearIfInactive()
		return ""
	}
	return obfuscationruntime.Current().State
}
