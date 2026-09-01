package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	goreecloudstatus "github.com/netbirdio/netbird/goreecloud/status"
)

type fakeGoreeCloudStatusServer struct {
	errors     chan error
	containers map[string]any
}

func newFakeGoreeCloudStatusServer() *fakeGoreeCloudStatusServer {
	return &fakeGoreeCloudStatusServer{
		errors:     make(chan error),
		containers: make(map[string]any),
	}
}

func (s *fakeGoreeCloudStatusServer) Start(context.Context) error { return nil }
func (s *fakeGoreeCloudStatusServer) Stop() error                 { return nil }
func (s *fakeGoreeCloudStatusServer) Errors() <-chan error        { return s.errors }
func (s *fakeGoreeCloudStatusServer) GetContainer(key string) (any, bool) {
	value, ok := s.containers[key]
	return value, ok
}
func (s *fakeGoreeCloudStatusServer) SetContainer(key string, value any) {
	s.containers[key] = value
}

func TestGoreeCloudStatusServerPublishesBoundedManagementEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "network-status.json")
	wrapped := newGoreeCloudStatusServer(newFakeGoreeCloudStatusServer(), path)

	if err := wrapped.Start(context.Background()); err != nil {
		t.Fatalf("start wrapped server: %v", err)
	}

	snapshot := readGoreeCloudStatusSnapshot(t, path)
	if snapshot.State != "partial" {
		t.Fatalf("running management status = %q, want partial", snapshot.State)
	}
	states := make(map[string]string, len(snapshot.Capabilities))
	for _, capability := range snapshot.Capabilities {
		states[capability.ID] = capability.State
	}
	if states["peer-coordination"] != "verified" || states["access-policy"] != "verified" {
		t.Fatalf("management capabilities were not verified: %#v", states)
	}
	if states["private-connectivity"] != "attention" || states["network-dns"] != "attention" {
		t.Fatalf("unproven capabilities must remain attention: %#v", states)
	}

	if err := wrapped.Stop(); err != nil {
		t.Fatalf("stop wrapped server: %v", err)
	}
	stopped := readGoreeCloudStatusSnapshot(t, path)
	if stopped.State != "unavailable" {
		t.Fatalf("stopped management status = %q, want unavailable", stopped.State)
	}
}

func readGoreeCloudStatusSnapshot(t *testing.T, path string) goreecloudstatus.Snapshot {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat status file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("status permissions = %o, want 600", got)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read status file: %v", err)
	}
	var snapshot goreecloudstatus.Snapshot
	if err = json.Unmarshal(data, &snapshot); err != nil {
		t.Fatalf("decode status file: %v", err)
	}
	return snapshot
}
