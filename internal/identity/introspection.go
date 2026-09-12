package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const maxIntrospectionResponseBytes = 64 << 10

var (
	ErrNotConfigured   = errors.New("GoreeCloud Identity authentication is not configured")
	ErrUnauthenticated = errors.New("GoreeCloud Identity did not authenticate the bearer token")
	ErrForbidden       = errors.New("GoreeCloud Identity authority is insufficient for Network administration")
	ErrUnavailable     = errors.New("GoreeCloud Identity authentication is unavailable")
)

type Config struct {
	IntrospectionURL string
	ClientID         string
	ClientSecret     string
	ExpectedIssuer   string
	ExpectedAudience string
	RequiredScope    string
	HTTPClient       *http.Client
}

type Status struct {
	Authority          string `json:"authority"`
	Protocol           string `json:"protocol"`
	State              string `json:"state"`
	RuntimeConfigured  bool   `json:"runtimeConfigured"`
	RequiredScope      string `json:"requiredScope"`
	IssuerValidation   string `json:"issuerValidation"`
	AudienceValidation string `json:"audienceValidation"`
	ProductionAccepted bool   `json:"productionAccepted"`
	Detail             string `json:"detail"`
}

type Principal struct {
	Subject  string   `json:"subject"`
	Username string   `json:"username,omitempty"`
	Scopes   []string `json:"scopes"`
}

type AdminAuthenticator interface {
	Status() Status
	AuthenticateAdmin(context.Context, string) (Principal, error)
}

type Authenticator struct {
	configured       bool
	introspectionURL string
	clientID         string
	clientSecret     string
	expectedIssuer   string
	expectedAudience string
	requiredScope    string
	httpClient       *http.Client
}

