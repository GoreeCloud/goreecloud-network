# GoreeCloud Identity Runtime Acceptance Verifier

**Lifecycle:** Development

This verifier is a bounded evidence tool for the existing GoreeCloud Network → GoreeCloud Identity administrative-authentication consumer boundary. It exists because source-level OAuth token-introspection support is not sufficient evidence that an approved Identity application registration and runtime behave correctly.

## Purpose

`cmd/network-identity-verify` exercises the same `identity.AdminAuthenticator` boundary used by the Network administrative-session probe. It accepts the existing Identity configuration variables plus one short-lived probe token supplied at execution time through `GOREECLOUD_NETWORK_IDENTITY_ACCEPTANCE_TOKEN`.

The verifier emits a JSON evidence report containing only configuration state, validation-mode state, the required Network scope, probe result, timestamp, and an explicit `productionAccepted: false` value. It does not return or retain the bearer token, client secret, subject identifier, username, or other reusable credential material.

## Success semantics

A successful probe means only that, at that moment:

- the Network Identity consumer boundary was fully configured;
- the configured introspection authority responded through the existing fail-closed authentication path;
- the supplied short-lived token was active under the configured issuer/audience requirements; and
- the required Network administrative scope was accepted.

The success state is intentionally named `runtime_probe_passed_production_acceptance_pending`.

A successful probe does **not** establish production SSO acceptance, disabled-account propagation, session-expiration behavior, logout behavior, MFA behavior, account mapping, recovery, rollback, outage handling, or any other application-integration gate required by the authoritative GoreeCloud Identity architecture.

## Failure states

The report distinguishes:

- unconfigured runtime;
- missing ephemeral probe token;
- Identity rejection of the token;
- insufficient Network administrative scope;
- Identity authority unavailability; and
- unclassified verifier failure.

All failure states remain fail-closed.

## Execution

Configure the existing Development Identity variables and inject a short-lived acceptance token only for the command invocation. Do not commit or persist the token in repository configuration.

```bash
GOREECLOUD_NETWORK_IDENTITY_ACCEPTANCE_TOKEN='short-lived-token' \
  go run ./cmd/network-identity-verify
```

Exit status is `0` only for the successful Development runtime probe state, `1` for a completed/blocked probe that does not pass, and `2` for verifier initialization or evidence-encoding errors.

## Authority boundary

GoreeCloud Identity remains authoritative for identity establishment and identity-provided authority claims. GoreeCloud Network remains responsible for mapping accepted claims to Network-domain administrative operations. This verifier does not create users, sessions, tokens, client registrations, permissions, policies, or Network configuration.

Administrative mutation remains blocked until the Identity application/runtime integration receives the required independent acceptance evidence and the Network revision-management transaction layer is implemented.
