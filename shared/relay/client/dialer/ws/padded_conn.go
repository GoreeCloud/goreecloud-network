package ws

import (
	"crypto/rand"
	"fmt"
	"net"

	"github.com/coder/websocket"

	"github.com/netbirdio/netbird/native/conduit/obfuscation/paddedframe"
)

// PaddedConn keeps the existing relay protocol unchanged while carrying each
// relay message in a bounded padded Conduit frame inside authenticated WSS.
type PaddedConn struct {
	*Conn
}

func NewPaddedConn(wsConn *websocket.Conn, serverAddress string, underlying net.Conn) net.Conn {
	var addr net.Addr = WebsocketAddr{serverAddress}
	if underlying != nil {
		if ra := underlying.RemoteAddr(); ra != nil {
			addr = ra
		}
	}
	return &PaddedConn{Conn: &Conn{
		ctx:        contextBackground(),
		Conn:       wsConn,
		remoteAddr: addr,
	}}
}

func (c *PaddedConn) Protocol() string {
	return paddedframe.TransportID
}

func (c *PaddedConn) Read(b []byte) (int, error) {
	t, r, err := c.Conn.Conn.Reader(c.ctx)
	if err != nil {
		return 0, err
	}
	if t != websocket.MessageBinary {
		return 0, fmt.Errorf("unexpected message type")
	}
	return paddedframe.Read(r, b)
}

func (c *PaddedConn) Write(b []byte) (int, error) {
	frame, err := paddedframe.Encode(b, rand.Reader)
	if err != nil {
		return 0, err
	}
	if err := c.Conn.Conn.Write(c.ctx, websocket.MessageBinary, frame); err != nil {
		return 0, err
	}
	return len(b), nil
}

// Kept as a tiny helper so construction stays parallel to NewConn without
// exposing Conn.ctx outside this package.
func contextBackground() context.Context { return context.Background() }
