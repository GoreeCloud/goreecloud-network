# Current Features

Verified in the current Development implementation:

- Native Go Network Server process with loopback-safe default binding.
- Versioned JSON API under `/api/v1`.
- Liveness endpoint.
- Truthful product/surface status endpoint.
- Truthful seven-Platform-System state endpoint.
- Capability inventory endpoint that distinguishes implemented Development primitives from unavailable capabilities.
- Native volatile control-plane state kernel.
- Read-only Development device inventory; the initial inventory is empty until state is explicitly supplied by Development code.
- Development overview reporting device/resource/policy counts and authority-state labels.
- Deterministic deny-by-default access-decision evaluator with explicit matching allow-policy support.
- Browser-native Web Dashboard consuming live status and control-plane overview state.
- Native Android Development client consuming live server/control-plane state.
- Native Google TV Development client consuming live server/control-plane state.
- Native iOS Swift core and SwiftUI Development client consuming live server/control-plane state.
- Repository-level platform contract and CI definition.

The access evaluator is **decision-only**. It does not enforce packets, create network reachability, authenticate callers, persist policy, or provision tunnels.

Not implemented by the current Development state:

- WireGuard tunnel lifecycle or packet-flow enforcement.
- Device enrollment or key issuance.
- Durable policy/resource/device persistence and migrations.
- Authenticated policy administration.
- Production Zero Trust enforcement on live network traffic.
- Relay/signaling operation.
- Obfuscation transports.
- Private route advertisement/withdrawal.
- Posture enforcement.
- Production Platform System integrations.
- Production deployment or Stable release.
