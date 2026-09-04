# GoreeCloud Network Benefits

These benefits are limited to implemented Conduit behavior and the accepted migration architecture.

## Native-by-destination migration

Conduit provides a controlled path from inherited NetBird authority toward GoreeCloud-owned network capabilities without requiring an unsafe all-at-once replacement. Capability state, migration stage, authority, evidence, and cutover are explicit rather than implicit.

## Safer transitions and recovery

Protected snapshots, compare-and-swap transitions, immutable receipts, deterministic reconciliation, staging evidence, backup/restore gates, rollback gates, and production-cutover=false contracts reduce ambiguity during migration and recovery.

## Privacy-minimized operational boundaries

Conduit status and acceptance surfaces are intentionally coarse. The platform evidence bundle carries only platform identity, immutable evidence hash, acceptance result, and exact source binding rather than peers, routes, policies, packet data, DNS queries, credentials, devices, users, or raw diagnostics.

## Explicit transport semantics

`conduit-padded-wss` is a distinct first-party transport with explicit negotiation and required-mode failure behavior. It does not silently rename or collapse ordinary relay, QUIC, Force Relay, P2P, or WireGuard behavior.

## Stronger evidence discipline

Exact-source runtime checks, immutable artifact identities, container-boundary testing, minimized evidence, and platform acceptance gates support traceable engineering decisions while keeping production authority separate from CI success.

## Platform integration without authority collapse

Conduit can consume acceptance results from Privacy Shield, Wardveil Security, Everkeep, GoreeCloud Mesh, GoreeCloud Identity, and governance without taking ownership of those systems or importing unnecessary private state.
