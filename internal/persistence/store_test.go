package persistence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoreeCloud/goreecloud-network/internal/controlplane"
)

func TestOpenCreatesRevisionedPersistentState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "network-state.json")
	store, state, status, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if status.Persistence != "development_file_store" || status.SchemaVersion != 1 || status.Revision != 1 || status.MigrationCount != 1 || status.Integrity != "sha256" {
		t.Fatalf("unexpected initial storage status: %+v", status)
	}
	if err := state.PutDevice(controlplane.Device{ID: "device-1", Name: "Phone", Platform: "android", State: "approved"}); err != nil {
		t.Fatal(err)
	}
	status, err = store.Save(state)
	if err != nil {
		t.Fatal(err)
	}
	if status.Revision != 2 {
		t.Fatalf("expected revision 2 after save, got %+v", status)
	}

	_, restored, restoredStatus, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if restoredStatus.Revision != 2 || len(restored.Devices()) != 1 || restored.Devices()[0].ID != "device-1" {
		t.Fatalf("persistent state did not survive reopen: status=%+v devices=%+v", restoredStatus, restored.Devices())
	}
}

func TestOpenMigratesLegacyUnversionedEnvelope(t *testing.T) {
	path := filepath.Join(t.TempDir(), "network-state.json")
	legacy := envelope{
		Format:   fileFormat,
		Revision: 7,
		State: controlplane.Snapshot{Devices: []controlplane.Device{{
			ID: "device-legacy", Name: "Legacy device", Platform: "linux", State: "approved",
		}}},
	}
	encoded, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}

	_, state, status, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if status.SchemaVersion != 1 || status.Revision != 8 || status.MigrationCount != 1 {
		t.Fatalf("unexpected migrated status: %+v", status)
	}
	if got := state.Devices(); len(got) != 1 || got[0].ID != "device-legacy" {
		t.Fatalf("legacy state was not restored: %+v", got)
	}
}

func TestOpenRejectsChecksumMismatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "network-state.json")
	_, _, _, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var env envelope
	if err := json.Unmarshal(encoded, &env); err != nil {
		t.Fatal(err)
	}
	env.Checksum = "deadbeef"
	encoded, err = json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, _, err := Open(path); err == nil {
		t.Fatal("expected checksum mismatch to fail closed")
	}
}

func TestOpenRejectsPersistedTimestampTampering(t *testing.T) {
	path := filepath.Join(t.TempDir(), "network-state.json")
	_, _, _, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var env envelope
	if err := json.Unmarshal(encoded, &env); err != nil {
		t.Fatal(err)
	}
	env.WrittenAt = "2000-01-01T00:00:00Z"
	encoded, err = json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, _, err := Open(path); err == nil {
		t.Fatal("expected timestamp tampering to fail checksum verification")
	}
}

func TestSaveIsDeterministicForMapBackedState(t *testing.T) {
	state := controlplane.NewState()
	for _, device := range []controlplane.Device{
		{ID: "device-z", Name: "Z", Platform: "linux", State: "approved"},
		{ID: "device-a", Name: "A", Platform: "android", State: "approved"},
	} {
		if err := state.PutDevice(device); err != nil {
			t.Fatal(err)
		}
	}
	snapshot := state.Snapshot()
	if snapshot.Devices[0].ID != "device-a" || snapshot.Devices[1].ID != "device-z" {
		t.Fatalf("snapshot must be deterministic: %+v", snapshot.Devices)
	}
}
