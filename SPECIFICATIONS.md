# GoreeCloud Network Repository Specifications

## Product contract

GoreeCloud Network is a native GoreeCloud product. Product-defining control-plane, client, policy, routing, diagnostics, lifecycle, and administrative behavior belongs to this repository and future GoreeCloud-owned dependencies. Mature standards-based primitives may remain bounded dependencies where justified.

## Required product surfaces

1. GoreeCloud Network Server.
2. Web Dashboard.
3. Android client.
4. Google TV client.
5. iOS client.

The five surfaces share the same API versioning, lifecycle vocabulary, capability state, and authority boundaries.

## Current Development API contract

The `v1` Development API exposes:

- `/healthz` — process liveness only.
- `/api/v1/status` — product version, lifecycle, and surface implementation state.
- `/api/v1/overview` — truthful control-plane counts plus persistence/authentication/policy-mode and revision metadata.
- `/api/v1/storage` — read-only persistence metadata.
- `/api/v1/authentication` — read-only GoreeCloud Identity consumer-boundary status without credentials.
- `/api/v1/admin/session` — bearer-token-protected, read-only administrative session probe using the configured GoreeCloud Identity introspection boundary and Network administrator scope.
- `/api/v1/devices` — read-only Development device inventory.
- `/api/v1/access/evaluate` — deterministic decision-only access evaluation; deny by default unless an explicit current allow rule matches.
- `/api/v1/platform-systems` — truthful integration state for all seven Integral Platform Systems.
- `/api/v1/capabilities` — capability availability without implying implementation that does not exist.

No current endpoint enrolls a device, issues keys, mutates policy/resources/devices, provisions routes, establishes a tunnel, selects a relay, enables obfuscation, or creates a security/privacy claim.

## GoreeCloud Identity administrative boundary

GoreeCloud Identity is the identity authority. Network must not establish a parallel user/password or token-issuer system.

The current Development consumer boundary uses OAuth 2.0 token introspection. It requires the introspection URL, confidential client ID, and client secret to be configured together. Non-loopback introspection uses HTTPS. Optional expected issuer and audience checks can be configured, and the Network-specific administrator scope defaults to `goreecloud.network.admin`.

A successfully introspected token must be active, contain a subject, satisfy configured issuer/audience checks, and contain the required Network administrator scope. Failure of any requirement denies the administrative session. Provider unavailability or missing runtime configuration fails closed.

Network returns only the minimum current session information required by the Development probe: subject, optional preferred username, and normalized scopes. Client credentials and bearer tokens are never returned by status APIs.

This source boundary is not production Identity acceptance. Independent runtime validation against the approved GoreeCloud Identity deployment is required before the integration may be represented as accepted or before Network exposes mutating administrative APIs.

## Development persistence contract

The current Development persistence package provides a bounded first revision/configuration-store slice with schema version 1, monotonic local revisions, migration ledger, deterministic snapshots, SHA-256 integrity verification, atomic replacement, and fail-closed loading. It remains Development-only and single-process.

## Access-decision contract

The Development evaluator is deliberately fail-closed:

1. Missing principal/resource context returns deny with `INVALID_CONTEXT`.
2. An explicit matching allow policy returns allow with `MATCHING_ALLOW_POLICY`.
3. Every other request returns deny with `NO_MATCHING_ALLOW_POLICY`.

This is an early native policy primitive, not production Zero Trust enforcement.

## Authority boundaries

- GoreeCloud Identity: platform identity/authentication and authority claims; Network maps accepted claims to Network-domain operations.
- Privacy Shield: privacy authorization and data-governance requirements.
- Wardveil Security: security/trust integration and evidence presentation.
- Everkeep: backup, restoration, migration, and continuity contracts.
- Glaze UI: interface semantics and accessibility requirements.
- GoreeCloud Mesh: approved service/capability coordination.
- GoreeCloud Manager: inventory, lifecycle, health, and administration surfaces.

GoreeCloud Network remains responsible for Network policy enforcement, private connectivity, routing, path selection, relay coordination, obfuscation orchestration, and Network-specific diagnostics.

## Development status

This repository is Development. Persistence, the Identity consumer boundary, control-plane kernel, inventory, decision evaluator, clients, dashboard, and CI do not prove production SSO, administrative mutation, tunnel, routing, relay, enrollment, obfuscation, production policy enforcement, production persistence/recovery, or full Platform System conformance.
