# GoreeCloud Network isolated staging harness

This directory defines the non-production deployment harness used to execute the GoreeCloud Network acceptance plans.

## Role

The staging stack exists only to validate a specific GoreeCloud Network source revision, its dashboard, and approved test clients. It is not a production deployment template and must not be pointed at the accepted production NetBird state.

The harness is intentionally fail-closed:

- production GoreeCloud NetBird hostnames and known production peer addresses are rejected by the preflight validator;
- every container image must be supplied as an explicit immutable tag or digest rather than an implicit `latest` value;
- all persistent volumes use a `goreecloud-network-staging-` prefix;
- the default bind address is `127.0.0.1`;
- no production setup key, API token, OAuth secret, signing secret, peer database, management database, DNS configuration, policy export, or other credential/state is stored in this directory;
- the Compose project name is fixed to `goreecloud-network-staging` so staging resources remain visibly separated from production resources;
- destructive staging cleanup is limited to this Compose project and its staging volumes.

This harness implements the repository acceptance boundary described in `docs/goreecloud/isolated-acceptance-plan.md` and `docs/goreecloud/rollback-and-recovery-plan.md`.

## Files

- `compose.yaml` — isolated container topology.
- `.env.example` — non-secret variable contract. Copy it outside Git or to an ignored `.env` file before use.
- `management.json.template` — management configuration template rendered from staging-only variables.
- `render-config.sh` — renders `management.json` without committing secrets.
- `validate-staging.sh` — fail-closed static preflight for the staging environment and generated configuration.

## Required deployment model

Use a disposable or dedicated non-production Linux host or virtual machine. Do not run this stack on the current production NetBird host unless a separate, documented acceptance decision proves that port, volume, firewall, DNS, and failure-domain separation are sufficient. A separate VM is preferred.

The first repository-controlled image contract targets `linux/amd64`. This keeps management's CGO-enabled build on a native CI architecture and avoids treating unvalidated cross-compilation as acceptance evidence. Additional architectures must receive their own build and runtime evidence before use.

The host must have an independent administrative recovery path that does not depend on this staging network. Production peers and production credentials must never be required to recover the staging host.

## Exact-revision image contract

The harness does not select upstream images automatically. Before startup, provide exact image references for:

- management;
- signal;
- relay;
- dashboard;
- coturn.

The main repository contains `.github/workflows/goreecloud-staging-images.yml`. On pull requests and `agent/**` pushes it compiles the Linux/amd64 management, signal, and relay binaries and builds their container images without publishing them. A controlled `workflow_dispatch` run with `publish=true` may publish only the exact source revision under these names:

```text
ghcr.io/goreecloud/goreecloud-network-management:staging-<full-source-sha>
ghcr.io/goreecloud/goreecloud-network-signal:staging-<full-source-sha>
ghcr.io/goreecloud/goreecloud-network-relay:staging-<full-source-sha>
```

The workflow never creates `latest` and records the source revision plus binary SHA-256 values as a short-retention evidence artifact.

The dashboard repository contains `.github/workflows/goreecloud-staging-image.yml`. It runs GoreeCloud source-integrity validation, performs the production static build, builds a Linux/amd64 container, and may publish only through an explicit manual `publish=true` dispatch using:

```text
ghcr.io/goreecloud/goreecloud-network-dashboard:staging-<full-dashboard-source-sha>
```

The dashboard source SHA is intentionally independent from the control-plane source SHA. Record both revisions in every acceptance run.

For GoreeCloud-controlled components, prefer immutable registry digests (`image@sha256:...`) after publication. A commit-SHA tag is an exact source identifier but is not by itself an immutable registry identity; capture the registry digest before treating a run as formal release evidence.

Coturn remains an external supporting component and must use an explicitly pinned approved version or digest. Do not use `latest`.

## Image evidence transfer

After the two staging-image workflows complete, combine their generated dotenv fragments with the approved coturn reference in the ignored staging `.env` file. The expected variables are:

```dotenv
GOREECLOUD_NETWORK_MANAGEMENT_IMAGE=ghcr.io/goreecloud/goreecloud-network-management:staging-<main-source-sha>
GOREECLOUD_NETWORK_SIGNAL_IMAGE=ghcr.io/goreecloud/goreecloud-network-signal:staging-<main-source-sha>
GOREECLOUD_NETWORK_RELAY_IMAGE=ghcr.io/goreecloud/goreecloud-network-relay:staging-<main-source-sha>
GOREECLOUD_NETWORK_DASHBOARD_IMAGE=ghcr.io/goreecloud/goreecloud-network-dashboard:staging-<dashboard-source-sha>
GOREECLOUD_NETWORK_COTURN_IMAGE=<approved-pinned-coturn-reference>
```

Do not commit a populated `.env`, authentication material, relay secret, TURN password, or generated runtime configuration.

## Initial workflow

1. Provision a dedicated non-production Linux/amd64 host or VM and preserve independent SSH/console access.
2. Run the main repository staging-image workflow for the exact source revision under test; publish only through a deliberate manual dispatch when the images are needed remotely.
3. Run the dashboard staging-image workflow for the exact dashboard revision under test under the same publication rule.
4. Record the control-plane SHA, dashboard SHA, binary hashes, image tags, and registry digests.
5. Copy `.env.example` to `.env` and fill only staging values plus the exact image references.
6. Run `./render-config.sh`.
7. Run `./validate-staging.sh`.
8. Inspect the rendered Compose model with `docker compose config`.
9. Start the stack only after the preflight passes.
10. Create dedicated staging identities, setup keys, groups, policies, resources, routes, and peers. Do not import production state.
11. Execute the isolated control-plane, dashboard, and Android acceptance plans.
12. Retain exact source SHAs, image digests/IDs, configuration hashes, test timestamps, results, and rollback evidence.

## Publication boundary

The default configuration binds published ports to loopback so an accidental `docker compose up` does not create a broadly reachable service.

When remote test clients are required, change `STAGING_BIND_ADDRESS` only after the staging host firewall, DNS name, Caddy/reverse-proxy boundary if applicable, and permitted source networks are explicitly defined for the test. Do not reuse `netbird.goreecloud.com`. Use a staging-only hostname if one is approved.

Caddy remains the GoreeCloud HTTPS publication authority. This harness does not make the GoreeCloud Network dashboard a competing public reverse proxy.

## State and cleanup

Staging state is disposable, but evidence is not. Before a destructive reset, retain the evidence required by the acceptance plan.

Normal teardown:

```sh
docker compose --project-name goreecloud-network-staging down
```

Destructive staging reset, only after evidence retention:

```sh
docker compose --project-name goreecloud-network-staging down --volumes
```

Never substitute a production project name, production volume, or broad Docker prune operation for these commands.

## Production gate

A successful staging run does not authorize production migration by itself. Production consideration still requires the complete acceptance matrix, backup/restore proof, rollback rehearsal, artifact provenance, dashboard build evidence, Android real-device evidence, and an explicit migration decision.
