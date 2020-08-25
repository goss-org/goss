#!/usr/bin/env bash
set -euo pipefail

os_name="$(go env GOOS)"

# darwin & windows do not support integration-testing approach via docker, so on those, just run fast tests.
if [[ "${os_name}" == "darwin" || "${os_name}" == "windows" ]]; then
  make "test-short-all" "release/goss-alpha-${os_name}-amd64"
  make "alpha-test-alpha-${os_name}-amd64"
  make "test-serve-alpha-${os_name}-amd64"
else
  # linux runs all tests; unit and integration.
  make all
  make "test-serve-${os_name}-amd64"
fi
