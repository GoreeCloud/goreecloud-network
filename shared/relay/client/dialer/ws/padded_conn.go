package ws

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"sync"

	"github.com/coder/websocket"

	"github.com/netbirdio/netbird/native/conduit/obfuscation/paddedframe"
	obfuscationruntime "github.com/netbirdio/netbird/native/conduit/obfuscation/runtime"
	"github.com/netbirdio/netbird/shared/relay/messages"
)

// PaddedConn keeps the existing relay protocol unchanged while carrying each
// relay message in a bounded padded Conduit frame inside authenticated WSS.
// Runtime obfuscation becomes active only after this connection observes the
// relay authentication response, not merely after the WebSocket upgrade.
type PaddedConn struct {
	*Conn

	stateMu  sync.Mutex
	active   bool
	released bool
}

func NewPaddedConn(wsConn *websocket.Conn, serverAddress string, underlying net.Conn) net.Conn {
	var addr net.Addr = WebsocketAddr{serverAddress}
	if underlying != nil {
		if ra := underlying.RemoteAddr(); ra != nil {
			addr = ra
		}
	}
	return &PaddedConn{Conn: &Conn{
		ctx:        context.Background(),
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
		c.releaseRuntime(true)
		return 0, err
	}
	if t != websocket.MessageBinary {
		c.releaseRuntime(true)
		return 0, fmt.Errorf("unexpected message type")
	}

	n, err := paddedframe.Read(r, b)
	if err != nil {
		c.releaseRuntime(true)
		return 0, err
	}
	if n > 0 {
		if msgType, msgErr := messages.DetermineServerMessageType(b[:n]); msgErr == nil && msgType == messages.MsgTypeAuthResponse {
			if err := c.confirmActive(); err != nil {
				c.releaseRuntime(true)
				return 0, err
			}
		}
	}
	return n, nil
}

func (c *PaddedConn) Write(b []byte) (int, error) {
	frame, err := paddedframe.Encode(b, rand.Reader)
	if err != nil {
		c.releaseRuntime(true)
		return 0, err
	}
	if err := c.Conn.Conn.Write(c.ctx, websocket.MessageBinary, frame); err != nil {
		c.releaseRuntime(true)
		return 0, err
	}
	return len(b), nil
}

func (c *PaddedConn) Close() error {
	// A close before the authenticated relay response means the transport never
	// reached active. Once active, an explicit close is treated as graceful;
	// read/write failures release the state as failed before Close is called.
	c.stateMu.Lock()
	active := c.active
	alreadyReleased := c.released
	c.stateMu.Unlock()
	if !alreadyReleased {
		c.releaseRuntime(!active)
	}
	return c.Conn.Close()
}

func (c *PaddedConn) confirmActive() error {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	if c.released || c.active {
		return nil
	}
	if err := obfuscationruntime.Active(c.Protocol()); err != nil {
		return err
	}
	c.active = true
	return nil
}

func (c *PaddedConn) releaseRuntime(failed bool) {
	c.stateMu.Lock()
	if c.released {
		c.stateMu.Unlock()
		return
	}
	wasActive := c.active
	c.active = false
	c.released = true
	c.stateMu.Unlock()

	if wasActive {
		obfuscationruntime.ReleaseActive(failed)
		return
	}
	if failed {
		obfuscationruntime.Failed()
	}
}
