#!/bin/sh

{
set -e

LATEST="v0.2.6"
GOSS_VER=${GOSS_VER:-$LATEST}
GOSS_DST=${GOSS_DST:-/usr/local/bin}
INSTALL_LOC="${GOSS_DST%/}/goss"
touch "$INSTALL_LOC" || { echo "ERROR: Cannot write to $GOSS_DST set GOSS_DST elsewhere or use sudo"; exit 1; }

arch=""
if [ "$(uname -m)" = "x86_64" ]; then
    arch="amd64"
else
    arch="386"
fi

url="https://github.com/aelsabbahy/goss/releases/download/$GOSS_VER/goss-linux-$arch"

echo "Downloading $url"
curl -L "https://github.com/aelsabbahy/goss/releases/download/$GOSS_VER/goss-linux-$arch" -o "$INSTALL_LOC"
chmod +rx "$INSTALL_LOC"
echo "Goss $GOSS_VER has been installed to $INSTALL_LOC"
echo "goss --version"
"$INSTALL_LOC" --version
}
