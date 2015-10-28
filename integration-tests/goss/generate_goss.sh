#!/usr/bin/env bash

SCRIPT_DIR=$(readlink -f $(dirname $0))

OS=$1
ARCH=$2
[[ $3 == "-q" ]] && args=("--exclude-attr" "*")

goss() {
  $SCRIPT_DIR/$OS/goss-linux-$ARCH -g $SCRIPT_DIR/${OS}/goss-generated-$ARCH.json "$@"
  # Validate that duplicates are ignored
  $SCRIPT_DIR/$OS/goss-linux-$ARCH -g $SCRIPT_DIR/${OS}/goss-generated-$ARCH.json "$@"
}

rm -f $SCRIPT_DIR/${OS}/goss*generated*-$ARCH.json

for x in /etc/passwd /tmp/goss/foobar;do
  goss a "${args[@]}" file $x
done

[[ $OS == "centos7" ]] && package="httpd" || package="apache2"
[[ $OS == "centos7" ]] && user="apache" || user="www-data"
goss a "${args[@]}" package $package foobar vim-tiny

goss a "${args[@]}" addr --timeout 1s google.com:443 google.com:22

goss a "${args[@]}" port tcp:80 tcp6:80 9999

goss a "${args[@]}" service $package foobar

goss a "${args[@]}" user $user foobar

goss a "${args[@]}" group $user foobar

goss a "${args[@]}" command "echo 'hi'" foobar

goss a "${args[@]}" dns --timeout 1s localhost

goss a "${args[@]}" process $package foobar

goss a "${args[@]}" kernel-param kernel.ostype

goss a "${args[@]}" mount /dev

goss a "${args[@]}" http https://www.google.com

# Auto-add
$SCRIPT_DIR/$OS/goss-linux-$ARCH -g $SCRIPT_DIR/${OS}/goss-aa-generated-$ARCH.json aa $package
# Validate that duplicates are ignored
$SCRIPT_DIR/$OS/goss-linux-$ARCH -g $SCRIPT_DIR/${OS}/goss-aa-generated-$ARCH.json aa $package
