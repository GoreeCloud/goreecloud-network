# GoreeCloud Network

GoreeCloud Network is GoreeCloud's native private networking, encrypted connectivity, remote-access, device-enrollment, private-routing, obfuscation, and Zero Trust network-access platform.

## Current state

**Version:** `0.1.0-dev.3`  
**Lifecycle:** Development  
**Repository model:** GoreeCloud-created native monorepository  
**Production status:** Not production-ready. No Stable claim is made.

The repository currently contains five coordinated product surfaces:

- `cmd/network-server` + `internal` — GoreeCloud Network Server and native Development control plane.
- `apps/web` — Web Dashboard.
- `apps/android` — Android client.
- `apps/google-tv` — Google TV client.
- `apps/ios` — iOS client and portable Swift core.

The Development server now has a revisioned JSON file store for its control-plane snapshot. The store uses schema versioning, an explicit migration ledger, deterministic snapshots, SHA-256 integrity verification, restrictive local file permissions, and atomic same-directory replacement. This is a Development persistence foundation, not production database or high-availability storage.

The access evaluator remains decision-only and deny-by-default. No packet/tunnel enforcement is implied. Platform System adapters and security/privacy/recovery integrations remain explicitly Planned until implementation and independent validation exist.

## Run the server

```bash
go run ./cmd/network-server
```

The server binds to `127.0.0.1:8080` by default. Set `GOREECLOUD_NETWORK_ADDR` only when an intentional non-loopback Development binding is required.

The default Development state file is `var/network-state.json`. Override it with `GOREECLOUD_NETWORK_DATA_FILE`. The server creates an empty schema-v1 state file on first startup and fails closed if an existing supported state file cannot be parsed or passes neither schema nor checksum validation.

Key endpoints:

- `GET /healthz`
- `GET /api/v1/status`
- `GET /api/v1/overview`
- `GET /api/v1/storage`
- `GET /api/v1/devices`
- `POST /api/v1/access/evaluate`
- `GET /api/v1/platform-systems`
- `GET /api/v1/capabilities`

`GET /api/v1/storage` reports Development persistence metadata only; it does not expose state-file contents or a mutation capability.

`POST /api/v1/access/evaluate` evaluates the currently loaded Development policy state. It does not authenticate a caller, mutate policy, issue credentials, establish a tunnel, or enforce packet flow.

The Development server also serves the Web Dashboard from `apps/web` when launched from the repository root. Override the path with `GOREECLOUD_NETWORK_WEB_ROOT`.

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
- Authenticated administrative sessions or production authorization.
- Mutating administrative APIs for devices, resources, or policy.
- Device enrollment or key issuance.
- WireGuard tunnel lifecycle or packet enforcement.
- Private routing, signal/path coordination, relay, or obfuscation.
- Posture enforcement.
- Runtime-integrated GoreeCloud Manager, Privacy Shield, Wardveil Security, Everkeep, Glaze UI, GoreeCloud Mesh, or GoreeCloud Identity contracts.
- Production deployment or Stable qualification.

Repository documentation and manifests describe evidence; they do not substitute for runtime validation.
