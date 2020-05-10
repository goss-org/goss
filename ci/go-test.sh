#!/usr/bin/env bash
set -euo pipefail

command -v go

go test -coverprofile="c.out" "${@}"

sed 's|github.com/aelsabbahy/goss/||' <"c.out" >"c.out.tmp"

mv "c.out.tmp" "c.out"
