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

# Ubuntu releases to target (add/remove as needed)
RELEASES=("questing")  # 25.10

# PPA revision — increment this when rebuilding the same upstream version
PPA_REV="${1:-1}"

echo "=== Building gmt-mail ${VERSION} (ppa${PPA_REV}) for: ${RELEASES[*]} ==="
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/gmt-mail-${VERSION}"

# Copy source + vendor into build dir
cd "$SRC_DIR"
tar cf - --exclude='.git' --exclude='ppa' --exclude='ai' --exclude='.claude' \
         --exclude='gmt' --exclude='gmt-mail' --exclude='gmt-mail.spec' . \
  | tar xf - -C "$BUILD_DIR/gmt-mail-${VERSION}"

# Copy PPA-specific debian dir (overwrite the upstream debian/)
rm -rf "$BUILD_DIR/gmt-mail-${VERSION}/debian"
cp -a "$SCRIPT_DIR/debian" "$BUILD_DIR/gmt-mail-${VERSION}/debian"

# Create orig tarball
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

    # Update changelog for this release
    sed -i "1s/.*$/gmt-mail (${FULL_VER}) ${RELEASE}; urgency=medium/" debian/changelog

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
