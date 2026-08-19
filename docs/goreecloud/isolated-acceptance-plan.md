# GoreeCloud Network Isolated Acceptance Plan

## Purpose

This plan defines the minimum controlled test environment and evidence required before GoreeCloud Network may be considered for production migration. It does not authorize changes to the accepted production NetBird deployment.

## Test environment boundary

Use an isolated or otherwise controlled non-production environment with separate test identities, setup keys, devices, routes, networks, and DNS data. Production peers, production setup keys, production access policies, production DNS state, production routing peers, and production application credentials must not be reused as test fixtures.

The test topology should include:

- one isolated GoreeCloud Network management/control-plane instance built from the current GoreeCloud branch;
- one dashboard instance connected only to that isolated control plane;
- at least two test peers capable of direct connectivity;
- at least one scenario that requires relay or otherwise validates relayed connectivity;
- one Linux routing peer for a controlled private subnet or test resource;
- one Android test device using the GoreeCloud Network client;
- an isolated DNS resolver path that can validate managed DNS delivery without changing production AdGuard Home or Unbound state;
- at least one deliberately unavailable resource to validate failure presentation and recovery.

## Control-plane acceptance

Record evidence for all of the following:

1. Clean startup from documented configuration.
2. Administrative authentication succeeds against the intended identity-provider configuration.
3. Test device enrollment succeeds using both ordinary browser/SSO enrollment and a one-off setup key where supported.
4. Setup-key expiry, single-use limits, revocation, and failed reuse behave as expected.
5. Device removal prevents subsequent access without affecting unrelated devices.
6. Purpose-based groups can be created, updated, and removed without unintended membership changes.
7. Access Policies deny access by default and allow only the intended source groups, destinations, protocols, and ports.
8. Posture conditions affect access only when configured and do not silently broaden access.
9. Networks and resources can be created, edited, disabled, and removed with expected API behavior.
10. Routing-peer enable/disable/removal produces the expected reachability impact and does not delete the underlying enrolled device unless separately revoked.
11. Direct and relayed peer connectivity can both be observed where the topology permits.
12. DNS resolver delivery, DNS zones, and excluded groups behave as configured.
13. Audit Events record material administrative changes when the underlying platform exposes them.
14. Backup of required control-plane state completes successfully.
15. Restore into a clean isolated instance reproduces required users, devices, policies, routes, resources, and configuration without relying on undocumented state.

## Network-path acceptance

For each tested path, record source, destination, expected policy, observed path, and result.

Required cases:

- approved device to approved device;
- approved device to denied device/resource;
- approved device to allowed private resource through routing peer;
- approved device to private resource with no matching Access Policy;
- resource reachable at network layer but denied by the destination application's own authentication/authorization;
- routing peer disabled while traffic is active;
- routing peer restored or replaced;
- DNS resolution with managed DNS enabled;
- DNS resolution for an excluded group;
- reconnect after temporary network loss;
- reconnect after control-plane restart where supported;
- revocation while the client is connected.

## Safety assertions

The test is a failure if any of the following occur:

- a new user/device receives broad access without an explicit policy;
- a resource becomes reachable merely because it exists in the inventory;
- deleting a GoreeCloud Network resource deletes or modifies destination application data;
- routing-peer removal revokes or deletes the enrolled device unexpectedly;
- DNS testing changes production resolver, rewrite, or DNSSEC configuration;
- private-network reachability is presented as proof of application authorization;
- Wardveil UI claims protected/secure/trusted state without runtime evidence;
- a test requires production setup keys, private keys, tokens, or reusable credentials.

## Evidence to retain

For every acceptance run retain:

- GoreeCloud repository commit SHAs and relevant release/artifact digests;
- configuration version or sanitized configuration snapshot;
- test date/time and tester;
- topology description;
- pass/fail result for each case;
- relevant logs with secrets redacted;
- screenshots where UI state matters;
- backup/restore evidence;
- rollback evidence;
- unresolved defects and severity.

## Exit criteria

The isolated environment is accepted only when all critical and high-severity defects are resolved or explicitly deferred with a documented safety rationale, rollback works, and the dashboard/Android clients use artifacts that have passed their repository validation contracts.

Passing this plan does not by itself authorize production migration. Production cutover remains a separate controlled change requiring backup, recovery, rollback, and final approval.
