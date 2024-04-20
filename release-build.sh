#!/usr/bin/env bash
set -euo pipefail

platform_spec="${1:?"Must supply name of release binary to build e.g. goss-linux-amd64"}"
version_stamp="${TRAVIS_TAG:-"0.0.0"}"

# Split platform_spec into platform/arch segments
IFS='- ' read -r -a segments <<< "${platform_spec}"

os="${segments[0]}"
arch="${segments[1]}"
if [[ "${segments[0]}" == "alpha" ]]; then
  os="${segments[1]}"
  arch="${segments[2]}"
fi

output_dir="release/"
output_fname="goss-${platform_spec}"
if [[ "${os}" == "windows" ]]; then
  output_fname="${output_fname}.exe"
fi
output="${output_dir}/${output_fname}"

GOOS="${os}" GOARCH="${arch}" CGO_ENABLED=0 go build \
  -ldflags "-X main.version=${version_stamp} -X github.com/goss-org/goss/util.Version=${version_stamp} -s -w" \
  -o "${output}" \
  github.com/goss-org/goss/cmd/goss

chmod +x "${output}"

function __sha256sum {
  if [[ "$OSTYPE" == "darwin"* ]]; then
    shasum -a 256 "$1"
  else
    sha256sum "$1"
  fi
}

(cd "$output_dir" && __sha256sum "${output_fname}" > "${output_fname}.sha256")
