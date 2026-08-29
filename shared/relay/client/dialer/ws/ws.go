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
	"github.com/netbirdio/netbird/shared/relay"
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
	path := d.urlPath
	if path == "" {
		path = relay.WebSocketURLPath
	}
	wsURL, err := prepareURLPath(address, path)
	if err != nil {
		return nil, err
	}

	var underlying net.Conn
	opts := createDialOptions(serverName, &underlying)

	wsConn, resp, err := websocket.Dial(ctx, wsURL, opts)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, err
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
		return NewPaddedConn(wsConn, address, underlying), nil
	}
	return NewConn(wsConn, address, underlying), nil
}

// prepareURL rewrites a rel://host[:port] or rels://host[:port] address into a
// ws://host[:port]/relay or wss://host[:port]/relay URL, preserving any
// non-standard port from the input.
func prepareURL(address string) (string, error) {
	return prepareURLPath(address, relay.WebSocketURLPath)
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
