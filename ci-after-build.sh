#!/usr/bin/env bash
set -euo pipefail

os_name="${1:?"No value from TRAVIS_OS_NAME in 1st arg. This is meant to be run in Travis CI, see also https://docs.travis-ci.com/user/environment-variables/#convenience-variables"}"

if [[ "${os_name}" != "windows" ]]; then
  ./cc-test-reporter after-build --exit-code "${TRAVIS_TEST_RESULT}" -d
fi
