#!/bin/bash
# Tests for the helpers in copr-build.sh: make test-copr
# No network: feeds releases.json-, Bodhi- and COPR-shaped JSON on stdin.
set -uo pipefail

# shellcheck source=copr-build.sh
source "$(dirname "$0")/copr-build.sh"
set +e # copr-build.sh sets -e; the checks below must keep going after a failure

fail=0
check() { # check NAME EXPECTED ACTUAL
    if [ "$2" == "$3" ]; then echo "ok   - $1"; else echo "FAIL - $1: want '$2', got '$3'"; fail=1; fi
}
flat() { paste -sd' '; }

# Mirrors 2026-10-31: 42 released but no longer maintained, 45 maintained but beta only.
RELEASES_JSON='[{"version": "45 Beta", "variant": "Workstation"}, {"version": "44", "variant": "Workstation"},
 {"version": "44", "variant": "Server"}, {"version": "43", "variant": "Workstation"}, {"version": "42", "variant": "KDE"}]'
BODHI='{"releases": [
 {"name": "EPEL-9", "id_prefix": "FEDORA-EPEL", "version": "9", "state": "current"},
 {"name": "F43", "id_prefix": "FEDORA", "version": "43", "state": "current"},
 {"name": "F43C", "id_prefix": "FEDORA-CONTAINER", "version": "43", "state": "current"},
 {"name": "F44", "id_prefix": "FEDORA", "version": "44", "state": "current"},
 {"name": "F44F", "id_prefix": "FEDORA-FLATPAK", "version": "44", "state": "current"},
 {"name": "F45", "id_prefix": "FEDORA", "version": "45", "state": "current"},
 {"name": "ELN", "id_prefix": "FEDORA", "version": "eln", "state": "pending"}]}'
BUILDS='{"items": [
 {"id": 6, "state": "failed",    "chroots": ["fedora-44-x86_64"], "source_package": {"version": "0.8.0-1"}},
 {"id": 5, "state": "succeeded", "chroots": ["fedora-43-x86_64"], "source_package": {"version": "0.8.0-1"}},
 {"id": 4, "state": "running",   "chroots": ["fedora-44-x86_64"], "source_package": {"version": "0.8.1-1"}},
 {"id": 3, "state": "succeeded", "chroots": ["fedora-44-x86_64"], "source_package": {"version": "0.8.0-10"}},
 {"id": 2, "state": "importing", "chroots": [], "source_package": null}]}'

RELEASED=$(released_versions <<<"$RELEASES_JSON")
check "released: final images only, no beta" "42 43 44" "$(flat <<<"$RELEASED")"
TARGETS=$(supported_chroots "$RELEASED" <<<"$BODHI")
check "supported: released and maintained" "fedora-43-x86_64 fedora-44-x86_64" "$(flat <<<"$TARGETS")"
check "supported: picks up 45 once released" "fedora-43-x86_64 fedora-44-x86_64 fedora-45-x86_64" \
    "$(supported_chroots "$(printf '43\n44\n45')" <<<"$BODHI" | flat)"

check "built: succeeded counts, failed does not" "fedora-43-x86_64" "$(built_chroots 0.8.0-1 <<<"$BUILDS" | flat)"
check "built: running counts" "fedora-44-x86_64" "$(built_chroots 0.8.1-1 <<<"$BUILDS" | flat)"
check "built: 0.8.0-1 does not match 0.8.0-10" "" "$(built_chroots 0.8.0-1 <<<'{"items":[{"state":"succeeded","chroots":["fedora-44-x86_64"],"source_package":{"version":"0.8.0-10"}}]}')"
check "built: unknown version" "" "$(built_chroots 1.0.0-1 <<<"$BUILDS")"

check "todo: only the missing release" "fedora-44-x86_64" "$(missing "$TARGETS" "$(built_chroots 0.8.0-1 <<<"$BUILDS")" | flat)"
check "todo: nothing when all are built" "" "$(missing "$TARGETS" "$(printf 'fedora-43-x86_64\nfedora-44-x86_64\nfedora-42-x86_64')")"
check "todo: everything when nothing is built" "fedora-43-x86_64 fedora-44-x86_64" "$(missing "$TARGETS" "" | flat)"

check "project chroots" "fedora-43-x86_64" "$(project_chroots <<<'{"chroot_repos": {"fedora-43-x86_64": "https://x"}}')"
check "release from spec template" 1 "$(spec_release)"

exit "$fail"
