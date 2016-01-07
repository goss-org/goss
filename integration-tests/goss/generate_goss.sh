#!/usr/bin/env bash

SCRIPT_DIR=$(readlink -f $(dirname $0))

OS=$1
[[ $2 == "-q" ]] && args=("--exclude-attr" "*")

goss() {
  $SCRIPT_DIR/goss -g $SCRIPT_DIR/${OS}/goss-generated.json "$@"
  # Validate that duplicates are ignored
  $SCRIPT_DIR/goss -g $SCRIPT_DIR/${OS}/goss-generated.json "$@"
}

rm -f $SCRIPT_DIR/${OS}/goss*generated*.json

for x in /etc/passwd /tmp/goss/foobar;do
  goss a "${args[@]}" file $x
done

[[ $OS == "centos6" ]] && package="httpd" || package="apache2"
[[ $OS == "centos6" ]] && user="apache" || user="www-data"
for x in $package foobar vim-tiny;do
  goss a "${args[@]}" package $x
done

for x in google.com:443 google.com:22;do
  goss a "${args[@]}" addr $x
done

for x in tcp:80 tcp6:80 9999;do
  goss a "${args[@]}" port $x
done

for x in $package foobar;do
  goss a "${args[@]}" service $x
done

for x in $user foobar;do
  goss a "${args[@]}" user $x
done

for x in $user foobar;do
  goss a "${args[@]}" group $x
done

for x in "$package -v" foobar;do
  goss a "${args[@]}" command "$x"
done

for x in localhost;do
  goss a "${args[@]}" dns $x
done

for x in $package foobar;do
  goss a "${args[@]}" process $x
done


# Auto-add
$SCRIPT_DIR/goss -g $SCRIPT_DIR/${OS}/goss-aa-generated.json aa $package
# Validate that duplicates are ignored
$SCRIPT_DIR/goss -g $SCRIPT_DIR/${OS}/goss-aa-generated.json aa $package
