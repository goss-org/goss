#!/usr/bin/env bash
set -euo pipefail

# gofmt must be on PATH
command -v gofmt

fmt="$(gofmt -l "${GO_FILES}")"

if [[ -z "${fmt}" ]]; then
  echo "valid gofmt"
else
  echo "invalid gofmt:"
  echo "${fmt}"
  exit 1
fi
