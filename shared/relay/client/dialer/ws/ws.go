package ws

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/coder/websocket"
	log "github.com/sirupsen/logrus"

	nbnet "github.com/netbirdio/netbird/client/net"
	"github.com/netbirdio/netbird/native/conduit/obfuscation/paddedframe"
	obfuscationruntime "github.com/netbirdio/netbird/native/conduit/obfuscation/runtime"
	"github.com/netbirdio/netbird/shared/relay"
	relaydialer "github.com/netbirdio/netbird/shared/relay/client/dialer"
	"github.com/netbirdio/netbird/util/embeddedroots"
)

type Dialer struct {
	urlPath  string
	protocol string
	padded   bool
}

// NewConduitObfuscationDialer returns the client-side transport for the
// first-party Conduit padded-WSS profile. It is not part of the ordinary relay
// dialer race and therefore cannot be selected accidentally as a fallback.
func NewConduitObfuscationDialer() Dialer {
	return Dialer{
		urlPath:  paddedframe.URLPath,
		protocol: paddedframe.TransportID,
		padded:   true,
	}
}

func (d Dialer) Protocol() string {
	if d.protocol != "" {
		return d.protocol
	}
	return Network
}

func (d Dialer) Dial(ctx context.Context, address, serverName string) (net.Conn, error) {
	if d.padded {
		if err := obfuscationruntime.Begin(true); err != nil {
			return nil, err
		}
	}

	var (
		wsURL string
		err   error
	)
	if d.padded {
		wsURL, err = prepareConduitObfuscationURL(address)
	} else {
		path := d.urlPath
		if path == "" {
			path = relay.WebSocketURLPath
		}
		wsURL, err = prepareURLPath(address, path)
	}
	if err != nil {
		if d.padded {
			obfuscationruntime.Failed()
		}
		return nil, err
	}

	var underlying net.Conn
	opts := createDialOptions(serverName, &underlying)

	wsConn, resp, err := websocket.Dial(ctx, wsURL, opts)
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		if d.padded && conduitTransportUnavailableResponse(resp) {
			obfuscationruntime.Unavailable()
			return nil, fmt.Errorf("%w: Conduit padded WSS endpoint is not offered", relaydialer.ErrTransportUnavailable)
		}
		if errors.Is(err, context.Canceled) {
			return nil, err
		}
		if d.padded {
			obfuscationruntime.Failed()
		}
		// websocket.Dial wraps the cause in verbose layers; surface the
		// underlying network error when present.
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			return nil, opErr
		}
		return nil, err
	}
	if resp.Body != nil {
		_ = resp.Body.Close()
	}

	if d.padded {
		if err := obfuscationruntime.Negotiating(d.Protocol()); err != nil {
			obfuscationruntime.Failed()
			_ = wsConn.Close(websocket.StatusPolicyViolation, "invalid Conduit transport identity")
			return nil, err
		}
		return NewPaddedConn(wsConn, address, underlying), nil
	}
	return NewConn(wsConn, address, underlying), nil
}

// conduitTransportUnavailableResponse recognizes only explicit HTTP responses
// that mean the dedicated Conduit transport endpoint is absent. Network, TLS,
// proxy, and generic server failures remain transport failures rather than being
// mislabeled as a capability absence.
func conduitTransportUnavailableResponse(resp *http.Response) bool {
	if resp == nil {
		return false
	}
	return resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed
}

// prepareURL rewrites a rel://host[:port] or rels://host[:port] address into a
// ws://host[:port]/relay or wss://host[:port]/relay URL, preserving any
// non-standard port from the input.
func prepareURL(address string) (string, error) {
	return prepareURLPath(address, relay.WebSocketURLPath)
}

// prepareConduitObfuscationURL builds the dedicated Conduit padded transport
// endpoint and refuses plaintext rel:// inputs. The padded profile is traffic
// shaping, not encryption, so it must only operate inside authenticated WSS.
func prepareConduitObfuscationURL(address string) (string, error) {
	parsed, err := url.Parse(address)
	if err != nil {
		return "", fmt.Errorf("parse relay address %q: %w", address, err)
	}
	if parsed.Scheme != "rels" {
		return "", fmt.Errorf("conduit padded transport requires rels/WSS, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("missing host in relay address %q", address)
	}
	parsed.Scheme = "wss"
	parsed.Path = paddedframe.URLPath
	return parsed.String(), nil
}

func prepareURLPath(address, path string) (string, error) {
	parsed, err := url.Parse(address)
	if err != nil {
		return "", fmt.Errorf("parse relay address %q: %w", address, err)
	}
	switch parsed.Scheme {
	case "rel":
		parsed.Scheme = "ws"
	case "rels":
		parsed.Scheme = "wss"
	default:
		return "", fmt.Errorf("unsupported scheme: %s", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("missing host in relay address %q", address)
	}
	if path == "" || path[0] != '/' {
		return "", fmt.Errorf("invalid websocket path %q", path)
	}
	parsed.Path = path
	return parsed.String(), nil
}

// httpClientNbDialer builds the http client used by the websocket library.
// underlyingOut, when non-nil, is populated with the raw conn from the
// transport's DialContext so the caller can read its RemoteAddr.
func httpClientNbDialer(serverName string, underlyingOut *net.Conn) *http.Client {
	customDialer := nbnet.NewDialer()

	certPool, err := x509.SystemCertPool()
	if err != nil || certPool == nil {
		log.Debugf("System cert pool not available; falling back to embedded cert, error: %v", err)
		certPool = embeddedroots.Get()
	}

	customTransport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			c, err := customDialer.DialContext(ctx, network, addr)
			if err == nil && underlyingOut != nil {
				*underlyingOut = c
			}
			return c, err
		},
		TLSClientConfig: &tls.Config{
			RootCAs:    certPool,
			ServerName: serverName,
		},
	}

	return &http.Client{
		Transport: customTransport,
	}
}
