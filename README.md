# GoreeCloud Network

GoreeCloud Network is GoreeCloud’s first-party private networking platform. Its native capability system is **GoreeCloud Conduit**.

Conduit is the coordination and connectivity layer for approved people, devices, services, servers, virtual machines, routed networks, and private resources. The long-term product is GoreeCloud-owned and native by destination. This repository still contains inherited NetBird compatibility code while individual capabilities are replaced, validated, migrated, and made recoverable.

WireGuard and other mature standards-based cryptographic, tunneling, protocol, operating-system, and interoperability foundations may remain where replacing them would materially increase risk or reduce compatibility. That exception does not authorize retaining inherited product architecture, product UI, branding, management workflows, or general application logic as the final GoreeCloud product.

## Native Conduit direction

The active Conduit migration model is machine-enforced through `native/conduit/features.json` and `native/conduit/migrations.json`. Replacement-required capabilities progress through explicit stages:

`inventory → contract → implementation → isolated-validation → migration-ready → production-candidate → native-accepted → inherited-path retirement`

Source metadata does not authorize production cutover. Compatibility bridges remain active until their native replacements pass all required gates.

Current first-party server-side work includes:

- Conduit Control and Conduit API contracts with inherited-authority coercion and `production_cutover_authorized=false`;
- a separate opt-in loopback-only management companion status endpoint for isolated staging;
- protected capability inventory snapshots and compare-and-swap transition semantics;
- immutable transition receipts and deterministic reconciliation behavior;
- immutable staging evidence bound to exact source and runtime-artifact identity;
- `goreecloud-conduit-capability-staging-status/v1`, a minimized durable-evidence contract for approved central GoreeCloud consumers;
- validation that rejects incomplete, unreconciled, malformed, non-inherited-authority, bridge-disabled, or cutover-authorizing staging status;
- a GET-only, `Cache-Control: no-store` staging-status handler that omits capability IDs, source revisions, runtime-artifact digests, peers, routes, policies, credentials, packet data, DNS-query detail, and raw operational state;
- an isolated staging harness with source-bound configuration and acceptance-evidence tooling.

The minimized staging-status handler is a source-level read-only boundary only. It is not mounted on a production listener by the current milestone.

## Migration acceptance boundary

GoreeCloud Network is not migration-ready merely because source, unit, integration, or container-build workflows pass. Capability migration still requires applicable retained evidence for:

- exact source revisions and immutable runtime artifacts;
- isolated runtime behavior and exposure boundaries;
- state migration and reconciliation;
- backup and restoration for authoritative state;
- rollback and emergency recovery;
- direct and relayed connectivity;
- enrollment, revocation, access policy, posture, routes, routed networks, DNS delivery, signaling, relay, and path behavior;
- signed client artifacts and approved signer identity where applicable;
- real-device networking acceptance for client applications;
- accessibility and current Stable Glaze UI conformance for GoreeCloud-controlled interfaces;
- Privacy Shield, Wardveil Security, Everkeep, and GoreeCloud Mesh integration where required;
- explicit production approval before inherited-path retirement.

The currently validated Glaze UI Stable baseline is 1.6.0. Candidate 2.0.0 is not accepted as the Stable migration baseline.

## Responsibility boundaries

- **GoreeCloud Network / Conduit** owns private connectivity, peer/device enrollment and lifecycle, groups, network access policy, posture, resources, routes, signaling/relay/path coordination, client networking, and privacy-safe network diagnostics.
- **GoreeCloud DNS / Beacon** owns DNS filtering, private rewrites, recursive resolution, caching, DNSSEC, authoritative/private DNS, encrypted DNS, and DNS policy. Conduit may deliver approved DNS configuration but does not absorb resolver authority.
- **GoreeCloud Gateway** owns HTTPS ingress, reverse proxying, TLS termination, and controlled service publication.
- **GoreeCloud Identity** remains authoritative for identity-provider and authentication functions where integrated.
- **Wardveil Security** supplies evidence-backed security posture and security experiences around Network without replacing Network policy enforcement.
- **Privacy Shield** supplies privacy contracts and data-minimization requirements.
- **Everkeep** supplies backup, restoration, recovery, portability, and continuity requirements.
- **GoreeCloud Mesh** coordinates approved platform-level integration and governance without becoming the networking data plane.
- **Glaze UI** governs GoreeCloud-controlled user interfaces.

Network reachability never grants destination-application authorization by itself.

## Isolated staging

`deploy/goreecloud-staging/` defines a non-production staging topology with loopback-first host publication, exact image-reference requirements, staging-only configuration, source-bound evidence manifests, and explicit rejection of known production markers.

The management container may expose the Conduit status companion only on `127.0.0.1:9097` inside that container. Port 9097 is not published to the staging host. Formal isolated-runtime evidence must therefore prove the intended local-only boundary from within the staging runtime.

## Dedicated clients

GoreeCloud Network uses dedicated GoreeCloud client applications rather than treating upstream NetBird clients as the permanent user experience. The Android client has separate source, artifact, signing, emulator, and real-device acceptance requirements. Inherited clients may remain only as compatibility and recovery paths until native clients are accepted.

## Upstream provenance

The compatibility tree originated from NetBird and remains subject to all applicable licenses, notices, source history, and attribution. This repository includes code under the BSD-3-Clause license and directories/components with additional licensing such as AGPLv3; the repository’s existing license files and third-party notices remain authoritative for inherited code.

The native transition must preserve required provenance while progressively replacing inherited product features with original GoreeCloud Conduit implementations.

## Production boundary

The accepted production NetBird deployment remains authoritative. No source change in this development branch transfers production networking authority or authorizes client cutover. Production peers, setup keys, Access Policies, routes, DNS state, firewall state, credentials, persistent operational state, signing material, client configuration, WireGuard/TUN behavior, and compatibility-bridge state remain unchanged until exact retained acceptance evidence passes and a production migration is explicitly approved.
