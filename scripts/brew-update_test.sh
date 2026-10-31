#!/bin/bash
# Tests for brew-update.sh: make test-brew
# No network: the end-to-end checks use a local tap and a formula already at the version.
set -uo pipefail

# shellcheck source=brew-update.sh
source "$(dirname "$0")/brew-update.sh"
set +e # brew-update.sh sets -e; the checks below must keep going after a failure

fail=0
check() { # check NAME EXPECTED ACTUAL
    if [ "$2" == "$3" ]; then echo "ok   - $1"; else echo "FAIL - $1: want '$2', got '$3'"; fail=1; fi
}

OLD_SHA=54e67abf51c7ac3251a50a9077e7e07bc723df39d44e3f0be8350b9be3c22a78
NEW_SHA=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
FORMULA='class GmtMail < Formula
  desc "x"
  url "https://github.com/al-maisan/gmt-mail/archive/refs/tags/v0.9.0.tar.gz"
  sha256 "'$OLD_SHA'"
  head "https://github.com/al-maisan/gmt-mail.git", branch: "master"
end'

check "formula version" "0.9.0" "$(formula_version <<<"$FORMULA")"
OUT=$(bump_formula 1.10.0 "$NEW_SHA" <<<"$FORMULA")
check "url bumped" '  url "https://github.com/al-maisan/gmt-mail/archive/refs/tags/v1.10.0.tar.gz"' "$(grep '^  url' <<<"$OUT")"
check "sha256 bumped" "  sha256 \"$NEW_SHA\"" "$(grep '^  sha256' <<<"$OUT")"
check "head untouched" '  head "https://github.com/al-maisan/gmt-mail.git", branch: "master"' "$(grep '^  head' <<<"$OUT")"
check "only url and sha256 changed" "2" "$(diff <(echo "$FORMULA") <(echo "$OUT") | grep -c '^>')"

# End to end against a local tap whose formula is already at the version: no network, no commit.
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
git init -q -b main "$TMP/src" && mkdir "$TMP/src/Formula" && echo "$FORMULA" >"$TMP/src/Formula/gmt-mail.rb"
git -C "$TMP/src" add -A && git -C "$TMP/src" -c commit.gpgsign=false -c user.name=t -c user.email=t@t commit -qm init
git clone -q --bare "$TMP/src" "$TMP/tap.git"
check "already at version" "brew: formula already at 0.9.0" \
    "$(BREW_TAP_URL="$TMP/tap.git" "$(dirname "$0")/brew-update.sh" 0.9.0 2>&1)"
check "refuses a downgrade" "brew: formula is at 0.9.0, refusing to go back to 0.8.10" \
    "$(BREW_TAP_URL="$TMP/tap.git" "$(dirname "$0")/brew-update.sh" 0.8.10 2>&1)"
check "tap untouched" "1" "$(git -C "$TMP/tap.git" rev-list --count main)"
check "usage error" "1" "$("$(dirname "$0")/brew-update.sh" >/dev/null 2>&1; echo $?)"

exit "$fail"
