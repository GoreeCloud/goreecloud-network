# GoreeCloud Network Features

This file distinguishes GoreeCloud-owned Conduit behavior from inherited NetBird compatibility behavior and from future accepted scope.

## Implemented Conduit foundations

- Native capability and migration registries.
- Privacy-safe capability inventory and aggregate status.
- Fail-closed `implementation -> isolated-validation` transition rules.
- Protected inventory snapshots.
- Compare-and-swap transitions and immutable transition receipts.
- Deterministic retry/reconciliation behavior.
- Exact source/artifact staging evidence and minimized staging status.
- First-party obfuscation negotiation state machine.
- Required-mode no-silent-downgrade behavior.
- First-party `conduit-padded-wss` relay transport with bounded randomized message padding inside authenticated WSS.
- Android coarse transport/runtime state exposure without relay URL, peer, route, packet, DNS, credential, or network metadata.
- Exact-source loopback status runtime acceptance.
- Hardened container-boundary status acceptance with an unpublished private listener.
- Required-mode fail-closed startup when the Conduit private status listener cannot bind.
- Isolated acceptance gates for exact source, immutable artifact, state migration, recovery, client networking, security/privacy, Privacy Shield, Wardveil Security, Everkeep, Mesh, Identity, and governance.
- Minimized `goreecloud-conduit-platform-acceptance-evidence/v1` consumption contract with exact source binding and SHA-256 evidence identities.

## Accepted but not yet complete

- Full staging-runtime backup/restore and rollback rehearsal.
- Final immutable registry image digest retention for dedicated-host acceptance.
- Real target-environment platform evidence production and consumption.
- Representative restrictive-network coverage for obfuscation mode.
- Security/privacy review and controlled rollback evidence for the isolated obfuscation runtime.
- Source-bound signed Android APK/AAB acceptance evidence.
- Approved real-device acceptance.
- Capability-by-capability transition from compatibility bridge authority to native authority.
- Final production migration and removal of obsolete inherited authority only after acceptance.

## Explicit non-claims

- A valid platform evidence bundle is not proof that the referenced platform produced valid evidence; the referenced evidence must exist and be accepted separately.
- NetBird remains production-authoritative.
- Current Conduit source does not authorize production cutover.
- Broad inherited platform CI failures are not automatically Conduit feature failures, but Conduit-owned regressions must still be repaired.
