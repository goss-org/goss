#!/usr/bin/env bash

set -xe
cp ../release/goss-linux-amd64 goss/goss

for os in centos6 wheezy;do
  if ! docker images | grep aelsabbahy/goss_$os;then
    docker build -t aelsabbahy/goss_$os - < Dockerfile_$os
  fi

  if ! docker ps | grep goss_int_test_$os;then
    if docker ps -a | grep goss_int_test_$os;then
      docker rm -vf goss_int_test_$os
    fi
    docker run --privileged -v $PWD/goss:/tmp/goss  -d --name goss_int_test_$os aelsabbahy/goss_$os /sbin/init
    # Give httpd time to start up
    sleep 10
  fi

  out=$(docker exec goss_int_test_$os bash -c 'time /tmp/goss/goss -g /tmp/goss/goss.json validate')
  echo "$out"

  grep -q 'Count: 37 failed: 0' <<<"$out"

  docker exec goss_int_test_$os bash -c 'time /tmp/goss/generate_goss.sh > /dev/null'

  docker exec goss_int_test_$os bash -c 'diff -u /tmp/goss/goss-expected.json /tmp/goss/goss-generated.json'

  docker exec goss_int_test_$os bash -c 'diff -u /tmp/goss/goss-aa-expected.json /tmp/goss/goss-aa-generated.json'

  docker exec goss_int_test_$os bash -c 'diff -u /tmp/goss/goss-expected.json <(/tmp/goss/goss -g /tmp/goss/goss-render.json render)'

  #docker rm -vf goss_int_test_$os
done
