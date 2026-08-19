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

The host must have an independent administrative recovery path that does not depend on this staging network. Production peers and production credentials must never be required to recover the staging host.

## Image policy

The harness does not select upstream images automatically. Before startup, provide exact image references for:

- management;
- signal;
- relay;
- dashboard;
- coturn.

For GoreeCloud-controlled components, the image should be built from the exact Git commit under test and recorded in the acceptance evidence. Prefer immutable registry digests (`image@sha256:...`) once an artifact exists. If a temporary tag is used during local development, record the image ID and source commit before testing and do not treat that run as release evidence.

## Initial workflow

1. Provision a dedicated non-production host/VM and preserve independent SSH/console access.
2. Copy `.env.example` to `.env` and fill only staging values.
3. Build or obtain images for the exact source revisions under test.
4. Run `./render-config.sh`.
5. Run `./validate-staging.sh`.
6. Inspect the rendered Compose model with `docker compose config`.
7. Start the stack only after the preflight passes.
8. Create dedicated staging identities, setup keys, groups, policies, resources, routes, and peers. Do not import production state.
9. Execute the isolated control-plane, dashboard, and Android acceptance plans.
10. Retain exact source SHAs, image digests/IDs, configuration hashes, test timestamps, results, and rollback evidence.

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