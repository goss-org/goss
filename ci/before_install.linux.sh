#!/usr/bin/env bash
# -ET: propagate DEBUG/RETURN/ERR traps to functions and subshells
# -e: exit on unhandled error
# pipefail: any failure in a pipe causes the pipe to fail
set -eET -o pipefail
[[ -n "${DEBUG:-}" ]] && set -x
if ! cd "$(dirname "${BASH_SOURCE[0]}")/.."; then
  echo -e "Failed to cd to repository root"
  return 1
fi

curl -L https://github.com/Masterminds/glide/releases/download/v0.11.0/glide-v0.11.0-linux-amd64.zip --output glide.zip
unzip glide.zip
go get -u golang.org/x/lint/golint
