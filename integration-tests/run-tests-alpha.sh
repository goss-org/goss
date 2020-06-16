#!/usr/bin/env bash
set -euo pipefail

os_name="${1:?Enter the OS name you want to run (windows|darwin)}"

repo_root="$(git rev-parse --show-toplevel)"
export GOSS_BINARY="${repo_root}/release/goss-alpha-${os_name}-amd64"
echo "Using: '${GOSS_BINARY}', cwd: '$(pwd)'"
readarray -t goss_test_files < <(find integration-tests -type f -name "*.goss.yaml" | grep "${os_name}" | sort | uniq)

export GOSS_USE_ALPHA=1
for file in "${goss_test_files[@]}"; do
  args=(
    "-g=${file}"
    "validate"
  )
  echo -e "\nTesting \`${GOSS_BINARY} ${args[*]}\` ...\n"
  "${GOSS_BINARY}" "${args[@]}"
done
