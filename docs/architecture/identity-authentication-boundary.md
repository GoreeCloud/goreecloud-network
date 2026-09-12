# GoreeCloud Identity Administrative Authentication Boundary

**Lifecycle:** Development  
**Network version:** `0.1.0-dev.4`

GoreeCloud Network consumes GoreeCloud Identity as the platform identity authority. Network does not create a parallel password database, login authority, user directory, or bearer-token issuer.

## Current Development boundary

The server implements an OAuth 2.0 token-introspection consumer boundary for administrative requests. The boundary:

- accepts only bearer tokens supplied to protected Network administrative routes;
- submits the token to the configured GoreeCloud Identity introspection endpoint using confidential client authentication;
- requires an active identity result with a non-empty subject;
- can require exact issuer and audience values when configured;
- requires the application-specific `goreecloud.network.admin` authority scope by default;
- returns only the minimum Network session view: subject, optional preferred username, and scopes;
- does not consume or expose email, profile, group, or other identity claims that are not required by this slice;
- uses HTTPS for non-loopback introspection endpoints and rejects partial Identity configuration;
- fails closed when Identity is unconfigured, unavailable, rejects the token, or does not provide Network administrator authority.

The client secret is accepted only through protected runtime configuration and is never included in API status responses or normal log output.

## Protected Development probe

`GET /api/v1/admin/session` is the first route behind the Identity boundary. It proves the authentication/authorization seam without creating an administrative mutation capability.

The endpoint does not modify devices, resources, policies, enrollment, routes, tunnel state, or persistent configuration.

## Authority split

GoreeCloud Identity establishes the actor and supplies authority claims. GoreeCloud Network remains responsible for mapping those claims to Network-specific operations and for enforcing Network-domain policy. Authentication does not imply privacy authorization, Wardveil trust, Network reachability, or application-level authorization outside Network.

## Runtime acceptance boundary

Source implementation and automated tests do not establish that the current GoreeCloud Identity deployment is configured, reachable, production-approved, or accepted for Network SSO. Runtime integration must separately validate application registration, confidential-client handling, token lifetime/session behavior, disabled-account behavior, outage behavior, issuer/audience expectations, scope mapping, logout/session revocation expectations, recovery, and rollback.

Until that evidence exists, the repository reports the Identity consumer boundary as implemented Development source with runtime acceptance pending.
