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

# PPA revision — increment this when rebuilding the same upstream version
PPA_REV="${1:-1}"

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
    PREV_TAG=$(git -C "$SRC_DIR" describe --tags --abbrev=0 "${TAG}^" 2>/dev/null || echo "")
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
    debuild -S -sa -d ${SIGN_FLAG}

    echo ""
    echo "=== Uploading to PPA for ${RELEASE} ==="
    cd "$BUILD_DIR"
    dput "$PPA" "gmt-mail_${FULL_VER}_source.changes"
done

echo ""
echo "=== Done ==="
echo "Check build status at: https://launchpad.net/~al-maisan/+archive/ubuntu/gmt-mail"
