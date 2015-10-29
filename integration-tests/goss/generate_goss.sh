#!/usr/bin/env bash

SCRIPT_DIR=$(readlink -f $(dirname $0))

OS=$1

goss() {
  $SCRIPT_DIR/goss -g $SCRIPT_DIR/${OS}/goss-generated.json "$@"
  # Validate that duplicates are ignored
  $SCRIPT_DIR/goss -g $SCRIPT_DIR/${OS}/goss-generated.json "$@"
}

rm -f $SCRIPT_DIR/goss*generated*.json

for x in /etc/passwd /tmp/goss/foobar;do
  goss a file $x
done

[[ $OS == "centos6" ]] && package="httpd" || package="apache2"
[[ $OS == "centos6" ]] && user="apache" || user="www-data"
for x in $package foobar vim-tiny;do
  goss a package $x
done

for x in google.com:443 google.com:22;do
  goss a addr $x
done

for x in tcp6:80 9999;do
  goss a port $x
done

for x in $package foobar;do
  goss a service $x
done

for x in $user foobar;do
  goss a user $x
done

for x in $user foobar;do
  goss a group $x
done

for x in "$package -v" foobar;do
  goss a command "$x"
done

for x in localhost;do
  goss a dns $x
done

for x in $package foobar;do
  goss a process $x
done


# Auto-add
$SCRIPT_DIR/goss -g $SCRIPT_DIR/${OS}/goss-aa-generated.json aa $package
# Validate that duplicates are ignored
$SCRIPT_DIR/goss -g $SCRIPT_DIR/${OS}/goss-aa-generated.json aa $package
