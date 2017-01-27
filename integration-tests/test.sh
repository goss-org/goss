#!/usr/bin/env bash

set -xeu

os=$1
arch=$2

seccomp_opts() {
  local docker_ver minor_ver
  docker_ver=$(docker version -f '{{.Client.Version}}')
  minor_ver=$(cut -d'.' -f2 <<<$docker_ver)
  if ((minor_ver>=10)); then
    echo '--security-opt seccomp:unconfined'
  fi
}

cp "../release/goss-linux-$arch" "goss/$os/"
# Run build if Dockerfile has changed but hasn't been pushed to dockerhub
if ! md5sum -c "Dockerfile_${os}.md5"; then
  docker build -t "aelsabbahy/goss_${os}:latest" - < "Dockerfile_$os"
# Pull if image doesn't exist locally
elif ! docker images | grep "aelsabbahy/goss_$os";then
  docker pull "aelsabbahy/goss_$os"
fi

container_name="goss_int_test_${os}_${arch}"
docker_exec() {
  docker exec "$container_name" "$@"
}

# Cleanup any old containers
if docker ps -a | grep "$container_name";then
  docker rm -vf "$container_name"
fi
opts=(--cap-add SYS_ADMIN -v "$PWD/goss:/goss"  -d --name "$container_name" $(seccomp_opts))
id=$(docker run "${opts[@]}" "aelsabbahy/goss_$os" /sbin/init)
ip=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' "$id")
trap "rv=\$?; docker rm -vf $id; exit \$rv" INT TERM EXIT
# Give httpd time to start up, adding 1 second to see if it helps with intermittent CI failures
[[ $os != "arch" ]] && docker_exec "/goss/$os/goss-linux-$arch" -g "/goss/goss-wait.json" validate -r 10s -s 100ms && sleep 1

#out=$(docker exec "$container_name" bash -c "time /goss/$os/goss-linux-$arch -g /goss/$os/goss.json validate")
out=$(docker_exec "/goss/$os/goss-linux-$arch" -g "/goss/$os/goss.json" validate)
echo "$out"

if [[ $os == "arch" ]]; then
  egrep -q 'Count: 37, Failed: 0' <<<"$out"
else
  egrep -q 'Count: 53, Failed: 0' <<<"$out"
fi

if [[ ! $os == "arch" ]]; then
  docker_exec /goss/generate_goss.sh "$os" "$arch"

  #docker exec goss_int_test_$os bash -c "cp /goss/${os}/goss-generated-$arch.json /goss/${os}/goss-expected.json"
  docker_exec diff -wu "/goss/${os}/goss-expected.json" "/goss/${os}/goss-generated-$arch.json"

  #docker exec goss_int_test_$os bash -c "cp /goss/${os}/goss-aa-generated-$arch.json /goss/${os}/goss-aa-expected.json"
  docker_exec diff -wu "/goss/${os}/goss-aa-expected.json" "/goss/${os}/goss-aa-generated-$arch.json"

  docker_exec /goss/generate_goss.sh "$os" "$arch" -q

  #docker exec goss_int_test_$os bash -c "cp /goss/${os}/goss-generated-$arch.json /goss/${os}/goss-expected-q.json"
  docker_exec diff -wu "/goss/${os}/goss-expected-q.json" "/goss/${os}/goss-generated-$arch.json"
fi

#docker rm -vf goss_int_test_$os
