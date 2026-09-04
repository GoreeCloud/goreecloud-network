# Mandatory Native and Platform Conformance

Effective August 24, 2026, GoreeCloud Network must be built and maintained toward an original GoreeCloud-owned product architecture rather than retaining a complete third-party application as its final architecture.

Small, technically necessary foundational dependencies remain permitted where independent reimplementation would reduce security, correctness, interoperability, standards compliance, or maintainability. Examples include WireGuard, cryptographic and encryption libraries, protocol libraries, codecs, database engines, operating-system APIs, and comparable critical foundations. Such dependencies must not become the application shell or define the GoreeCloud product identity.

## Integral Platform Systems

Platform Contract v0.2 requires substantive, current, evidence-backed conformance with all seven Integral Platform Systems where applicable:

- GoreeCloud Manager
- GoreeCloud Privacy Shield
- Wardveil Security
- Everkeep
- Glaze UI
- GoreeCloud Mesh
- GoreeCloud Identity

For Glaze UI, the current approved Stable target is GLAZE UI V1.1 / 1.1.0. Superseded pre-reset version lines are historical or migration context and do not replace the current Stable contract.

No integral platform system may substitute for another. A decorative identity, placeholder declaration, planned adapter, or source-only stub does not satisfy an integration gate.

No GoreeCloud Network release or service state may be classified or retained as Stable unless applicable native qualification and current validated conformance with the integral platform systems are complete. A missing, materially incomplete, unvalidated, or materially outdated required integration is a Stable blocker.

If this repository contains inherited or upstream-derived application code, that code is transitional or historical. It may be maintained for security, continuity, migration, compatibility, recovery, and rollback while the native replacement is developed, but it is not the approved final architecture.

Repository CI, release documentation, project specifications, change logs, migration evidence, and `goreecloud.platform.yaml` must progressively enforce and record this contract. Source presence, a successful build, or isolated staging evidence does not by itself authorize production migration or transfer NetBird's current production authority.
