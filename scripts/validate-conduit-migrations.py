#!/usr/bin/env python3

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
FEATURES_PATH = ROOT / "native" / "conduit" / "features.json"
MIGRATIONS_PATH = ROOT / "native" / "conduit" / "migrations.json"

BASE_GATES = {
    "source",
    "unit",
    "integration",
    "security-privacy",
    "migration",
    "rollback",
    "isolated-runtime",
}
OPTIONAL_GATES = {
    "backup-restore",
    "network-path",
    "real-device",
    "dns-resolution",
    "signed-artifact",
    "accessibility",
    "restrictive-network",
}
ALLOWED_STATE_AUTHORITY = {"inherited", "transitional", "native", "none"}


def fail(message: str) -> None:
    raise SystemExit(f"Conduit migration validation failed: {message}")


def load(path: Path) -> dict:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        fail(f"missing required registry: {path.relative_to(ROOT)}")
    except json.JSONDecodeError as exc:
        fail(f"invalid JSON in {path.relative_to(ROOT)}: {exc}")


def main() -> None:
    features_doc = load(FEATURES_PATH)
    migrations_doc = load(MIGRATIONS_PATH)

    if migrations_doc.get("schema") != "goreecloud-conduit-migrations/v1":
        fail("unexpected migration schema")
    if migrations_doc.get("product") != "GoreeCloud Network":
        fail("product must be GoreeCloud Network")
    if migrations_doc.get("capability_identity") != "GoreeCloud Conduit":
        fail("capability identity must be GoreeCloud Conduit")
    if migrations_doc.get("production_cutover_authorized") is not False:
        fail("source migration contracts must not authorize production cutover")

    allowed_stages = set(migrations_doc.get("allowed_stages", []))
    if not {"inventory", "contract", "implementation", "isolated-validation", "migration-ready", "production-candidate", "native-accepted", "retired"}.issubset(allowed_stages):
        fail("required migration stages are missing")

    feature_map = {item.get("id"): item for item in features_doc.get("features", [])}
    migrations = migrations_doc.get("migrations", [])
    migration_map = {item.get("feature_id"): item for item in migrations}

    if len(migration_map) != len(migrations):
        fail("duplicate feature_id in migration registry")

    if set(feature_map) != set(migration_map):
        missing = sorted(set(feature_map) - set(migration_map))
        extra = sorted(set(migration_map) - set(feature_map))
        fail(f"migration/feature registry mismatch; missing={missing}, extra={extra}")

    allowed_gates = BASE_GATES | OPTIONAL_GATES

    for feature_id, feature in feature_map.items():
        migration = migration_map[feature_id]
        stage = migration.get("stage")
        if stage not in allowed_stages:
            fail(f"{feature_id}: invalid stage {stage!r}")

        authority = migration.get("state_authority")
        if authority not in ALLOWED_STATE_AUTHORITY:
            fail(f"{feature_id}: invalid state_authority {authority!r}")

        stateful = migration.get("stateful")
        if not isinstance(stateful, bool):
            fail(f"{feature_id}: stateful must be boolean")

        gates = migration.get("gates")
        if not isinstance(gates, list) or not gates:
            fail(f"{feature_id}: gates must be a non-empty list")
        if len(gates) != len(set(gates)):
            fail(f"{feature_id}: duplicate acceptance gate")
        unknown_gates = set(gates) - allowed_gates
        if unknown_gates:
            fail(f"{feature_id}: unknown gates {sorted(unknown_gates)}")

        retirement_condition = migration.get("retirement_condition")
        if not isinstance(retirement_condition, str) or not retirement_condition.strip():
            fail(f"{feature_id}: retirement_condition is required")

        if feature.get("replacement_required"):
            bridge = migration.get("bridge")
            if not isinstance(bridge, str) or not bridge.strip():
                fail(f"{feature_id}: replacement-required features need a compatibility bridge")
            if stage == "retired":
                fail(f"{feature_id}: replacement-required feature cannot be retired before native acceptance")
            if not BASE_GATES.issubset(set(gates)):
                missing_gates = sorted(BASE_GATES - set(gates))
                fail(f"{feature_id}: missing mandatory gates {missing_gates}")
            if stateful and "backup-restore" not in gates:
                fail(f"{feature_id}: stateful replacement requires backup-restore gate")
            if authority == "native" and stage != "native-accepted":
                fail(f"{feature_id}: native state authority requires native-accepted stage")
            if stage == "native-accepted" and authority != "native":
                fail(f"{feature_id}: native-accepted stage requires native state authority")
        else:
            if feature_id != "commercial-telemetry":
                fail(f"{feature_id}: unexpected non-replacement feature")
            if stage != "retired" or authority != "none" or migration.get("bridge") is not None:
                fail("commercial-telemetry must be retired with no state authority or compatibility bridge")

    print(f"Conduit migration registry valid: {len(migrations)} migration contracts")


if __name__ == "__main__":
    main()
