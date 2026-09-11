package main

import (
	"log"
	"net/http"
	"os"

	"github.com/GoreeCloud/goreecloud-network/internal/httpapi"
	"github.com/GoreeCloud/goreecloud-network/internal/persistence"
)

func main() {
	addr := getenv("GOREECLOUD_NETWORK_ADDR", "127.0.0.1:8080")
	webRoot := getenv("GOREECLOUD_NETWORK_WEB_ROOT", "apps/web")
	dataFile := getenv("GOREECLOUD_NETWORK_DATA_FILE", "var/network-state.json")

	_, state, storage, err := persistence.Open(dataFile)
	if err != nil {
		log.Fatalf("initialize GoreeCloud Network state: %v", err)
	}

	server := &http.Server{
		Addr:    addr,
		Handler: httpapi.NewRouterWithRuntime(webRoot, state, storage),
	}

	log.Printf(
		"GoreeCloud Network Server %s (%s) listening on http://%s; persistence=%s schema=%d revision=%d",
		httpapi.Version,
		httpapi.Lifecycle,
		addr,
		storage.Persistence,
		storage.SchemaVersion,
		storage.Revision,
	)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
