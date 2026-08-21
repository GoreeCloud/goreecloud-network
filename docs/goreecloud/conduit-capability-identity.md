# GoreeCloud Conduit

## Purpose

GoreeCloud Conduit is the umbrella capability identity for GoreeCloud Network.

Conduit represents the controlled private paths that connect approved people, devices, services, servers, virtual machines, routed networks, and private resources across GoreeCloud. It describes the complete GoreeCloud Network capability layer without replacing the GoreeCloud Network product name or absorbing responsibilities owned by other platform components.

## Identity

- Official capability name: **GoreeCloud Conduit**
- Short name: **Conduit**
- Parent product: **GoreeCloud Network**
- Product role: Private networking, encrypted connectivity, access paths, routing, enrollment, and network access control
- Design language: **Glaze UI**
- Security identity: **Wardveil Security by GoreeCloud**
- Privacy identity: **GoreeCloud Privacy Shield**
- DNS capability identity: **GoreeCloud Beacon**

## Capability Scope

Conduit covers the first-party GoreeCloud Network experience for:

- Encrypted peer-to-peer and relayed private connectivity.
- Device and peer enrollment, approval, ownership, revocation, and lifecycle management.
- People, device, service-identity, and resource relationships used for network access.
- Purpose-based groups and deny-by-default Access Policies.
- Posture-based access conditions.
- Private resources, networks, routed subnets, and routing peers.
- Site-to-site, routing-device, clientless-device, travel, and remote-access paths where approved.
- Network-managed DNS delivery to approved clients while preserving GoreeCloud Beacon as the DNS authority.
- Direct-versus-relayed path visibility and privacy-safe connection diagnostics.
- Administrative network workflows in the GoreeCloud Network dashboard.
- Dedicated GoreeCloud Network client experiences, beginning with Android and expanding to other supported platforms.
- Privacy Shield `network-privacy` integration.
- Wardveil Security network-security presentation and actionable network-risk workflows.
- GoreeCloud Manager and Monitor integrations that consume minimized, purpose-limited Network status.

## Responsibility Boundaries

Conduit does not turn GoreeCloud Network into a general platform catch-all.

- **GoreeCloud Beacon** remains authoritative for DNS filtering, private rewrites, recursive resolution, caching, DNSSEC, authoritative DNS, encrypted DNS, and DNS policy.
- **Wardveil Security** remains the platform security identity and security-oriented analysis/presentation layer.
- **GoreeCloud Privacy Shield** remains the shared privacy identity and privacy-capability contract layer.
- **GoreeCloud Identity** remains authoritative for identity-provider, authentication, and SSO responsibilities where integrated.
- **Caddy** remains the approved HTTPS gateway and reverse-proxy authority.
- **GoreeCloud Monitor** remains authoritative for service availability and operational monitoring.
- **GoreeCloud Manager** remains the broader infrastructure administration and operations console.
- Destination applications remain authoritative for their own authentication, authorization, and application data.

Network reachability is not application authorization. Conduit may create an approved path to a destination, but it does not grant permission inside the destination application.

## Product Presentation

The primary application name remains **GoreeCloud Network**. Conduit is the named capability system beneath that product, comparable to other GoreeCloud umbrella identities such as Glaze UI, Wardveil Security, Privacy Shield, Everkeep, Waypoint, Beacon, and Quill.

Recommended presentation patterns include:

- `GoreeCloud Network — powered by Conduit`
- `Conduit connection`
- `Conduit access path`
- `Conduit routing`
- `Conduit device enrollment`
- `Conduit Access Policies`
- `Conduit network privacy`

The short name should be used only where the surrounding interface already makes the GoreeCloud Network context clear.

## Design Direction

Conduit should feel like a calm, dependable private connective fabric rather than a generic consumer VPN brand. Glaze UI presentation should emphasize:

- clear connected/disconnected state;
- understandable path and resource relationships;
- minimal cognitive load;
- strong accessibility;
- restrained topology and route visualization;
- privacy-safe diagnostics;
- explicit security and authorization boundaries;
- clear recovery and revocation controls.

## Technical Direction

The Conduit identity does not authorize protocol rewrites by itself. During the maintained-fork phase, mature NetBird and WireGuard behavior remains compatibility-sensitive. Product, integration, workflow, privacy, security, observability, and client layers should become progressively more GoreeCloud-native where that creates a measurable improvement and passes the existing acceptance, recovery, and production-readiness gates.

## Acceptance Boundary

The Conduit identity is a product-capability naming and architecture decision. It does not change production networking state and does not authorize production migration.

Any Conduit-branded runtime claim must remain bound to the same exact-source, artifact, isolated-runtime, signed-client, real-device, backup/restore, rollback, and production-acceptance requirements already defined for GoreeCloud Network.
