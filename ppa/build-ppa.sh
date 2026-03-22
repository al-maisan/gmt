#!/bin/bash
set -e

# Configuration — adjust these
PPA="ppa:al-maisan/gmt-mail"
GPG_KEY="753B6ECF2B458FF3D19D568C1E0A288397AE739E"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SRC_DIR="$(dirname "$SCRIPT_DIR")"
VERSION="0.2.3"
BUILD_DIR="/tmp/gmt-ppa-build"

# Ubuntu releases to target (add/remove as needed)
RELEASES=("noble")  # 24.04 LTS

echo "=== Preparing source tree ==="
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
    echo ""
    echo "=== Building source package for ${RELEASE} ==="
    cd "$BUILD_DIR"
    rm -rf "gmt-mail-${VERSION}-${RELEASE}"
    cp -a "gmt-mail-${VERSION}" "gmt-mail-${VERSION}-${RELEASE}"
    cd "gmt-mail-${VERSION}-${RELEASE}"

    # Update changelog for this release
    sed -i "1s/.*$/gmt-mail (${VERSION}-1ppa1~${RELEASE}1) ${RELEASE}; urgency=medium/" debian/changelog

    # Build signed source package
    debuild -S -sa ${SIGN_FLAG}

    echo ""
    echo "=== Uploading to PPA for ${RELEASE} ==="
    cd "$BUILD_DIR"
    dput "$PPA" "gmt-mail_${VERSION}-1ppa1~${RELEASE}1_source.changes"
done

echo ""
echo "=== Done ==="
echo "Check build status at: https://launchpad.net/~al-maisan/+archive/ubuntu/gmt-mail"
