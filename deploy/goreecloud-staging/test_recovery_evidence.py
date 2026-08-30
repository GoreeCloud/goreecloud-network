#!/usr/bin/env python3

from __future__ import annotations

from copy import deepcopy
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from recovery_evidence import (  # noqa: E402
    RECOVERY_PLAN_SCHEMA,
    RECOVERY_RECEIPT_SCHEMA,
    REQUIRED_COMPONENTS,
    canonical_sha256,
    validate_recovery_plan,
    validate_recovery_receipt,
)


def fixture_plan() -> dict:
    components = []
    for index, component in enumerate(sorted(REQUIRED_COMPONENTS), start=1):
        components.append(
            {
                "component": component,
                "artifact_sha256": f"{index:x}" * 64,
                "encrypted": True,
                "restore_required": True,
            }
        )
    return {
        "schema": RECOVERY_PLAN_SCHEMA,
        "created_at": "2026-08-30T13:55:00Z",
        "environment": "isolated-staging",
        "source_revision": "a" * 40,
        "backup_set_sha256": "b" * 64,
        "components": components,
        "credentials_included": False,
        "production_cutover_authorized": False,
    }


def fixture_receipt(plan: dict) -> dict:
    return {
        "schema": RECOVERY_RECEIPT_SCHEMA,
        "started_at": "2026-08-30T14:00:00Z",
        "completed_at": "2026-08-30T14:03:00Z",
        "environment": "isolated-staging",
        "source_revision": plan["source_revision"],
        "plan_sha256": canonical_sha256(plan),
        "restored_components": {
            component["component"]: component["artifact_sha256"]
            for component in plan["components"]
        },
        "state_equivalence_sha256_before": "c" * 64,
        "state_equivalence_sha256_after": "c" * 64,
        "all_components_restored": True,
        "rollback_rehearsed": True,
        "credentials_included": False,
        "production_cutover_authorized": False,
    }


def expect_failure(fn, contains: str) -> None:
    try:
        fn()
    except ValueError as exc:
        if contains not in str(exc):
            raise AssertionError(f"unexpected error: {exc}") from exc
    else:
        raise AssertionError(f"expected failure containing {contains!r}")


def main() -> None:
    plan = fixture_plan()
    validate_recovery_plan(deepcopy(plan))
    validate_recovery_receipt(fixture_receipt(plan), deepcopy(plan))

    missing = fixture_plan()
    missing["components"].pop()
    expect_failure(lambda: validate_recovery_plan(missing), "missing required components")

    unencrypted = fixture_plan()
    unencrypted["components"][0]["encrypted"] = False
    expect_failure(lambda: validate_recovery_plan(unencrypted), "must be encrypted")

    cutover = fixture_plan()
    cutover["production_cutover_authorized"] = True
    expect_failure(lambda: validate_recovery_plan(cutover), "cannot authorize production cutover")

    mismatch_plan = fixture_plan()
    mismatch_receipt = fixture_receipt(mismatch_plan)
    mismatch_receipt["restored_components"][sorted(REQUIRED_COMPONENTS)[0]] = "d" * 64
    expect_failure(
        lambda: validate_recovery_receipt(mismatch_receipt, mismatch_plan),
        "restored component digest mismatch",
    )

    inequivalent_plan = fixture_plan()
    inequivalent_receipt = fixture_receipt(inequivalent_plan)
    inequivalent_receipt["state_equivalence_sha256_after"] = "e" * 64
    expect_failure(
        lambda: validate_recovery_receipt(inequivalent_receipt, inequivalent_plan),
        "not equivalent",
    )

    no_rollback_plan = fixture_plan()
    no_rollback_receipt = fixture_receipt(no_rollback_plan)
    no_rollback_receipt["rollback_rehearsed"] = False
    expect_failure(
        lambda: validate_recovery_receipt(no_rollback_receipt, no_rollback_plan),
        "rollback rehearsal",
    )

    print("GoreeCloud Network isolated-staging recovery evidence contract: PASS")


if __name__ == "__main__":
    main()
