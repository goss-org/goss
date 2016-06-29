#!/usr/bin/env bash

set -xeu

os=$1

for arch in amd64 386;do
  cp ../release/goss-linux-$arch goss/$os/
  if ! docker images | grep aelsabbahy/goss_$os;then
    docker build -t aelsabbahy/goss_$os - < Dockerfile_$os
  fi

  if ! docker ps | grep goss_int_test_$os;then
    if docker ps -a | grep goss_int_test_$os;then
      docker rm -vf goss_int_test_$os
    fi
    #id=$(docker run --cap-add SYS_ADMIN --security-opt seccomp:unconfined -v "$PWD/goss:/goss"  -d --name "goss_int_test_$os" "aelsabbahy/goss_$os" /sbin/init)
    id=$(docker run --cap-add SYS_ADMIN -v "$PWD/goss:/goss"  -d --name "goss_int_test_$os" "aelsabbahy/goss_$os" /sbin/init)
    ip=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' "$id")
    trap "rv=\$?; docker rm -vf $id; exit \$rv" INT TERM EXIT
    # Give httpd time to start up
    for i in {1..10};do curl -sL -o /dev/null -m 1 "$ip" && break;sleep 1;done
    #sleep 10
  fi

  out=$(docker exec goss_int_test_$os bash -c "time /goss/$os/goss-linux-$arch -g /goss/$os/goss.json validate")
  echo "$out"

  if [[ $os == "arch" ]]; then
    egrep -q 'Count: 25, Failed: 0' <<<"$out"
  else
    egrep -q 'Count: 44, Failed: 0' <<<"$out"
  fi

  if [[ ! $os == "arch" ]]; then
    docker exec goss_int_test_$os bash -c "bash -x /goss/generate_goss.sh $os $arch"

    #docker exec goss_int_test_$os bash -c "cp /goss/${os}/goss-generated.json /goss/${os}/goss-expected.json"
    docker exec goss_int_test_$os bash -c "diff -wu /goss/${os}/goss-expected.json /goss/${os}/goss-generated.json"

    #docker exec goss_int_test_$os bash -c "cp /goss/${os}/goss-aa-generated.json /goss/${os}/goss-aa-expected.json"
    docker exec goss_int_test_$os bash -c "diff -wu /goss/${os}/goss-aa-expected.json /goss/${os}/goss-aa-generated.json"

    docker exec goss_int_test_$os bash -c "bash -x /goss/generate_goss.sh $os $arch -q"

    #docker exec goss_int_test_$os bash -c "cp /goss/${os}/goss-generated.json /goss/${os}/goss-expected-q.json"
    docker exec goss_int_test_$os bash -c "diff -wu /goss/${os}/goss-expected-q.json /goss/${os}/goss-generated.json"
  fi

  #docker rm -vf goss_int_test_$os
done
