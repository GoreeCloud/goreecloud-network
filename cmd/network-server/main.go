package main

import (
	"log"
	"net/http"
	"os"

	"github.com/GoreeCloud/goreecloud-network/internal/httpapi"
)

func main() {
	addr := getenv("GOREECLOUD_NETWORK_ADDR", "127.0.0.1:8080")
	webRoot := getenv("GOREECLOUD_NETWORK_WEB_ROOT", "apps/web")

	server := &http.Server{
		Addr:    addr,
		Handler: httpapi.NewRouter(webRoot),
	}

	log.Printf("GoreeCloud Network Server %s (%s) listening on http://%s", httpapi.Version, httpapi.Lifecycle, addr)
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
