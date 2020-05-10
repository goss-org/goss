#!/usr/bin/env bash
set -euo pipefail

os_name="${1:?"No value from TRAVIS_OS_NAME in 1st arg. This is meant to be run in Travis CI, see also https://docs.travis-ci.com/user/environment-variables/#convenience-variables"}"
goos="${os_name}"
if [[ "${goos}" == "osx" ]]; then
  goos="darwin"
fi

go get -u golang.org/x/lint/golint

if [[ "${goos}" != "windows" ]]; then
  curl -L "https://codeclimate.com/downloads/test-reporter/test-reporter-latest-${goos}-amd64" > "./cc-test-reporter"
  chmod +x "./cc-test-reporter"
fi
