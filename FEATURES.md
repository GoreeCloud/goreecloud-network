# Current Features

Verified in the current Development implementation:

- Native Go Network Server process with loopback-safe default binding.
- Versioned JSON API under `/api/v1`.
- Liveness, product/surface status, Platform System state, and capability inventory endpoints.
- Native control-plane state kernel with deterministic snapshot serialization.
- Revisioned Development JSON file store with schema version 1, migration ledger, SHA-256 integrity verification, restrictive local permissions, and atomic replacement.
- Read-only Development storage and device inventory reporting.
- Deterministic deny-by-default access-decision evaluator with explicit matching allow-policy support.
- GoreeCloud Identity OAuth token-introspection consumer boundary for protected administrative routes.
- Fail-closed Identity configuration: partial core configuration is rejected; non-loopback introspection requires HTTPS; unconfigured/unavailable Identity cannot authorize an administrator.
- Optional exact issuer and audience validation plus required Network administrator scope.
- Read-only `/api/v1/authentication` status that does not expose client credentials.
- Protected read-only `/api/v1/admin/session` probe returning only subject, optional preferred username, and scopes after successful Identity authentication/Network authorization.
- Browser-native Web Dashboard plus native Android, Google TV, and iOS Development clients consuming truthful server/control-plane state.
- Repository-level platform contract and exact-revision CI definition.

The Identity boundary is **not** proof that a GoreeCloud Identity runtime is configured or production accepted. No Network password database, token issuer, or unauthenticated administrative mutation path is introduced.

The Development file store is a single-process foundation. It is not a production database, distributed consensus system, backup system, or Everkeep integration.

The access evaluator is **decision-only**. It does not enforce packets, create network reachability, or provision tunnels.

Not implemented by the current Development state:

- Runtime-accepted/production GoreeCloud Identity SSO integration.
- Authenticated mutating policy/device/resource administration.
- Production database/high-availability persistence, snapshots, restore, rollback, or multi-writer coordination.
- WireGuard tunnel lifecycle or packet-flow enforcement.
- Device enrollment or key issuance.
- Production Zero Trust enforcement on live network traffic.
- Relay/signaling operation, obfuscation transports, private routing, or posture enforcement.
- Production integration/acceptance for the remaining Integral Platform Systems.
- Production deployment or Stable release.
