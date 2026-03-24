#!/bin/bash
set -e

# Configuration
PPA="ppa:al-maisan/gmt-mail"
GPG_KEY="753B6ECF2B458FF3D19D568C1E0A288397AE739E"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SRC_DIR="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="/tmp/gmt-ppa-build"

# Read version from Makefile
VERSION=$(grep '^VERSION' "$SRC_DIR/Makefile" | head -1 | awk -F':= ' '{print $2}' | tr -d ' ')
TAG="v${VERSION}"

# Ubuntu releases to target (add/remove as needed)
RELEASES=("questing")  # 25.10

# Determine PPA revision — query Launchpad for existing uploads of this version
# and auto-increment, or use explicit argument if provided.
if [ -n "${1:-}" ]; then
    PPA_REV="$1"
else
    LP_USER="al-maisan"
    LP_PPA="gmt-mail"
    # Find the highest ppa revision already uploaded for this version
    EXISTING=$(curl -s "https://api.launchpad.net/1.0/~${LP_USER}/+archive/ubuntu/${LP_PPA}?ws.op=getPublishedSources&source_name=gmt-mail&exact_match=true" \
        | grep -oP "\"source_package_version\": \"${VERSION}-1ppa\K[0-9]+" \
        | sort -n | tail -1)
    if [ -n "$EXISTING" ]; then
        PPA_REV=$((EXISTING + 1))
        echo "Found existing ppa${EXISTING} for ${VERSION}, using ppa${PPA_REV}"
    else
        PPA_REV=1
    fi
fi

# Verify the git tag exists
cd "$SRC_DIR"
if ! git rev-parse "$TAG" >/dev/null 2>&1; then
    echo "Error: git tag '$TAG' not found. Create it first:"
    echo "  git tag -s $TAG -m '$TAG'"
    echo "  git push origin $TAG"
    exit 1
fi

echo "=== Building gmt-mail ${VERSION} (ppa${PPA_REV}) from tag ${TAG} for: ${RELEASES[*]} ==="
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Generate orig tarball from the git tag — this is reproducible and immutable
git archive --format=tar.gz --prefix="gmt-mail-${VERSION}/" "$TAG" \
    -o "$BUILD_DIR/gmt-mail_${VERSION}.orig.tar.gz"

# Extract it so we can add vendor and debian
cd "$BUILD_DIR"
tar xzf "gmt-mail_${VERSION}.orig.tar.gz"

# Add vendored dependencies (not in git, must be generated)
cd "gmt-mail-${VERSION}"
go mod vendor 2>/dev/null || GOTOOLCHAIN=local go mod vendor

# Recreate orig tarball with vendor included
cd "$BUILD_DIR"
tar czf "gmt-mail_${VERSION}.orig.tar.gz" "gmt-mail-${VERSION}"

SIGN_FLAG=""
if [ -n "$GPG_KEY" ]; then
    SIGN_FLAG="-k${GPG_KEY}"
fi

for RELEASE in "${RELEASES[@]}"; do
    FULL_VER="${VERSION}-1ppa${PPA_REV}~${RELEASE}1"
    echo ""
    echo "=== Building source package: ${FULL_VER} ==="
    cd "$BUILD_DIR"
    rm -rf "gmt-mail-${VERSION}-${RELEASE}"
    cp -a "gmt-mail-${VERSION}" "gmt-mail-${VERSION}-${RELEASE}"
    cd "gmt-mail-${VERSION}-${RELEASE}"

    # Add PPA-specific debian dir
    cp -a "$SCRIPT_DIR/debian" debian

    # Generate changelog from git log between previous tag and current tag
    # Find previous released tag (not just any tag)
    PREV_TAG=$(gh release list --limit 100 --json tagName --jq '.[].tagName' 2>/dev/null \
        | grep -v "^${TAG}$" | head -1 || echo "")
    TIMESTAMP=$(date -R)
    {
        echo "gmt-mail (${FULL_VER}) ${RELEASE}; urgency=medium"
        echo ""
        if [ -n "$PREV_TAG" ]; then
            git -C "$SRC_DIR" log --format="  * %s" "${PREV_TAG}..${TAG}" --no-merges
        else
            git -C "$SRC_DIR" log --format="  * %s" "${TAG}" --no-merges --max-count=20
        fi
        echo ""
        echo " -- Muharem Hrnjadovic <muharem@linux.com>  ${TIMESTAMP}"
    } > debian/changelog

    # Build signed source package
    # Use -sd (debian-only) for rebuilds to avoid orig tarball mismatch
    if [ "$PPA_REV" -gt 1 ]; then
        SA_FLAG="-sd"
    else
        SA_FLAG="-sa"
    fi
    debuild -S $SA_FLAG -d ${SIGN_FLAG}

    echo ""
    echo "=== Uploading to PPA for ${RELEASE} ==="
    cd "$BUILD_DIR"
    dput "$PPA" "gmt-mail_${FULL_VER}_source.changes"
done

echo ""
echo "=== Done ==="
echo "Check build status at: https://launchpad.net/~al-maisan/+archive/ubuntu/gmt-mail"
