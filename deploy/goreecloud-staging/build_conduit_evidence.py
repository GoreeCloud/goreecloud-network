#!/usr/bin/env python3
"""Build a bounded Conduit isolated-runtime evidence envelope from explicit accepted checks."""

from __future__ import annotations

import argparse
import json
from pathlib import Path

from validate_conduit_evidence import validate

CHECK_NAMES = ("source", "unit", "integration", "security_privacy", "isolated_runtime")


def build(feature_id: str, source_sha: str, artifact_digest: str, checks: dict[str, bool]) -> dict[str, object]:
    payload: dict[str, object] = {
        "schema": "goreecloud-conduit-isolated-evidence/v1",
        "feature_id": feature_id.strip(),
        "source_sha": source_sha.strip().lower(),
        "artifact_digest": artifact_digest.strip().lower(),
        "environment": "isolated-staging",
        "authority": "inherited",
        "compatibility_bridge_active": True,
        "production_cutover_authorized": False,
        "checks": {name: checks.get(name) is True for name in CHECK_NAMES},
    }
    validate(payload)
    return payload


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--feature-id", required=True)
    parser.add_argument("--source-sha", required=True)
    parser.add_argument("--artifact-digest", required=True)
    parser.add_argument("--output", type=Path, required=True)
    for name in CHECK_NAMES:
        parser.add_argument(f"--{name.replace('_', '-')}", action="store_true")
    args = parser.parse_args()
    checks = {name: bool(getattr(args, name)) for name in CHECK_NAMES}
    payload = build(args.feature_id, args.source_sha, args.artifact_digest, checks)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(f"wrote validated Conduit isolated evidence to {args.output}")


if __name__ == "__main__":
    main()
