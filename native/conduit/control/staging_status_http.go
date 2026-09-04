package control

import (
	"context"
	"encoding/json"
	"net/http"
)

// CapabilityStagingStatusProvider supplies minimized, durable Conduit staging
// status without exposing the underlying transition receipt or staging record.
type CapabilityStagingStatusProvider interface {
	CapabilityStagingStatus(context.Context) (CapabilityStagingStatus, error)
}

// CapabilityStagingStatusHandler exposes minimized read-only migration evidence
// to central GoreeCloud consumers. It never performs a migration transition,
// changes authority, disables the compatibility bridge, or authorizes cutover.
type CapabilityStagingStatusHandler struct {
	Provider CapabilityStagingStatusProvider
}

func (h CapabilityStagingStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.Provider == nil {
		http.Error(w, "conduit staging status provider unavailable", http.StatusServiceUnavailable)
		return
	}
	status, err := h.Provider.CapabilityStagingStatus(r.Context())
	if err != nil {
		http.Error(w, "conduit staging status unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := ValidateCapabilityStagingStatus(status); err != nil {
		http.Error(w, "conduit staging status rejected", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
}
