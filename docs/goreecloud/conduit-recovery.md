# GoreeCloud Conduit migration inventory recovery

Conduit migration state must be reversible before any capability can advance toward production authority. `native/conduit/control/inventory_recovery.go` adds an immutable recovery-point boundary for the privacy-safe native capability inventory.

A recovery point captures one validated `CapabilityInventorySnapshot`, its exact fingerprint, the exact Conduit source revision that created the recovery evidence, and a creation timestamp. The inventory contains migration stage, authority class, compatibility-bridge state, and cutover=false only; it does not contain peers, routes, policy contents, packets, DNS queries, credentials, devices, users, or unrestricted diagnostics.

Recovery records are stored separately from active inventory state below an owner-protected directory. Each record uses the snapshot SHA-256 fingerprint as its deterministic filename, is created exclusively rather than overwritten, and is written with owner-only permissions on supported platforms. Windows intentionally fails this owner-only filesystem persistence contract rather than claiming Unix-style protection that is not available through the same mechanism.

`RestoreCapabilityInventoryRecoveryPoint` is explicit and compare-and-swap protected. The operator or migration orchestrator must provide the exact fingerprint of the currently active inventory. If active state changed after rollback was planned, restoration fails closed. The recovered inventory is revalidated and must reproduce the exact retained recovery fingerprint before it replaces active inventory state.

This recovery mechanism restores only Conduit migration-control inventory. It does not restore NetBird peers, routes, keys, WireGuard state, DNS state, relay sessions, credentials, host firewall configuration, container volumes, or application data. Full staging-runtime backup/restore remains a broader target-environment acceptance gate and must use Everkeep where applicable.

A recovery point can never set `production_cutover_authorized=true`. NetBird remains production-authoritative until capability-specific migration evidence, platform acceptance, backup/restore, rollback rehearsal, deployment provenance, and explicit production approval are complete.