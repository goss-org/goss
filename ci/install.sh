#!/usr/bin/env bash
set -euo pipefail

os_name="$(go env GOOS)"

go get -u golang.org/x/lint/golint

if [[ "${os_name}" != "windows" ]]; then
  curl -L "https://codeclimate.com/downloads/test-reporter/test-reporter-latest-${os_name}-amd64" > "./cc-test-reporter"
  chmod +x "./cc-test-reporter"
fi
