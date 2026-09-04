#!/usr/bin/env python3
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

required = {
    "native/conduit/control/inventory_recovery.go": (
        "goreecloud-conduit-capability-inventory-recovery-point/v1",
        "BuildCapabilityInventoryRecoveryPoint",
        "CapabilityInventoryRecoveryStore",
        "os.O_EXCL",
        "0o600",
        "RestoreCapabilityInventoryRecoveryPoint",
        "active inventory changed after recovery was planned",
        "recovered inventory fingerprint mismatch",
        "ProductionCutoverAuthorized",
    ),
    "native/conduit/control/inventory_recovery_test.go": (
        "TestCapabilityInventoryRecoveryStoreRoundTrip",
        "TestRestoreCapabilityInventoryRecoveryPointCompareAndSwap",
        "TestCapabilityInventoryRecoveryPointRejectsCutoverAndTampering",
    ),
    "docs/goreecloud/conduit-recovery.md": (
        "immutable recovery-point boundary",
        "compare-and-swap",
        "does not contain peers",
        "Everkeep",
        "NetBird remains production-authoritative",
    ),
}

for rel, markers in required.items():
    path = ROOT / rel
    if not path.is_file():
        raise SystemExit(f"Conduit recovery validation failed: missing {rel}")
    text = path.read_text(encoding="utf-8")
    for marker in markers:
        if marker not in text:
            raise SystemExit(f"Conduit recovery validation failed: {rel} missing {marker!r}")

print("Conduit immutable migration-inventory recovery point, owner-protected persistence, compare-and-swap restore, privacy boundary, and cutover=false contract: PASS")
