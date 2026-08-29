package paddedframe_test

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/netbirdio/netbird/native/conduit/obfuscation/paddedframe"
	serverws "github.com/netbirdio/netbird/relay/server/listener/ws"
	clientws "github.com/netbirdio/netbird/shared/relay/client/dialer/ws"
)

func TestPaddedWSSClientAndServerInteroperate(t *testing.T) {
	requestPayload := bytes.Repeat([]byte("request-"), 180)
	responsePayload := bytes.Repeat([]byte("response-"), 120)
	serverErr := make(chan error, 1)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != paddedframe.URLPath {
			http.NotFound(w, r)
			return
		}
		wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
		if err != nil {
			serverErr <- err
			return
		}
		rAddr, err := net.ResolveTCPAddr("tcp", r.RemoteAddr)
		if err != nil {
			_ = wsConn.Close(websocket.StatusInternalError, "internal error")
			serverErr <- err
			return
		}
		conn := serverws.NewPaddedConn(wsConn, rAddr)
		defer conn.Close()

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		buf := make([]byte, paddedframe.MaxPayload)
		n, err := conn.Read(ctx, buf)
		if err != nil {
			serverErr <- err
			return
		}
		if !bytes.Equal(buf[:n], requestPayload) {
			serverErr <- errPayloadMismatch
			return
		}
		if _, err := conn.Write(ctx, responsePayload); err != nil {
			serverErr <- err
			return
		}
		serverErr <- nil
	})

	srv := httptest.NewServer(h)
	defer srv.Close()
	relayURL := strings.Replace(srv.URL, "http://", "rel://", 1)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	dialer := clientws.NewConduitObfuscationDialer()
	conn, err := dialer.Dial(ctx, relayURL, "")
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer conn.Close()

	if tc, ok := conn.(interface{ Protocol() string }); !ok || tc.Protocol() != paddedframe.TransportID {
		t.Fatalf("transport protocol = %#v, want %q", tc, paddedframe.TransportID)
	}
	if _, err := conn.Write(requestPayload); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	buf := make([]byte, paddedframe.MaxPayload)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if !bytes.Equal(buf[:n], responsePayload) {
		t.Fatal("response payload differs from server payload")
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("server profile error = %v", err)
	}
}

type payloadMismatchError struct{}

func (payloadMismatchError) Error() string { return "server received different payload" }

var errPayloadMismatch payloadMismatchError
