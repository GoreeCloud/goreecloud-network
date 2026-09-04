#!/usr/bin/env python3
"""Collect sanitized, source-bound GoreeCloud Network staging acceptance evidence.

The evidence manifest intentionally contains no credentials or secret-bearing configuration
values. Formal evidence requires immutable image digests and exact source revisions so an
acceptance result cannot be detached from the artifacts that were actually evaluated.
"""

from __future__ import annotations

import argparse
from datetime import datetime, timezone
import hashlib
import json
import platform
from pathlib import Path
import re
import shutil
import subprocess

from config_tool import (
    ENV_PATH,
    IMAGE_VALUES,
    ROOT,
    RUNTIME,
    ensure_nonproduction,
    fail,
    parse_dotenv,
    require,
)

SOURCE_VALUES = (
    "GOREECLOUD_NETWORK_SOURCE_SHA",
    "GOREECLOUD_NETWORK_DASHBOARD_SOURCE_SHA",
)

ANDROID_VALUES = (
    "GOREECLOUD_NETWORK_ANDROID_SOURCE_SHA",
    "GOREECLOUD_NETWORK_ANDROID_ARTIFACT_SHA256",
)

FULL_SHA = re.compile(r"^[0-9a-fA-F]{40}$")
SHA256 = re.compile(r"^[0-9a-fA-F]{64}$")
IMAGE_DIGEST = re.compile(r"@sha256:([0-9a-fA-F]{64})$")
CONDUIT_STATUS_SCHEMA = "goreecloud-conduit-control-status/v1"
CONDUIT_STATUS_FIELDS = {
    "schema",
    "generated_at",
    "authority",
    "migration_stage",
    "compatibility_bridge_active",
    "production_cutover_authorized",
}


def sha256_file(path: Path) -> str:
    if not path.is_file():
        fail(f"required evidence input is missing: {path.relative_to(ROOT)}")
    return hashlib.sha256(path.read_bytes()).hexdigest()


def command_output(command: list[str]) -> str | None:
    try:
        completed = subprocess.run(
            command,
            cwd=ROOT,
            check=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            text=True,
        )
    except (OSError, subprocess.CalledProcessError):
        return None
    value = completed.stdout.strip()
    return value or None


def validate_source_sha(name: str, value: str) -> str:
    if not FULL_SHA.fullmatch(value):
        fail(f"{name} must be an exact 40-character Git commit SHA")
    return value.lower()


def validate_sha256(name: str, value: str) -> str:
    if not SHA256.fullmatch(value):
        fail(f"{name} must be a 64-character SHA-256 value")
    return value.lower()


def immutable_images(values: dict[str, str]) -> dict[str, dict[str, str]]:
    require(values, IMAGE_VALUES)
    images: dict[str, dict[str, str]] = {}
    for name in IMAGE_VALUES:
        reference = values[name]
        match = IMAGE_DIGEST.search(reference)
        if not match:
            fail(
                f"{name} must use an immutable image@sha256:<digest> reference before formal acceptance evidence is collected"
            )
        images[name] = {
            "reference": reference,
            "digest": match.group(1).lower(),
        }
    return images


def verify_repository_head(expected_sha: str) -> str | None:
    git = shutil.which("git")
    if not git:
        return None
    actual = command_output([git, "rev-parse", "HEAD"])
    if actual and actual.lower() != expected_sha.lower():
        fail(
            f"checked-out repository revision {actual} does not match GOREECLOUD_NETWORK_SOURCE_SHA {expected_sha}"
        )
    return actual.lower() if actual else None


def validate_conduit_status(path: Path) -> dict[str, object]:
    if not path.is_file():
        fail(f"Conduit runtime status evidence is missing: {path}")
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, UnicodeError, json.JSONDecodeError) as exc:
        fail(f"Conduit runtime status evidence is not valid JSON: {exc}")

    if not isinstance(payload, dict):
        fail("Conduit runtime status evidence must be a JSON object")
    fields = set(payload)
    if fields != CONDUIT_STATUS_FIELDS:
        missing = sorted(CONDUIT_STATUS_FIELDS - fields)
        extra = sorted(fields - CONDUIT_STATUS_FIELDS)
        fail(f"Conduit runtime status field boundary mismatch; missing={missing}, extra={extra}")
    if payload["schema"] != CONDUIT_STATUS_SCHEMA:
        fail("Conduit runtime status schema is not the accepted v1 contract")
    if payload["authority"] != "inherited":
        fail("Conduit runtime status must retain inherited authority during this acceptance stage")
    if payload["migration_stage"] != "implementation":
        fail("Conduit runtime status must remain at implementation until isolated validation is accepted")
    if payload["compatibility_bridge_active"] is not True:
        fail("Conduit compatibility bridge must remain active during isolated runtime evidence collection")
    if payload["production_cutover_authorized"] is not False:
        fail("Conduit runtime evidence must never authorize production cutover")
    generated_at = payload["generated_at"]
    if not isinstance(generated_at, str) or not generated_at.strip():
        fail("Conduit runtime status must include a non-empty generation timestamp")
    try:
        parsed = datetime.fromisoformat(generated_at.replace("Z", "+00:00"))
    except ValueError:
        fail("Conduit runtime status generation timestamp must be ISO-8601")
    if parsed.tzinfo is None:
        fail("Conduit runtime status generation timestamp must include a timezone")

    return payload


