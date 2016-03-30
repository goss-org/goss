#!/usr/bin/env bash

set -xeu

os=$1

for arch in amd64 386;do
  cp ../release/goss-linux-$arch goss/$os/
  if ! sudo docker images | grep aelsabbahy/goss_$os;then
    sudo docker build -t aelsabbahy/goss_$os - < Dockerfile_$os
  fi

  if ! sudo docker ps | grep goss_int_test_$os;then
    if sudo docker ps -a | grep goss_int_test_$os;then
      sudo docker rm -vf goss_int_test_$os
    fi
    id=$(sudo docker run -v "$PWD/goss:/tmp/goss"  -d --name "goss_int_test_$os" "aelsabbahy/goss_$os" /sbin/init)
    ip=$(sudo docker inspect --format '{{ .NetworkSettings.IPAddress }}' "$id")
    trap "rv=\$?; sudo docker rm -vf $id; exit \$rv" INT TERM EXIT
    # Give httpd time to start up
    for i in {1..10};do curl -sL -o /dev/null -m 1 "$ip" && break;sleep 1;done
    #sleep 10
  fi

  out=$(sudo docker exec goss_int_test_$os bash -c "time /tmp/goss/$os/goss-linux-$arch -g /tmp/goss/$os/goss.json validate")
  echo "$out"

  grep -q 'Count: 39, Failed: 0' <<<"$out"

  sudo docker exec goss_int_test_$os bash -c "bash -x /tmp/goss/generate_goss.sh $os $arch"

  sudo docker exec goss_int_test_$os bash -c "diff -wu /tmp/goss/${os}/goss-expected.json /tmp/goss/${os}/goss-generated.json"

  sudo docker exec goss_int_test_$os bash -c "diff -wu /tmp/goss/${os}/goss-aa-expected.json /tmp/goss/${os}/goss-aa-generated.json"

  sudo docker exec goss_int_test_$os bash -c "bash -x /tmp/goss/generate_goss.sh $os $arch -q"

  sudo docker exec goss_int_test_$os bash -c "diff -wu /tmp/goss/${os}/goss-expected-q.json /tmp/goss/${os}/goss-generated.json"

  #docker rm -vf goss_int_test_$os
done
