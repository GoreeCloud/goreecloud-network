package control

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const CapabilityInventoryStatusSchemaV1 = "goreecloud-conduit-capability-inventory-status/v1"

// CapabilityInventoryStatus exposes aggregate migration state without
// individual capability identifiers or inherited operational data.
type CapabilityInventoryStatus struct {
	Schema                      string            `json:"schema"`
	GeneratedAt                 time.Time         `json:"generated_at"`
	SnapshotUpdatedAt           string            `json:"snapshot_updated_at"`
	SnapshotFingerprint         string            `json:"snapshot_fingerprint"`
	Summary                     CapabilitySummary `json:"summary"`
	ProductionCutoverAuthorized bool              `json:"production_cutover_authorized"`
}

func BuildCapabilityInventoryStatus(snapshot CapabilityInventorySnapshot, now time.Time) (CapabilityInventoryStatus, error) {
	if err := validateCapabilityInventorySnapshot(snapshot); err != nil {
		return CapabilityInventoryStatus{}, err
	}
	if now.IsZero() {
		return CapabilityInventoryStatus{}, errors.New("conduit control: capability inventory status generation time is required")
	}
	summary, err := SummarizeCapabilities(snapshot.Inventory)
	if err != nil {
		return CapabilityInventoryStatus{}, err
	}
	return CapabilityInventoryStatus{
		Schema:                      CapabilityInventoryStatusSchemaV1,
		GeneratedAt:                 now.UTC(),
		SnapshotUpdatedAt:           snapshot.UpdatedAt,
		SnapshotFingerprint:         snapshot.Fingerprint,
		Summary:                     summary,
		ProductionCutoverAuthorized: false,
	}, nil
}

// InventoryStatusHandler provides a read-only aggregate endpoint suitable for
// Manager, Privacy Shield, and Wardveil Security status consumption. It never
// returns capability IDs, peers, routes, policies, credentials, or packet data.
type InventoryStatusHandler struct {
	Store *CapabilityInventoryStore
	Now   func() time.Time
}

func (h InventoryStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.Store == nil {
		http.Error(w, "conduit capability inventory unavailable", http.StatusServiceUnavailable)
		return
	}
	snapshot, err := h.Store.Load()
	if err != nil {
		http.Error(w, "conduit capability inventory unavailable", http.StatusServiceUnavailable)
		return
	}
	now := time.Now
	if h.Now != nil {
		now = h.Now
	}
	status, err := BuildCapabilityInventoryStatus(snapshot, now())
	if err != nil {
		http.Error(w, "conduit capability inventory status rejected", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
}
