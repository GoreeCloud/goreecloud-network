# GoreeCloud Network

GoreeCloud Network is GoreeCloud's native private networking, encrypted connectivity, remote-access, device-enrollment, private-routing, obfuscation, and Zero Trust network-access platform.

## Current state

**Version:** `0.1.0-dev.2`  
**Lifecycle:** Development  
**Repository model:** GoreeCloud-created native monorepository  
**Production status:** Not production-ready. No Stable claim is made.

The repository currently contains five coordinated product surfaces:

- `cmd/network-server` + `internal` — GoreeCloud Network Server and native Development control plane.
- `apps/web` — Web Dashboard.
- `apps/android` — Android client.
- `apps/google-tv` — Google TV client.
- `apps/ios` — iOS client and portable Swift core.

The Development control plane now exposes truthful read-only inventory/overview state plus a deterministic deny-by-default access-decision evaluator. Its state is volatile in-memory state, access evaluation is decision-only, and no packet/tunnel enforcement is implied.

Platform System adapters and security/privacy/recovery integrations remain explicitly Planned until implementation and independent validation exist.

## Run the server

```bash
go run ./cmd/network-server
```

The server binds to `127.0.0.1:8080` by default. Set `GOREECLOUD_NETWORK_ADDR` only when an intentional non-loopback development binding is required.

Key endpoints:

- `GET /healthz`
- `GET /api/v1/status`
- `GET /api/v1/overview`
- `GET /api/v1/devices`
- `POST /api/v1/access/evaluate`
- `GET /api/v1/platform-systems`
- `GET /api/v1/capabilities`

`POST /api/v1/access/evaluate` evaluates the current volatile Development policy state. It does not authenticate a caller, mutate policy, issue credentials, establish a tunnel, or enforce packet flow.

The development server also serves the Web Dashboard from `apps/web` when launched from the repository root. Override the path with `GOREECLOUD_NETWORK_WEB_ROOT`.

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

- Durable control-plane persistence and migrations.
- Authenticated administrative sessions or production authorization.
- Device enrollment or key issuance.
- WireGuard tunnel lifecycle or packet enforcement.
- Private routing, signal/path coordination, relay, or obfuscation.
- Posture enforcement.
- Runtime-integrated GoreeCloud Manager, Privacy Shield, Wardveil Security, Everkeep, Glaze UI, GoreeCloud Mesh, or GoreeCloud Identity contracts.
- Production deployment or Stable qualification.

Repository documentation and manifests describe evidence; they do not substitute for runtime validation.
