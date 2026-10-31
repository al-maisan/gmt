#!/bin/bash
# Checks that the man page version/date placeholders get filled in everywhere.
# Usage: man_test.sh RENDERED_MAN GENERATED_SPEC VERSION   (run via: make test-man)
set -uo pipefail

MAN="$1" SPEC="$2" VERSION="$3"
SRC_DIR="$(cd "$(dirname "$0")/.." && pwd)"
fail=0
check() { # check NAME EXPECTED ACTUAL
    if [ "$2" == "$3" ]; then echo "ok   - $1"; else echo "FAIL - $1: want '$2', got '$3'"; fail=1; fi
}

check "template header has both placeholders" 1 "$(grep -c -F '.TH GMT\-MAIL 1 "@DATE@" "gmt-mail @VERSION@"' "$SRC_DIR/gmt-mail.1")"
check "make man: no placeholders left" 0 "$(grep -c -E '@(VERSION|DATE)@' "$MAN")"
check "make man: version in header" 1 "$(grep -c -E "^\.TH GMT\\\\-MAIL 1 \"[0-9]{4}-[0-9]{2}-[0-9]{2}\" \"gmt-mail ${VERSION//./\\.}\"" "$MAN")"

# gen-spec replaces @VERSION@ in the spec; the man page sed there must survive it.
check "spec: man sed survives gen-spec" 1 "$(grep -c -F "s/@VERSION[@]/%{version}/g" "$SPEC")"

# Run the spec's and debian/rules' sed expressions on the template, as the builds do.
RPM_OUT=$(sed -e "s/@VERSION[@]/$VERSION/g" -e "s/@DATE[@]/2026-01-02/g" "$SRC_DIR/gmt-mail.1" | grep '^\.TH')
check "rpm render" ".TH GMT\\-MAIL 1 \"2026-01-02\" \"gmt-mail $VERSION\" \"User Commands\"" "$RPM_OUT"
RULE=$(grep -E "^\s+sed -e 's/@VERSION@/\\\$\(VERSION\)/g' -e 's/@DATE@/\\\$\(MANDATE\)/g' gmt-mail.1" "$SRC_DIR/ppa/debian/rules")
check "debian rules: sed present" 1 "$([ -n "$RULE" ] && echo 1 || echo 0)"
DEB_OUT=$(sed -e "s/@VERSION@/$VERSION/g" -e "s/@DATE@/2026-01-02/g" "$SRC_DIR/gmt-mail.1" | grep '^\.TH')
check "debian render" "$RPM_OUT" "$DEB_OUT"

exit "$fail"
