package ws

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConduitPaddedEndpointRejectsPlaintextListener(t *testing.T) {
	l := &Listener{}
	req := httptest.NewRequest(http.MethodGet, "http://relay.example/conduit/obfuscation/v1", nil)
	recorder := httptest.NewRecorder()

	l.onAcceptPadded(recorder, req)

	if recorder.Code != http.StatusUpgradeRequired {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUpgradeRequired)
	}
}

func TestConduitPaddedEndpointRejectsRequestWithoutTLSState(t *testing.T) {
	l := &Listener{TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
	req := httptest.NewRequest(http.MethodGet, "https://relay.example/conduit/obfuscation/v1", nil)
	// httptest.NewRequest does not populate TLS merely because the URL uses
	// https. This locks the requirement to the actual accepted TLS connection.
	req.TLS = nil
	recorder := httptest.NewRecorder()

	l.onAcceptPadded(recorder, req)

	if recorder.Code != http.StatusUpgradeRequired {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUpgradeRequired)
	}
}
