# GoreeCloud Network Privacy Shield Integration

## Purpose

GoreeCloud Network is an enforcement adapter for the `network-privacy` capability of GoreeCloud Privacy Shield.

Privacy Shield supplies the shared platform privacy identity, capability vocabulary, adapter/status contracts, and conformance boundary. GoreeCloud Network remains authoritative for its encrypted private-networking runtime, peer connectivity, routes, signaling, relay behavior, access-policy plumbing, and other compatibility-sensitive networking behavior inherited from and maintained on the NetBird foundation.

## Authority boundaries

The initial Privacy Shield adapter does not change GoreeCloud Network's existing component boundaries:

- WireGuard/TUN and peer-connectivity behavior remain networking-core responsibilities.
- GoreeCloud Network management, signal, relay, endpoint, and client components remain authoritative for their own runtime behavior.
- GoreeCloud DNS and Unbound retain DNS filtering, recursive resolution, caching, and DNSSEC responsibilities.
- Caddy remains the separate HTTPS publication gateway.
- GoreeCloud Identity remains the identity/authentication authority where integrated.
- Wardveil Security remains the separate security/protection presentation identity.
- Privacy Shield is not a VPN engine, route controller, firewall, identity provider, or peer-management service.

## Declared capability

The initial adapter declares only:

- `network-privacy`

It does not declare Browser `content-blocking`, `tracking-resistance`, or `url-cleaning`; DNS `dns-privacy`; or generic application telemetry, data-minimization, retention, deletion, export, privacy-status, or user-visible-exception capabilities merely because networking software may expose adjacent settings or metadata.

Additional Privacy Shield capabilities require explicit implementation and runtime-specific acceptance before declaration.

## What `network-privacy` means here

Within GoreeCloud Network, `network-privacy` identifies the privacy role of the encrypted private-networking path and privacy-preserving remote connectivity provided by the accepted runtime. It may describe high-level protection such as whether an approved peer is using the GoreeCloud private-networking path and whether the intended encrypted connectivity capability is accepted for that runtime.

The declaration does not claim anonymity, tracker blocking, DNS filtering, application-level confidentiality beyond the networking transport, endpoint trust, malware protection, firewall enforcement, or application authorization. Network reachability and application authorization remain separate concerns.

## Privacy-safe status boundary

A future Privacy Shield status producer may expose minimized network-privacy state, but it must not export raw network flows or other private activity merely to populate Privacy Shield, Manager, or Wardveil dashboards.

The central status path must not require packet contents, browsing history, DNS queries, source/destination flow logs, unrestricted peer inventories, private routes, setup keys, access tokens, private keys, credentials, user identifiers, device identifiers, IP-address inventories, policy contents, or raw diagnostic logs solely for presentation.

A producer must explicitly declare that raw private activity, credentials, and identifying content are absent; runtime acceptance remains required; and production approval is not inferred.

Where operational troubleshooting legitimately requires detailed network evidence, that evidence remains within the authoritative Network administrative boundary and follows separate access, retention, and security controls.

## Acceptance boundary

`privacy-shield/adapter.json` intentionally records `production_approved=false` and `runtime_acceptance_required=true`.

The declaration recognizes a planned capability boundary; it does not approve the current development branch, a staging image, Android client, dashboard, control-plane build, or production NetBird replacement.

Before production approval, the exact GoreeCloud Network runtime must demonstrate the declared encrypted private-networking behavior, compatibility, failure/recovery behavior, privacy-safe status output if enabled, backup/restore, rollback, and the applicable real-device/client acceptance gates.

Shared Privacy Shield contract validation is not runtime acceptance.

## Current implementation state

This work is source-level adapter groundwork stacked on the active GoreeCloud Network stable-foundation branch. It does not intentionally change WireGuard/TUN behavior, peer connectivity, routing, signal/relay behavior, management APIs, DNS mechanics, firewall mechanics, setup keys, production peers, policies, routes, DNS state, hostnames, credentials, or production migration state.

The next Privacy Shield runtime step is a minimized Network status producer exercised only in the isolated GoreeCloud staging/acceptance environment. Until that producer and runtime acceptance exist, Manager and Wardveil must treat Network Privacy Shield runtime state as unavailable rather than derive it from raw peer or flow data.
