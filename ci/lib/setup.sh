# configure cwd, vars and logging
_setup_env() {
  # -ET: propagate DEBUG/RETURN/ERR traps to functions and subshells
  set -ET
  # exit on unhandled error
  set -o errexit
  # exit on unset variable
  set -o nounset
  # pipefail: any failure in a pipe causes the pipe to fail
  set -o pipefail

  if [[ -n "${SCRIPT_DEBUG:-}" ]]; then
    set -o xtrace
    # http://www.skybert.net/bash/debugging-bash-scripts-on-the-command-line/
    export PS4='# ${BASH_SOURCE}:${LINENO}: ${FUNCNAME[0]:-}() - [${SHLVL},${BASH_SUBSHELL},$?] '
  fi
  trap _err_trap ERR
  # shellcheck disable=SC2034
  # START_DIR is used elsewhere.
  START_DIR="${PWD}"
  export START_DIR
  readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[2]}")" && pwd)"
  readonly TOP_SCRIPT="${SCRIPT_DIR}/$(basename "${BASH_SOURCE[2]}")"
  if [[ -z "${SCRIPT_DIR}" ]]; then
    echo >&2 -e "setup.sh:\tFailed to determine directory containing executed script."
    return 1
  fi
  if ! cd "$(dirname "${BASH_SOURCE[0]}")/../.."; then
    echo >&2 -e "setup.sh:\tFailed to cd to repository root"
    return 1
  fi
  REPO_ROOT="$(pwd)"
  export REPO_ROOT
  if ! source ci/lib/log.sh; then
    echo >&2 -e "setup.sh:\tFailed to source logging library"
    return 1
  fi
}

_err_trap() {
  local err=$?
  local cmd="${BASH_COMMAND:-}"
  # Disable echoing all commands as this makes the traceback really hard to follow
  set +x
  if [[ -n "${SKIP_BASH_STACKTRACE:-}" ]]; then
    log_debug "SKIP_BASH_STACKTRACE was set to something; silencing bash stack-trace."
    exit "${err}"
  fi

  echo >&2 "panic: uncaught error" 1>&2
  print_traceback 1
  echo >&2 "${cmd} exited ${err}" 1>&2
}

_setup_constants() {
  export EXIT_SUCCESS=0
  export EXIT_INVALID_ARGUMENT=66
  export EXIT_FAILED_TO_SOURCE=67
  export EXIT_FAILED_TO_CD=68
  export EXIT_FAILED_AFTER_RETRY=69
  export EXIT_NOT_FOUND=70
}

# Print traceback of call stack, starting from the call location.
# An optional argument can specify how many additional stack frames to skip.
print_traceback() {
  local skip=${1:-0}
  local start=$((skip + 1))
  local end=${#BASH_SOURCE[@]}
  local curr=0
  echo >&2 "Traceback (most recent call first):" 1>&2
  for ((curr = start; curr < end; curr++)); do
    local prev=$((curr - 1))
    local func="${FUNCNAME[$curr]}"
    local file="${BASH_SOURCE[$curr]}"
    local line="${BASH_LINENO[$prev]}"
    echo >&2 "  at ${file}:${line} in ${func}()" 1>&2
  done
}

_setup_env || exit $?
