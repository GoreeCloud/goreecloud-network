package cmd

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/netbirdio/netbird/management/internals/server"
	"github.com/netbirdio/netbird/native/conduit/control"
)

type fakeManagementServer struct {
	started bool
	stopped bool
	errors  chan error
	values  map[string]any
}

func newFakeManagementServer() *fakeManagementServer {
	return &fakeManagementServer{errors: make(chan error, 1), values: make(map[string]any)}
}

func (s *fakeManagementServer) Start(context.Context) error { s.started = true; return nil }
func (s *fakeManagementServer) Stop() error                 { s.stopped = true; return nil }
func (s *fakeManagementServer) Errors() <-chan error        { return s.errors }
func (s *fakeManagementServer) GetContainer(key string) (any, bool) {
	v, ok := s.values[key]
	return v, ok
}
func (s *fakeManagementServer) SetContainer(key string, value any) { s.values[key] = value }

var _ server.Server = (*fakeManagementServer)(nil)

func TestConduitStatusDisabledReturnsInheritedServer(t *testing.T) {
	inner := newFakeManagementServer()
	got := newConduitStatusServer(inner, conduitStatusSettings{enabled: false, addr: "127.0.0.1:0"})
	if got != inner {
		t.Fatal("disabled Conduit status must not wrap the inherited management server")
	}
}

func TestConduitStatusRejectsNonLoopbackListener(t *testing.T) {
	inner := newFakeManagementServer()
	wrapped := newConduitStatusServer(inner, conduitStatusSettings{enabled: true, addr: "0.0.0.0:9097"})
	if err := wrapped.Start(context.Background()); err == nil {
		t.Fatal("expected non-loopback status listener to be rejected")
	}
	if inner.started {
		t.Fatal("inherited server must not start when Conduit listener configuration is unsafe")
	}
}

func TestConduitStatusStartsAfterInheritedRuntimeAndIsReadOnly(t *testing.T) {
	inner := newFakeManagementServer()
	wrapped := newConduitStatusServer(inner, conduitStatusSettings{enabled: true, addr: "127.0.0.1:0"})
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
	if !inner.started {
		t.Fatal("inherited management runtime must start before Conduit status is exposed")
	}

	conduit := wrapped.(*conduitStatusServer)
	baseURL := "http://" + conduit.listener.Addr().String() + conduitStatusPath
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL)
	if err != nil {
		t.Fatalf("GET status: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d, body=%s", resp.StatusCode, body)
	}

	var status control.Status
	if err := json.Unmarshal(body, &status); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if status.Authority != control.AuthorityInherited {
		t.Fatalf("authority = %q, want inherited", status.Authority)
	}
	if status.MigrationStage != "implementation" {
		t.Fatalf("migration_stage = %q, want implementation", status.MigrationStage)
	}
	if !status.CompatibilityBridgeActive {
		t.Fatal("compatibility bridge must remain active")
	}
	if status.ProductionCutoverAuthorized {
		t.Fatal("status must not authorize production cutover")
	}
	if status.Availability != control.AvailabilityUnknown {
		t.Fatalf("availability = %q, want %q", status.Availability, control.AvailabilityUnknown)
	}
	if status.AvailabilityReason != control.AvailabilityReasonRuntimeHealthNotObserved {
		t.Fatalf(
			"availability_reason = %q, want %q",
			status.AvailabilityReason,
			control.AvailabilityReasonRuntimeHealthNotObserved,
		)
	}

	var fields map[string]any
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatalf("decode fields: %v", err)
	}
	allowed := map[string]bool{
		"schema": true, "generated_at": true, "authority": true,
		"migration_stage": true, "compatibility_bridge_active": true,
		"production_cutover_authorized": true, "availability": true,
		"availability_reason": true,
	}
	for field := range fields {
		if !allowed[field] {
			t.Fatalf("unexpected status field %q", field)
		}
	}
	if len(fields) != len(allowed) {
		t.Fatalf("status fields = %d, want exactly %d privacy-safe fields", len(fields), len(allowed))
	}

	req, err := http.NewRequest(http.MethodPost, baseURL, nil)
	if err != nil {
		t.Fatalf("new POST request: %v", err)
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("POST status: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("POST status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}
