#!/usr/bin/env bash

set -xeu

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
INTEGRATION_TEST_DIR="$SCRIPT_DIR/../integration-tests/"

LABEL_DATE=$(date -u +'%Y-%m-%dT%H:%M:%S.%3NZ')
LABEL_URL="https://github.com/goss-org/goss"
LABEL_REVISION=$(git rev-parse HEAD)

for docker_file in $INTEGRATION_TEST_DIR/Dockerfile_*; do
    [[ $docker_file == *.md5 ]] && continue
    os=$(cut -d '_' -f2 <<<"$docker_file")
    docker build \
        --label "org.opencontainers.image.created=$LABEL_DATE" \
        --label "org.opencontainers.image.description=Quick and Easy server testing/validation" \
        --label "org.opencontainers.image.licenses=Apache-2.0" \
        --label "org.opencontainers.image.revision=$LABEL_REVISION" \
        --label "org.opencontainers.image.source=$LABEL_URL" \
        --label "org.opencontainers.image.title=goss" \
        --label "org.opencontainers.image.url=$LABEL_URL" \
        --label "org.opencontainers.image.version=manual" \
        -t "aelsabbahy/goss_${os}:latest" - < "$docker_file"
done
