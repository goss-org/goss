#!/usr/bin/env bash
set -euo pipefail

os_name="$(go env GOOS)"

go get -u golang.org/x/lint/golint

if [[ "${goos}" != "windows" ]]; then
  curl -L "https://codeclimate.com/downloads/test-reporter/test-reporter-latest-${goos}-amd64" > "./cc-test-reporter"
  chmod +x "./cc-test-reporter"
fi
