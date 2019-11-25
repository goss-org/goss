#!/bin/sh

{
set -e

LATEST_URL="https://github.com/aelsabbahy/goss/releases/latest"

if [[ ! -x $(which cut) ]]; then
    echo "cut is not present; backfalling then to v0.3.7"
    LATEST="v0.3.7"
fi

if [[ -x $(which curl) ]]; then 
    LATEST="$(/usr/bin/curl -s -L -o /dev/null ${LATEST_URL} -w '%{url_effective}' | /usr/bin/cut -f 8 -d '/')"
elif [[ -x $(which wget) ]]; then
    LATEST="$(/usr/bin/wget -S  https://github.com/aelsabbahy/goss/releases/latest -O /dev/null 2>&1 | grep "Location" | /usr/bin/cut -f 8 -d '/')"
else 
    echo "neither wget nor curl is installed; backfalling to v0.3.7"
    LATEST="v0.3.7"
fi
    

DGOSS_VER=$GOSS_VER

if [ -z "$GOSS_VER" ]; then
    GOSS_VER=${GOSS_VER:-$LATEST}
    DGOSS_VER='master'
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
chmod +rx "$INSTALL_LOC"
echo "Goss $GOSS_VER has been installed to $INSTALL_LOC"
echo "goss --version"
"$INSTALL_LOC" --version

dgoss_url="https://raw.githubusercontent.com/aelsabbahy/goss/$DGOSS_VER/extras/dgoss/dgoss"
echo "Downloading $dgoss_url"
curl -L "$dgoss_url" -o "$DGOSS_INSTALL_LOC"
chmod +rx "$DGOSS_INSTALL_LOC"
echo "dgoss $DGOSS_VER has been installed to $DGOSS_INSTALL_LOC"
}
