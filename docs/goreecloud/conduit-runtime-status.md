# Conduit isolated runtime status integration

## Purpose

This document records the first GoreeCloud Conduit server-side integration with the inherited GoreeCloud Network management runtime.

The integration is intentionally read-only and staging-oriented. It proves that a first-party Conduit API contract can be attached to a running inherited management process without transferring authoritative state, changing networking protocols, or exposing private management data.

## Runtime model

`management/cmd/conduit_status.go` decorates the existing management `server.Server` factory without modifying the inherited management command implementation.

When Conduit status is disabled, the decorator returns the inherited server unchanged.

When explicitly enabled, the wrapper:

1. validates the companion listener address before management startup;
2. requires an explicit loopback IP address;
3. starts the inherited management runtime first;
4. creates a separate HTTP listener only after inherited startup succeeds;
5. serves the first-party `goreecloud-conduit-control-status/v1` contract at `/goreecloud/conduit/v1/status`;
6. forwards inherited runtime errors through the wrapper;
7. shuts down the companion listener before delegating management shutdown.

The Conduit listener does not replace or multiplex the inherited management HTTP/gRPC listener.

## Enablement

The integration is disabled by default.

It is enabled only when:

```text
GOREECLOUD_CONDUIT_STATUS_ENABLED=true
```

The optional listener variable is:

```text
GOREECLOUD_CONDUIT_STATUS_ADDR=127.0.0.1:9097
```

If no address is provided, `127.0.0.1:9097` is used. Non-loopback addresses such as `0.0.0.0:9097` are rejected before the inherited management server starts.

The isolated staging Compose model enables this companion endpoint inside the management container. Port 9097 is not published to the staging host. Formal runtime evidence therefore must be collected from within the isolated management container/network namespace or through another separately approved loopback-local diagnostic mechanism.

## Contract boundary

The response contains only:

- schema identifier;
- generation timestamp;
- authority classification;
- migration stage;
- compatibility-bridge state;
- production-cutover authorization state.

The integration does not expose peer or device inventories, user identifiers, group memberships, Access Policy contents, private resources, route definitions, setup keys, API tokens, private keys, credentials, packet contents, DNS queries, raw flow logs, or raw diagnostic bundles.

The compatibility bridge forces:

- `authority=inherited`;
- `compatibility_bridge_active=true`;
- `production_cutover_authorized=false`.

The endpoint accepts GET only and sends `Cache-Control: no-store`.

## Persisted staging evidence status

`native/conduit/control/staging_status.go` adds a separate evidence-status builder for isolated validation. It emits status only after reloading the exact immutable transition receipt and exact immutable staging-evidence record from their protected stores.

The builder fails closed when transition reconciliation is unresolved, either persisted record is unreadable, the transition receipt no longer binds the inventory snapshot, the staging record does not bind the same transition identity, the staging record predates the transition, or inherited-authority/active-bridge/cutover-false safety invariants diverge.

The resulting `goreecloud-conduit-capability-staging-status/v1` record is minimized for Manager, Wardveil Security, Privacy Shield, and GoreeCloud Mesh consumption. It includes the inventory fingerprint and evidence-state booleans but omits capability IDs, source revisions, artifact digests, peers, routes, policies, credentials, packet data, DNS queries, and other operational detail. This builder is evidence plumbing only; it does not expose a new listener and does not change migration authority.

## Migration classification

This source/runtime wiring is an implementation milestone, not isolated-runtime acceptance.

Conduit Control and Conduit API may advance from `contract` to `implementation` only after the exact-head focused integration tests pass. They must remain below `isolated-validation` until a real isolated staging management container is started from exact source-bound images and the endpoint is exercised as part of retained runtime evidence.

The inherited management/control plane remains authoritative throughout this stage.

## Acceptance requirements before isolated-validation

The next gate requires an exact-revision staging run that demonstrates:

- the management container starts successfully with Conduit status enabled;
- the status endpoint is reachable only from its intended loopback-local context;
- the returned schema and authority values match the source contract;
- POST and other mutation attempts remain rejected;
- no sensitive or private network state appears in the payload;
- disabling the Conduit environment switch removes the companion listener without changing inherited management behavior;
- management startup failure prevents the Conduit companion listener from being exposed;
- staging teardown cleanly closes both listeners;
- the source SHA, image digest, configuration hashes, and result timestamps are retained in the acceptance evidence set.

Successful source/unit/integration CI does not satisfy this runtime gate by itself.

## Production boundary

This integration does not authorize production migration and does not require any change to the accepted production NetBird deployment. No production peer, setup key, Access Policy, route, DNS state, Caddy configuration, firewall rule, credential, persistent state, signing material, or client configuration is changed by this source work.
