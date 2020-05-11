#!/usr/bin/env bash

# comment out the unused-ones so far until they're needed. Otherwise it's a google search to find them again.
NOCOLOUR='\033[0m'
RED='\033[0;31m'
GREEN='\033[0;32m'
ORANGE='\033[0;33m'
# BLUE='\033[0;34m'
# PURPLE='\033[0;35m'
CYAN='\033[0;36m'
# LIGHTGRAY='\033[0;37m'
# DARKGRAY='\033[1;30m'
LIGHTRED='\033[1;31m'
LIGHTGREEN='\033[1;32m'
# YELLOW='\033[1;33m'
# LIGHTBLUE='\033[1;34m'
# LIGHTPURPLE='\033[1;35m'
LIGHTCYAN='\033[1;36m'
# WHITE='\033[1;37m'

is_ci() {
  if [[ "${CI:-}" == "true" ]]; then
    echo "true"
  else
    echo "false"
  fi
}

log_action() {
  echo -e "${LIGHTGREEN}${*}${NOCOLOUR}" >&2
}
log_warn() {
  echo -e "${ORANGE}${*}${NOCOLOUR}" >&2
}
log_error() {
  echo -e "${LIGHTRED}${*}${NOCOLOUR}" >&2
}
log_debug() {
  if [[ -n "${SCRIPT_LOG_LEVEL:-}" && "${SCRIPT_LOG_LEVEL}" == "debug" ]]; then
    echo -e "${CYAN}${*}${NOCOLOUR}" >&2
  fi
}
log_info() {
  echo -e "${LIGHTCYAN}${*}${NOCOLOUR}" >&2
}
log_success() {
  echo -e "${GREEN}${*}${NOCOLOUR}" >&2
}
log_fatal() {
  echo -e "${RED}${*}${NOCOLOUR}" >&2
  exit "${2:-"1"}"
}
