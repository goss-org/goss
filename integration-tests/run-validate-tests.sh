#!/usr/bin/env bash
# shellcheck source=../ci/lib/setup.sh
source "$(dirname "${BASH_SOURCE[0]}")/../ci/lib/setup.sh" || exit 67

platform_spec="${1:?"Must supply name of release binary to build e.g. goss-linux-amd64"}"
# Split platform_spec into platform/arch segments
IFS='- ' read -r -a segments <<< "${platform_spec}"

os="${segments[0]}"
arch="${segments[1]}"

if [[ "${os}" == "linux" ]]; then
  echo "OS is ${os}. This script is not for running tests on the different flavours of linux."
  echo "Linux is exercised via the integration-tests/test.sh currently, because linux can be"
  echo "verified via docker containers; macOS and Windows cannot."
  echo "This script is for macOS and Windows, and runs tests that are expected to pass on"
  echo "Travis-CI provided images, running nakedly (no containerisation) on the hosts there."
  exit 1
fi

repo_root="$(git rev-parse --show-toplevel)"
export GOSS_BINARY="${repo_root}/release/goss-${platform_spec}"
log_info "Using: '${GOSS_BINARY}', cwd: '$(pwd)', os: ${os}"

export GOSS_USE_ALPHA=1
for file in `find integration-tests -type f -name "*.goss.yaml" | grep "${os}" | sort | uniq`; do
  args=(
    "-g=${file}"
    "validate"
  )
  log_action "\nTesting \`${GOSS_BINARY} ${args[*]}\` ...\n"
  "${GOSS_BINARY}" "${args[@]}"
done
