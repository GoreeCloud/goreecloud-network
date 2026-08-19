#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

if [[ ! -f .env ]]; then
  echo "error: deploy/goreecloud-staging/.env is required; copy .env.example and populate staging-only values" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1091
source ./.env
set +a

required=(
  STAGING_DOMAIN
  STAGING_DNS_DOMAIN
  STAGING_AUTH_AUDIENCE
  STAGING_AUTH_CLIENT_ID
  STAGING_AUTH_AUTHORITY
  STAGING_AUTH_JWT_CERTS
  STAGING_AUTH_TOKEN_ENDPOINT
  STAGING_AUTH_DEVICE_AUTH_PROVIDER
  STAGING_AUTH_DEVICE_AUTH_CLIENT_ID
  STAGING_AUTH_DEVICE_AUTH_ENDPOINT
  STAGING_AUTH_PKCE_AUTHORIZATION_ENDPOINT
  STAGING_AUTH_PKCE_REDIRECT_URLS
  STAGING_DATASTORE_ENC_KEY
  STAGING_RELAY_AUTH_SECRET
  STAGING_TURN_USER
  STAGING_TURN_PASSWORD
)

for name in "${required[@]}"; do
  if [[ -z "${!name:-}" ]]; then
    echo "error: $name must be set to a staging-only value" >&2
    exit 1
  fi
done

if ! command -v envsubst >/dev/null 2>&1; then
  echo "error: envsubst is required to render staging configuration" >&2
  exit 1
fi

umask 077
mkdir -p runtime

envsubst < management.json.template > runtime/management.json
envsubst < turnserver.conf.template > runtime/turnserver.conf

printf 'Rendered staging configuration with restrictive permissions:\n'
printf '  %s\n' "$(pwd)/runtime/management.json" "$(pwd)/runtime/turnserver.conf"
printf 'Run ./validate-staging.sh before docker compose up.\n'
