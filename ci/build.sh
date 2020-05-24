#!/usr/bin/env bash
set -euo pipefail

os_name="${1:?"No value from TRAVIS_OS_NAME in 1st arg. This is meant to be run in Travis CI, see also https://docs.travis-ci.com/user/environment-variables/#convenience-variables"}"

# darwin & windows do not support integration-testing approach via docker, so on those, just run fast tests.
# linux runs all tests; unit and integration.
if [[ "${os_name}" == "osx" ]]; then
  make test-short-all
  # darwin is the GOOS value which is easier to work with
  integration-tests/run-tests-alpha.sh "darwin"
elif [[ "${os_name}" == "windows" ]]; then
  make test-short-all
  integration-tests/run-tests-alpha.sh "windows"
else
  make all
fi
