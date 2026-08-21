package control

import (
	"encoding/json"
	"net/http"
)

// StatusHandler exposes the first-party read-only Conduit Control status API.
// It deliberately implements GET only; authoritative mutations remain outside
// this compatibility-phase surface.
type StatusHandler struct {
	Provider Provider
}

func (h StatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.Provider == nil {
		http.Error(w, "conduit control provider unavailable", http.StatusServiceUnavailable)
		return
	}
	status, err := h.Provider.Status(r.Context())
	if err != nil {
		http.Error(w, "conduit control status unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := ValidateStatus(status); err != nil {
		http.Error(w, "conduit control status rejected", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
}
