#!/usr/bin/env python3
"""Render and validate GoreeCloud Network isolated-staging configuration.

This utility deliberately treats .env values as data. It never sources the file as shell code.
Only Python's standard library is required; Docker Compose validation is used when available.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys

ROOT = Path(__file__).resolve().parent
ENV_PATH = ROOT / ".env"
RUNTIME = ROOT / "runtime"
PLACEHOLDER = re.compile(r"\$\{([A-Z0-9_]+)\}")

PRODUCTION_MARKERS = (
    "netbird.goreecloud.com",
    "100.71.27.119",
)

REQUIRED_RENDER_VALUES = (
    "STAGING_DOMAIN",
    "STAGING_DNS_DOMAIN",
    "STAGING_MANAGEMENT_PORT",
    "STAGING_SIGNAL_PORT",
    "STAGING_TURN_PORT",
    "STAGING_SIGNAL_ENDPOINT",
    "STAGING_RELAY_ENDPOINT",
    "STAGING_AUTH_AUDIENCE",
    "STAGING_AUTH_CLIENT_ID",
    "STAGING_AUTH_AUTHORITY",
    "STAGING_AUTH_JWT_CERTS",
    "STAGING_AUTH_TOKEN_ENDPOINT",
    "STAGING_AUTH_DEVICE_AUTH_PROVIDER",
    "STAGING_AUTH_DEVICE_AUTH_CLIENT_ID",
    "STAGING_AUTH_DEVICE_AUTH_ENDPOINT",
    "STAGING_AUTH_DEVICE_AUTH_SCOPE",
    "STAGING_AUTH_PKCE_AUTHORIZATION_ENDPOINT",
    "STAGING_AUTH_PKCE_REDIRECT_URLS",
    "STAGING_AUTH_SUPPORTED_SCOPES",
    "STAGING_AUTH_USER_ID_CLAIM",
    "STAGING_IDP_MANAGER_TYPE",
    "STAGING_IDP_MGMT_EXTRA_CONFIG",
    "STAGING_DATASTORE_ENC_KEY",
    "STAGING_RELAY_AUTH_SECRET",
    "STAGING_TURN_USER",
    "STAGING_TURN_PASSWORD",
    "STAGING_DISABLE_DEFAULT_POLICY",
    "STAGING_STORE_ENGINE",
)

IMAGE_VALUES = (
    "GOREECLOUD_NETWORK_MANAGEMENT_IMAGE",
    "GOREECLOUD_NETWORK_SIGNAL_IMAGE",
    "GOREECLOUD_NETWORK_RELAY_IMAGE",
    "GOREECLOUD_NETWORK_DASHBOARD_IMAGE",
    "GOREECLOUD_NETWORK_COTURN_IMAGE",
)


def fail(message: str) -> None:
    raise SystemExit(f"error: {message}")


def parse_dotenv(path: Path) -> dict[str, str]:
    if not path.is_file():
        fail(f"{path.relative_to(ROOT)} is missing; copy .env.example and populate staging-only values")

    values: dict[str, str] = {}
    for number, raw in enumerate(path.read_text(encoding="utf-8").splitlines(), start=1):
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        if line.startswith("export "):
            line = line[7:].lstrip()
        if "=" not in line:
            fail(f"invalid dotenv entry on line {number}")
        key, value = line.split("=", 1)
        key = key.strip()
        value = value.strip()
        if not re.fullmatch(r"[A-Z_][A-Z0-9_]*", key):
            fail(f"invalid environment variable name {key!r} on line {number}")
        if len(value) >= 2 and value[0] == value[-1] and value[0] in ("'", '"'):
            quote = value[0]
            value = value[1:-1]
            if quote == '"':
                value = bytes(value, "utf-8").decode("unicode_escape")
        values[key] = value
    return values


def require(values: dict[str, str], names: tuple[str, ...]) -> None:
    for name in names:
        if not values.get(name, ""):
            fail(f"{name} must be set to a staging-only value")


def render_text(template: Path, values: dict[str, str]) -> str:
    text = template.read_text(encoding="utf-8")

    def replace(match: re.Match[str]) -> str:
        name = match.group(1)
        if name not in values:
            fail(f"{name} is referenced by {template.name} but is not defined in .env")
        return values[name]

    rendered = PLACEHOLDER.sub(replace, text)
    remaining = PLACEHOLDER.findall(rendered)
    if remaining:
        fail(f"unrendered placeholders remain in {template.name}: {', '.join(sorted(set(remaining)))}")
    return rendered


def ensure_nonproduction(values: dict[str, str], rendered: list[str] | None = None) -> None:
    domain = values.get("STAGING_DOMAIN", "")
    if domain in ("", "goreecloud.com", "www.goreecloud.com", "netbird.goreecloud.com"):
        fail("STAGING_DOMAIN must be a dedicated non-production hostname")

    haystacks = list(values.values())
    if rendered:
        haystacks.extend(rendered)
    joined = "\n".join(haystacks)
    for marker in PRODUCTION_MARKERS:
        if marker in joined:
            fail(f"production marker {marker!r} detected in staging configuration")


def validate_images(values: dict[str, str]) -> None:
    require(values, IMAGE_VALUES)
    for name in IMAGE_VALUES:
        image = values[name]
        if image.endswith(":latest"):
            fail(f"{name} must not use the latest tag")
        if "@sha256:" not in image and ":" not in image.rsplit("/", 1)[-1]:
            fail(f"{name} must use an explicit tag or digest")


def render() -> None:
    values = parse_dotenv(ENV_PATH)
    require(values, REQUIRED_RENDER_VALUES)
    ensure_nonproduction(values)

    management = render_text(ROOT / "management.json.template", values)
    turn = render_text(ROOT / "turnserver.conf.template", values)

    try:
        json.loads(management)
    except json.JSONDecodeError as exc:
        fail(f"rendered management.json is invalid JSON: {exc}")

    ensure_nonproduction(values, [management, turn])

    RUNTIME.mkdir(mode=0o700, parents=True, exist_ok=True)
    management_path = RUNTIME / "management.json"
    turn_path = RUNTIME / "turnserver.conf"
    management_path.write_text(management, encoding="utf-8")
    turn_path.write_text(turn, encoding="utf-8")
    management_path.chmod(0o600)
    turn_path.chmod(0o600)

    print("Rendered staging configuration:")
    print(f"  {management_path}")
    print(f"  {turn_path}")
    print("Run: python3 config_tool.py validate")


def validate() -> None:
    values = parse_dotenv(ENV_PATH)
    require(values, REQUIRED_RENDER_VALUES)
    require(values, ("STAGING_MANAGEMENT_ENDPOINT", "STAGING_RELAY_ENDPOINT"))
    validate_images(values)
    ensure_nonproduction(values)

    if values.get("COMPOSE_PROJECT_NAME") != "goreecloud-network-staging":
        fail("COMPOSE_PROJECT_NAME must be goreecloud-network-staging")
    if values.get("STAGING_DISABLE_DEFAULT_POLICY") != "true":
        fail("STAGING_DISABLE_DEFAULT_POLICY must remain true during baseline acceptance")
    if values.get("STAGING_DISABLE_ANONYMOUS_METRICS") != "true":
        fail("anonymous metrics must remain disabled")

    management_path = RUNTIME / "management.json"
    turn_path = RUNTIME / "turnserver.conf"
    if not management_path.is_file() or not turn_path.is_file():
        fail("rendered runtime configuration is missing; run python3 config_tool.py render")

    management_text = management_path.read_text(encoding="utf-8")
    turn_text = turn_path.read_text(encoding="utf-8")
    try:
        json.loads(management_text)
    except json.JSONDecodeError as exc:
        fail(f"runtime/management.json is invalid JSON: {exc}")

    ensure_nonproduction(values, [management_text, turn_text])

    docker = shutil.which("docker")
    if docker:
        result = subprocess.run(
            [docker, "compose", "--env-file", str(ENV_PATH), "-f", str(ROOT / "compose.yaml"), "config", "--quiet"],
            cwd=ROOT,
            check=False,
        )
        if result.returncode != 0:
            fail("docker compose configuration validation failed")
    else:
        print("warning: Docker not found; Compose model validation skipped", file=sys.stderr)

    bind = values.get("STAGING_BIND_ADDRESS", "127.0.0.1")
    if bind not in ("127.0.0.1", "::1"):
        print(
            f"warning: non-loopback staging bind {bind!r}; verify firewall and publication approval before startup",
            file=sys.stderr,
        )

    for path in (management_path, turn_path):
        digest = hashlib.sha256(path.read_bytes()).hexdigest()
        print(f"sha256 {path.name}: {digest}")

    print("GoreeCloud Network staging preflight passed.")
    print("This validates configuration boundaries only; it is not runtime or production acceptance.")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("command", choices=("render", "validate"))
    args = parser.parse_args()
    if args.command == "render":
        render()
    else:
        validate()


if __name__ == "__main__":
    main()