def collect(require_android: bool, output: Path, conduit_status_json: Path | None) -> None:
    values = parse_dotenv(ENV_PATH)
    require(values, SOURCE_VALUES)
    ensure_nonproduction(values)

    source_sha = validate_source_sha("GOREECLOUD_NETWORK_SOURCE_SHA", values[SOURCE_VALUES[0]])
    dashboard_sha = validate_source_sha(
        "GOREECLOUD_NETWORK_DASHBOARD_SOURCE_SHA", values[SOURCE_VALUES[1]]
    )
    images = immutable_images(values)
    repository_head = verify_repository_head(source_sha)

    android: dict[str, str | bool | None] = {
        "required_for_this_manifest": require_android,
        "source_sha": None,
        "artifact_sha256": None,
    }
    android_source = values.get("GOREECLOUD_NETWORK_ANDROID_SOURCE_SHA", "")
    android_artifact = values.get("GOREECLOUD_NETWORK_ANDROID_ARTIFACT_SHA256", "")
    if require_android:
        require(values, ANDROID_VALUES)
    if android_source:
        android["source_sha"] = validate_source_sha(
            "GOREECLOUD_NETWORK_ANDROID_SOURCE_SHA", android_source
        )
    if android_artifact:
        android["artifact_sha256"] = validate_sha256(
            "GOREECLOUD_NETWORK_ANDROID_ARTIFACT_SHA256", android_artifact
        )

    conduit_status: dict[str, object] = {
        "attached": False,
        "payload_sha256": None,
        "payload": None,
    }
    if conduit_status_json is not None:
        conduit_status["attached"] = True
        conduit_status["payload_sha256"] = sha256_file(conduit_status_json)
        conduit_status["payload"] = validate_conduit_status(conduit_status_json)

    files = {
        "compose_yaml": ROOT / "compose.yaml",
        "management_json": RUNTIME / "management.json",
        "turnserver_conf": RUNTIME / "turnserver.conf",
    }

    docker_version = command_output(["docker", "--version"]) if shutil.which("docker") else None
    compose_version = (
        command_output(["docker", "compose", "version"]) if shutil.which("docker") else None
    )

    manifest = {
        "schema": "goreecloud-network-acceptance-evidence/v1",
        "collected_at_utc": datetime.now(timezone.utc).isoformat(),
        "project": "GoreeCloud Network",
        "environment": "isolated-staging",
        "production_authorization": False,
        "source": {
            "control_plane_sha": source_sha,
            "dashboard_sha": dashboard_sha,
            "repository_head_sha": repository_head,
        },
        "android": android,
        "images": images,
        "configuration_sha256": {
            key: sha256_file(path) for key, path in files.items()
        },
        "conduit_status": conduit_status,
        "runtime": {
            "platform": platform.platform(),
            "python": platform.python_version(),
            "docker": docker_version,
            "docker_compose": compose_version,
        },
        "assertions": {
            "immutable_image_digests_required": True,
            "production_markers_rejected_by_staging_preflight": True,
            "credentials_included": False,
            "production_migration_authorized": False,
            "conduit_runtime_status_does_not_advance_migration_stage": True,
        },
    }

    ensure_nonproduction(values, [json.dumps(manifest, sort_keys=True)])

    output.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
    output.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    output.chmod(0o600)

    print(f"Wrote sanitized acceptance evidence manifest: {output}")
    print(f"control-plane source: {source_sha}")
    print(f"dashboard source: {dashboard_sha}")
    if android["source_sha"]:
        print(f"android source: {android['source_sha']}")
    else:
        print("android source: not yet attached")
    if conduit_status["attached"]:
        print(f"Conduit status evidence: attached ({conduit_status['payload_sha256']})")
    else:
        print("Conduit status evidence: not attached")
    print("This manifest records evidence identity only; it does not authorize production migration.")


def main() -> None:
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(dest="command", required=True)
    collect_parser = subparsers.add_parser("collect")
    collect_parser.add_argument(
        "--require-android",
        action="store_true",
        help="fail unless exact Android source and APK/AAB SHA-256 evidence is supplied",
    )
    collect_parser.add_argument(
        "--conduit-status-json",
        type=Path,
        help="attach a captured loopback-local Conduit status response after strict contract validation",
    )
    collect_parser.add_argument(
        "--output",
        type=Path,
        default=RUNTIME / "acceptance-evidence.json",
        help="sanitized manifest destination",
    )
    args = parser.parse_args()

    if args.command == "collect":
        collect(args.require_android, args.output, args.conduit_status_json)


if __name__ == "__main__":
    main()
