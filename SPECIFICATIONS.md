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
- `/api/v1/storage` — read-only persistence mode, schema version, revision, migration count, last persisted timestamp, and integrity scheme.
- `/api/v1/devices` — read-only Development device inventory.
- `/api/v1/access/evaluate` — deterministic decision-only access evaluation; deny by default unless an explicit current allow rule matches.
- `/api/v1/platform-systems` — truthful integration state for all seven Integral Platform Systems.
- `/api/v1/capabilities` — capability availability without implying implementation that does not exist.

Normal server startup uses `development_file_store`. Authentication remains `not_implemented`. The access evaluator returns a decision and reason only; it does not enforce packet flow or establish connectivity.

No current endpoint enrolls a device, issues keys, mutates policy/resources/devices, provisions routes, establishes a tunnel, selects a relay, enables obfuscation, or creates a security/privacy claim.

## Development persistence contract

The current Development persistence package provides a bounded first revision/configuration-store slice:

1. JSON envelope format `goreecloud.network.controlplane`.
2. Explicit schema version; current version is `1`.
3. Monotonically increasing local revision number for successful writes.
4. Migration ledger with version, name, and application timestamp.
5. A tested migration from legacy unversioned (`schemaVersion = 0`) state to schema v1.
6. Deterministic device/resource/policy snapshot ordering.
7. SHA-256 integrity verification over format, schema, revision, migrations, and state.
8. Atomic same-directory temporary-file replacement and restrictive created-file permissions.
9. Fail-closed startup on unsupported newer schema, malformed state, invalid restored objects, or checksum mismatch.

This store is deliberately **Development-only** and single-process. It does not establish a production database, transactional multi-writer semantics, distributed locking/consensus, high availability, backup/restore, Everkeep conformance, or production rollback evidence.

No unauthenticated mutating administrative API is added by this persistence milestone. GoreeCloud Identity-backed authentication and explicit authorization boundaries must precede network-accessible mutation operations.

## Access-decision contract

The Development evaluator is deliberately fail-closed:

1. Missing principal/resource context returns deny with `INVALID_CONTEXT`.
2. An explicit matching allow policy returns allow with `MATCHING_ALLOW_POLICY`.
3. Every other request returns deny with `NO_MATCHING_ALLOW_POLICY`.

This is an early native policy primitive, not production Zero Trust enforcement. Authenticated administration, identity context, posture, runtime adapters, audit evidence, and packet-path enforcement remain required before a production claim.

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

This repository is Development. The implemented file-store foundation, control-plane kernel, inventory, decision evaluator, clients, dashboard, and CI do not prove tunnel, routing, relay, enrollment, obfuscation, production policy enforcement, production persistence/recovery, or Platform System conformance.
