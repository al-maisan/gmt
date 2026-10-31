#!/bin/bash
# Bump the gmt-mail formula in the Homebrew tap to VERSION and push it.
# Usage: brew-update.sh VERSION   (called by `make brew`)
#
# The formula builds the tagged source archive, so only its url and sha256 change.
# Does nothing if the formula is already at VERSION and refuses an older VERSION.
# BREW_TAP_URL overrides the tap remote (default: SSH, like `make tag`).
set -euo pipefail

TAP_URL=${BREW_TAP_URL:-git@github.com:al-maisan/homebrew-tap.git}
FORMULA=Formula/gmt-mail.rb

# formula_version: print the version in the formula's url (formula on stdin).
formula_version() {
    sed -nE 's#^  url "https://github.com/al-maisan/gmt-mail/archive/refs/tags/v([^"]+)\.tar\.gz"#\1#p'
}

# bump_formula VERSION SHA256: print the formula from stdin with url and sha256 updated.
bump_formula() {
    sed -E -e "s#^(  url \"https://github.com/al-maisan/gmt-mail/archive/refs/tags/v)[^\"]+(\.tar\.gz\")#\1$1\2#" \
        -e "s#^(  sha256 \")[0-9a-f]{64}(\")#\1$2\2#"
}

# Sourced by the tests: stop here.
if [[ "${BASH_SOURCE[0]}" != "$0" ]]; then return 0; fi

VERSION=${1:?usage: $0 VERSION}
URL="https://github.com/al-maisan/gmt-mail/archive/refs/tags/v${VERSION}.tar.gz"
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

git clone -q "$TAP_URL" "$WORK/tap"
cd "$WORK/tap"
CURRENT=$(formula_version <"$FORMULA")
if [ "$CURRENT" == "$VERSION" ]; then
    echo "brew: formula already at ${VERSION}"
    exit 0
fi
# Never downgrade (e.g. `make publish VERSION=<older>` to re-run COPR/PPA).
if [ -n "$CURRENT" ] && [ "$(printf '%s\n' "$CURRENT" "$VERSION" | sort -V | tail -1)" == "$CURRENT" ]; then
    echo "brew: formula is at ${CURRENT}, refusing to go back to ${VERSION}" >&2
    exit 1
fi
SHA=$(curl -fsSL "$URL" | sha256sum | cut -d' ' -f1)
bump_formula "$VERSION" "$SHA" <"$FORMULA" >"$FORMULA.new"
if ! grep -qF "/v${VERSION}.tar.gz\"" "$FORMULA.new" || ! grep -qF "\"$SHA\"" "$FORMULA.new"; then
    echo "brew: could not update url/sha256 in $FORMULA" >&2
    exit 1
fi
mv "$FORMULA.new" "$FORMULA"
git commit -q -am "gmt-mail ${VERSION}"
git push -q origin HEAD
echo "brew: formula bumped to ${VERSION} (sha256 ${SHA})"
