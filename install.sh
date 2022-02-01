#!/usr/bin/env sh

{
  set -e

  if command -v curl >/dev/null; then
    HTTP_GET="curl -sSfL"
  elif command -v wget >/dev/null; then
    HTTP_GET="wget -qO-"
  else
    echo "[ERROR] This script needs wget or curl to be installed."
    exit 1
  fi

  # shellcheck disable=SC2153
  DGOSS_VER="${GOSS_VER}"

  GOSS_DST=${GOSS_DST:-/usr/local/bin}
  INSTALL_LOC="${GOSS_DST%/}/goss"
  DGOSS_INSTALL_LOC="${GOSS_DST%/}/dgoss"

  touch "$INSTALL_LOC" || {
    echo "ERROR: Cannot write to $GOSS_DST set GOSS_DST elsewhere or use sudo"
    exit 1
  }

  case "$(uname -m)" in
  x86_64)
    arch="amd64"
    ;;
  aarch64)
    arch="arm"
    ;;
  *)
    arch="386"
    ;;
  esac

  if [ -n "$GOSS_VER" ]; then
    GOSS_URL="https://github.com/aelsabbahy/goss/releases/download/$GOSS_VER/goss-linux-$arch"
    DGOSS_URL="https://raw.githubusercontent.com/aelsabbahy/goss/$DGOSS_VER/extras/dgoss/dgoss"
  else
    GOSS_URL="https://github.com/aelsabbahy/goss/releases/latest/download/goss-linux-$arch"
    DGOSS_URL="https://raw.githubusercontent.com/aelsabbahy/goss/master/extras/dgoss/dgoss"
  fi

  echo "Downloading ${GOSS_URL}"
  if ! $HTTP_GET "$GOSS_URL" > "$INSTALL_LOC"; then
    echo "ERROR: Cannot download goss from $GOSS_URL"
    exit 1
  fi

  chmod +rx "$INSTALL_LOC"

  echo "Goss has been installed to $INSTALL_LOC"
  echo "goss --version"
  "$INSTALL_LOC" --version

  echo "Downloading ${DGOSS_URL}"
  if ! $HTTP_GET "$DGOSS_URL" > "$DGOSS_INSTALL_LOC"; then
    echo "ERROR: Cannot download goss from $GOSS_URL"
    exit 1
  fi

  chmod +rx "$DGOSS_INSTALL_LOC"
  echo "dgoss has been installed to $DGOSS_INSTALL_LOC"
}
