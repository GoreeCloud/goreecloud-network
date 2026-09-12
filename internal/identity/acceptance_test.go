package identity

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

type acceptanceFakeAuth struct {
	status    Status
	principal Principal
	err       error
}

func (f acceptanceFakeAuth) Status() Status { return f.status }
func (f acceptanceFakeAuth) AuthenticateAdmin(context.Context, string) (Principal, error) {
	return f.principal, f.err
}

func TestVerifyAdminRuntimeBlocksUnconfiguredRuntime(t *testing.T) {
	report := VerifyAdminRuntime(context.Background(), acceptanceFakeAuth{status: Status{Authority: "goreecloud_identity", Protocol: "oauth2_token_introspection", RequiredScope: "goreecloud.network.admin"}}, "token", time.Unix(0, 0))
	if report.State != "blocked_runtime_unconfigured" || report.ProbeExecuted || report.ProductionAccepted {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestVerifyAdminRuntimeRequiresEphemeralProbeToken(t *testing.T) {
	report := VerifyAdminRuntime(context.Background(), acceptanceFakeAuth{status: configuredAcceptanceStatus()}, " ", time.Unix(0, 0))
	if report.State != "blocked_probe_token_missing" || report.ProbeExecuted || report.AdminScopeAccepted {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestVerifyAdminRuntimeRecordsSuccessfulProbeWithoutProductionClaim(t *testing.T) {
	report := VerifyAdminRuntime(context.Background(), acceptanceFakeAuth{status: configuredAcceptanceStatus(), principal: Principal{Subject: "person-1", Username: "private-user-name", Scopes: []string{"goreecloud.network.admin"}}}, "ephemeral-secret-token", time.Unix(123, 456))
	if report.State != "runtime_probe_passed_production_acceptance_pending" || !report.ProbeExecuted || !report.AdminScopeAccepted || report.ProductionAccepted {
		t.Fatalf("unexpected report: %+v", report)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	if strings.Contains(text, "ephemeral-secret-token") || strings.Contains(text, "person-1") || strings.Contains(text, "private-user-name") {
		t.Fatalf("acceptance report leaked credential or principal data: %s", text)
	}
}

func TestVerifyAdminRuntimeClassifiesAuthorityAndAvailabilityFailures(t *testing.T) {
	cases := []struct {
		err   error
		state string
	}{
		{ErrForbidden, "runtime_probe_failed_insufficient_scope"},
		{ErrUnauthenticated, "runtime_probe_failed_identity_rejected_token"},
		{ErrUnavailable, "runtime_probe_failed_identity_unavailable"},
		{errors.New("unexpected"), "runtime_probe_failed_unclassified"},
	}
	for _, tc := range cases {
		report := VerifyAdminRuntime(context.Background(), acceptanceFakeAuth{status: configuredAcceptanceStatus(), err: tc.err}, "token", time.Unix(0, 0))
		if report.State != tc.state || !report.ProbeExecuted || report.ProductionAccepted {
			t.Fatalf("err %v: unexpected report %+v", tc.err, report)
		}
	}
}

func configuredAcceptanceStatus() Status {
	return Status{Authority: "goreecloud_identity", Protocol: "oauth2_token_introspection", State: "identity_introspection_configured_runtime_unaccepted", RuntimeConfigured: true, RequiredScope: "goreecloud.network.admin", IssuerValidation: "configured", AudienceValidation: "configured", ProductionAccepted: false}
}
