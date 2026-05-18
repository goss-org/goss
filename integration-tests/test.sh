#!/usr/bin/env bash
# shellcheck source=../ci/lib/setup.sh
source "$(dirname "${BASH_SOURCE[0]}")/../ci/lib/setup.sh" || exit 67
# preserve current behaviour
set -x

: "${DOCKER_BIN:=docker}"

os="${1:?"Need OS as 1st arg. e.g. alpine arch rockylinux9 jammy bullseye"}"
arch="${2:?"Need arch as 2nd arg. e.g. amd64 386"}"

vars_inline="{inline: bar, overwrite: bar}"
container_repository="aelsabbahy"

# setup places us inside repo-root; this preserves current behaviour with least change.
cd integration-tests

cp "../release/goss-linux-$arch" "goss/$os/"
# Run build if Dockerfile has changed but hasn't been pushed to dockerhub
if ! md5sum -c "Dockerfile_${os}.md5"; then
  $DOCKER_BIN build -t "$container_repository/goss_${os}:latest" --file "Dockerfile_$os" .
# Pull if image doesn't exist locally
elif ! $DOCKER_BIN images | grep "$container_repository/goss_$os";then
  $DOCKER_BIN pull "$container_repository/goss_$os"
fi

container_name="goss_int_test_${os}_${arch}"
docker_exec() {
  $DOCKER_BIN exec "$container_name" "$@"
}

# Cleanup any old containers
if $DOCKER_BIN ps -a | grep "$container_name";then
  $DOCKER_BIN rm -vf "$container_name"
fi

# Setup local httbin
# FIXME: this is a quick hack to fix intermittent CI issues
network=goss-test
$DOCKER_BIN network create --driver bridge --subnet '172.19.0.0/16' $network
$DOCKER_BIN run -d --name httpbin --network $network docker.io/kennethreitz/httpbin
opts=(--env OS=$os --cap-add SYS_ADMIN -v "$PWD/goss:/goss" -d --name "$container_name" --security-opt seccomp:unconfined --security-opt label:disable --privileged)
id=$($DOCKER_BIN run "${opts[@]}" --network $network "$container_repository/goss_$os" /sbin/init)
ip=$($DOCKER_BIN inspect --format '{{ .NetworkSettings.IPAddress }}' "$id")
trap "rv=\$?; $DOCKER_BIN rm -vf $id || :;$DOCKER_BIN rm -vf httpbin || :;$DOCKER_BIN network rm $network || :; exit \$rv" INT TERM EXIT
# Give httpd time to start up, adding 1 second to see if it helps with intermittent CI failures
docker_exec "/goss/$os/goss-linux-$arch" -g "/goss/goss-wait.yaml" validate -r 10s -s 100ms && sleep 1
out=$(docker_exec "/goss/$os/goss-linux-$arch" --vars "/goss/vars.yaml" --vars-inline "$vars_inline" -g "/goss/$os/goss.yaml" validate)
echo "$out"

if [[ $os == "arch" ]]; then
    egrep -q 'Count: 106, Failed: 0, Skipped: 3' <<<"$out"
else
    egrep -q 'Count: 127, Failed: 0, Skipped: 5' <<<"$out"
fi

if [[ ! $os == "arch" ]]; then
  docker_exec /goss/generate_goss.sh "$os" "$arch"

  # docker exec $container_name bash -c "cp /goss/${os}/goss-generated-$arch.yaml /goss/${os}/goss-expected.yaml"
  docker_exec diff -wu "/goss/${os}/goss-expected.yaml" "/goss/${os}/goss-generated-$arch.yaml"

  # docker exec $container_name bash -c "cp /goss/${os}/goss-aa-generated-$arch.yaml /goss/${os}/goss-aa-expected.yaml"
  docker_exec diff -wu "/goss/${os}/goss-aa-expected.yaml" "/goss/${os}/goss-aa-generated-$arch.yaml"

  docker_exec /goss/generate_goss.sh "$os" "$arch" -q

  # docker exec $container_name bash -c "cp /goss/${os}/goss-generated-$arch.yaml /goss/${os}/goss-expected-q.yaml"
  docker_exec diff -wu "/goss/${os}/goss-expected-q.yaml" "/goss/${os}/goss-generated-$arch.yaml"
fi

#docker rm -vf goss_int_test_$os
