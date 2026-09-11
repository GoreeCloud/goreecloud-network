# Feature Roadmap

## P0 — Native foundation

- [x] Establish coordinated Server, Web, Android, Google TV, and iOS source surfaces.
- [x] Establish versioned API discovery contract.
- [x] Establish truthful Platform System status reporting.
- [x] Establish native control-plane state and truthful overview reporting.
- [~] Add persistent revision/configuration store with migration framework. Development file persistence, schema versioning, migration ledger, checksum verification, and revision metadata are implemented; production database/rollback/HA evidence remains incomplete.
- [ ] Add GoreeCloud Identity authenticated administrative sessions.
- [ ] Add Privacy Shield authorization context to data-affecting operations.
- [ ] Add Wardveil trust/evidence adapter.
- [ ] Add Everkeep snapshot/restore integration.
- [ ] Add GoreeCloud Mesh capability registration and event contracts.
- [ ] Add GoreeCloud Manager inventory/health/admin contract.
- [ ] Validate current applicable Glaze UI contracts on all user-facing surfaces.

## P1 — Native network control

- [~] Device model exists as a Development control-plane primitive; enrollment workflow remains unimplemented.
- [~] Deterministic deny-by-default access evaluation exists; authenticated groups/resources/revisioned policy administration remains unimplemented.
- [~] Decision reason codes exist for the current Development evaluator; complete production decision evidence remains unimplemented.
- [ ] WireGuard lifecycle boundary.
- [ ] Signal/path coordination.
- [ ] Relay inventory and selection.
- [ ] Route controller and private routing.

## P2 — Resilient connectivity

- [ ] Obfuscation provider contract and fail-closed required mode.
- [ ] Restrictive-network detection with evidence.
- [ ] Direct/relay/obfuscated path transitions.
- [ ] Android foreground connection service.
- [ ] iOS Network Extension tunnel provider.
- [ ] Google TV connection experience and remote-first diagnostics.

## P3 — Production qualification

- [~] Exact-revision CI covers server/web, portable iOS core, Android, and Google TV Development builds; full Apple application/release validation remains required.
- [ ] Network-lab integration and failure-injection tests.
- [~] Development persistence migration and corruption tests exist; production migration/rollback/restore testing remains incomplete.
- [ ] Security/privacy/recovery validation.
- [ ] Accessibility validation.
- [ ] Release provenance/signing evidence.
- [ ] Stable qualification review.

`[~]` denotes a partially implemented Development milestone that is not complete for production/release qualification.

This file records repository-local implementation progress. It must be reconciled with the canonical GoreeCloud feature-roadmap record before release qualification.
