#!/usr/bin/env python3

from __future__ import annotations

import json
from pathlib import Path
import tempfile
import unittest

from acceptance_evidence import validate_conduit_status


VALID_STATUS = {
    "schema": "goreecloud-conduit-control-status/v1",
    "generated_at": "2026-08-24T15:00:00Z",
    "authority": "inherited",
    "migration_stage": "implementation",
    "compatibility_bridge_active": True,
    "production_cutover_authorized": False,
}


class ConduitRuntimeEvidenceTests(unittest.TestCase):
    def write_status(self, payload: object) -> Path:
        temporary = tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", suffix=".json", delete=False)
        with temporary:
            json.dump(payload, temporary)
        self.addCleanup(Path(temporary.name).unlink, missing_ok=True)
        return Path(temporary.name)

    def test_accepts_exact_privacy_safe_contract(self) -> None:
        path = self.write_status(VALID_STATUS)
        self.assertEqual(validate_conduit_status(path), VALID_STATUS)

    def test_rejects_extra_private_or_unapproved_fields(self) -> None:
        payload = dict(VALID_STATUS)
        payload["peer_count"] = 4
        with self.assertRaises(SystemExit):
            validate_conduit_status(self.write_status(payload))

    def test_rejects_native_authority_claim(self) -> None:
        payload = dict(VALID_STATUS)
        payload["authority"] = "native"
        with self.assertRaises(SystemExit):
            validate_conduit_status(self.write_status(payload))

    def test_rejects_migration_stage_self_promotion(self) -> None:
        payload = dict(VALID_STATUS)
        payload["migration_stage"] = "isolated-validation"
        with self.assertRaises(SystemExit):
            validate_conduit_status(self.write_status(payload))

    def test_rejects_production_cutover_claim(self) -> None:
        payload = dict(VALID_STATUS)
        payload["production_cutover_authorized"] = True
        with self.assertRaises(SystemExit):
            validate_conduit_status(self.write_status(payload))

    def test_rejects_disabled_compatibility_bridge(self) -> None:
        payload = dict(VALID_STATUS)
        payload["compatibility_bridge_active"] = False
        with self.assertRaises(SystemExit):
            validate_conduit_status(self.write_status(payload))

    def test_requires_timezone_aware_timestamp(self) -> None:
        payload = dict(VALID_STATUS)
        payload["generated_at"] = "2026-08-24T15:00:00"
        with self.assertRaises(SystemExit):
            validate_conduit_status(self.write_status(payload))


if __name__ == "__main__":
    unittest.main()
