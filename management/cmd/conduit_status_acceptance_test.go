package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/netbirdio/netbird/native/conduit/control"
)

const conduitStatusAcceptanceSchema = "goreecloud-conduit-status-isolated-runtime-evidence/v1"

var exactSourceRevision = regexp.MustCompile(`^[0-9a-f]{40}$`)

type conduitStatusAcceptanceEvidence struct {
	Schema                      string `json:"schema"`
	SourceRevision              string `json:"source_revision"`
	RuntimeArtifactSHA256       string `json:"runtime_artifact_sha256"`
	ListenerScope               string `json:"listener_scope"`
	StatusSchema                string `json:"status_schema"`
	Authority                   string `json:"authority"`
	MigrationStage              string `json:"migration_stage"`
	CompatibilityBridgeActive   bool   `json:"compatibility_bridge_active"`
	Availability                string `json:"availability"`
	AvailabilityReason          string `json:"availability_reason"`
	ReadOnlyValidated           bool   `json:"read_only_validated"`
	MinimizedFieldsValidated    bool   `json:"minimized_fields_validated"`
	CredentialsIncluded         bool   `json:"credentials_included"`
	ProductionCutoverAuthorized bool   `json:"production_cutover_authorized"`
}

func TestConduitStatusIsolatedRuntimeEvidence(t *testing.T) {
	evidencePath := os.Getenv("GOREECLOUD_CONDUIT_ACCEPTANCE_EVIDENCE")
	if evidencePath == "" {
		t.Skip("isolated runtime evidence path is not configured")
	}
	sourceRevision := os.Getenv("GOREECLOUD_NETWORK_SOURCE_SHA")
	if !exactSourceRevision.MatchString(sourceRevision) {
		t.Fatalf("GOREECLOUD_NETWORK_SOURCE_SHA = %q, want exact lowercase 40-character SHA", sourceRevision)
	}

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
	if status.Schema != control.SchemaV1 {
		t.Fatalf("schema = %q, want %q", status.Schema, control.SchemaV1)
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

	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("resolve test executable: %v", err)
	}
	artifactDigest, err := fileSHA256(executable)
	if err != nil {
		t.Fatalf("hash test executable: %v", err)
	}

	evidence := conduitStatusAcceptanceEvidence{
		Schema:                      conduitStatusAcceptanceSchema,
		SourceRevision:              sourceRevision,
		RuntimeArtifactSHA256:       artifactDigest,
		ListenerScope:               "127.0.0.1",
		StatusSchema:                status.Schema,
		Authority:                   string(status.Authority),
		MigrationStage:              status.MigrationStage,
		CompatibilityBridgeActive:   status.CompatibilityBridgeActive,
		Availability:                status.Availability,
		AvailabilityReason:          status.AvailabilityReason,
		ReadOnlyValidated:           true,
		MinimizedFieldsValidated:    true,
		CredentialsIncluded:         false,
		ProductionCutoverAuthorized: false,
	}
	encoded, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Fatalf("encode evidence: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(evidencePath), 0o700); err != nil {
		t.Fatalf("create evidence directory: %v", err)
	}
	if err := os.WriteFile(evidencePath, append(encoded, '\n'), 0o600); err != nil {
		t.Fatalf("write evidence: %v", err)
	}
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}
