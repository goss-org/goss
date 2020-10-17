#!/usr/bin/env bash
# shellcheck source=../ci/lib/setup.sh
source "$(dirname "${BASH_SOURCE[0]}")/../ci/lib/setup.sh" || exit 67

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
    <(ss -tan | tail -n +2 | awk '{print $4}' | cut -d':' -f2 | sort -u) |
    shuf -n "${how_many}" ||
    shuf -i "${from}-${to}" -n "${how_many}"
}

cleanup() {
  binary_name="$(basename "${GOSS_BINARY}")"
  if [[ "${os}" == "darwin" ]]; then
    killall "${binary_name}"
  elif [[ "${os}" == "linux" ]]; then
    killall "${binary_name}"
  elif [[ "${os}" == "windows" ]]; then
    # Can't use killall, doesn't exist on Windows. Also would interfere with concurrent runs.
    ps -W |
      awk "/${binary_name}/,NF=1" |
      xargs kill
  fi
  exit "${ret:-0}"
}
trap cleanup EXIT

repo_root="$(git rev-parse --show-toplevel)"
export GOSS_BINARY="${repo_root}/release/goss-${platform_spec}"
log_info "Using: '${GOSS_BINARY}', cwd: '$(pwd)'"

export GOSS_USE_ALPHA=1
open_port="$(find_open_port 1025 65335)"
args=(
  "-g=${repo_root}/integration-tests/goss/goss-serve.yaml"
  "serve"
  "--listen-addr=127.0.0.1:${open_port}"
)
log_action -e "\nTesting \`${GOSS_BINARY} ${args[*]}\` ...\n"
"${GOSS_BINARY}" "${args[@]}" &
url="http://127.0.0.1:${open_port}/healthz"

assert_response_contains() {
  local url="${1:?"1st arg: url"}"
  local test_name="${2:?"2nd arg: test name"}"
  local expectation="${3:?"3rd arg: response body match"}"
  local accept_header="${4:-""}"

  curl_args=("--silent")
  [[ -n "${accept_header:-}" ]] && curl_args+=("-H" "Accept: ${accept_header}")
  curl_args+=("${url}")
  log_info "curl ${curl_args[*]}"
  curl="curl"
  [[ "$(go env GOOS)" == "windows" ]] && curl="curl.exe"
  response="$(${curl} "${curl_args[@]}")"
  if grep --quiet "${expectation}" <<<"${response}"; then
    log_success "Passed: ${test_name}"
    return 0
  fi
  log_error "Failed: ${test_name}"
  log_error "  Expected: ${expectation}"
  log_error "  Response: ${response}"
  return 1
}
failure="false"
on_test_failure() {
  failure="true"
}
assert_response_contains "${url}" "no accept header" "Count: 2, Failed: 0, Skipped: 0" "" || on_test_failure
assert_response_contains "${url}" "tap accept header" "Count: 2, Failed: 0, Skipped: 0" "application/vnd.goss-documentation" || on_test_failure
assert_response_contains "${url}" "json accept header" "\"failed-count\":0" "application/json" || on_test_failure

[[ "${failure}" == "true" ]] && log_fatal "Test(s) failed, check output above."
