# User Manual — Development Bootstrap

GoreeCloud Network is not yet released for production use.

## Development server

Run from the repository root:

```bash
go run ./cmd/network-server
```

Open `http://127.0.0.1:8080/` to view the Development Web Dashboard.

## What the dashboard means

The dashboard reports the server's current declared lifecycle and surface state. A `bootstrap`, `planned`, `not implemented`, or similar state is not an error and must not be interpreted as active protection or connectivity.

## Native clients

Android, Google TV, and iOS source projects currently provide Development status experiences and API connectivity scaffolding. They do not yet establish a VPN tunnel or private-network access.

A central canonical GoreeCloud Network user manual must be created/synchronized before release qualification.
