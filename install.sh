#!/bin/sh

{
set -e

LATEST="v0.2.2"
GOSS_VER=${GOSS_VER:-$LATEST}

arch=""
if [ "$(uname -m)" = "x86_64" ]; then
    arch="amd64"
else
    arch="386"
fi

url="https://github.com/aelsabbahy/goss/releases/download/$GOSS_VER/goss-linux-$arch"

echo "Downloading $url"
curl -sSL "https://github.com/aelsabbahy/goss/releases/download/$GOSS_VER/goss-linux-$arch" -o /usr/local/bin/goss
chmod +rx /usr/local/bin/goss
echo "Goss $GOSS_VER has been installed at /usr/local/bin/goss"
echo "goss --version"
/usr/local/bin/goss --version
}
