#!/usr/bin/env bash
set -euo pipefail

os_name="$(go env GOOS)"

# darwin & windows do not support integration-testing approach via docker.
# platform support is coupled to the travis CI environment, which is stable 'enough'.
if [[ "${os_name}" == "darwin" || "${os_name}" == "windows" ]]; then
  make "test-${os_name}-all"
else
  # linux runs all tests; unit and integration.
  make all
fi
