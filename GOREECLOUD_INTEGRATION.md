# GoreeCloud Network Integration Boundary

GoreeCloud Network currently retains the NetBird-derived runtime as its connectivity data plane. GoreeCloud Manager must not depend on NetBird-specific peer APIs as the long-term product boundary.

## Infrastructure Status v1

`./goreecloud/statuscmd` emits the GoreeCloud-owned, privacy-minimized status envelope that Manager can consume without receiving NetBird API credentials or peer inventory.

```bash
go run ./goreecloud/statuscmd
```

Set `GOREECLOUD_NETWORK_STATUS_FILE=/path/to/network-status.json` to write the envelope atomically instead of printing it.

The initial adapter reports `development` with `pending` capabilities until it is connected to accepted local runtime evidence.

## Privacy boundary

The status document must not contain access tokens, setup keys, user identities, peer names, peer IDs, peer IP addresses, public keys, routes, ACL rules, DNS labels, activity history, raw logs, or personal records. Aggregate/coarse state may be added only when it cannot be used to reconstruct those identifiers.

## Manager migration

Manager's existing direct NetBird adapter is transitional. The target boundary is:

`NetBird-derived runtime -> GoreeCloud Network adapter -> sanitized Infrastructure Status v1 -> GoreeCloud Manager`

Manager remains a read-only observer; Network remains authoritative for network state and policy.

## Next implementation slice

1. Derive connectivity/coordination/policy/DNS state from bounded local Network runtime interfaces.
2. Normalize results without peer or route inventory.
3. Add fail-closed malformed-state and privacy tests.
4. Migrate Manager's visible Network integration from direct NetBird API access to this contract.
5. Complete target-environment connectivity, policy, outage, upgrade/rollback, and privacy acceptance before any production approval claim.

## Licensing

This root-level GoreeCloud adapter follows the repository's existing root BSD-3-Clause boundary. Changes inside `management/`, `signal/`, `relay/`, or `combined/` remain under their existing AGPL-3.0 boundary.
