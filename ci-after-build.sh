#!/usr/bin/env bash
set -euo pipefail
set -x
goos="${TRAVIS_OS_NAME:?"No value for TRAVIS_OS_NAME. This is meant to be run in Travis CI, see also https://docs.travis-ci.com/user/environment-variables/#convenience-variables"}"

if [[ "${goos}" != "windows" ]]; then
  ./cc-test-reporter after-build --exit-code "${TRAVIS_TEST_RESULT}" -d
fi
