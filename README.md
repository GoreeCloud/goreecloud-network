# GoreeCloud Network

GoreeCloud Network is GoreeCloud's native private networking, encrypted connectivity, remote-access, device-enrollment, private-routing, obfuscation, and Zero Trust network-access platform.

## Current state

**Version:** `0.1.0-dev.1`  
**Lifecycle:** Development  
**Repository model:** GoreeCloud-created native monorepository  
**Production status:** Not production-ready. No Stable claim is made.

This bootstrap establishes five coordinated product surfaces:

- `cmd/network-server` + `internal/httpapi` — GoreeCloud Network Server and versioned control API.
- `apps/web` — Web Dashboard.
- `apps/android` — Android client.
- `apps/google-tv` — Google TV client.
- `apps/ios` — iOS client.

All surfaces consume the same versioned Network API contract. Platform System adapters and security/privacy/recovery integrations remain explicitly Development or Planned until implementation and independent validation exist.

## Run the server

```bash
go run ./cmd/network-server
```

The server binds to `127.0.0.1:8080` by default. Set `GOREECLOUD_NETWORK_ADDR` only when an intentional non-loopback development binding is required.

Key endpoints:

- `GET /healthz`
- `GET /api/v1/status`
- `GET /api/v1/platform-systems`
- `GET /api/v1/capabilities`

The development server also serves the Web Dashboard from `apps/web` when launched from the repository root. Override the path with `GOREECLOUD_NETWORK_WEB_ROOT`.

## Validation

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
node --check apps/web/app.js
```

Android and Google TV require an Android SDK with API 37 and Gradle 9.5.0. iOS requires an Apple toolchain with SwiftUI support. Their native project sources are included in this bootstrap, but platform builds must be validated on the corresponding supported toolchains before any release claim.

## Repository controls

The root governance files are intentionally explicit about current gaps. A documentation or manifest declaration does not establish Platform System conformance, production readiness, security, privacy, or Stable qualification.
