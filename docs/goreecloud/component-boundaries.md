# GoreeCloud Network Component Boundaries

## Purpose

This record defines the initial compatibility and customization boundaries of the upstream-derived GoreeCloud Network repository. The objective is to separate product presentation and GoreeCloud integration work from security-sensitive networking behavior during the maintained-fork foundation phase.

## Repository licensing boundary

The inherited repository uses a mixed-license model:

- Most repository code is covered by the inherited BSD-3-Clause license.
- `management/`, `signal/`, `relay/`, and `combined/` are covered by the inherited GNU Affero General Public License v3 terms recorded in those directories.

GoreeCloud must preserve upstream notices, provenance, modification history, and the applicable license for every component.

## Client and endpoint boundary

The `client/` tree contains compatibility-critical endpoint behavior and platform-specific integration. The current tree includes areas for Android and iOS integration, command-line behavior, configuration, firewall handling, gRPC, interface/TUN-related work, and other client runtime functions.

During the initial GoreeCloud phases, this area is treated as a networking-core boundary rather than a branding surface. Product strings, packaging, diagnostics presentation, and GoreeCloud integrations may be changed only when they do not alter protocol or tunnel semantics.

Changes requiring dedicated regression testing include:

- WireGuard/interface lifecycle.
- TUN creation and teardown.
- DNS application and restoration.
- Firewall behavior.
- Route calculation and application.
- Peer connectivity.
- Management/control-plane communication.
- Signal/relay negotiation.
- Network-change recovery.
- Setup-key and interactive enrollment behavior.

## Management boundary

The `management/` component is the control-plane authority for network state and administration behavior inherited from NetBird. GoreeCloud dashboard/API integration depends on this boundary.

During the foundation phase, GoreeCloud should prefer compatibility-preserving extensions and presentation changes over broad control-plane rewrites. Any modification affecting identity, peers, groups, policies, routes, DNS, setup keys, posture, networks, or authorization must be validated against an isolated self-hosted deployment before production consideration.

## Signal boundary

The `signal/` component participates in peer connection coordination. It is protocol-sensitive infrastructure and is not a Glaze UI or branding target.

GoreeCloud should retain upstream-compatible signaling behavior until there is a specific, tested reason to diverge.

## Relay boundary

The `relay/` component provides connectivity support when direct peer paths are unavailable or unsuitable. Relay behavior is availability- and connectivity-sensitive and remains outside early branding work.

Any future Wardveil visibility around relay usage should observe and report state without changing relay selection or transport semantics unless separately designed and regression tested.

## Combined-server boundary

The `combined/` component packages multiple server responsibilities into the inherited combined deployment model. GoreeCloud may use or adapt this deployment model where appropriate, but deployment simplification must not erase logical responsibility boundaries among management, signaling, relay, and other supporting services.

## Dashboard relationship

The separate `GoreeCloud/goreecloud-network-dashboard` repository is the primary user-interface transformation surface. Glaze UI, GoreeCloud administrator workflows, Wardveil presentation, and privacy cleanup should normally be implemented there before considering equivalent changes in protocol-sensitive server code.

The dashboard must continue to use documented management APIs rather than gaining direct access to networking internals.

## Android relationship

The separate `GoreeCloud/goreecloud-network-android` repository is the dedicated GoreeCloud Android product surface. It currently delegates compatibility-critical networking behavior into the inherited Go mobile client. Mobile presentation and enrollment UX may evolve independently while that core boundary remains stable.

## Wardveil Security boundary

Wardveil Security may initially provide:

- Connection and peer trust/status presentation.
- Enrollment and device-security context.
- Policy-risk and broad-access warnings.
- Relay/direct-path visibility.
- DNS-protection status.
- Security-event presentation and links to appropriate GoreeCloud administration workflows.

Wardveil must not be implemented by weakening, bypassing, or silently replacing existing authentication, authorization, firewall, WireGuard, DNS, or routing controls.

## Glaze UI boundary

Glaze UI belongs primarily in user-facing dashboard and client repositories. Networking services, wire protocols, CLI contracts, and server APIs should not be changed merely to support visual branding.

Where a UI requires new data, GoreeCloud should first determine whether the required state is already available through existing APIs. New API surface should be minimal, authenticated, authorization-aware, documented, and covered by tests.

## Production safety boundary

The current GoreeCloud production NetBird deployment remains authoritative during development. No fork component replaces production until a controlled acceptance process verifies at minimum:

- Enrollment and revocation.
- Peer connectivity.
- Direct and relayed connectivity.
- DNS behavior and recovery.
- Routes and network resources.
- Groups and access policies.
- Administrative authentication.
- Dashboard/API compatibility.
- Android client compatibility.
- Restart and upgrade behavior.
- Backup and recovery.
- Emergency administrative access independent of the new deployment.

## Fork-to-native direction

GoreeCloud may progressively replace inherited components when doing so provides a clear security, privacy, maintainability, usability, or independence benefit. Replacement should occur one bounded component at a time with a compatibility contract and regression evidence.

GoreeCloud should not reimplement WireGuard or other mature cryptographic primitives merely for ownership or branding purposes.

## Next engineering gates

1. Record an exact tested upstream release/commit baseline.
2. Establish repeatable source build validation for server and client targets.
3. Validate the privacy-clean dashboard and Android branches.
4. Test the fork against an isolated self-hosted management/control-plane environment.
5. Begin Glaze UI and Wardveil presentation work at the dashboard/client layer.
6. Introduce control-plane changes only when a documented GoreeCloud requirement cannot be satisfied safely through existing interfaces.
