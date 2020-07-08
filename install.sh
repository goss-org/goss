#!/bin/sh

{
set -e

LATEST_URL="https://github.com/aelsabbahy/goss/releases/latest"
LATEST_EFFECTIVE=$(curl -s -L -o /dev/null ${LATEST_URL} -w '%{url_effective}')
LATEST=${LATEST_EFFECTIVE##*/}


if [ -z "$GOSS_VER" ]; then
    GOSS_VER=${GOSS_VER:-$LATEST}
fi
if [ -z "$GOSS_VER" ]; then
    echo "ERROR: Could not automatically detect latest version, set GOSS_VER env var and re-run"
    exit 1
fi

GOSS_DST=${GOSS_DST:-/usr/local/bin}
INSTALL_LOC="${GOSS_DST%/}/goss"
DGOSS_INSTALL_LOC="${GOSS_DST%/}/dgoss"
touch "$INSTALL_LOC" || { echo "ERROR: Cannot write to $GOSS_DST set GOSS_DST elsewhere or use sudo"; exit 1; }

arch=""
if [ "$(uname -m)" = "x86_64" ]; then
    arch="amd64"
elif [ "$(uname -m)" = "aarch64" ]; then
    arch="arm"
else
    arch="386"
fi

url="https://github.com/aelsabbahy/goss/releases/download/$GOSS_VER/goss-linux-$arch"

echo "Downloading $url"
curl -L "$url" -o "$INSTALL_LOC"
curl -sSL "${url}.sha256" | sed "s#goss-linux-${arch}#${INSTALL_LOC}#" | sha256sum -c -
chmod +rx "$INSTALL_LOC"
echo "Goss $GOSS_VER has been installed to $INSTALL_LOC"
echo "goss --version"
"$INSTALL_LOC" --version

dgoss_url="https://github.com/aelsabbahy/goss/releases/download/$GOSS_VER/dgoss"
echo "Downloading $dgoss_url"
curl -L "$dgoss_url" -o "$DGOSS_INSTALL_LOC"
curl -sSL "${dgoss_url}.sha256" | sed "s#dgoss#${DGOSS_INSTALL_LOC}#" | sha256sum -c -
chmod +rx "$DGOSS_INSTALL_LOC"
echo "dgoss $GOSS_VER has been installed to $DGOSS_INSTALL_LOC"
}
