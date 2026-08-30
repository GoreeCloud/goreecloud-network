#!/usr/bin/env python3
"""Validate source-bound GoreeCloud Network isolated-staging recovery evidence.

This module defines the evidence contract used after an external backup/restore
executor such as Everkeep has actually preserved and restored staging runtime
state. It intentionally carries artifact identities and restore results only; it
does not embed credentials, private keys, peer data, routes, packet data, DNS
queries, or unrestricted runtime configuration.
"""

from __future__ import annotations

from datetime import datetime
import hashlib
import json
import re
from typing import Any

RECOVERY_PLAN_SCHEMA = "goreecloud-network-staging-recovery-plan/v1"
RECOVERY_RECEIPT_SCHEMA = "goreecloud-network-staging-recovery-receipt/v1"
FULL_SHA = re.compile(r"^[0-9a-f]{40}$")
SHA256 = re.compile(r"^[0-9a-f]{64}$")

REQUIRED_COMPONENTS = {
    "management-state",
    "relay-state",
    "network-runtime-config",
    "conduit-migration-inventory",
    "dashboard-runtime-config",
}


def _fail(message: str) -> None:
    raise ValueError(f"goreecloud network recovery: {message}")


def _timestamp(value: Any, field: str) -> str:
    if not isinstance(value, str) or not value.strip():
        _fail(f"{field} is required")
    try:
        parsed = datetime.fromisoformat(value.replace("Z", "+00:00"))
    except ValueError as exc:
        raise ValueError(f"goreecloud network recovery: {field} must be ISO-8601") from exc
    if parsed.tzinfo is None:
        _fail(f"{field} must include a timezone")
    return value


def _sha256(value: Any, field: str) -> str:
    if not isinstance(value, str):
        _fail(f"{field} must be a SHA-256 string")
    normalized = value.strip().lower()
    if not SHA256.fullmatch(normalized):
        _fail(f"{field} must be exactly 64 lowercase hexadecimal characters")
    return normalized


def _source_sha(value: Any) -> str:
    if not isinstance(value, str):
        _fail("source_revision must be a Git commit SHA")
    normalized = value.strip().lower()
    if not FULL_SHA.fullmatch(normalized):
        _fail("source_revision must be exactly 40 lowercase hexadecimal characters")
    return normalized


def canonical_sha256(payload: dict[str, Any]) -> str:
    encoded = json.dumps(payload, sort_keys=True, separators=(",", ":")).encode("utf-8")
    return hashlib.sha256(encoded).hexdigest()


def validate_recovery_plan(plan: dict[str, Any]) -> dict[str, Any]:
    expected_fields = {
        "schema",
        "created_at",
        "environment",
        "source_revision",
        "backup_set_sha256",
        "components",
        "credentials_included",
        "production_cutover_authorized",
    }
    if not isinstance(plan, dict) or set(plan) != expected_fields:
        _fail("recovery plan field boundary mismatch")
    if plan["schema"] != RECOVERY_PLAN_SCHEMA:
        _fail("unsupported recovery plan schema")
    _timestamp(plan["created_at"], "created_at")
    if plan["environment"] != "isolated-staging":
        _fail("recovery plan environment must be isolated-staging")
    plan["source_revision"] = _source_sha(plan["source_revision"])
    plan["backup_set_sha256"] = _sha256(plan["backup_set_sha256"], "backup_set_sha256")
    if plan["credentials_included"] is not False:
        _fail("recovery plan evidence must not include credentials")
    if plan["production_cutover_authorized"] is not False:
        _fail("recovery plan cannot authorize production cutover")

    components = plan["components"]
    if not isinstance(components, list) or not components:
        _fail("recovery plan components must be a non-empty list")
    seen: set[str] = set()
    for component in components:
        if not isinstance(component, dict) or set(component) != {
            "component",
            "artifact_sha256",
            "encrypted",
            "restore_required",
        }:
            _fail("recovery component field boundary mismatch")
        name = component["component"]
        if not isinstance(name, str) or name not in REQUIRED_COMPONENTS:
            _fail("recovery plan contains an unsupported component")
        if name in seen:
            _fail(f"duplicate recovery component: {name}")
        seen.add(name)
        component["artifact_sha256"] = _sha256(component["artifact_sha256"], f"{name}.artifact_sha256")
        if component["encrypted"] is not True:
            _fail(f"{name} recovery artifact must be encrypted")
        if component["restore_required"] is not True:
            _fail(f"{name} must be required for staging restore")
    if seen != REQUIRED_COMPONENTS:
        missing = sorted(REQUIRED_COMPONENTS - seen)
        _fail(f"recovery plan is missing required components: {missing}")
    return plan


def validate_recovery_receipt(receipt: dict[str, Any], plan: dict[str, Any]) -> dict[str, Any]:
    validate_recovery_plan(plan)
    expected_fields = {
        "schema",
        "started_at",
        "completed_at",
        "environment",
        "source_revision",
        "plan_sha256",
        "restored_components",
        "state_equivalence_sha256_before",
        "state_equivalence_sha256_after",
        "all_components_restored",
        "rollback_rehearsed",
        "credentials_included",
        "production_cutover_authorized",
    }
    if not isinstance(receipt, dict) or set(receipt) != expected_fields:
        _fail("recovery receipt field boundary mismatch")
    if receipt["schema"] != RECOVERY_RECEIPT_SCHEMA:
        _fail("unsupported recovery receipt schema")
    started = _timestamp(receipt["started_at"], "started_at")
    completed = _timestamp(receipt["completed_at"], "completed_at")
    if datetime.fromisoformat(completed.replace("Z", "+00:00")) < datetime.fromisoformat(started.replace("Z", "+00:00")):
        _fail("recovery receipt completed_at precedes started_at")
    if receipt["environment"] != "isolated-staging":
        _fail("recovery receipt environment must be isolated-staging")
    receipt["source_revision"] = _source_sha(receipt["source_revision"])
    if receipt["source_revision"] != plan["source_revision"]:
        _fail("recovery receipt source revision does not match the plan")
    receipt["plan_sha256"] = _sha256(receipt["plan_sha256"], "plan_sha256")
    if receipt["plan_sha256"] != canonical_sha256(plan):
        _fail("recovery receipt plan SHA-256 does not match the exact plan")

    restored = receipt["restored_components"]
    if not isinstance(restored, dict) or set(restored) != REQUIRED_COMPONENTS:
        _fail("recovery receipt must identify every required restored component")
    expected_digests = {item["component"]: item["artifact_sha256"] for item in plan["components"]}
    for component, digest in restored.items():
        normalized = _sha256(digest, f"restored_components.{component}")
        if normalized != expected_digests[component]:
            _fail(f"restored component digest mismatch for {component}")
        restored[component] = normalized

    before = _sha256(receipt["state_equivalence_sha256_before"], "state_equivalence_sha256_before")
    after = _sha256(receipt["state_equivalence_sha256_after"], "state_equivalence_sha256_after")
    if before != after:
        _fail("restored staging state is not equivalent to the retained pre-change state")
    if receipt["all_components_restored"] is not True:
        _fail("recovery receipt must confirm every required component was restored")
    if receipt["rollback_rehearsed"] is not True:
        _fail("recovery receipt must confirm rollback rehearsal")
    if receipt["credentials_included"] is not False:
        _fail("recovery receipt evidence must not include credentials")
    if receipt["production_cutover_authorized"] is not False:
        _fail("recovery receipt cannot authorize production cutover")
    return receipt
