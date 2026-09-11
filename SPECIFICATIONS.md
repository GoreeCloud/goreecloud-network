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

## Initial API contract

The Development bootstrap exposes read-only discovery endpoints:

- `/healthz` — process liveness only.
- `/api/v1/status` — product version, lifecycle, and surface implementation state.
- `/api/v1/platform-systems` — truthful integration state for all seven Integral Platform Systems.
- `/api/v1/capabilities` — capability availability without implying implementation that does not exist.

No endpoint in this bootstrap grants tunnel access, enrolls a device, changes policy, provisions routes, or creates a security/privacy claim.

## Authority boundaries

- GoreeCloud Identity: identity and authorization context.
- Privacy Shield: privacy authorization and data-governance requirements.
- Wardveil Security: security/trust integration and evidence presentation.
- Everkeep: backup, restoration, migration, and continuity contracts.
- Glaze UI: interface semantics and accessibility requirements.
- GoreeCloud Mesh: approved service/capability coordination.
- GoreeCloud Manager: inventory, lifecycle, health, and administration surfaces.

GoreeCloud Network remains responsible for Network policy enforcement, private connectivity, routing, path selection, relay coordination, obfuscation orchestration, and Network-specific diagnostics.

## Development status

This repository is Development. Native client shells and server discovery contracts do not prove tunnel, policy, relay, routing, enrollment, obfuscation, or Platform System implementation.
