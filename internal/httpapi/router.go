package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

const (
	Version   = "0.1.0-dev.1"
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

func NewRouter(webRoot string) http.Handler {
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
				Server:   "bootstrap",
				Web:      "bootstrap",
				Android:  "source_bootstrap",
				GoogleTV: "source_bootstrap",
				IOS:      "source_bootstrap",
			},
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
			{ID: "web.dashboard", State: "bootstrap", Detail: "Live Development dashboard consuming the status API."},
			{ID: "device.enrollment", State: "not_implemented", Detail: "No enrollment or key issuance exists in this bootstrap."},
			{ID: "tunnel.wireguard", State: "not_implemented", Detail: "No tunnel lifecycle exists in this bootstrap."},
			{ID: "access.policy", State: "not_implemented", Detail: "No network authorization decision engine exists in this bootstrap."},
			{ID: "routing.private", State: "not_implemented", Detail: "No route advertisement or forwarding controller exists in this bootstrap."},
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
