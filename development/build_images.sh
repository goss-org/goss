#!/usr/bin/env bash

set -xeu

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INTEGRATION_TEST_DIR="$SCRIPT_DIR/../integration-tests/"


for docker_file in $INTEGRATION_TEST_DIR/Dockerfile_*; do
    [[ $docker_file == *.md5 ]] && continue
    os=$(cut -d '_' -f2 <<<"$docker_file")
    docker build -t "aelsabbahy/goss_${os}:latest" - < "$docker_file"
done
