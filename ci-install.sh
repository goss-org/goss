#!/usr/bin/env bash
set -euo pipefail

goos="${TRAVIS_OS_NAME:?"No value for TRAVIS_OS_NAME. This is meant to be run in Travis CI, see also https://docs.travis-ci.com/user/environment-variables/#convenience-variables"}"
if [[ "${goos}" == "osx" ]]; then
  goos="darwin"
fi

go get -u golang.org/x/lint/golint

if [[ "${goos}" != "windows" ]]; then
  curl -L "https://codeclimate.com/downloads/test-reporter/test-reporter-latest-${goos}-amd64${extension}" > "./cc-test-reporter"
  chmod +x "./cc-test-reporter"
fi
