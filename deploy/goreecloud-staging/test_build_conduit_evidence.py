#!/usr/bin/env python3

from build_conduit_evidence import CHECK_NAMES, build


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

    failed = dict(checks)
    failed["isolated_runtime"] = False
    try:
        build("conduit-control-status", "a" * 40, "sha256:" + "b" * 64, failed)
    except SystemExit:
        pass
    else:
        raise AssertionError("incomplete isolated evidence unexpectedly accepted")

    print("Conduit isolated evidence generator tests: PASS")


if __name__ == "__main__":
    main()
