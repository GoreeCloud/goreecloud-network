package android

import (
	"os"
	"testing"

	"github.com/netbirdio/netbird/native/conduit/control"
	obfuscationruntime "github.com/netbirdio/netbird/native/conduit/obfuscation/runtime"
)

func TestConduitObfuscationRuntimeStateBinding(t *testing.T) {
	old, hadOld := os.LookupEnv(EnvKeyNBRelayTransport)
	t.Cleanup(func() {
		obfuscationruntime.ClearIfInactive()
		if hadOld {
			_ = os.Setenv(EnvKeyNBRelayTransport, old)
		} else {
			_ = os.Unsetenv(EnvKeyNBRelayTransport)
		}
	})

	if err := os.Setenv(EnvKeyNBRelayTransport, ConduitPaddedWSSTransportID); err != nil {
		t.Fatal(err)
	}
	if err := obfuscationruntime.Begin(true); err != nil {
		t.Fatal(err)
	}
	if got := GetConduitObfuscationRuntimeState(); got != string(control.ObfuscationStateRequested) {
		t.Fatalf("state = %q, want requested", got)
	}

	if err := os.Setenv(EnvKeyNBRelayTransport, ""); err != nil {
		t.Fatal(err)
	}
	if got := GetConduitObfuscationRuntimeState(); got != "" {
		t.Fatalf("state with transport disabled = %q, want empty", got)
	}
}

func TestConduitObfuscationStateExports(t *testing.T) {
	want := map[string]string{
		"requested":   ConduitObfuscationStateRequested,
		"negotiating": ConduitObfuscationStateNegotiating,
		"active":      ConduitObfuscationStateActive,
		"unavailable": ConduitObfuscationStateUnavailable,
		"failed":      ConduitObfuscationStateFailed,
	}
	for expected, got := range want {
		if got != expected {
			t.Fatalf("exported state = %q, want %q", got, expected)
		}
	}
}
