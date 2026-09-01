# GoreeCloud Network Integration Boundary

GoreeCloud Network currently retains the NetBird-derived runtime as its connectivity data plane. GoreeCloud Manager must not depend on NetBird-specific peer APIs as the long-term product boundary.

## Infrastructure Status v1

`./goreecloud/statuscmd` emits the GoreeCloud-owned, privacy-minimized status envelope that Manager can consume without receiving NetBird API credentials or peer inventory.

```bash
go run ./goreecloud/statuscmd
```

Set `GOREECLOUD_NETWORK_STATUS_FILE=/path/to/network-status.json` to enable the local status handoff. The management-server adapter refreshes the same owner-only atomic file every 30 seconds while that runtime is active.

The in-process management adapter is deliberately narrow:

- `peer-coordination` is verified only after the existing management `Server.Start` succeeds;
- `access-policy` is verified on the same boundary because the management store, managers, API/gRPC server, and listener must initialize before `Start` returns;
- `network-dns` is verified on that same successful-start boundary because construction of the management API handler initializes the zones and DNS-record managers before startup can succeed; the adapter reads none of their zone, record, label, address, account, or tenant content;
- `private-connectivity` remains unverified because end-to-end connectivity depends on signal/relay and peer runtimes outside the management process.

A successfully started management service therefore reports `partial`, not `ready`. On graceful stop the adapter writes `unavailable`; unexpected process loss is handled by Manager's stale-status fail-closed rule. Runtime evidence never sets `production_approved`.

## Privacy boundary

The status document must not contain access tokens, setup keys, user identities, peer names, peer IDs, peer IP addresses, public keys, routes, ACL rules, DNS labels, activity history, raw logs, or personal records. Aggregate/coarse state may be added only when it cannot be used to reconstruct those identifiers.

The management adapter wraps the existing server lifecycle only. It adds no network listener and does not query the management API, store, peer inventory, route inventory, DNS zones/records, or user data.

## Manager migration

Manager's existing direct NetBird adapter is transitional. The target boundary is:

`NetBird-derived runtime -> GoreeCloud Network adapter -> sanitized Infrastructure Status v1 -> GoreeCloud Manager`

Manager remains a read-only observer; Network remains authoritative for network state and policy. The direct NetBird adapter is not removed until private-connectivity evidence is independently available and accepted alongside the management-side boundaries.

## Next implementation slice

1. Add bounded signal/relay lifecycle evidence for end-to-end private-connectivity readiness without peer or route inventory.
2. Complete target-environment connectivity, Network DNS behavior, policy, outage, upgrade/rollback, and privacy acceptance.
3. Migrate Manager away from the direct NetBird API only after those runtime boundaries are accepted.
4. Keep `production_approved` false until all target-environment gates pass.

## Licensing

The root-level `goreecloud/status` adapter follows the repository's existing root BSD-3-Clause boundary. The management lifecycle wrapper under `management/cmd/` is inside that directory's existing AGPL-3.0 boundary. No upstream-derived code is relicensed by this work.
