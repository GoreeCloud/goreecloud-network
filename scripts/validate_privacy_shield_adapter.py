#!/usr/bin/env python3
"""Validate the GoreeCloud Network Privacy Shield adapter declaration."""

from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
ADAPTER = ROOT / "privacy-shield" / "adapter.json"
DOC = ROOT / "docs" / "goreecloud" / "privacy-shield-integration.md"

EXPECTED_CAPABILITIES = ["network-privacy"]
FORBIDDEN_CAPABILITIES = {
    "content-blocking",
    "tracking-resistance",
    "url-cleaning",
    "dns-privacy",
    "telemetry-minimization",
    "data-minimization",
    "retention-controls",
    "deletion-controls",
    "portable-export",
    "privacy-status",
    "user-visible-exceptions",
}


def fail(message: str) -> None:
    print(f"ERROR: {message}", file=sys.stderr)
    raise SystemExit(1)


def main() -> None:
    if not ADAPTER.is_file():
        fail("missing privacy-shield/adapter.json")
    if not DOC.is_file():
        fail("missing docs/goreecloud/privacy-shield-integration.md")

    try:
        adapter = json.loads(ADAPTER.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        fail(f"invalid Privacy Shield adapter JSON: {exc}")

    if adapter.get("schema_version") != 1:
        fail("unsupported Privacy Shield adapter schema version")

    metadata = adapter.get("adapter")
    if not isinstance(metadata, dict):
        fail("adapter metadata is missing")
    expected_metadata = {
        "id": "network",
        "product": "GoreeCloud Network",
        "runtime_authority": "GoreeCloud/goreecloud-network",
        "contract_version": 1,
    }
    for key, expected in expected_metadata.items():
        if metadata.get(key) != expected:
            fail(f"adapter metadata {key} drifted from {expected!r}")

    capabilities = adapter.get("capabilities")
    if capabilities != EXPECTED_CAPABILITIES:
        fail(f"Network adapter must declare only {EXPECTED_CAPABILITIES!r}")
    leaked = set(capabilities or []) & FORBIDDEN_CAPABILITIES
    if leaked:
        fail(f"Network adapter claims unsupported Privacy Shield capabilities: {sorted(leaked)}")

    privacy = adapter.get("privacy")
    if not isinstance(privacy, dict):
        fail("privacy guarantees are missing")
    required_privacy = {
        "local_first": True,
        "raw_private_activity_exported_for_status": False,
        "remote_tracker_learning": False,
        "remote_tracker_telemetry": False,
    }
    for key, expected in required_privacy.items():
        if privacy.get(key) is not expected:
            fail(f"privacy guarantee {key} must remain {expected!r}")

    acceptance = adapter.get("acceptance")
    if not isinstance(acceptance, dict):
        fail("acceptance boundary is missing")
    if acceptance.get("runtime_acceptance_required") is not True:
        fail("runtime acceptance must remain required")
    if acceptance.get("production_approved") is not False:
        fail("Network Privacy Shield adapter must remain unapproved until runtime acceptance")

    text = DOC.read_text(encoding="utf-8").lower()
    for phrase in (
        "enforcement adapter for the `network-privacy` capability",
        "privacy shield is not a vpn engine",
        "does not declare browser",
        "must not export raw network flows",
        "network reachability and application authorization remain separate concerns",
        "production_approved=false",
        "shared privacy shield contract validation is not runtime acceptance",
        "does not intentionally change wireguard/tun behavior",
    ):
        if phrase not in text:
            fail(f"Privacy Shield Network integration document missing boundary: {phrase}")

    print("GoreeCloud Network Privacy Shield adapter validation passed.")


if __name__ == "__main__":
    main()
