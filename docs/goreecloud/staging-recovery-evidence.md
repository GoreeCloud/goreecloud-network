# Conduit Isolated-Staging Recovery Evidence

GoreeCloud Network now defines a minimized evidence contract for staging-runtime backup/restore and rollback rehearsal. The contract is designed to consume evidence after an external recovery executor, such as Everkeep, has actually preserved and restored the isolated staging environment.

The recovery plan schema is `goreecloud-network-staging-recovery-plan/v1`. It is bound to an exact GoreeCloud Network source revision and requires encrypted backup-artifact identities for all of these staging components:

- management state;
- relay state;
- network runtime configuration;
- Conduit migration inventory; and
- dashboard runtime configuration.

Only SHA-256 artifact identities and coarse recovery metadata cross the evidence boundary. Credentials, private keys, peer data, routes, packets, DNS queries, users, device identifiers, and unrestricted runtime configuration are not embedded in the evidence record.

The recovery receipt schema is `goreecloud-network-staging-recovery-receipt/v1`. It binds the result to the exact recovery plan SHA-256 and source revision, requires every planned component to be restored with the same retained artifact identity, requires an aggregate before/after state-equivalence SHA-256 match, and requires explicit successful rollback rehearsal. Both schemas require `credentials_included=false` and `production_cutover_authorized=false`.

This implementation validates recovery evidence; it does not claim that CI itself backed up or restored a live staging host. Formal acceptance still requires the owning Everkeep/operations process to execute backup, restore, state-equivalence measurement, and rollback in the target isolated-staging environment and retain the resulting evidence.

NetBird remains production-authoritative. The staging recovery evidence contract cannot change peers, routes, WireGuard/TUN authority, DNS, firewall state, credentials, compatibility-bridge state, or production migration authority.
