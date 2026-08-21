# GoreeCloud Network

GoreeCloud Network is the GoreeCloud-maintained private-networking fork based on NetBird.

Its umbrella capability identity is **GoreeCloud Conduit**. Conduit represents the controlled private paths that connect approved people, devices, services, servers, virtual machines, routed networks, and private resources across GoreeCloud. The application/product name remains GoreeCloud Network; Conduit names the capability system beneath it.

See `docs/goreecloud/conduit-capability-identity.md` for the authoritative repository-local capability identity, scope, presentation guidance, and responsibility boundaries.

## Current development state

This repository is in the fork-foundation phase. The inherited NetBird networking behavior remains the compatibility baseline while GoreeCloud establishes its own product identity, governance, security presentation, integrations, release process, and long-term fork-to-native path.

The existing production NetBird deployment is not replaced or modified by source-development work in this repository. Production migration requires separate compatibility, security, backup, recovery, DNS, routing, relay, enrollment, policy, and rollback validation.

## Upstream relationship

Upstream project: `netbirdio/netbird`

GoreeCloud preserves upstream history and will continue to review upstream security fixes, networking corrections, protocol changes, dependency changes, and relevant compatibility work. Upstream changes must be reviewed before adoption rather than merged blindly.

## Licensing boundary

The inherited repository uses a mixed-license structure. The repository-level BSD-3-Clause license applies except to `management/`, `signal/`, `relay/`, and `combined/`, which are licensed under AGPLv3 according to the inherited upstream license notice. GoreeCloud will preserve applicable copyright, attribution, source-availability, and license requirements.

## Development principles

- Preserve working WireGuard and peer-connectivity behavior before changing low-level networking code.
- Prefer product-shell, integration, administration, security-presentation, and client improvements before replacing mature networking primitives.
- Keep network reachability separate from application authorization.
- Keep Caddy, DNS services, identity, application authentication, and host/VLAN firewall responsibilities distinct from the private-networking platform.
- Maintain privacy-first and self-hosted-first defaults.
- Use controlled branches and pull requests for GoreeCloud modifications.
- Keep reusable credentials, setup keys, tokens, private keys, and other secrets out of source control.

## GoreeCloud product integration

The planned GoreeCloud-controlled experience uses:

- **GoreeCloud Conduit** as the Network umbrella capability identity for encrypted connectivity, enrollment, access paths, policies, private resources, routing, client connectivity, and network diagnostics.
- **Glaze UI** for GoreeCloud graphical interfaces.
- **Wardveil Security by GoreeCloud** for security-state presentation, risk indicators, enrollment review, policy-risk visibility, and related security workflows.
- **GoreeCloud Privacy Shield** as the shared platform privacy identity and contract authority for the explicitly declared `network-privacy` capability. GoreeCloud Network remains the networking runtime authority, and `privacy-shield/adapter.json` remains `production_approved=false` until Network-specific runtime acceptance is complete.
- **GoreeCloud Identity** where centralized authentication or identity-provider integration is appropriate.
- **GoreeCloud Beacon** for DNS filtering, private rewrites, recursive resolution, caching, DNSSEC, authoritative DNS, encrypted DNS, and DNS policy rather than absorbing DNS-service roles into this project.
- **Caddy** as the separate approved HTTPS gateway for web-service publication.

Privacy Shield does not turn GoreeCloud Network into a Browser blocker, DNS filtering service, firewall, identity provider, or application-authorization layer. A future Privacy Shield status producer must remain minimized and must not export raw private network activity merely for central presentation.

Conduit does not change these responsibility boundaries. It names the GoreeCloud Network capability layer; it does not make Network authoritative for DNS, reverse proxying, identity-provider authentication, destination-application authorization, monitoring, or platform-wide security analysis.

## Initial repository family

- `GoreeCloud/goreecloud-network` — control plane, agents, shared networking foundation, and server-side components.
- `GoreeCloud/goreecloud-network-dashboard` — dedicated GoreeCloud administration web interface.
- `GoreeCloud/goreecloud-network-android` — dedicated Android client.
- `GoreeCloud/goreecloud-network-ios` — future Apple-platform client when implementation is approved.

## Current branch purpose

`agent/stable-foundation` is the initial controlled GoreeCloud development branch. The goal of this branch is to establish provenance, governance, build/CI understanding, licensing boundaries, safe product customization points, and release-readiness gates before material networking behavior is changed.

The focused `agent/privacy-shield-network-adapter` branch is stacked on this stable foundation and adds Privacy Shield adapter governance only. It does not intentionally modify the inherited networking core or authorize production migration.

The focused `agent/conduit-capability-identity` branch is stacked on the Privacy Shield adapter branch and adds the Conduit capability identity and documentation only. It does not intentionally modify the inherited networking core or authorize production migration.
