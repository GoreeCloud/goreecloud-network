# GoreeCloud Network Specifications

## Status

GoreeCloud Network is in active development. GoreeCloud Conduit is the first-party native-by-destination network capability system. The repository retains NetBird as a controlled compatibility bridge while native Conduit capabilities are implemented and accepted.

The canonical product record is maintained in Google Drive under `GoreeCloud/Projects/Project Specification — Network`. This file is the repository-local, version-coupled implementation specification and must not override the canonical project record.

Production authority has **not** migrated. NetBird remains production-authoritative until each capability clears isolated validation and an explicitly approved production migration gate.

## Native Conduit implementation boundary

Current GoreeCloud-owned Conduit source includes:

- native feature and migration registries;
- privacy-safe capability inventory and aggregate status;
- fail-closed capability transition contracts;
- protected inventory snapshots and compare-and-swap state transitions;
- immutable transition receipts and deterministic reconciliation;
- exact-source/artifact staging evidence and minimized staging status;
- first-party obfuscation negotiation with explicit requested/negotiating/active/unavailable/failed states;
- first-party `conduit-padded-wss` transport with bounded randomized relay-message padding inside authenticated WSS;
- Android exposure limited to coarse Conduit transport/runtime state;
- exact-source isolated status runtime acceptance;
- container-boundary runtime acceptance with no host publication of the private status listener;
- required-mode fail-closed startup behavior;
- minimized cross-platform acceptance-evidence consumption for Privacy Shield, Wardveil Security, Everkeep, GoreeCloud Mesh, GoreeCloud Identity, and governance.

WireGuard remains a narrow networking primitive where justified; it is not the GoreeCloud product identity or control plane.

## Platform acceptance-evidence contract

`goreecloud-conduit-platform-acceptance-evidence/v1` binds a platform evidence bundle to an exact 40-character Conduit source revision. Each required platform record contains only an expected platform identity, immutable SHA-256 evidence identity, and explicit accepted boolean.

The contract fails closed for a wrong schema, invalid source identity, wrong platform identity, rejected/missing platform acceptance, malformed evidence digest, or any attempt to set `production_cutover_authorized=true`. When the complete minimized bundle validates, it can set the corresponding platform booleans in the existing isolated-acceptance contract.

This source contract does not prove that a target environment has produced or approved those platform-system evidence artifacts. It defines how Conduit may consume them without importing private peer, route, policy, packet, DNS, credential, user, or unrestricted diagnostic state.

## Migration acceptance boundary

Production migration remains blocked by capability-specific evidence including state migration, backup/restore, rollback, client networking, security/privacy, platform systems, exact runtime artifacts, representative restrictive-network behavior, signed Android artifacts where applicable, approved real-device acceptance, final immutable staging/deployment provenance, and explicit production approval.

## Authority boundaries

- GoreeCloud Network / Conduit governs private connectivity and network capability orchestration within its accepted scope.
- GoreeCloud DNS / Beacon governs DNS resolution and DNS policy.
- GoreeCloud Gateway governs ingress and controlled service publication.
- Privacy Shield, Wardveil Security, Everkeep, GoreeCloud Mesh, GoreeCloud Identity, and governance retain their own platform responsibilities.

## Production non-claims

- NetBird authority has not been transferred.
- Platform evidence consumption does not itself prove platform acceptance.
- `conduit-padded-wss` does not make ordinary relay, QUIC, Force Relay, P2P, or WireGuard equivalent transports.
- Current isolated/container runtime evidence does not prove production network parity, migration completion, or production cutover.
