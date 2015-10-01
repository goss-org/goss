#!/usr/bin/env bash

SCRIPT_DIR=$(readlink -f $(dirname $0))

goss() {
  $SCRIPT_DIR/goss -f $SCRIPT_DIR/goss-generated.json "$@"
}

rm -f $SCRIPT_DIR/goss-generated.json

for x in /etc/httpd/conf.d/welcome.conf /tmp/goss/foobar;do
  goss a file $x
done

for x in httpd foobar;do
  goss a package $x
done

for x in google.com:443 google.com:22;do
  goss a addr $x
done

for x in tcp6:80 9999;do
  goss a port $x
done

for x in httpd foobar;do
  goss a service $x
done

for x in apache foobar;do
  goss a user $x
done

for x in apache foobar;do
  goss a group $x
done

for x in "httpd -v" foobar;do
  goss a command "$x"
done

for x in localhost;do
  goss a dns $x
done

for x in httpd foobar;do
  goss a process $x
done
