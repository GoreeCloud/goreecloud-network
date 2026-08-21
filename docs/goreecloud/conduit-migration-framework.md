# GoreeCloud Conduit Migration Framework

## Purpose

This document defines the repository-level transition framework used to replace inherited NetBird product behavior with first-party GoreeCloud Conduit capabilities without treating a source rewrite as production acceptance.

The authoritative strategic direction remains `Strategy — Network Native Feature Replacement`. This repository document translates that strategy into enforceable source contracts.

## Governing model

Every inherited product feature follows this progression:

`inventory -> contract -> implementation -> isolated-validation -> migration-ready -> production-candidate -> native-accepted -> retired inherited path`

A feature may move forward only when its declared acceptance evidence supports the next stage. Production cutover is never authorized by the source registry itself.

## Registries

`native/conduit/features.json` defines the required first-party destination for each inherited product feature.

`native/conduit/migrations.json` defines how each feature is allowed to transition. Each migration record identifies:

- the Conduit feature family;
- the current migration stage;
- the compatibility bridge that preserves accepted behavior;
- the current state authority;
- whether the feature carries persistent or security-critical state;
- the acceptance gates required before retirement; and
- the explicit retirement condition for the inherited path.

## State authority

The migration registry uses four state-authority values:

- `inherited` — the accepted inherited implementation remains authoritative;
- `transitional` — GoreeCloud-native behavior exists, but the migration is not yet complete;
- `native` — the Conduit implementation has passed the required acceptance gates and is authoritative;
- `none` — no state authority exists because the inherited feature is intentionally retired rather than replaced.

A replacement-required feature cannot declare `native` authority until its stage is `native-accepted`.

## Mandatory acceptance gates

All replacement-required features must declare these gates:

- source validation;
- unit testing;
- integration testing;
- security and privacy review;
- migration testing;
- rollback testing; and
- isolated-runtime acceptance.

Stateful features must also declare backup/restore acceptance.

Additional gates are required when relevant, including network-path validation, DNS-resolution validation, signed-artifact provenance, real-device testing, and accessibility validation.

## Compatibility bridges

A compatibility bridge is temporary by definition. It exists to preserve currently accepted behavior while a Conduit-native implementation is introduced and validated.

A bridge must not become an excuse to deepen unnecessary coupling to inherited code. New GoreeCloud behavior should prefer stable adapters, explicit contracts, and first-party domain models whenever practical.

## Production boundary

The migration registry is intentionally fail-closed:

- `production_cutover_authorized` must remain `false`;
- the source contract cannot mark production migration complete;
- production acceptance requires separate runtime evidence;
- stateful replacements require backup and restoration evidence;
- connectivity-critical replacements require failure and rollback evidence; and
- client replacements require signed and real-device evidence where applicable.

The currently accepted production NetBird deployment remains the rollback and continuity boundary until a separately approved Conduit production migration is completed.

## Standards-based dependencies

WireGuard remains an approved standards-based encrypted-tunnel primitive. The native-by-destination rule applies to GoreeCloud product behavior surrounding that primitive, including orchestration, enrollment, policy, routes, signaling, relay behavior, path selection, diagnostics, clients, administration, and release engineering.

A mature cryptographic primitive is not rewritten merely to remove an upstream name.

## Retirement rule

Inherited code may be retired only when one of these conditions is true:

1. an accepted Conduit-native implementation provides the required behavior;
2. the inherited behavior has been intentionally removed from GoreeCloud scope; or
3. the component is explicitly classified as a retained standards-based dependency.

Licensing, attribution, provenance, and historical records must remain preserved even when an inherited implementation is no longer operational.