func NewAuthenticator(config Config) (*Authenticator, error) {
	config.IntrospectionURL = strings.TrimSpace(config.IntrospectionURL)
	config.ClientID = strings.TrimSpace(config.ClientID)
	config.ClientSecret = strings.TrimSpace(config.ClientSecret)
	config.ExpectedIssuer = strings.TrimSpace(config.ExpectedIssuer)
	config.ExpectedAudience = strings.TrimSpace(config.ExpectedAudience)
	config.RequiredScope = strings.TrimSpace(config.RequiredScope)
	if config.RequiredScope == "" {
		return nil, errors.New("required Network administrator scope is required")
	}

	present := 0
	for _, value := range []string{config.IntrospectionURL, config.ClientID, config.ClientSecret} {
		if value != "" {
			present++
		}
	}
	if present != 0 && present != 3 {
		return nil, errors.New("Identity introspection URL, client ID, and client secret must be configured together")
	}

	auth := &Authenticator{
		configured:       present == 3,
		introspectionURL: config.IntrospectionURL,
		clientID:         config.ClientID,
		clientSecret:     config.ClientSecret,
		expectedIssuer:   config.ExpectedIssuer,
		expectedAudience: config.ExpectedAudience,
		requiredScope:    config.RequiredScope,
		httpClient:       config.HTTPClient,
	}
	if auth.httpClient == nil {
		auth.httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	if auth.configured {
		if err := validateIntrospectionURL(auth.introspectionURL); err != nil {
			return nil, err
		}
	}
	return auth, nil
}

func (a *Authenticator) Status() Status {
	status := Status{
		Authority:          "goreecloud_identity",
		Protocol:           "oauth2_token_introspection",
		State:              "identity_boundary_runtime_unconfigured",
		RuntimeConfigured:  false,
		RequiredScope:      a.requiredScope,
		IssuerValidation:   validationState(a.expectedIssuer),
		AudienceValidation: validationState(a.expectedAudience),
		ProductionAccepted: false,
		Detail:             "Network contains a GoreeCloud Identity OAuth introspection boundary, but no Identity runtime is configured for this process.",
	}
	if a.configured {
		status.State = "identity_introspection_configured_runtime_unaccepted"
		status.RuntimeConfigured = true
		status.Detail = "GoreeCloud Identity OAuth introspection is configured for this Development process. Production SSO/runtime acceptance has not been established."
	}
	return status
}

func (a *Authenticator) AuthenticateAdmin(ctx context.Context, bearerToken string) (Principal, error) {
	if !a.configured {
		return Principal{}, ErrNotConfigured
	}
	bearerToken = strings.TrimSpace(bearerToken)
	if bearerToken == "" {
		return Principal{}, ErrUnauthenticated
	}

	form := url.Values{}
	form.Set("token", bearerToken)
	form.Set("token_type_hint", "access_token")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.introspectionURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Principal{}, fmt.Errorf("%w: create introspection request: %v", ErrUnavailable, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "GoreeCloud-Network/0.1.0-dev.4")
	req.SetBasicAuth(a.clientID, a.clientSecret)

	response, err := a.httpClient.Do(req)
	if err != nil {
		return Principal{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return Principal{}, fmt.Errorf("%w: introspection returned HTTP %d", ErrUnavailable, response.StatusCode)
	}

	decoder := json.NewDecoder(io.LimitReader(response.Body, maxIntrospectionResponseBytes+1))
	var payload introspectionResponse
	if err := decoder.Decode(&payload); err != nil {
		return Principal{}, fmt.Errorf("%w: decode introspection response: %v", ErrUnavailable, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Principal{}, fmt.Errorf("%w: introspection response must contain exactly one JSON value", ErrUnavailable)
	}
	if !payload.Active || strings.TrimSpace(payload.Subject) == "" {
		return Principal{}, ErrUnauthenticated
	}
	if a.expectedIssuer != "" && strings.TrimSpace(payload.Issuer) != a.expectedIssuer {
		return Principal{}, ErrUnauthenticated
	}
	if a.expectedAudience != "" && !payload.Audience.Contains(a.expectedAudience) {
		return Principal{}, ErrUnauthenticated
	}

	scopes := normalizeScopes(payload.Scope)
	if !contains(scopes, a.requiredScope) {
		return Principal{}, ErrForbidden
	}
	username := strings.TrimSpace(payload.PreferredUsername)
	if username == "" {
		username = strings.TrimSpace(payload.Username)
	}
	return Principal{Subject: strings.TrimSpace(payload.Subject), Username: username, Scopes: scopes}, nil
}

type introspectionResponse struct {
	Active            bool     `json:"active"`
	Subject           string   `json:"sub"`
	Scope             string   `json:"scope"`
	Username          string   `json:"username"`
	PreferredUsername string   `json:"preferred_username"`
	Issuer            string   `json:"iss"`
	Audience          audience `json:"aud"`
}

type audience []string

func (a *audience) UnmarshalJSON(data []byte) error {
	var one string
	if err := json.Unmarshal(data, &one); err == nil {
		if one != "" {
			*a = []string{one}
		}
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return errors.New("audience must be a string or string array")
	}
	*a = many
	return nil
}

func (a audience) Contains(expected string) bool {
	for _, value := range a {
		if value == expected {
			return true
		}
	}
	return false
}

func normalizeScopes(raw string) []string {
	set := make(map[string]struct{})
	for _, scope := range strings.Fields(raw) {
		set[scope] = struct{}{}
	}
	scopes := make([]string, 0, len(set))
	for scope := range set {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)
	return scopes
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func validationState(value string) string {
	if value == "" {
		return "not_configured"
	}
	return "configured"
}

func validateIntrospectionURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid Identity introspection URL: %w", err)
	}
	if parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return errors.New("Identity introspection URL must contain a host and must not contain userinfo or a fragment")
	}
	if parsed.Scheme == "https" {
		return nil
	}
	if parsed.Scheme != "http" {
		return errors.New("Identity introspection URL must use HTTPS, except loopback Development endpoints may use HTTP")
	}
	host := parsed.Hostname()
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return nil
	}
	return errors.New("non-loopback Identity introspection URL must use HTTPS")
}
