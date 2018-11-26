#!/usr/bin/env bash
set -euo pipefail

# Bash replacement for glide novendor command
# Returns all directories which include go files

DIRS=$(ls -ld */ . | awk {'print $9'} | grep -v vendor)

for DIR in ${DIRS}; do
    GOFILES=$(git ls-files ${DIR} | grep ".*\.go$") || true

    if [[ ${DIR} == "."  ]]; then
        echo "."
        continue
    fi

    if [[ ${GOFILES} != "" ]]; then
        echo "./"${DIR}"..."
    fi
done

exit 0
