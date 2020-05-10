#!/usr/bin/env bash
set -euo pipefail

os_name="${TRAVIS_OS_NAME:?"No value from TRAVIS_OS_NAME. This is meant to be run in Travis CI, see also https://docs.travis-ci.com/user/environment-variables/#convenience-variables"}"

# gofmt must be on PATH
command -v gofmt

if [[ "${os_name}" == "windows" ]]; then
  echo "Skipping go-fmt on Windows because line-endings cause every file to need formatting."
  echo "Linux is treated as authoritative."
  echo "Exiting 0..."
  exit 0
fi

fmt="$(go fmt github.com/aelsabbahy/goss/...)"

if [[ -z "${fmt}" ]]; then
  echo "valid gofmt"
else
  echo "invalid gofmt:"
  echo "${fmt}"
  exit 1
fi
