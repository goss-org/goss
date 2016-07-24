#!/usr/bin/env bash

set -xeu

images=$(docker images | grep '^aelsabbahy/goss_.*latest' | awk '$0=$1')

for image in $images; do
  docker push "${image}:latest"
done
