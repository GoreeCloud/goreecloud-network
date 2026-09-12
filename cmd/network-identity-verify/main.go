package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/GoreeCloud/goreecloud-network/internal/identity"
)

const defaultAdminScope = "goreecloud.network.admin"

func main() {
	auth, err := identity.NewAuthenticator(identity.Config{
		IntrospectionURL: os.Getenv("GOREECLOUD_NETWORK_IDENTITY_INTROSPECTION_URL"),
		ClientID:         os.Getenv("GOREECLOUD_NETWORK_IDENTITY_CLIENT_ID"),
		ClientSecret:     os.Getenv("GOREECLOUD_NETWORK_IDENTITY_CLIENT_SECRET"),
		ExpectedIssuer:   os.Getenv("GOREECLOUD_NETWORK_IDENTITY_EXPECTED_ISSUER"),
		ExpectedAudience: os.Getenv("GOREECLOUD_NETWORK_IDENTITY_EXPECTED_AUDIENCE"),
		RequiredScope:    getenv("GOREECLOUD_NETWORK_IDENTITY_ADMIN_SCOPE", defaultAdminScope),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialize GoreeCloud Identity acceptance verifier: %v\n", err)
		os.Exit(2)
	}

	report := identity.VerifyAdminRuntime(
		context.Background(),
		auth,
		os.Getenv("GOREECLOUD_NETWORK_IDENTITY_ACCEPTANCE_TOKEN"),
		time.Now(),
	)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "encode acceptance report: %v\n", err)
		os.Exit(2)
	}
	if report.State != "runtime_probe_passed_production_acceptance_pending" {
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
