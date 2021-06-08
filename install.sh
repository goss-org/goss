#!/bin/sh

{
set -e

LATEST_URL="https://github.com/aelsabbahy/goss/releases/latest"
LATEST_EFFECTIVE=$(curl -s -L -o /dev/null ${LATEST_URL} -w '%{url_effective}')
LATEST=${LATEST_EFFECTIVE##*/}

EXTRA_VER=$GOSS_VER

if [ -z "$GOSS_VER" ]; then
    GOSS_VER=${GOSS_VER:-$LATEST}
    EXTRA_VER='master'
fi
if [ -z "$GOSS_VER" ]; then
    echo "ERROR: Could not automatically detect latest version, set GOSS_VER env var and re-run"
    exit 1
fi
GOSS_DST=${GOSS_DST:-/usr/local/bin}
INSTALL_LOC="${GOSS_DST%/}/goss"
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
chmod +rx "$INSTALL_LOC"
echo "Goss $GOSS_VER has been installed to $INSTALL_LOC"
echo "goss --version"
"$INSTALL_LOC" --version

install_extra() {
    name=$1
    extra_url="https://raw.githubusercontent.com/aelsabbahy/goss/$EXTRA_VER/extras/$name/$name"
    dest="${GOSS_DST%/}/${name}"
    echo "Downloading $extra_url"
    curl -L "$extra_url" -o "$dest"
    chmod +rx "$dest"
    echo "$name $EXTRA_VER has been installed to $dest"
}

install_extra "dgoss"
install_extra "dcgoss"
install_extra "kgoss"

}
