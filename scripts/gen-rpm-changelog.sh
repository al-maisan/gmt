#!/bin/bash
# Generate RPM %changelog from GitHub releases.
# Each release produces one changelog entry with commits since the previous release.
set -e

MAINTAINER="Muharem Hrnjadovic <muharem@linux.com>"

# List released tags newest first (only tags with actual GitHub releases)
TAGS=($(gh release list --json tagName --jq '.[].tagName' 2>/dev/null))

if [ ${#TAGS[@]} -eq 0 ]; then
    # Fallback to git tags if gh is unavailable
    TAGS=($(git tag -l 'v*' --sort=-version:refname))
fi

for i in "${!TAGS[@]}"; do
    TAG="${TAGS[$i]}"
    VER="${TAG#v}"

    # RPM changelog date format: "Day Mon DD YYYY"
    DATE=$(git log -1 --format='%ad' --date=format:'%a %b %d %Y' "$TAG")

    echo "* ${DATE} ${MAINTAINER} - ${VER}-1"

    # Get commits between this release and the next older one
    if [ $((i + 1)) -lt ${#TAGS[@]} ]; then
        PREV="${TAGS[$((i + 1))]}"
        git log --format="- %s" "${PREV}..${TAG}" --no-merges
    else
        git log --format="- %s" "${TAG}" --no-merges --max-count=10
    fi
    echo ""
done
