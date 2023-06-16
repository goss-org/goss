#!/usr/bin/env bash
set -euo pipefail

command -v go

go test -coverprofile="c.out" "${1}"

sed 's|github.com/goss-org/goss/||' <"c.out" >"c.out.tmp"

mv "c.out.tmp" "c.out"
