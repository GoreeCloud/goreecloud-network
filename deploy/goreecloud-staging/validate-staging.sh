#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

fail() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

[[ -f .env ]] || fail ".env is missing; copy .env.example and populate staging-only values"

set -a
# shellcheck disable=SC1091
source ./.env
set +a

[[ "${COMPOSE_PROJECT_NAME:-}" == "goreecloud-network-staging" ]] || fail "COMPOSE_PROJECT_NAME must be goreecloud-network-staging"
[[ "${STAGING_DISABLE_DEFAULT_POLICY:-}" == "true" ]] || fail "STAGING_DISABLE_DEFAULT_POLICY must remain true during baseline acceptance"
[[ "${STAGING_DISABLE_ANONYMOUS_METRICS:-}" == "true" ]] || fail "anonymous metrics must remain disabled"

case "${STAGING_DOMAIN:-}" in
  ""|netbird.goreecloud.com|goreecloud.com|www.goreecloud.com)
    fail "STAGING_DOMAIN must be a dedicated non-production hostname"
    ;;
esac

for value in \
  "${STAGING_MANAGEMENT_ENDPOINT:-}" \
  "${STAGING_SIGNAL_ENDPOINT:-}" \
  "${STAGING_RELAY_ENDPOINT:-}" \
  "${STAGING_TURN_ENDPOINT:-}"; do
  [[ "$value" != *"netbird.goreecloud.com"* ]] || fail "production NetBird hostname detected in staging endpoint"
done

# Known production NetBird/private service address from the current GoreeCloud records.
for file in .env runtime/management.json runtime/turnserver.conf; do
  [[ -e "$file" ]] || continue
  if grep -Fq '100.71.27.119' "$file"; then
    fail "known production NetBird/private-service address detected in $file"
  fi
done

images=(
  GOREECLOUD_NETWORK_MANAGEMENT_IMAGE
  GOREECLOUD_NETWORK_SIGNAL_IMAGE
  GOREECLOUD_NETWORK_RELAY_IMAGE
  GOREECLOUD_NETWORK_DASHBOARD_IMAGE
  GOREECLOUD_NETWORK_COTURN_IMAGE
)

for name in "${images[@]}"; do
  value="${!name:-}"
  [[ -n "$value" ]] || fail "$name is empty"
  [[ "$value" != *":latest" ]] || fail "$name must not use the latest tag"
  if [[ "$value" != *@sha256:* && "$value" != *:* ]]; then
    fail "$name must use an explicit tag or digest"
  fi
done

secrets=(
  STAGING_DATASTORE_ENC_KEY
  STAGING_RELAY_AUTH_SECRET
  STAGING_TURN_PASSWORD
)
for name in "${secrets[@]}"; do
  [[ -n "${!name:-}" ]] || fail "$name is empty"
done

[[ -f runtime/management.json ]] || fail "runtime/management.json is missing; run ./render-config.sh"
[[ -f runtime/turnserver.conf ]] || fail "runtime/turnserver.conf is missing; run ./render-config.sh"

if grep -R -n -E '\$\{STAGING_[A-Z0-9_]+\}' runtime/management.json runtime/turnserver.conf >/dev/null; then
  fail "unrendered STAGING_* placeholders remain in runtime configuration"
fi

if command -v jq >/dev/null 2>&1; then
  jq -e . runtime/management.json >/dev/null || fail "runtime/management.json is not valid JSON"
else
  printf 'warning: jq not found; JSON syntax check skipped\n' >&2
fi

if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  docker compose --env-file .env -f compose.yaml config --quiet || fail "docker compose configuration validation failed"
else
  printf 'warning: Docker Compose not found; Compose model validation skipped\n' >&2
fi

if [[ "${STAGING_BIND_ADDRESS:-127.0.0.1}" != "127.0.0.1" && "${STAGING_BIND_ADDRESS:-}" != "::1" ]]; then
  printf 'warning: staging ports are configured for non-loopback binding (%s). Verify firewall and publication approval before startup.\n' "${STAGING_BIND_ADDRESS}" >&2
fi

printf 'GoreeCloud Network staging preflight passed.\n'
printf 'This validates configuration boundaries only; it is not runtime or production acceptance.\n'
