package ws

import (
	"net/http"
	"testing"

	"github.com/netbirdio/netbird/native/conduit/obfuscation/paddedframe"
)

func TestPrepareURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "rel scheme with non-standard port",
			input: "rel://test-domain-2:45678",
			want:  "ws://test-domain-2:45678/relay",
		},
		{
			name:  "rels scheme with non-standard port",
			input: "rels://test-domain-2:45678",
			want:  "wss://test-domain-2:45678/relay",
		},
		{
			name:  "rel scheme without port",
			input: "rel://test-domain-2",
			want:  "ws://test-domain-2/relay",
		},
		{
			name:  "rels scheme without port",
			input: "rels://test-domain-2",
			want:  "wss://test-domain-2/relay",
		},
		{
			name:  "rel scheme with IP and port",
			input: "rel://1.2.3.4:45678",
			want:  "ws://1.2.3.4:45678/relay",
		},
		{
			name:  "rel scheme with hostname starting with rel",
			input: "rel://relay.example.com:45678",
			want:  "ws://relay.example.com:45678/relay",
		},
		{
			name:  "rel scheme with IPv6 and port",
			input: "rel://[2001:db8::1]:45678",
			want:  "ws://[2001:db8::1]:45678/relay",
		},
		{
			name:  "rels scheme with IPv6 loopback and port",
			input: "rels://[::1]:45678",
			want:  "wss://[::1]:45678/relay",
		},
		{
			name:    "unsupported scheme",
			input:   "http://test-domain-2:45678",
			wantErr: true,
		},
		{
			name:    "no scheme",
			input:   "test-domain-2:45678",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prepareURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("prepareURL(%q) err = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("prepareURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPrepareConduitObfuscationURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "secure relay maps to dedicated padded endpoint",
			input: "rels://relay.example.com:443",
			want:  "wss://relay.example.com:443" + paddedframe.URLPath,
		},
		{
			name:    "plaintext relay is rejected",
			input:   "rel://relay.example.com:80",
			wantErr: true,
		},
		{
			name:    "ordinary https input is rejected",
			input:   "https://relay.example.com",
			wantErr: true,
		},
		{
			name:    "missing host is rejected",
			input:   "rels://",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prepareConduitObfuscationURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("prepareConduitObfuscationURL(%q) err = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("prepareConduitObfuscationURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConduitObfuscationDialerIdentity(t *testing.T) {
	d := NewConduitObfuscationDialer()
	if got := d.Protocol(); got != paddedframe.TransportID {
		t.Fatalf("Protocol() = %q, want %q", got, paddedframe.TransportID)
	}
}

func TestConduitTransportUnavailableResponse(t *testing.T) {
	tests := []struct {
		name string
		resp *http.Response
		want bool
	}{
		{name: "nil response is a transport failure", resp: nil, want: false},
		{name: "not found means transport is absent", resp: &http.Response{StatusCode: http.StatusNotFound}, want: true},
		{name: "method not allowed means transport is absent", resp: &http.Response{StatusCode: http.StatusMethodNotAllowed}, want: true},
		{name: "server error remains a transport failure", resp: &http.Response{StatusCode: http.StatusInternalServerError}, want: false},
		{name: "upgrade required remains a transport failure", resp: &http.Response{StatusCode: http.StatusUpgradeRequired}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := conduitTransportUnavailableResponse(tt.resp); got != tt.want {
				t.Fatalf("conduitTransportUnavailableResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
