# Current Features

Verified in the current Development implementation:

- Native Go Network Server process with loopback-safe default binding.
- Versioned JSON API under `/api/v1`.
- Liveness endpoint.
- Truthful product/surface status endpoint.
- Truthful seven-Platform-System state endpoint.
- Capability inventory endpoint that distinguishes implemented Development primitives from unavailable capabilities.
- Native control-plane state kernel with deterministic snapshot serialization.
- Revisioned Development JSON file store with schema version 1.
- Explicit persistence migration ledger and a tested legacy-unversioned-to-v1 migration path.
- SHA-256 integrity verification for stored control-plane envelopes.
- Atomic same-directory file replacement with `0600` state-file permissions and `0700` created state directories.
- Read-only Development storage-status endpoint exposing persistence mode, schema, revision, migration count, last persistence timestamp, and integrity scheme.
- Read-only Development device inventory restored from the Development file store on normal server startup.
- Development overview reporting device/resource/policy counts plus storage and authority state.
- Deterministic deny-by-default access-decision evaluator with explicit matching allow-policy support.
- Browser-native Web Dashboard consuming live status, control-plane overview, and persisted revision metadata.
- Native Android Development client consuming live server/control-plane persistence state.
- Native Google TV Development client consuming live server/control-plane persistence state.
- Native iOS Swift core and SwiftUI Development client consuming live server/control-plane persistence state.
- Repository-level platform contract and exact-revision CI definition.

The Development file store is a single-process foundation. It is not a production database, distributed consensus system, backup system, or Everkeep integration.

The access evaluator is **decision-only**. It does not enforce packets, create network reachability, authenticate callers, or provision tunnels.

Not implemented by the current Development state:

- Authenticated policy/device/resource administration.
- Production database/high-availability persistence, snapshots, restore, rollback, or multi-writer coordination.
- WireGuard tunnel lifecycle or packet-flow enforcement.
- Device enrollment or key issuance.
- Production Zero Trust enforcement on live network traffic.
- Relay/signaling operation.
- Obfuscation transports.
- Private route advertisement/withdrawal.
- Posture enforcement.
- Production Platform System integrations.
- Production deployment or Stable release.
