# Current Features

Verified in this bootstrap branch:

- Native Go Network Server process with loopback-safe default binding.
- Versioned read-only JSON API under `/api/v1`.
- Liveness endpoint.
- Truthful product/surface status endpoint.
- Truthful seven-Platform-System state endpoint.
- Capability inventory endpoint that distinguishes bootstrap from unavailable capabilities.
- Browser-native Web Dashboard consuming the live status API.
- Native Android source bootstrap.
- Native Google TV source bootstrap.
- Native iOS SwiftUI source bootstrap.
- Repository-level platform contract and CI definition.

Not implemented by this bootstrap:

- WireGuard tunnel lifecycle.
- Device enrollment or key issuance.
- Zero Trust policy enforcement.
- Relay/signaling operation.
- Obfuscation transports.
- Private route advertisement/withdrawal.
- Posture enforcement.
- Production authentication/authorization.
- Production persistence or migrations.
- Production Platform System integrations.
- Production deployment or Stable release.
