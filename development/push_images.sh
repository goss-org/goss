#!/usr/bin/env bash

set -xeu

SCRIPT_DIR=$(readlink -f "$(dirname "$0")")
# "docker images" does not show the docker registry "docker.io/"
CONTAINER_REPOSITORY="aelsabbahy"
images=$(docker images | grep "^$CONTAINER_REPOSITORY/goss_.*latest" | awk '$0=$1')

# Use md5sum to determine if CI needs to do a docker build
pushd "$SCRIPT_DIR/../integration-tests";
  for dockerfile in Dockerfile_*;do
    [[ $dockerfile == *.md5 ]] && continue
    md5sum "$dockerfile" > "${dockerfile}.md5"
  done
popd

for image in $images; do
  docker push "${image}:latest"
done
