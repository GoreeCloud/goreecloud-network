package paddedframe

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

func TestEncodeDecodeUsesCanonicalBucketAndPreservesPayload(t *testing.T) {
	payload := bytes.Repeat([]byte{0x42}, 1500)
	frame, err := Encode(payload, zeroReader{})
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if got, want := len(frame), 2048; got != want {
		t.Fatalf("frame size = %d, want %d", got, want)
	}
	if got := int(binary.BigEndian.Uint16(frame[0:2])); got != len(payload) {
		t.Fatalf("payload length header = %d, want %d", got, len(payload))
	}

	dst := make([]byte, MaxPayload)
	n, err := Decode(frame, dst)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if !bytes.Equal(dst[:n], payload) {
		t.Fatal("decoded payload differs from original")
	}
}

func TestSmallPayloadIsLengthHiddenInsideMinimumBucket(t *testing.T) {
	for _, size := range []int{1, 8, 64, 128, 500} {
		payload := bytes.Repeat([]byte{byte(size)}, size)
		frame, err := Encode(payload, zeroReader{})
		if err != nil {
			t.Fatalf("Encode(%d) error = %v", size, err)
		}
		want := 512
		if size+headerSize > 512 {
			want = 1024
		}
		if len(frame) != want {
			t.Fatalf("Encode(%d) size = %d, want %d", size, len(frame), want)
		}
	}
}

func TestEncodeRejectsUnsafeInputs(t *testing.T) {
	if _, err := Encode(nil, zeroReader{}); err == nil {
		t.Fatal("Encode(nil) succeeded")
	}
	if _, err := Encode([]byte{1}, nil); err == nil {
		t.Fatal("Encode() with nil random source succeeded")
	}
	if _, err := Encode(make([]byte, MaxPayload+1), zeroReader{}); err == nil {
		t.Fatal("Encode() accepted oversized payload")
	}
}

func TestEncodeFailsClosedWhenPaddingRandomnessFails(t *testing.T) {
	_, err := Encode([]byte("hello"), errorReader{})
	if err == nil {
		t.Fatal("Encode() succeeded with failing random source")
	}
}

func TestDecodeRejectsMalformedFrames(t *testing.T) {
	good, err := Encode([]byte("hello"), zeroReader{})
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	tests := map[string]func([]byte) []byte{
		"truncated": func(frame []byte) []byte { return frame[:headerSize-1] },
		"version": func(frame []byte) []byte {
			out := append([]byte(nil), frame...)
			out[4] = 2
			return out
		},
		"flags": func(frame []byte) []byte {
			out := append([]byte(nil), frame...)
			out[5] = 1
			return out
		},
		"length": func(frame []byte) []byte {
			out := append([]byte(nil), frame...)
			binary.BigEndian.PutUint16(out[0:2], 6)
			return out
		},
		"non-canonical": func(frame []byte) []byte {
			out := append([]byte(nil), frame...)
			out = append(out, 0)
			binary.BigEndian.PutUint16(out[2:4], binary.BigEndian.Uint16(out[2:4])+1)
			return out
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Decode(mutate(good), make([]byte, MaxPayload)); err == nil {
				t.Fatal("Decode() accepted malformed frame")
			}
		})
	}
}

func TestDecodeReturnsShortBufferWithoutPartialSuccess(t *testing.T) {
	payload := bytes.Repeat([]byte{1}, 128)
	frame, err := Encode(payload, zeroReader{})
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if _, err := Decode(frame, make([]byte, len(payload)-1)); !errors.Is(err, io.ErrShortBuffer) {
		t.Fatalf("Decode() error = %v, want io.ErrShortBuffer", err)
	}
}

func TestReadRejectsOversizedMessageBeforeDecode(t *testing.T) {
	n, err := Read(bytes.NewReader(make([]byte, MaxFrameSize+1)), make([]byte, MaxPayload))
	if n != 0 || err == nil {
		t.Fatalf("Read() = (%d, %v), want failure", n, err)
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0xA5
	}
	return len(p), nil
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("entropy unavailable") }
