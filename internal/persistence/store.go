package persistence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/GoreeCloud/goreecloud-network/internal/controlplane"
)

const (
	fileFormat           = "goreecloud.network.controlplane"
	CurrentSchemaVersion = 1
	maxStateFileBytes    = 8 << 20
)

type MigrationRecord struct {
	Version   int    `json:"version"`
	Name      string `json:"name"`
	AppliedAt string `json:"appliedAt"`
}

type envelope struct {
	Format        string                `json:"format"`
	SchemaVersion int                   `json:"schemaVersion"`
	Revision      uint64                `json:"revision"`
	WrittenAt     string                `json:"writtenAt"`
	Migrations    []MigrationRecord     `json:"migrations"`
	State         controlplane.Snapshot `json:"state"`
	Checksum      string                `json:"checksum,omitempty"`
}

type checksumPayload struct {
	Format        string                `json:"format"`
	SchemaVersion int                   `json:"schemaVersion"`
	Revision      uint64                `json:"revision"`
	WrittenAt     string                `json:"writtenAt"`
	Migrations    []MigrationRecord     `json:"migrations"`
	State         controlplane.Snapshot `json:"state"`
}

type Store struct {
	mu         sync.Mutex
	path       string
	revision   uint64
	migrations []MigrationRecord
	writtenAt  string
}

func Open(path string) (*Store, *controlplane.State, controlplane.StorageStatus, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, nil, controlplane.StorageStatus{}, errors.New("state file path is required")
	}
	cleanPath := filepath.Clean(path)
	store := &Store{path: cleanPath}

	env, err := readEnvelope(cleanPath)
	if errors.Is(err, os.ErrNotExist) {
		state := controlplane.NewState()
		store.migrations = []MigrationRecord{{
			Version:   CurrentSchemaVersion,
			Name:      "initial_schema_v1",
			AppliedAt: time.Now().UTC().Format(time.RFC3339Nano),
		}}
		status, saveErr := store.Save(state)
		if saveErr != nil {
			return nil, nil, controlplane.StorageStatus{}, fmt.Errorf("initialize persistent state: %w", saveErr)
		}
		return store, state, status, nil
	}
	if err != nil {
		return nil, nil, controlplane.StorageStatus{}, err
	}

	if env.Format != fileFormat {
		return nil, nil, controlplane.StorageStatus{}, fmt.Errorf("unsupported state file format %q", env.Format)
	}
	if env.SchemaVersion > CurrentSchemaVersion {
		return nil, nil, controlplane.StorageStatus{}, fmt.Errorf("state schema %d is newer than supported schema %d", env.SchemaVersion, CurrentSchemaVersion)
	}

	store.revision = env.Revision
	store.migrations = append([]MigrationRecord(nil), env.Migrations...)
	store.writtenAt = env.WrittenAt

	migrated := false
	for env.SchemaVersion < CurrentSchemaVersion {
		switch env.SchemaVersion {
		case 0:
			env.SchemaVersion = 1
			store.migrations = append(store.migrations, MigrationRecord{
				Version:   1,
				Name:      "legacy_unversioned_to_v1",
				AppliedAt: time.Now().UTC().Format(time.RFC3339Nano),
			})
			migrated = true
		default:
			return nil, nil, controlplane.StorageStatus{}, fmt.Errorf("no migration path from schema %d", env.SchemaVersion)
		}
	}

	if !migrated {
		expected, checksumErr := checksumFor(env)
		if checksumErr != nil {
			return nil, nil, controlplane.StorageStatus{}, fmt.Errorf("calculate state checksum: %w", checksumErr)
		}
		if env.Checksum == "" || !strings.EqualFold(env.Checksum, expected) {
			return nil, nil, controlplane.StorageStatus{}, errors.New("state file checksum verification failed")
		}
	}

	state, err := controlplane.NewStateFromSnapshot(env.State)
	if err != nil {
		return nil, nil, controlplane.StorageStatus{}, fmt.Errorf("restore control-plane state: %w", err)
	}

	if migrated {
		status, saveErr := store.Save(state)
		if saveErr != nil {
			return nil, nil, controlplane.StorageStatus{}, fmt.Errorf("persist migrated state: %w", saveErr)
		}
		return store, state, status, nil
	}

	return store, state, store.Status(), nil
}

func (s *Store) Save(state *controlplane.State) (controlplane.StorageStatus, error) {
	if state == nil {
		return controlplane.StorageStatus{}, errors.New("state is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return controlplane.StorageStatus{}, fmt.Errorf("create state directory: %w", err)
	}

	nextRevision := s.revision + 1
	writtenAt := time.Now().UTC().Format(time.RFC3339Nano)
	env := envelope{
		Format:        fileFormat,
		SchemaVersion: CurrentSchemaVersion,
		Revision:      nextRevision,
		WrittenAt:     writtenAt,
		Migrations:    append([]MigrationRecord(nil), s.migrations...),
		State:         state.Snapshot(),
	}
	checksum, err := checksumFor(env)
	if err != nil {
		return controlplane.StorageStatus{}, fmt.Errorf("calculate state checksum: %w", err)
	}
	env.Checksum = checksum

	if err := writeEnvelopeAtomic(s.path, env); err != nil {
		return controlplane.StorageStatus{}, err
	}

	s.revision = nextRevision
	s.writtenAt = writtenAt
	return s.statusLocked(), nil
}

func (s *Store) Status() controlplane.StorageStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.statusLocked()
}

func (s *Store) statusLocked() controlplane.StorageStatus {
	return controlplane.StorageStatus{
		Persistence:    "development_file_store",
		SchemaVersion:  CurrentSchemaVersion,
		Revision:       s.revision,
		MigrationCount: len(s.migrations),
		LastPersisted:  s.writtenAt,
		Integrity:      "sha256",
	}
}

func readEnvelope(path string) (envelope, error) {
	info, err := os.Stat(path)
	if err != nil {
		return envelope{}, err
	}
	if info.Size() > maxStateFileBytes {
		return envelope{}, fmt.Errorf("state file exceeds %d bytes", maxStateFileBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return envelope{}, err
	}
	defer file.Close()

	decoder := json.NewDecoder(io.LimitReader(file, maxStateFileBytes+1))
	decoder.DisallowUnknownFields()
	var env envelope
	if err := decoder.Decode(&env); err != nil {
		return envelope{}, fmt.Errorf("decode state file: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return envelope{}, errors.New("state file must contain exactly one JSON value")
		}
		return envelope{}, fmt.Errorf("decode trailing state data: %w", err)
	}
	return env, nil
}

func checksumFor(env envelope) (string, error) {
	payload := checksumPayload{
		Format:        env.Format,
		SchemaVersion: env.SchemaVersion,
		Revision:      env.Revision,
		WrittenAt:     env.WrittenAt,
		Migrations:    env.Migrations,
		State:         env.State,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func writeEnvelopeAtomic(path string, env envelope) error {
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, ".network-state-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary state file: %w", err)
	}
	tempPath := temp.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("set temporary state permissions: %w", err)
	}
	encoder := json.NewEncoder(temp)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(env); err != nil {
		_ = temp.Close()
		return fmt.Errorf("encode state file: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("sync temporary state file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary state file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace state file: %w", err)
	}
	removeTemp = false

	if directory, err := os.Open(dir); err == nil {
		_ = directory.Sync()
		_ = directory.Close()
	}
	return nil
}
