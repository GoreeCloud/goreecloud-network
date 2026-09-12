# GoreeCloud Network

GoreeCloud Network is GoreeCloud's native private networking, encrypted connectivity, remote-access, device-enrollment, private-routing, obfuscation, and Zero Trust network-access platform.

## Current state

**Version:** `0.1.0-dev.4`  
**Lifecycle:** Development  
**Repository model:** GoreeCloud-created native monorepository  
**Production status:** Not production-ready. No Stable claim is made.

The repository currently contains five coordinated product surfaces:

- `cmd/network-server` + `internal` — GoreeCloud Network Server and native Development control plane.
- `apps/web` — Web Dashboard.
- `apps/android` — Android client.
- `apps/google-tv` — Google TV client.
- `apps/ios` — iOS client and portable Swift core.

The Development server has a revisioned JSON file store for its control-plane snapshot and now contains the first GoreeCloud Identity administrative-authentication consumer boundary. Network consumes Identity through OAuth token introspection rather than creating a parallel password or token authority.

The Identity boundary is source-implemented and fail-closed, but runtime/production SSO acceptance has **not** been established. No mutating administrative API is exposed yet. The current protected `GET /api/v1/admin/session` route is a read-only boundary probe requiring an authenticated Identity token with Network administrator scope.

The access evaluator remains decision-only and deny-by-default. No packet/tunnel enforcement is implied. Remaining Platform System adapters and security/privacy/recovery integrations stay explicitly incomplete until implementation and independent validation exist.

## Run the server

```bash
go run ./cmd/network-server
```

The server binds to `127.0.0.1:8080` by default. Set `GOREECLOUD_NETWORK_ADDR` only when an intentional non-loopback Development binding is required.

The default Development state file is `var/network-state.json`. Override it with `GOREECLOUD_NETWORK_DATA_FILE`.

### Optional GoreeCloud Identity Development configuration

Configure the core Identity introspection values together:

- `GOREECLOUD_NETWORK_IDENTITY_INTROSPECTION_URL`
- `GOREECLOUD_NETWORK_IDENTITY_CLIENT_ID`
- `GOREECLOUD_NETWORK_IDENTITY_CLIENT_SECRET`

Optional issuer/audience checks are available through `GOREECLOUD_NETWORK_IDENTITY_EXPECTED_ISSUER` and `GOREECLOUD_NETWORK_IDENTITY_EXPECTED_AUDIENCE`. The required Network administrator scope defaults to `goreecloud.network.admin` and can be set with `GOREECLOUD_NETWORK_IDENTITY_ADMIN_SCOPE`.

Non-loopback introspection endpoints must use HTTPS. Real client secrets must not be committed to source control.

Key endpoints:

- `GET /healthz`
- `GET /api/v1/status`
- `GET /api/v1/overview`
- `GET /api/v1/storage`
- `GET /api/v1/authentication`
- `GET /api/v1/admin/session` — protected read-only Identity boundary probe
- `GET /api/v1/devices`
- `POST /api/v1/access/evaluate`
- `GET /api/v1/platform-systems`
- `GET /api/v1/capabilities`

`POST /api/v1/access/evaluate` evaluates the currently loaded Development policy state. It does not authenticate a caller, mutate policy, issue credentials, establish a tunnel, or enforce packet flow.

## Validation

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
node --check apps/web/app.js
(cd apps/ios && swift test)
```

Android and Google TV require an Android SDK with API 36 and Gradle 9.5.0. iOS application release validation requires an Apple toolchain with SwiftUI support. Exact-revision CI is required before merge/release claims.

## Explicit Development boundaries

Not implemented or production-validated yet:

- Production database, multi-writer, high-availability, backup, restore, or rollback persistence.
- Accepted GoreeCloud Identity runtime integration or production SSO.
- Mutating administrative APIs for devices, resources, or policy.
- Device enrollment or key issuance.
- WireGuard tunnel lifecycle or packet enforcement.
- Private routing, signal/path coordination, relay, or obfuscation.
- Posture enforcement.
- Runtime-integrated GoreeCloud Manager, Privacy Shield, Wardveil Security, Everkeep, Glaze UI, or GoreeCloud Mesh contracts.
- Production deployment or Stable qualification.

Repository documentation and manifests describe evidence; they do not substitute for runtime validation.
