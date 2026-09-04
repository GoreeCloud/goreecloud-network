package paddedframe

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	// TransportID is the negotiated Conduit transport/profile identifier. It is
	// intentionally distinct from ordinary WebSocket relay transport.
	TransportID = "conduit-padded-wss"
	Version     = "1"
	URLPath     = "/conduit/obfuscation/v1"

	frameVersion = 1
	headerSize   = 6
	MaxPayload   = 16 * 1024
	MaxFrameSize = 24 * 1024
)

var buckets = [...]int{512, 1024, 2048, 4096, 8192, 12288, 16384, MaxFrameSize}

// Encode wraps one already-encrypted-tunnel/relay protocol message in a padded
// application frame. This is traffic-shaping/length-hiding only: it does not
// provide encryption or authentication and must be carried inside the existing
// authenticated TLS/WebSocket channel.
func Encode(payload []byte, random io.Reader) ([]byte, error) {
	if len(payload) == 0 {
		return nil, errors.New("conduit padded frame: empty payload")
	}
	if len(payload) > MaxPayload {
		return nil, fmt.Errorf("conduit padded frame: payload too large: %d > %d", len(payload), MaxPayload)
	}
	if random == nil {
		return nil, errors.New("conduit padded frame: random source is required")
	}

	target, ok := paddedSize(headerSize + len(payload))
	if !ok {
		return nil, errors.New("conduit padded frame: no valid padding bucket")
	}
	paddingLen := target - headerSize - len(payload)
	if paddingLen > int(^uint16(0)) {
		return nil, errors.New("conduit padded frame: padding length overflow")
	}

	frame := make([]byte, target)
	binary.BigEndian.PutUint16(frame[0:2], uint16(len(payload)))
	binary.BigEndian.PutUint16(frame[2:4], uint16(paddingLen))
	frame[4] = frameVersion
	frame[5] = 0
	copy(frame[headerSize:], payload)
	if paddingLen > 0 {
		if _, err := io.ReadFull(random, frame[headerSize+len(payload):]); err != nil {
			return nil, fmt.Errorf("conduit padded frame: generate padding: %w", err)
		}
	}
	return frame, nil
}

// Decode validates and unwraps one complete padded frame into dst. The caller
// must preserve WebSocket message boundaries; concatenated or truncated frames
// are rejected rather than tolerated.
func Decode(frame, dst []byte) (int, error) {
	if len(frame) < headerSize {
		return 0, errors.New("conduit padded frame: truncated header")
	}
	if len(frame) > MaxFrameSize {
		return 0, fmt.Errorf("conduit padded frame: frame too large: %d > %d", len(frame), MaxFrameSize)
	}
	if frame[4] != frameVersion {
		return 0, fmt.Errorf("conduit padded frame: unsupported frame version %d", frame[4])
	}
	if frame[5] != 0 {
		return 0, errors.New("conduit padded frame: unsupported flags")
	}

	payloadLen := int(binary.BigEndian.Uint16(frame[0:2]))
	paddingLen := int(binary.BigEndian.Uint16(frame[2:4]))
	if payloadLen == 0 || payloadLen > MaxPayload {
		return 0, errors.New("conduit padded frame: invalid payload length")
	}
	if headerSize+payloadLen+paddingLen != len(frame) {
		return 0, errors.New("conduit padded frame: inconsistent frame lengths")
	}
	if expected, ok := paddedSize(headerSize + payloadLen); !ok || expected != len(frame) {
		return 0, errors.New("conduit padded frame: non-canonical padding bucket")
	}
	if len(dst) < payloadLen {
		return 0, io.ErrShortBuffer
	}
	copy(dst, frame[headerSize:headerSize+payloadLen])
	return payloadLen, nil
}

// Read decodes one bounded frame from a message reader without allowing an
// untrusted peer to force an unbounded allocation.
func Read(r io.Reader, dst []byte) (int, error) {
	if r == nil {
		return 0, errors.New("conduit padded frame: reader is required")
	}
	frame, err := io.ReadAll(io.LimitReader(r, MaxFrameSize+1))
	if err != nil {
		return 0, fmt.Errorf("conduit padded frame: read: %w", err)
	}
	if len(frame) > MaxFrameSize {
		return 0, errors.New("conduit padded frame: frame exceeds maximum size")
	}
	return Decode(frame, dst)
}

func paddedSize(minimum int) (int, bool) {
	for _, bucket := range buckets {
		if minimum <= bucket {
			return bucket, true
		}
	}
	return 0, false
}
