# GoreeCloud Network Upstream Baseline

## Purpose

This document records the exact inherited NetBird source baseline for the initial GoreeCloud Network stabilization branch. It separates the source snapshot that the fork actually started from from the most recent upstream release that can be used as a conservative comparison point.

## Fork source baseline

The `main` branch used as the base of GoreeCloud Network pull request #1 points to upstream NetBird commit:

- Commit: `ecfbd686b807f2e47295e1dfad3cd86b47af3c84`
- Upstream repository: `netbirdio/netbird`
- Upstream commit date: 2026-08-18
- Upstream subject: `[client, android] Expose ssh functionality for Android (#7156)`

The same full commit exists in the upstream NetBird repository. This means the initial GoreeCloud fork was not created from an unidentified or locally rewritten source snapshot; its starting point can be mapped exactly to upstream.

This commit is the **source baseline** for the current stabilization work. It is not yet the **accepted production baseline** because GoreeCloud has not completed its own build, integration, upgrade, privacy, and regression acceptance against it.

## Released comparison point

At the time this baseline was recorded, the latest stable NetBird release identified during the audit was `v0.76.1`, whose release commit is:

- Tag: `v0.76.1`
- Commit: `0780a806f2cc2e8a6a51782cfffe0591b7c3fa9c`
- Release date: 2026-07-31

The GoreeCloud source baseline is therefore newer than that release and contains unreleased upstream work. That difference matters when diagnosing regressions: a failure may originate from post-release upstream changes rather than from GoreeCloud modifications.

## Baseline policy

Until the stabilization branch passes GoreeCloud acceptance:

1. Treat `ecfbd686b807f2e47295e1dfad3cd86b47af3c84` as the immutable comparison point for changes made on `agent/stable-foundation`.
2. Do not describe the branch as production-ready merely because it matches a valid upstream commit.
3. Keep protocol-sensitive changes small and independently reviewable.
4. Compare suspected regressions against both the exact source baseline and the latest accepted release when useful.
5. Record any later upstream rebase or merge explicitly rather than silently moving the baseline.

## Acceptance still required

The source baseline becomes an accepted GoreeCloud baseline only after relevant validation succeeds, including:

- Go build and test coverage appropriate to the modified packages.
- Lint and license/dependency checks.
- Client/control-plane compatibility.
- Dashboard authentication and management API compatibility.
- Android tunnel, authentication, enrollment, DNS, route, and lifecycle compatibility.
- Privacy review for telemetry or external-service assumptions.
- Isolated self-hosted deployment testing before any production replacement.

## Relationship to product customization

Glaze UI, GoreeCloud workflow changes, and Wardveil Security presentation should continue to be developed primarily outside protocol-sensitive networking code until the compatibility baseline is accepted. The exact upstream commit recorded here provides the point against which future GoreeCloud-specific networking changes can be audited.
