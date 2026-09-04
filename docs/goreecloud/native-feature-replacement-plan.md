# GoreeCloud Conduit Native Feature Replacement Plan

## Decision

GoreeCloud Network will use **GoreeCloud Conduit** as the umbrella identity for its private-networking capabilities and will progressively replace inherited third-party product features with first-party GoreeCloud-native implementations.

This is a native-by-destination policy. An inherited feature may remain temporarily only as a compatibility bridge while its GoreeCloud replacement is implemented, tested, migrated, and proven recoverable. Temporary inheritance is not the long-term product state.

The policy distinguishes **product features** from **standards and low-level primitives**. GoreeCloud does not rewrite mature cryptography merely to remove a dependency name. WireGuard remains an acceptable standards-based tunnel primitive unless a separately approved technical decision replaces it. The GoreeCloud-owned feature layer around tunnel lifecycle, enrollment, policy, routing, relay selection, diagnostics, administration, and clients is nevertheless a native replacement target.

## Native destination

The long-term Conduit stack is intended to be GoreeCloud-owned at the feature and product-behavior layers:

| Current inherited capability | Native GoreeCloud destination | Migration contract |
| --- | --- | --- |
| NetBird management/control plane | Conduit Control | Native device, identity-reference, policy, network, route, posture, enrollment, audit, and configuration services |
| NetBird management REST/API behavior | Conduit API | GoreeCloud-owned versioned API with compatibility adapters during migration |
| NetBird peer/device model | Conduit Devices | Native device lifecycle, ownership, trust, approval, revocation, metadata minimization, and status |
| NetBird users/service users presentation | Conduit Principals | Native people-reference and service-identity network principals; authentication remains GoreeCloud Identity responsibility |
| NetBird groups | Conduit Groups | Native purpose-based grouping of devices, principals, resources, and policy targets |
| NetBird access policies | Conduit Access | Native deny-by-default network policy engine and review workflow |
| NetBird posture checks | Conduit Posture | Native device-condition policy gates with Wardveil presentation integration |
| NetBird networks/resources | Conduit Resources | Native private host, subnet, domain, and routed-resource definitions |
| NetBird routes/routing peers | Conduit Routes | Native route lifecycle, route priority, forwarding-device assignment, site-to-site and clientless-gateway workflows |
| NetBird setup keys | Conduit Enrollment | Native short-lived, purpose-limited enrollment credentials and approval flows |
| NetBird signaling | Conduit Signal | GoreeCloud-owned peer coordination/signaling service with compatibility transition support |
| NetBird relay | Conduit Relay | GoreeCloud-owned relay service and relay-selection/status logic |
| NetBird NAT traversal orchestration | Conduit Path | Native direct-path discovery/orchestration around approved standards and traversal mechanisms |
| NetBird DNS distribution | Conduit DNS Delivery | Native distribution of approved resolver/search-domain configuration; GoreeCloud Beacon remains DNS service authority |
| NetBird event/audit presentation | Conduit Activity | Native privacy-minimized network event records and review workflows |
| NetBird diagnostics bundles | Conduit Diagnostics | Native local-first diagnostics, redaction, bounded export, and evidence collection |
| NetBird desktop client product layer | Conduit Desktop | First-party GoreeCloud desktop client and system integration |
| NetBird Android product layer | Conduit Android | First-party Android UX, lifecycle, enrollment, resources, diagnostics, release identity, and platform integration |
| Future upstream Apple client dependency | Conduit Apple | First-party iOS/macOS client when implementation begins |
| NetBird dashboard product layer | Conduit Console | First-party Glaze UI administration application |
| NetBird update/release assumptions | Conduit Release | GoreeCloud-owned build, signing, provenance, update, migration, rollback, and support lifecycle |
| NetBird telemetry/commercial/cloud integrations | No replacement unless needed | Remove rather than recreate when the feature conflicts with GoreeCloud privacy, self-hosting, or non-commercial scope |

## Feature-equivalence rule

The objective is not byte-for-byte replication of NetBird. A replacement should preserve the useful capability and expected interoperability while adopting GoreeCloud requirements. Every replacement should normally provide the same core user outcome, then improve privacy, understandable administration, recovery, local-first operation, accessibility, observability, and GoreeCloud integration.

A third-party feature may be intentionally omitted rather than recreated when it is commercial-only, upstream-cloud-specific, analytics/marketing oriented, redundant with another GoreeCloud authority, or inconsistent with GoreeCloud privacy and ownership principles.

