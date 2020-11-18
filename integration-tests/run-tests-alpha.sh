#!/usr/bin/env bash
# shellcheck source=../ci/lib/setup.sh
source "$(dirname "${BASH_SOURCE[0]}")/../ci/lib/setup.sh" || exit 67

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
log_info "Using: '${GOSS_BINARY}', cwd: '$(pwd)', os: ${os}"
readarray -t goss_test_files < <(find integration-tests -type f -name "*.goss.yaml" | grep "${os}" | sort | uniq)

log_action "powershell reconnaisance inside CI:"
powershell.exe -noprofile -noninteractive -command "(get-itemproperty -path 'HKLM:/SYSTEM/CurrentControlSet/Control/Lsa/').restrictanonymous"

export GOSS_USE_ALPHA=1
for file in "${goss_test_files[@]}"; do
  args=(
    "-g=${file}"
    "validate"
  )
  log_action -e "\nTesting \`${GOSS_BINARY} ${args[*]}\` ...\n"
  "${GOSS_BINARY}" "${args[@]}"
done
