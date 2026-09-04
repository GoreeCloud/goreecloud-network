# GoreeCloud Network — Features

This file distinguishes current repository capabilities from planned GoreeCloud-native product work. Source presence does not by itself imply production acceptance.

## Current / inherited compatibility foundation

- Encrypted peer connectivity using the inherited NetBird/WireGuard networking foundation
- WireGuard tunnel management and VPN-service behavior
- Peer/device connectivity and status data
- Routing and routed-network mechanics
- NAT traversal, signaling, and relay behavior
- Management APIs and control-plane behavior inherited from the maintained-fork foundation
- People, service identities, groups, access policies, posture conditions, networks/resources, DNS presentation, activity, and settings in the administration experience
- Setup-key and self-hosted enrollment compatibility
- Dedicated GoreeCloud Network Android-client foundation
- GoreeCloud Network web-administration/dashboard foundation
- GoreeCloud-owned product identity and initial Glaze UI shell conversion
- GoreeCloud Devices and Access Policies presentation using least-privilege and deny-by-default terminology
- Privacy hardening that removes or disables inherited Google Analytics, Hotjar, Google Tag Manager, HubSpot, Firebase Analytics, Firebase Crashlytics, Advertising ID, and related commercial/attribution paths from the GoreeCloud product experience
- Self-hosted GoreeCloud management endpoint as the Android client's inherited default-management target
- Wardveil Security identity surfaces without unsupported static security claims

## Active first-party Conduit feature families

The long-term first-party feature system is organized under GoreeCloud Conduit. Implementation remains incremental and must be validated feature by feature.

- Conduit Control — control-plane orchestration and lifecycle coordination
- Conduit API — GoreeCloud-owned API contracts and compatibility adapters
- Conduit Devices — enrollment, ownership, approval, revocation, and device lifecycle
- Conduit Principals — people and service-identity references
- Conduit Groups — purpose-based organization
- Conduit Access — deny-by-default least-privilege policy management
- Conduit Posture — device-condition access gates
- Conduit Resources — private resources and reachability definitions
- Conduit Routes — routed networks, routing peers, priorities, failover, and site-to-site paths
- Conduit Enrollment — setup, approval, reauthentication, revocation, recovery, and migration workflows

## Planned / acceptance-gated capabilities

- Complete first-party replacement of inherited product-level control-plane and administrative features where technically justified
- Fully validated self-hosted authentication and management compatibility
- Complete Glaze UI conversion across owned administration and client surfaces
- Evidence-backed Wardveil Security status and security workflows
- Production-ready Android APK/AAB delivery and real-device validation
- Future Linux, desktop, and Apple clients where separately approved
- Controlled migration from the incumbent NetBird production environment
- Verified backup, recovery, upgrade, rollback, and release workflows for the GoreeCloud Network product family

## Scope boundary

GoreeCloud Network remains the private-networking platform. Caddy remains the reverse-proxy authority, GoreeCloud DNS remains the DNS authority, GoreeCloud Identity remains the authentication/identity-provider authority where integrated, and other Suite products remain authoritative for their own data and roles.
