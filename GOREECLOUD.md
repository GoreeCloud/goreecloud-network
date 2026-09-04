# GoreeCloud Network

GoreeCloud Network is the GoreeCloud-maintained private-networking platform based initially on NetBird.

Its umbrella capability identity is **GoreeCloud Conduit**. Conduit represents the controlled private paths that connect approved people, devices, services, servers, virtual machines, routed networks, and private resources across GoreeCloud. The application/product name remains GoreeCloud Network; Conduit names the capability system beneath it.

See `docs/goreecloud/conduit-capability-identity.md` for the capability identity and `docs/goreecloud/native-feature-replacement-plan.md` for the mandatory native-by-destination architecture.

## Current development state

This repository is in a controlled fork-to-native transition. The inherited NetBird implementation remains a temporary compatibility baseline while GoreeCloud progressively replaces inherited product features with first-party Conduit implementations.

The long-term destination is not a cosmetic or indefinitely maintained NetBird rebrand. Every inherited product feature is either a native replacement target or an explicitly retired feature. Standards-based low-level primitives such as WireGuard may remain dependencies when rewriting them would reduce safety or provide no technical benefit.

The existing production NetBird deployment is not replaced or modified merely because native replacement development has begun. Production migration requires separate compatibility, security, backup, recovery, DNS, routing, relay, enrollment, policy, client, and rollback validation.

## Upstream relationship

Upstream project: `netbirdio/netbird`

During migration, GoreeCloud preserves upstream history and continues reviewing upstream security fixes, networking corrections, protocol changes, dependency changes, and relevant compatibility work. Upstream changes are reviewed before adoption rather than merged blindly. The amount of inherited runtime code is expected to decrease as Conduit-native replacements reach acceptance.

## Licensing boundary

The inherited repository uses a mixed-license structure. The repository-level BSD-3-Clause license applies except to `management/`, `signal/`, `relay/`, and `combined/`, which are licensed under AGPLv3 according to the inherited upstream license notice. GoreeCloud will preserve applicable copyright, attribution, source-availability, and license requirements for inherited or derived code even as those components are progressively replaced.

## Development principles

- Use **native-by-destination** for inherited product features.
- Preserve working WireGuard and peer-connectivity behavior while replacing feature layers behind tested compatibility seams.
- Do not rewrite mature cryptographic primitives merely to make them appear first-party.
- Prefer GoreeCloud-owned contracts so new feature development does not deepen direct coupling to inherited NetBird product models.
- Keep network reachability separate from application authorization.
- Keep Caddy, DNS services, identity, application authentication, monitoring, and host/VLAN firewall responsibilities distinct from the private-networking platform.
- Maintain privacy-first, deny-by-default, self-hosted-first, and local-first defaults.
- Require migration, backup/restore, rollback, and exact-source evidence before retiring inherited stateful features.
- Use controlled branches and pull requests for GoreeCloud modifications.
- Keep reusable credentials, setup keys, tokens, private keys, and other secrets out of source control.

## Native feature ownership

The machine-readable migration registry is `native/conduit/features.json`. It tracks the current inherited feature, its Conduit-native destination, and migration status.

Primary native destinations include Conduit Control, API, Devices, Principals, Groups, Access, Posture, Resources, Routes, Enrollment, Signal, Relay, Path, DNS Delivery, Activity, Diagnostics, Console, Android, Desktop, Apple, and Release.

Analytics, marketing, upstream-cloud-only, commercial billing, marketplace, and similar features are not automatically reproduced. They are retired when they conflict with GoreeCloud privacy, self-hosting, ownership, or product scope.

## GoreeCloud product integration

The planned GoreeCloud-controlled experience uses:

- **GoreeCloud Conduit** as the Network umbrella capability identity for encrypted connectivity, enrollment, access paths, policies, private resources, routing, client connectivity, and network diagnostics.
- **Glaze UI** for GoreeCloud graphical interfaces.
- **Wardveil Security by GoreeCloud** for security-state presentation, risk indicators, enrollment review, policy-risk visibility, and related security workflows.
- **GoreeCloud Privacy Shield** as the shared platform privacy identity and contract authority for the explicitly declared `network-privacy` capability. GoreeCloud Network remains the networking runtime authority, and `privacy-shield/adapter.json` remains `production_approved=false` until Network-specific runtime acceptance is complete.
- **GoreeCloud Identity** where centralized authentication or identity-provider integration is appropriate.
- **GoreeCloud Beacon** for DNS filtering, private rewrites, recursive resolution, caching, DNSSEC, authoritative DNS, encrypted DNS, and DNS policy rather than absorbing DNS-service roles into this project.
- **Caddy** as the separate approved HTTPS gateway for web-service publication.
- **GoreeCloud Monitor** for broader availability and operational monitoring.
- **GoreeCloud Manager** for broader infrastructure administration and operational visibility.

Privacy Shield does not turn GoreeCloud Network into a Browser blocker, DNS filtering service, firewall, identity provider, or application-authorization layer. A future Privacy Shield status producer must remain minimized and must not export raw private network activity merely for central presentation.

Conduit does not change these responsibility boundaries. It names the GoreeCloud Network capability layer; it does not make Network authoritative for DNS, reverse proxying, identity-provider authentication, destination-application authorization, monitoring, or platform-wide security analysis.

## Repository family

- `GoreeCloud/goreecloud-network` — current inherited control-plane/agent foundation and the growing Conduit-native server-side replacement layer.
- `GoreeCloud/goreecloud-network-dashboard` — current dashboard-derived administration source moving toward Conduit Console.
- `GoreeCloud/goreecloud-network-android` — current Android-derived source moving toward Conduit Android.
- `GoreeCloud/goreecloud-network-ios` — future Conduit Apple repository when implementation is approved.

Additional first-party repositories may be created only when separation provides a clear lifecycle, security, release, or platform benefit. Native replacement does not require unnecessary repository fragmentation.

## Current branch purpose

`agent/stable-foundation` is the initial controlled GoreeCloud development branch. It establishes provenance, governance, build/CI understanding, licensing boundaries, safe product customization points, and release-readiness gates.

The focused `agent/privacy-shield-network-adapter` branch is stacked on this stable foundation and adds Privacy Shield adapter governance only. It does not intentionally modify the inherited networking core or authorize production migration.

The focused `agent/conduit-capability-identity` branch is stacked on the Privacy Shield adapter branch. It now establishes both the Conduit capability identity and the native-by-destination replacement contract. It does not claim that the inherited networking runtime has already been fully replaced and does not authorize production migration.
