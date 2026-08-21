#!/usr/bin/env python3

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
REGISTRY = ROOT / "native" / "conduit" / "features.json"

REQUIRED_IDS = {
    "control", "api", "devices", "principals", "groups", "access", "posture",
    "resources", "routes", "enrollment", "signal", "relay", "path", "dns-delivery",
    "activity", "diagnostics", "console", "android", "desktop", "apple", "release",
    "commercial-telemetry",
}
ALLOWED_STATUS = {"planned", "in-progress", "native", "retire"}


def fail(message: str) -> None:
    raise SystemExit(f"Conduit native-feature validation failed: {message}")


def main() -> None:
    data = json.loads(REGISTRY.read_text(encoding="utf-8"))

    if data.get("schema") != "goreecloud-conduit-native-features/v1":
        fail("unexpected schema")
    if data.get("policy") != "native-by-destination":
        fail("policy must be native-by-destination")
    if data.get("capability_identity") != "GoreeCloud Conduit":
        fail("capability identity must be GoreeCloud Conduit")
    if data.get("production_migration_authorized") is not False:
        fail("source planning must not authorize production migration")

    standards = data.get("standards_dependencies", [])
    wireguard = next((item for item in standards if item.get("name") == "WireGuard"), None)
    if not wireguard or wireguard.get("native_reimplementation_required") is not False:
        fail("WireGuard must remain classified as an allowed standards dependency unless separately approved")

    features = data.get("features", [])
    ids = [feature.get("id") for feature in features]
    if len(ids) != len(set(ids)):
        fail("duplicate feature id")
    missing = REQUIRED_IDS - set(ids)
    if missing:
        fail(f"missing required features: {', '.join(sorted(missing))}")

    for feature in features:
        feature_id = feature.get("id")
        status = feature.get("status")
        if status not in ALLOWED_STATUS:
            fail(f"{feature_id}: invalid status {status!r}")

        replacement_required = feature.get("replacement_required")
        native = feature.get("native")
        if replacement_required:
            if not isinstance(native, str) or not native.startswith("Conduit "):
                fail(f"{feature_id}: replacement-required feature needs a Conduit native destination")
            if status == "retire":
                fail(f"{feature_id}: replacement-required feature cannot be retire-only")
        else:
            if feature_id != "commercial-telemetry":
                fail(f"{feature_id}: only explicitly retired non-product features may omit replacement")
            if status != "retire" or native is not None:
                fail("commercial/telemetry surface must be retired rather than recreated")

    print(f"Conduit native-feature registry valid: {len(features)} feature records")


if __name__ == "__main__":
    main()
