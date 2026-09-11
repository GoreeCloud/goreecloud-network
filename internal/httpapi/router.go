package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/GoreeCloud/goreecloud-network/internal/controlplane"
)

const (
	Version   = "0.1.0-dev.2"
	Lifecycle = "Development"
)

type SurfaceStatus struct {
	Server   string `json:"server"`
	Web      string `json:"webDashboard"`
	Android  string `json:"android"`
	GoogleTV string `json:"googleTv"`
	IOS      string `json:"ios"`
}

type StatusResponse struct {
	Product   string        `json:"product"`
	Version   string        `json:"version"`
	Lifecycle string        `json:"lifecycle"`
	Surfaces  SurfaceStatus `json:"surfaces"`
}

type PlatformSystemState struct {
	Name     string `json:"name"`
	State    string `json:"state"`
	Evidence string `json:"evidence"`
}

type Capability struct {
	ID     string `json:"id"`
	State  string `json:"state"`
	Detail string `json:"detail"`
}

type AccessEvaluationResponse struct {
	controlplane.AccessDecision
	Enforcement string `json:"enforcement"`
}

func NewRouter(webRoot string) http.Handler {
	return NewRouterWithState(webRoot, controlplane.NewState())
}

func NewRouterWithState(webRoot string, state *controlplane.State) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /api/v1/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, StatusResponse{
			Product:   "GoreeCloud Network",
			Version:   Version,
			Lifecycle: Lifecycle,
			Surfaces: SurfaceStatus{
				Server:   "development_control_plane",
				Web:      "development_dashboard",
				Android:  "development_client",
				GoogleTV: "development_client",
				IOS:      "development_client",
			},
		})
	})

	mux.HandleFunc("GET /api/v1/overview", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, state.Overview())
	})

	mux.HandleFunc("GET /api/v1/devices", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, state.Devices())
	})

	mux.HandleFunc("POST /api/v1/access/evaluate", func(w http.ResponseWriter, r *http.Request) {
		var request controlplane.AccessRequest
		if err := decodeJSON(w, r, &request); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error":  "invalid_request",
				"detail": err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, AccessEvaluationResponse{
			AccessDecision: state.EvaluateAccess(request),
			Enforcement:    "decision_only",
		})
	})

	mux.HandleFunc("GET /api/v1/platform-systems", func(w http.ResponseWriter, r *http.Request) {
		systems := []PlatformSystemState{
			{Name: "GoreeCloud Manager", State: "planned", Evidence: "none"},
			{Name: "Privacy Shield", State: "planned", Evidence: "none"},
			{Name: "Wardveil Security", State: "planned", Evidence: "none"},
			{Name: "Everkeep", State: "planned", Evidence: "none"},
			{Name: "Glaze UI", State: "planned", Evidence: "none"},
			{Name: "GoreeCloud Mesh", State: "planned", Evidence: "none"},
			{Name: "GoreeCloud Identity", State: "planned", Evidence: "none"},
		}
		writeJSON(w, http.StatusOK, systems)
	})

	mux.HandleFunc("GET /api/v1/capabilities", func(w http.ResponseWriter, r *http.Request) {
		capabilities := []Capability{
			{ID: "status.discovery", State: "implemented", Detail: "Read-only Development status API."},
			{ID: "controlplane.overview", State: "implemented", Detail: "Read-only Development inventory and authority-state summary."},
			{ID: "device.inventory", State: "implemented", Detail: "Read-only in-memory Development device inventory. It starts empty and is not durable."},
			{ID: "access.policy.evaluation", State: "implemented", Detail: "In-memory decision kernel supports explicit allow rules and deny-by-default. It does not yet enforce network traffic."},
			{ID: "access.policy.management", State: "not_implemented", Detail: "No authenticated policy administration API or durable policy store exists yet."},
			{ID: "device.enrollment", State: "not_implemented", Detail: "No enrollment credential or device-key issuance exists yet."},
			{ID: "tunnel.wireguard", State: "not_implemented", Detail: "No tunnel lifecycle exists yet."},
			{ID: "routing.private", State: "not_implemented", Detail: "No route advertisement or forwarding controller exists yet."},
			{ID: "relay", State: "not_implemented", Detail: "No relay transport is active."},
			{ID: "obfuscation", State: "not_implemented", Detail: "No obfuscation transport is active."},
		}
		writeJSON(w, http.StatusOK, capabilities)
	})

	if info, err := os.Stat(webRoot); err == nil && info.IsDir() {
		fs := http.FileServer(http.Dir(filepath.Clean(webRoot)))
		mux.Handle("/", fs)
	} else {
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error":  "web_dashboard_unavailable",
				"detail": "Development dashboard files were not found at the configured web root.",
			})
		})
	}

	return securityHeaders(mux)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, value any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra any
	err := decoder.Decode(&extra)
	if err == nil {
		return fmt.Errorf("request body must contain exactly one JSON value")
	}
	if err != io.EOF {
		return fmt.Errorf("invalid trailing JSON data: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self' data:; base-uri 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}
