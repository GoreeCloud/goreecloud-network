package cmd

import (
	"context"
	"os"
	"testing"
	"time"
)

const (
	conduitContainerReadyEnv   = "GOREECLOUD_CONDUIT_CONTAINER_READY"
	conduitContainerReleaseEnv = "GOREECLOUD_CONDUIT_CONTAINER_RELEASE"
)

// TestConduitStatusContainerBoundaryServer is an opt-in acceptance fixture used
// by the container runtime workflow.  It exercises the same status-server
// wrapper as the management process while keeping the listener bound to an
// explicit container-local loopback address.  The fixture is deliberately
// unavailable unless both coordination paths are supplied by the workflow.
func TestConduitStatusContainerBoundaryServer(t *testing.T) {
	readyPath := os.Getenv(conduitContainerReadyEnv)
	releasePath := os.Getenv(conduitContainerReleaseEnv)
	if readyPath == "" || releasePath == "" {
		t.Skip("container-boundary acceptance coordination is not configured")
	}

	inner := newFakeManagementServer()
	wrapped := newConduitStatusServer(inner, conduitStatusSettings{
		enabled: true,
		addr:    conduitStatusDefaultAddr,
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := wrapped.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() {
		if err := wrapped.Stop(); err != nil {
			t.Errorf("Stop() error = %v", err)
		}
	}()

	if err := os.WriteFile(readyPath, []byte("ready\n"), 0o600); err != nil {
		t.Fatalf("write ready marker: %v", err)
	}

	deadline := time.NewTimer(90 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline.C:
			t.Fatal("timed out waiting for container-boundary probe release")
		case <-ticker.C:
			if _, err := os.Stat(releasePath); err == nil {
				return
			} else if !os.IsNotExist(err) {
				t.Fatalf("inspect release marker: %v", err)
			}
		}
	}
}
