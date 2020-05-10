#!/usr/bin/env bash
set -euo pipefail

platform_spec="${1:?Must supply name of release binary to build e.g. goss-linux-amd64}"
TRAVIS_TAG="${TRAVIS_TAG:-"local"}"

# Split platform_spec into platform/arch segments
IFS='- ' read -r -a segments <<< "${platform_spec}"

os="${segments[0]}"
arch="${segments[1]}"

output="release/goss-${platform_spec}"
if [[ "${os}" == "windows" ]]; then
  output="${output}.exe"
fi

GOOS="${os}" GOARCH="${arch}" CGO_ENABLED=0 go build \
  -ldflags "-X main.version=${TRAVIS_TAG} -s -w" \
  -o "${output}"

sha256sum "${output}" > "${output}".sha256
