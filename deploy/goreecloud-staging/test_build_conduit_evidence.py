#!/usr/bin/env python3

from build_conduit_evidence import CHECK_NAMES, build


def expect_rejected(checks: dict[str, bool], message: str) -> None:
    try:
        build("conduit-control-status", "a" * 40, "sha256:" + "b" * 64, checks)
    except SystemExit:
        return
    raise AssertionError(message)


def main() -> None:
    checks = {name: True for name in CHECK_NAMES}
    payload = build(
        "conduit-control-status",
        "a" * 40,
        "sha256:" + "b" * 64,
        checks,
    )
    assert payload["authority"] == "inherited"
    assert payload["compatibility_bridge_active"] is True
    assert payload["production_cutover_authorized"] is False
    assert set(payload["checks"]) == set(CHECK_NAMES)

    for gate in ("state_migration", "backup_restore", "rollback", "client_networking", "isolated_runtime"):
        failed = dict(checks)
        failed[gate] = False
        expect_rejected(failed, f"incomplete {gate} evidence unexpectedly accepted")

    print("Conduit isolated evidence generator tests: PASS")


if __name__ == "__main__":
    main()
