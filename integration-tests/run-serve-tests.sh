#!/usr/bin/env bash
set -euo pipefail

platform_spec="${1:?Must supply name of release binary to build e.g. goss-linux-amd64}"

# Split platform_spec into platform/arch segments
IFS='- ' read -r -a segments <<< "${platform_spec}"

os="${segments[0]}"
arch="${segments[1]}"
if [[ "${segments[0]}" == "alpha" ]]; then
  os="${segments[1]}"
  arch="${segments[2]}"
fi

find_open_port() {
  local from="${1:?"Supply start of port range"}"
  local to="${2:?"Supply end of port range"}"
  local how_many="${3:-"1"}"

  # Thanks to https://unix.stackexchange.com/questions/55913/whats-the-easiest-way-to-find-an-unused-local-port
  # ss doesn't exist on Windows, so fall back on just choosing a random number inside the range (since netstat is _slow_).
  comm -23 \
    <(seq "${from}" "${to}" | sort) \
    <(ss -Htan | awk '{print $4}' | cut -d':' -f2 | sort -u) |
    shuf -n "${how_many}" ||
    shuf -i "${from}-${to}" -n "${how_many}"
}

cleanup() {
  # Can't use killall, doesn't exist on Windows. Also would interfere with concurrent runs.
  binary_name="$(basename "${GOSS_BINARY}")"
  ps -W |
    awk "/${binary_name}/,NF=1" |
    xargs kill
}
trap cleanup EXIT

repo_root="$(git rev-parse --show-toplevel)"
export GOSS_BINARY="${repo_root}/release/goss-${platform_spec}"
echo "Using: '${GOSS_BINARY}', cwd: '$(pwd)'"

export GOSS_USE_ALPHA=1
open_port="$(find_open_port 1025 65335)"
args=(
  "-g=${repo_root}/integration-tests/goss/goss-serve.yaml"
  "serve"
  "--listen-addr=127.0.0.1:${open_port}"
)
echo -e "\nTesting \`${GOSS_BINARY} ${args[*]}\` ...\n"

"${GOSS_BINARY}" "${args[@]}" &
if curl --silent "http://127.0.0.1:${open_port}/healthz" | grep 'Count: 2, Failed: 0, Skipped: 0' ; then
  echo "passed"
else
  echo "failed, exit code $?"
fi
