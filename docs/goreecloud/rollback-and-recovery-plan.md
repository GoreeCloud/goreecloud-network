# GoreeCloud Network Rollback and Recovery Plan

## Purpose

This plan defines the minimum rollback and recovery controls required before GoreeCloud Network can replace the accepted production NetBird deployment. It does not authorize a production cutover.

## Core rule

A migration is not approved unless the previous accepted private-network state can be restored without depending on GoreeCloud Network itself being healthy.

## Pre-cutover requirements

Before any production migration:

1. Record the exact accepted NetBird server/client versions and configuration state.
2. Produce and verify a backup of all required server-side state.
3. Record current production peers, groups, policies, routes, DNS configuration, routing peers, and administrative access paths in a sanitized recovery record.
4. Preserve an administrative path that does not depend on the new private-network path.
5. Preserve the stock NetBird client or another accepted recovery client on at least one approved administrative endpoint until GoreeCloud Network has passed stabilization.
6. Ensure DNS, Caddy, firewall, VLAN, identity-provider, and application-authentication changes are not bundled into the same cutover unless separately approved and independently reversible.
7. Record APK/package/signing identity for the previously accepted Android client state and retain the installable recovery artifact where licensing and policy permit.
8. Verify restore procedures in an isolated environment before production use.

## Rollback triggers

Initiate rollback when a migration causes or risks any of the following and a safe forward fix is not immediately proven:

- loss of administrative access;
- widespread peer enrollment or connectivity failure;
- incorrect Access Policy enforcement;
- unintended broad access;
- routing loops or persistent route loss;
- managed DNS failure affecting required private resolution;
- routing-peer failure without a working replacement;
- Android client inability to reconnect or restore the previous accepted client state;
- control-plane corruption or failed upgrade;
- authentication failure that prevents administrators from recovering the service;
- inability to restore from the current GoreeCloud Network backup;
- any security defect that materially weakens the accepted production posture.

## Server rollback sequence

The exact commands are environment-specific and must be documented before cutover, but the logical sequence is:

1. Stop further configuration changes and preserve failure evidence.
2. Confirm an out-of-band or independent administrative path is working.
3. Disable the failed GoreeCloud Network production endpoint or redirect only the affected test/cutover path as appropriate.
4. Restore the previously accepted NetBird service/configuration state from verified backup or preserved deployment state.
5. Restore the previously accepted DNS/Caddy routing only if those components were intentionally changed as part of the approved cutover.
6. Verify management UI/API health.
7. Verify at least one administrative peer and one ordinary peer can enroll/connect as expected.
8. Verify critical routes, DNS, and Access Policies.
9. Revoke temporary migration setup keys or credentials.
10. Record the rollback outcome and freeze further migration until root cause is understood.

## Android rollback sequence

1. Confirm whether the GoreeCloud candidate and previous accepted client use the same or different application IDs/signing identities.
2. If they coexist, disconnect the failed candidate and reconnect using the accepted recovery client.
3. If replacement requires uninstall/reinstall, record the expected loss of local profiles/settings before cutover and preserve required enrollment material through approved secure channels.
4. Regrant Android VPN permission or restore Always On VPN configuration when package identity requires it.
5. Verify SSO/setup-key enrollment, DNS, routes, resources, and reconnection on the restored client.
6. Revoke any temporary test/migration device identity that should no longer remain enrolled.

## Recovery evidence

Retain:

- backup identifiers and verification results;
- exact source/release commit and artifact digests;
- configuration version;
- start/end time of recovery;
- observed failure trigger;
- steps performed;
- service/peer validation results after recovery;
- unresolved data or configuration differences;
- root-cause tracking reference.

Do not retain reusable secrets in the rollback record.

## Stabilization after rollback

After rollback, GoreeCloud Network returns to development or controlled-test status. Do not immediately retry production migration. Correct the defect, repeat repository validation, repeat isolated acceptance, repeat restore/rollback testing where relevant, and approve a new cutover separately.

## Acceptance requirement

Production migration cannot be approved until this rollback plan has been converted from a logical plan into environment-specific tested procedures and at least one controlled rollback rehearsal has succeeded.
