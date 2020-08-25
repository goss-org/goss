#!/usr/bin/env bash
set -euo pipefail

platform_spec="${1:?"Must supply name of release binary to build e.g. goss-linux-amd64"}"
# Split platform_spec into platform/arch segments
IFS='- ' read -r -a segments <<< "${platform_spec}"

os="${segments[0]}"
arch="${segments[1]}"
if [[ "${segments[0]}" == "alpha" ]]; then
  os="${segments[1]}"
  arch="${segments[2]}"
fi

repo_root="$(git rev-parse --show-toplevel)"
export GOSS_BINARY="${repo_root}/release/goss-${platform_spec}"
echo "Using: '${GOSS_BINARY}', cwd: '$(pwd)', os: ${os}"
readarray -t goss_test_files < <(find integration-tests -type f -name "*.goss.yaml" | grep "${os}" | sort | uniq)

export GOSS_USE_ALPHA=1
for file in "${goss_test_files[@]}"; do
  args=(
    "-g=${file}"
    "validate"
  )
  echo -e "\nTesting \`${GOSS_BINARY} ${args[*]}\` ...\n"
  "${GOSS_BINARY}" "${args[@]}"
done