## Architecture boundaries

Conduit remains the private-connectivity authority. It does not absorb unrelated platform responsibilities:

- **GoreeCloud Beacon** remains authoritative for DNS filtering, recursive and authoritative DNS functions, DNSSEC, blocking, caching, and DNS privacy features. Conduit only distributes approved DNS configuration to connected devices.
- **GoreeCloud Identity** remains authoritative for authentication, SSO, MFA, and identity-provider functions. Conduit consumes verified identity assertions and manages network authorization.
- **Wardveil Security** remains the security identity and security-analysis/presentation layer. Conduit remains authoritative for enforcement of network access and device/network state.
- **GoreeCloud Privacy Shield** remains the shared privacy identity and privacy-contract layer. Conduit implements the `network-privacy` capability without exporting raw private activity merely for presentation.
- **Caddy** remains the approved HTTPS gateway and reverse-proxy authority.
- **GoreeCloud Monitor** remains the broader availability and operational-monitoring authority.
- **GoreeCloud Manager** remains the broader infrastructure-administration authority.

## Implementation waves

### Wave 1 — Native contracts and compatibility seams

Create GoreeCloud-owned data contracts and interfaces for Devices, Groups, Access, Enrollment, Resources, Routes, Posture, Activity, Diagnostics, Signal, Relay, and DNS Delivery. Existing inherited services may implement these interfaces temporarily through adapters. New GoreeCloud code must depend on the Conduit contracts rather than directly adding new product coupling to inherited NetBird-specific presentation models.

### Wave 2 — Console and client feature ownership

Complete Conduit Console and Android ownership so normal users and administrators no longer depend on inherited NetBird product terminology, hosted documentation, analytics, commercial surfaces, release identity, or third-party presentation components. Finalize GoreeCloud package/application identity and signing after coexistence and migration testing.

### Wave 3 — Native policy, enrollment, device, and resource services

Replace management-layer feature implementations for Devices, Groups, Access, Enrollment, Posture, Resources, Routes, Activity, and privacy-safe diagnostics behind compatibility APIs. Maintain migration import/export and rollback tooling throughout this wave.

### Wave 4 — Native coordination services

Replace inherited signaling, relay orchestration, direct-path coordination, and related control-plane behavior with Conduit Signal, Relay, and Path components. These changes require isolated multi-peer regression testing across direct, relayed, NAT-constrained, routed, restart, and failure scenarios.

### Wave 5 — Native cross-platform clients

Complete first-party desktop, Android, and Apple clients. Preserve the WireGuard protocol primitive where appropriate while owning tunnel lifecycle orchestration, profile state, DNS delivery, route application, firewall integration, resource selection, reconnect behavior, diagnostics, updates, and recovery at the GoreeCloud product layer.

### Wave 6 — Retire inherited feature runtime

Remove inherited feature implementations only after equivalent Conduit behavior has passed source, unit, integration, isolated-runtime, migration, backup/restore, rollback, and applicable real-device acceptance. Preserve required licenses, copyright notices, source history, and attribution for code that remains derived or was historically incorporated.

## Acceptance requirements for each replacement

A feature is not considered natively replaced because a new name or wrapper exists. Replacement requires:

1. A GoreeCloud-owned implementation exists behind a documented Conduit contract.
2. Required user outcomes are demonstrated against representative scenarios.
3. Privacy and security boundaries are at least as strong as the accepted previous behavior.
4. Import/migration behavior is proven where persisted state is involved.
5. Backup and restoration are proven where server-side state is involved.
6. Rollback to the previously accepted state is documented and rehearsed when the change affects production connectivity.
7. Failure modes do not create unintended public exposure or broad allow access.
8. Conduit Console and applicable clients correctly reflect state and errors.
9. Exact-source and artifact provenance is retained for acceptance evidence.
10. Production cutover is separately approved; source completion alone never authorizes migration.

## Current state

The current GoreeCloud Network repositories still contain substantial inherited NetBird implementation. Several third-party presentation, telemetry, hosted-policy, commercial-cloud, release, and Android dependencies have already been removed or replaced, but the control plane, signaling, relay, routing/tunnel orchestration, APIs, and portions of the Android networking core remain inherited.

This document therefore records the mandatory destination and migration architecture; it does not falsely classify the current repository as fully native.
