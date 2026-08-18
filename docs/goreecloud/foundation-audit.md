# GoreeCloud Network Foundation Audit

## Scope

This audit records the first GoreeCloud review of the inherited NetBird source foundation. It is intentionally limited to repository structure, licensing, build/CI surface, and migration boundaries. It does not authorize production cutover.

## Verified baseline characteristics

- The repository remains a fork of `netbirdio/netbird`.
- The root license is BSD-3-Clause except for `management/`, `signal/`, `relay/`, and `combined/`, which are AGPLv3.
- The Go module remains `github.com/netbirdio/netbird` during the compatibility-first phase.
- The inherited build currently uses Go 1.25.x tooling.
- The repository includes broad upstream CI coverage including Linux, Windows, macOS, FreeBSD, linting, licensing checks, installation tests, and networking-oriented validation workflows.

## GoreeCloud boundaries

- WireGuard, routing, signaling, relay, peer negotiation, and core cryptographic behavior are compatibility-critical and must not be casually rewritten.
- GoreeCloud-specific branding must not erase upstream provenance or licensing notices.
- Production NetBird remains authoritative until a separately validated GoreeCloud release candidate passes networking, recovery, DNS, access-policy, and client compatibility gates.
- Caddy remains the HTTPS gateway; GoreeCloud Network remains the private-networking platform.
- AdGuard Home/Unbound DNS responsibilities remain separate from the network control plane.

## Next engineering gates

1. Select and record an exact upstream release and commit baseline.
2. Identify workflows that depend on NetBird-owned infrastructure or secrets and prevent accidental GoreeCloud execution where inappropriate.
3. Inventory control-plane configuration, state storage, migration behavior, and backup requirements.
4. Build a compatibility test matrix for existing GoreeCloud peers, DNS, policies, routes, and private SSH access.
5. Defer functional fork-to-native changes until the inherited source builds and tests cleanly under GoreeCloud-controlled CI.
