package identity

import (
	"context"
	"errors"
	"strings"
	"time"
)

const AcceptanceEvidenceVersion = "goreecloud.network.identity.acceptance/v1"

type AcceptanceReport struct {
	EvidenceVersion    string `json:"evidenceVersion"`
	Authority          string `json:"authority"`
	Protocol           string `json:"protocol"`
	State              string `json:"state"`
	RuntimeConfigured  bool   `json:"runtimeConfigured"`
	ProbeExecuted      bool   `json:"probeExecuted"`
	AdminScopeAccepted bool   `json:"adminScopeAccepted"`
	RequiredScope      string `json:"requiredScope"`
	IssuerValidation   string `json:"issuerValidation"`
	AudienceValidation string `json:"audienceValidation"`
	ProductionAccepted bool   `json:"productionAccepted"`
	CheckedAt          string `json:"checkedAt"`
	Detail             string `json:"detail"`
}

func VerifyAdminRuntime(ctx context.Context, auth AdminAuthenticator, bearerToken string, checkedAt time.Time) AcceptanceReport {
	report := AcceptanceReport{
		EvidenceVersion:    AcceptanceEvidenceVersion,
		Authority:          "goreecloud_identity",
		Protocol:           "oauth2_token_introspection",
		State:              "blocked_authenticator_unavailable",
		ProductionAccepted: false,
		CheckedAt:          checkedAt.UTC().Format(time.RFC3339Nano),
		Detail:             "No GoreeCloud Identity authenticator was supplied to the runtime acceptance verifier.",
	}
	if auth == nil {
		return report
	}

	status := auth.Status()
	report.Authority = status.Authority
	report.Protocol = status.Protocol
	report.RuntimeConfigured = status.RuntimeConfigured
	report.RequiredScope = status.RequiredScope
	report.IssuerValidation = status.IssuerValidation
	report.AudienceValidation = status.AudienceValidation

	if !status.RuntimeConfigured {
		report.State = "blocked_runtime_unconfigured"
		report.Detail = "The Network Identity consumer boundary is not configured for a runtime probe. No production acceptance is implied."
		return report
	}
	bearerToken = strings.TrimSpace(bearerToken)
	if bearerToken == "" {
		report.State = "blocked_probe_token_missing"
		report.Detail = "A short-lived GoreeCloud Identity access token is required to exercise the configured Network administrator boundary. The token is not retained or returned."
		return report
	}

	report.ProbeExecuted = true
	_, err := auth.AuthenticateAdmin(ctx, bearerToken)
	switch {
	case err == nil:
		report.State = "runtime_probe_passed_production_acceptance_pending"
		report.AdminScopeAccepted = true
		report.Detail = "The configured GoreeCloud Identity runtime authenticated the supplied short-lived token and Network administrator scope. This probe is evidence only; production SSO acceptance remains a separate gate."
	case errors.Is(err, ErrForbidden):
		report.State = "runtime_probe_failed_insufficient_scope"
		report.Detail = "GoreeCloud Identity authenticated the token but the required Network administrator scope was not accepted."
	case errors.Is(err, ErrUnauthenticated):
		report.State = "runtime_probe_failed_identity_rejected_token"
		report.Detail = "GoreeCloud Identity did not authenticate the supplied token under the configured issuer/audience boundary."
	case errors.Is(err, ErrNotConfigured):
		report.State = "blocked_runtime_unconfigured"
		report.RuntimeConfigured = false
		report.Detail = "The GoreeCloud Identity runtime configuration was unavailable when the acceptance probe executed."
	case errors.Is(err, ErrUnavailable):
		report.State = "runtime_probe_failed_identity_unavailable"
		report.Detail = "The configured GoreeCloud Identity introspection authority was unavailable or returned an unusable response."
	default:
		report.State = "runtime_probe_failed_unclassified"
		report.Detail = "The GoreeCloud Identity runtime probe failed without producing reusable credential material or a production acceptance claim."
	}
	return report
}
