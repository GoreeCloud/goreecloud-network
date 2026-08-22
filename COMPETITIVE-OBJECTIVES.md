# GoreeCloud Network — Competitive Objectives

This record defines the products and capability areas GoreeCloud Network uses as benchmarks for continuous improvement. It is a product-development target, not a claim that every benchmark capability is already implemented or production-accepted.

## Primary benchmark products

- NetBird
- Tailscale
- Twingate
- ZeroTier
- Headscale and other credible self-hosted private-network control-plane implementations

## Capabilities to match or improve

- Simple and secure device enrollment and revocation
- WireGuard-based or equivalently strong encrypted private connectivity
- Reliable direct-path establishment with NAT traversal and controlled relay fallback
- People, service identities, devices, groups, and ownership relationships
- Deny-by-default, least-privilege access policies
- Device posture and conditional access gates
- Private resources, subnet routing, site-to-site routing, and route failover
- Private DNS delivery and integration without turning Network into the DNS authority
- Web administration, stable APIs, automation, and diagnostics
- Android, desktop, and future Apple client experiences
- Activity history, auditability, troubleshooting, release, upgrade, rollback, and migration workflows

## Objectives GoreeCloud Network should exceed

- First-party control over the product experience, APIs, administrative model, clients, and release lifecycle
- Self-hosted operation without a required proprietary control plane
- Private-by-default and deny-by-default behavior for new people, devices, routes, and resources
- Clear separation between daily-use identities and administrative identities
- Strong integration with GoreeCloud Identity, GoreeCloud DNS, Wardveil Security, Privacy Shield, Monitor, Manager, and the wider Suite
- No required advertising, sponsorship, marketplace, billing, attribution, or analytics behavior
- Recoverable configuration, migration, rollback, and independently understandable operating state
- Glaze UI consistency and accessibility across GoreeCloud-owned client and administration surfaces

## Capabilities intentionally rejected or kept out of scope

- Commercial billing and marketplace workflows that do not serve GoreeCloud's private-network role
- Mandatory hosted-cloud attribution or analytics
- A general-purpose reverse proxy, DNS server, identity provider, password manager, or application authorization system inside Network
- Rewriting mature cryptographic or standards-based tunnel primitives merely to create artificial first-party ownership
- Broad network access granted automatically for convenience

## Competitive review rule

Competitor releases should be reviewed when they introduce a material private-networking capability. A gap should become a GoreeCloud objective only when it improves security, privacy, reliability, usability, accessibility, administrative control, portability, recovery, or interoperability without adding unjustified complexity.
