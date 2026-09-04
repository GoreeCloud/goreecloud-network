package ws

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"

	"github.com/coder/websocket"

	"github.com/netbirdio/netbird/native/conduit/obfuscation/paddedframe"
)

// PaddedConn carries existing relay protocol messages inside the first-party
// Conduit padded frame profile. TLS/WebSocket remains the authenticated carrier;
// paddedframe only reduces inner message-length/byte-pattern exposure.
type PaddedConn struct {
	*Conn
}

func NewPaddedConn(wsConn *websocket.Conn, rAddr *net.TCPAddr) *PaddedConn {
	return &PaddedConn{Conn: NewConn(wsConn, rAddr)}
}

func (c *PaddedConn) Read(ctx context.Context, b []byte) (int, error) {
	t, r, err := c.Reader(ctx)
	if err != nil {
		return 0, c.ioErrHandling(err)
	}
	if t != websocket.MessageBinary {
		return 0, fmt.Errorf("unexpected message type")
	}
	return paddedframe.Read(r, b)
}

func (c *PaddedConn) Write(ctx context.Context, b []byte) (int, error) {
	frame, err := paddedframe.Encode(b, rand.Reader)
	if err != nil {
		return 0, err
	}
	writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	if err := c.Conn.Conn.Write(writeCtx, websocket.MessageBinary, frame); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *PaddedConn) Protocol() string {
	return paddedframe.TransportID
}

var _ interface {
	Read(context.Context, []byte) (int, error)
	Write(context.Context, []byte) (int, error)
	RemoteAddr() net.Addr
	Close() error
	Protocol() string
} = (*PaddedConn)(nil)
