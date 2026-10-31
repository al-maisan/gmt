#!/bin/bash
# Tests for the upload-detection helpers in build-ppa.sh: make test-ppa
# No network, no Launchpad: feeds a getPublishedSources-shaped JSON on stdin.
set -uo pipefail

# shellcheck source=build-ppa.sh
source "$(dirname "$0")/build-ppa.sh"
set +e # build-ppa.sh sets -e; the checks below must keep going after a failure

JSON='{"entries": [
 {"source_package_name": "gmt-mail", "source_package_version": "0.8.0-1ppa1~resolute1"},
 {"source_package_name": "gmt-mail", "source_package_version": "0.6.2-1ppa1~questing1"},
 {"source_package_name": "gmt-mail", "source_package_version": "0.8.0-1ppa3~questing1"},
 {"source_package_name": "gmt-mail", "source_package_version": "0.8.01-1ppa9~resolute1"},
 {"source_package_name": "gmt-mail", "source_package_version": "0.2.4-1ppa1~noble1"}
]}'
fail=0
check() { # check NAME EXPECTED ACTUAL
    if [ "$2" == "$3" ]; then echo "ok   - $1"; else echo "FAIL - $1: want '$2', got '$3'"; fail=1; fi
}

VERSION=0.8.0
UPLOADED=$(uploaded_versions "$VERSION" <<<"$JSON")
check "only this version's uploads" "0.8.0-1ppa1~resolute1 0.8.0-1ppa3~questing1" "$(paste -sd' ' <<<"$UPLOADED")"
check "already uploaded for resolute" yes "$(uploaded_for resolute && echo yes || echo no)"
check "already uploaded for questing" yes "$(uploaded_for questing && echo yes || echo no)"
check "not uploaded for noble" no "$(uploaded_for noble && echo yes || echo no)"
check "next revision above the highest" 4 "$(next_rev)"

VERSION=0.9.0
UPLOADED=$(uploaded_versions "$VERSION" <<<"$JSON")
check "new version: nothing uploaded" "" "$UPLOADED"
check "new version: not uploaded for resolute" no "$(uploaded_for resolute && echo yes || echo no)"
check "new version: revision 1" 1 "$(next_rev)"

VERSION=0.8.0
UPLOADED=$(uploaded_versions "$VERSION" <<<'{"entries": []}')
check "empty PPA" "" "$UPLOADED"
check "dots are literal (0.8.0 does not match 0x8y0)" no "$(UPLOADED='0x8y0-1ppa1~resolute1' uploaded_for resolute && echo yes || echo no)"

exit "$fail"
