# Security

## Current security posture

This repository is Development and not production-ready.

The bootstrap server:

- binds to `127.0.0.1:8080` by default;
- exposes read-only development discovery endpoints;
- performs no device enrollment, key issuance, tunnel setup, route changes, or policy mutation;
- does not implement production authentication or authorization;
- sets basic browser security headers for served dashboard/API responses.

Do not expose the Development server to untrusted networks as if it were production-ready.

## Secrets

Never commit passwords, API tokens, private keys, WireGuard private keys, signing keys, recovery codes, production credentials, or active secret-bearing environment files.

## Future gates

Production use requires Wardveil Security integration, GoreeCloud Identity authorization, Privacy Shield data authorization, dependency review, secure key storage, network-lab validation, recovery validation, and release-signing/provenance evidence.
