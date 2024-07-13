#!/usr/bin/env bash

SCRIPT_DIR=$(readlink -f $(dirname $0))

OS=$1
ARCH=$2
[[ $3 == "-q" ]] && args=("--exclude-attr" "*")

goss() {
  $SCRIPT_DIR/$OS/goss-linux-$ARCH -g $SCRIPT_DIR/${OS}/goss-generated-$ARCH.yaml "$@"
  # Validate that duplicates are ignored
  $SCRIPT_DIR/$OS/goss-linux-$ARCH -g $SCRIPT_DIR/${OS}/goss-generated-$ARCH.yaml "$@"
}

rm -f $SCRIPT_DIR/${OS}/goss*generated*-$ARCH.yaml

for x in /etc/passwd /tmp/goss/foobar;do
  goss a "${args[@]}" file $x
done

[[ $OS == "centos7" || $OS == "rockylinux9" ]] && package="httpd" || package="apache2"
[[ $OS == "centos7" || $OS == "rockylinux9" ]] && user="apache" || user="www-data"
goss a "${args[@]}" package $package foobar vim-tiny

goss a "${args[@]}" addr --timeout 1s httpbin:80 httpbin:22

goss a "${args[@]}" addr --timeout 1s udp://8.8.8.8:53

goss a "${args[@]}" port tcp:80 tcp6:80 9999

goss a "${args[@]}" service $package foobar

goss a "${args[@]}" user $user foobar

goss a "${args[@]}" group $user foobar

goss a "${args[@]}" command "echo 'hi'" foobar

goss a "${args[@]}" dns --timeout 1s --server 8.8.8.8 CNAME:c.dnstest.io

goss a "${args[@]}" dns --timeout 1s --server 8.8.8.8 MX:dnstest.io

goss a "${args[@]}" dns --timeout 1s --server 8.8.8.8 NS:dnstest.io

goss a "${args[@]}" dns --timeout 1s --server 8.8.8.8 PTR:54.243.154.1

goss a "${args[@]}" dns --timeout 1s --server 8.8.8.8 SRV:_https._tcp.dnstest.io

goss a "${args[@]}" dns --timeout 1s --server 8.8.8.8 TXT:txt._test.dnstest.io

goss a "${args[@]}" dns --timeout 1s --server 8.8.8.8 CAA:dnstest.io

goss a "${args[@]}" dns --timeout 1s --server 8.8.8.8 ip6.dnstest.io

goss a "${args[@]}" dns --timeout 1s localhost

goss a "${args[@]}" process $package foobar

goss a "${args[@]}" kernel-param kernel.ostype

goss a "${args[@]}" mount /dev
# Make tests consistent across different docker setups
sed -i '/- seclabel/d' $SCRIPT_DIR/${OS}/goss-generated-$ARCH.yaml
sed -i '/- size=/d' $SCRIPT_DIR/${OS}/goss-generated-$ARCH.yaml
sed -i '/- mode=/d' $SCRIPT_DIR/${OS}/goss-generated-$ARCH.yaml
sed -i '/- inode64/d' $SCRIPT_DIR/${OS}/goss-generated-$ARCH.yaml

goss a "${args[@]}" http https://www.google.com

goss a "${args[@]}" http https://www.apple.com -x http://127.0.0.1:8888

goss a "${args[@]}" http http://google.com -r

# Auto-add
# Validate that empty configs don't get created
$SCRIPT_DIR/$OS/goss-linux-$ARCH -g $SCRIPT_DIR/${OS}/goss-aa-generated-$ARCH.yaml aa nosuchresource
if [[ -f $SCRIPT_DIR/${OS}/goss-aa-generated-$ARCH.yaml ]]
then
  echo "Error! Empty config file exists!" && exit 1
fi
$SCRIPT_DIR/$OS/goss-linux-$ARCH -g $SCRIPT_DIR/${OS}/goss-aa-generated-$ARCH.yaml aa $package
# Validate that duplicates are ignored
$SCRIPT_DIR/$OS/goss-linux-$ARCH -g $SCRIPT_DIR/${OS}/goss-aa-generated-$ARCH.yaml aa $package
# Validate that we can aa none existent resources without destroying the file
$SCRIPT_DIR/$OS/goss-linux-$ARCH -g $SCRIPT_DIR/${OS}/goss-aa-generated-$ARCH.yaml aa nosuchresource

if [[ ! -f $SCRIPT_DIR/${OS}/goss-aa-generated-$ARCH.yaml ]]
then
  echo "Error! Config file removed by aa!" && exit 1
fi
