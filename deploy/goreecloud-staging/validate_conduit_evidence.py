#!/usr/bin/env python3
"""Validate bounded GoreeCloud Conduit isolated-runtime evidence without external dependencies."""

from __future__ import annotations

import argparse
import json
from pathlib import Path
import re

FULL_SHA = re.compile(r"^[0-9a-f]{40}$")
ARTIFACT = re.compile(r"^sha256:[0-9a-f]{64}$")
EXPECTED_FIELDS = {
    "schema", "feature_id", "source_sha", "artifact_digest", "environment",
    "authority", "compatibility_bridge_active", "production_cutover_authorized", "checks",
}
EXPECTED_CHECKS = {"source", "unit", "integration", "security_privacy", "isolated_runtime"}


def fail(message: str) -> None:
    raise SystemExit(f"error: {message}")


def validate(payload: object) -> None:
    if not isinstance(payload, dict) or set(payload) != EXPECTED_FIELDS:
        fail("Conduit evidence must contain exactly the v1 evidence fields")
    if payload["schema"] != "goreecloud-conduit-isolated-evidence/v1":
        fail("unsupported Conduit evidence schema")
    if not isinstance(payload["feature_id"], str) or not payload["feature_id"].strip():
        fail("feature_id is required")
    if not isinstance(payload["source_sha"], str) or not FULL_SHA.fullmatch(payload["source_sha"]):
        fail("source_sha must be an exact lowercase 40-character Git SHA")
    if not isinstance(payload["artifact_digest"], str) or not ARTIFACT.fullmatch(payload["artifact_digest"]):
        fail("artifact_digest must be sha256:<64 lowercase hexadecimal characters>")
    if payload["environment"] != "isolated-staging":
        fail("Conduit evidence is accepted only from isolated-staging")
    if payload["authority"] not in {"inherited", "transitional"}:
        fail("Conduit evidence cannot claim native authority")
    if payload["compatibility_bridge_active"] is not True:
        fail("compatibility bridge must remain active")
    if payload["production_cutover_authorized"] is not False:
        fail("isolated evidence cannot authorize production cutover")
    checks = payload["checks"]
    if not isinstance(checks, dict) or set(checks) != EXPECTED_CHECKS:
        fail("Conduit evidence checks must contain exactly the v1 bounded checks")
    failed = sorted(name for name, value in checks.items() if value is not True)
    if failed:
        fail(f"Conduit isolated evidence contains incomplete checks: {failed}")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("evidence", type=Path)
    args = parser.parse_args()
    try:
        payload = json.loads(args.evidence.read_text(encoding="utf-8"))
    except (OSError, UnicodeError, json.JSONDecodeError) as exc:
        fail(f"cannot read Conduit evidence: {exc}")
    validate(payload)
    print("GoreeCloud Conduit isolated-runtime evidence contract: PASS")


if __name__ == "__main__":
    main()
