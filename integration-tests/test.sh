#!/usr/bin/env bash

cp ../release/goss-linux-amd64 goss/goss

if ! docker images | grep aelsabbahy/goss_centos;then
  docker build -t aelsabbahy/goss_centos .
fi

if ! docker ps | grep goss_int_test;then
  docker run -v $PWD/goss:/tmp/goss  -d --name goss_int_test aelsabbahy/goss_centos /sbin/init
  # Give httpd time to start up
  sleep 5
fi

out=$(docker exec goss_int_test bash -c 'time /tmp/goss/goss -f /tmp/goss/goss.json validate')
echo "$out"

grep -q 'Count: 36 failed: 0' <<<"$out"
exit_code=$?

docker exec goss_int_test bash -c 'time /tmp/goss/generate_goss.sh > /dev/null'

docker exec goss_int_test bash -c 'diff -u /tmp/goss/goss-expected.json /tmp/goss/goss-generated.json'
exit_code=$(($exit_code+$?))

docker exec goss_int_test bash -c 'diff -u <(/tmp/goss/goss -f /tmp/goss/goss-render.json render) /tmp/goss/goss-expected.json'
exit_code=$(($exit_code+$?))
#docker rm -vf goss_int_test

exit $exit_code
