# Conduit Platform Acceptance Evidence

`goreecloud-conduit-platform-acceptance-evidence/v1` is a minimized source-level consumption contract for required GoreeCloud platform-system acceptance.

The bundle is bound to an exact Conduit source revision and carries six required records:

- Privacy Shield;
- Wardveil Security;
- Everkeep;
- GoreeCloud Mesh;
- GoreeCloud Identity;
- GoreeCloud governance.

Each record contains only the expected platform identity, a 64-character SHA-256 identity for the external evidence artifact, and an explicit accepted result. The bundle cannot set production cutover authority.

Validation fails closed on schema mismatch, source-identity mismatch, platform-label substitution, a rejected record, malformed evidence identity, or any production-cutover claim. Only after the complete bundle validates are the existing Conduit isolated-acceptance platform gates marked true.

The bundle is designed to avoid importing private network state. It contains no peer, route, policy, packet, DNS-query, credential, device, user, or unrestricted diagnostic fields.

This contract does not verify the contents or producer authenticity of an external evidence artifact by hash alone. The referenced platform-system evidence must be produced, governed, retained, and accepted through the appropriate platform authority before its digest is consumed here.
