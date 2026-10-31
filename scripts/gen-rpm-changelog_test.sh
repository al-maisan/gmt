#!/bin/bash
# Tests for the helpers in gen-rpm-changelog.sh: make test-copr
set -uo pipefail

# shellcheck source=gen-rpm-changelog.sh
source "$(dirname "$0")/gen-rpm-changelog.sh"
set +e # gen-rpm-changelog.sh sets -e; the checks below must keep going after a failure

fail=0
check() { # check NAME EXPECTED ACTUAL
    if [ "$2" == "$3" ]; then echo "ok   - $1"; else echo "FAIL - $1: want '$2', got '$3'"; fail=1; fi
}

check "escape: plain text unchanged" "- Fix bug (#1)" "$(rpm_escape <<<'- Fix bug (#1)')"
check "escape: percent doubled" "- 100%% done" "$(rpm_escape <<<'- 100% done')"
check "escape: macros neutralised" "- %%(id) %%{version}" "$(rpm_escape <<<'- %(id) %{version}')"

# With rpm installed, check that an escaped subject renders literally and runs nothing.
if command -v rpmspec >/dev/null; then
    TMP=$(mktemp -d)
    trap 'rm -rf "$TMP"' EXIT
    {
        printf 'Name: x\nVersion: 1\nRelease: 1\nSummary: x\nLicense: MIT\n%%description\nx\n%%changelog\n'
        printf '* Mon Jan 05 2026 A <a@b.c> - 1-1\n- Fix %%(touch %s/pwned) at 100%%\n' "$TMP" | rpm_escape
    } >"$TMP/x.spec"
    check "rpmspec: renders literally" "- Fix %(touch $TMP/pwned) at 100%" "$(rpmspec -P "$TMP/x.spec" 2>/dev/null | tail -1)"
    check "rpmspec: command not run" "no" "$([ -e "$TMP/pwned" ] && echo yes || echo no)"
fi

exit "$fail"
