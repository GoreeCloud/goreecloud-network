package android

import (
	"github.com/netbirdio/netbird/client/internal/lazyconn"
	"github.com/netbirdio/netbird/client/internal/peer"
	"github.com/netbirdio/netbird/native/conduit/obfuscation/paddedframe"
	relayclient "github.com/netbirdio/netbird/shared/relay/client"
)

var (
	// EnvKeyNBForceRelay Exported for Android java client to force relay connections.
	// Force Relay remains ordinary relay path selection and must not be represented
	// as GoreeCloud Obfuscation Mode.
	EnvKeyNBForceRelay = peer.EnvKeyNBForceRelay

	// EnvKeyNBRelayTransport exports the relay transport selector used by the
	// first-party Conduit obfuscation profile.
	EnvKeyNBRelayTransport = relayclient.EnvRelayTransport

	// ConduitPaddedWSSTransportID and ConduitPaddedWSSTransportVersion are sourced
	// from the transport implementation so the generated Android binding cannot
	// drift from the protocol identity used by negotiation/runtime confirmation.
	ConduitPaddedWSSTransportID      = paddedframe.TransportID
	ConduitPaddedWSSTransportVersion = paddedframe.Version

	// EnvKeyNBLazyConn Exported for Android java client to configure lazy connection
	EnvKeyNBLazyConn = lazyconn.EnvLazyConn

	// EnvKeyNBInactivityThreshold Exported for Android java client to configure connection inactivity threshold
	EnvKeyNBInactivityThreshold = lazyconn.EnvInactivityThreshold
)

// EnvList wraps a Go map for export to Java
type EnvList struct {
	data map[string]string
}

// NewEnvList creates a new EnvList
func NewEnvList() *EnvList {
	return &EnvList{data: make(map[string]string)}
}

// Put adds a key-value pair
func (el *EnvList) Put(key, value string) {
	el.data[key] = value
}

// Get retrieves a value by key
func (el *EnvList) Get(key string) string {
	return el.data[key]
}

func (el *EnvList) AllItems() map[string]string {
	return el.data
}
