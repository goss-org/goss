# Goss - Quick and Easy server validation
[![Build Status](https://travis-ci.org/aelsabbahy/goss.svg?branch=master)](https://travis-ci.org/aelsabbahy/goss)
[![Github All Releases](https://img.shields.io/github/downloads/aelsabbahy/goss/total.svg?maxAge=604800)](https://github.com/aelsabbahy/goss/releases)
**
[![Twitter Follow](https://img.shields.io/twitter/follow/aelsabbahy1.svg?style=social&label=Follow&maxAge=2592000)]()
[![Blog](https://img.shields.io/badge/follow-blog-brightgreen.svg)](https://medium.com/@aelsabbahy)

## Goss in 45 seconds

**Note:** For an even faster way of doing this, see: [autoadd](https://github.com/aelsabbahy/goss/blob/master/docs/manual.md#autoadd-aa---auto-add-all-matching-resources-to-test-suite)

**Note:** For some Docker use cases, see my blog post [here](https://medium.com/@aelsabbahy/docker-1-12-kubernetes-simplified-health-checks-and-container-ordering-with-goss-fa8debbe676c) and [dockerhub repo](https://hub.docker.com/r/aelsabbahy/goss/)

<a href="https://asciinema.org/a/4suhr8p42qcn6r7crfzt6cc3e?autoplay=1" target="_blank"><img src="https://cloud.githubusercontent.com/assets/6783261/17330426/ce7ad066-5894-11e6-84ea-29fd4207af58.gif" alt="asciicast"></a>

## Introduction

### What is Goss?

Goss is a YAML based [serverspec](http://serverspec.org/)-like tool for validating a server’s configuration. It eases the process of writing tests by allowing the user to generate tests from the current system state. Once the test suite is written they can be executed, waited-on, or served as a health endpoint.

### Why use Goss?

* Goss is EASY!  - [Goss in 45 seconds](#goss-in-45-seconds)
* Goss is FAST!  - small-medium test suits are near instantaneous, see [benchmarks](https://github.com/aelsabbahy/goss/wiki/Benchmarks)
* Goss is SMALL! - <10MB single self-contained binary

## Installation

```bash
# Install latest version to /usr/local/bin
curl -fsSL https://goss.rocks/install | sh

# Install v0.2.5 version to ~/bin
curl -fsSL https://goss.rocks/install | GOSS_VER=v0.2.5 GOSS_DST=~/bin sh
```

### Manual installation
```bash
# See https://github.com/aelsabbahy/goss/releases for release versions
curl -L https://github.com/aelsabbahy/goss/releases/download/_VERSION_/goss-linux-amd64 -o /usr/local/bin/goss
chmod +rx /usr/local/bin/goss
```

## Full Documentation

Documentation is available here: https://github.com/aelsabbahy/goss/blob/master/docs/manual.md

## Quick start

### Writing a simple sshd test

An initial set of tests can be derived from the system state by using the [add](https://github.com/aelsabbahy/goss/blob/master/docs/manual.md#add-a---add-system-resource-to-test-suite) or [autoadd](https://github.com/aelsabbahy/goss/blob/master/docs/manual.md#autoadd-aa---auto-add-all-matching-resources-to-test-suite) commands.

Let's write a simple sshd test using autoadd.

```
# Running it as root will allow it to also detect ports
$ sudo goss autoadd sshd
```
Generated `goss.yaml`:
```yaml
$ cat goss.yaml
port:
  tcp:22:
    listening: true
    ip:
    - 0.0.0.0
  tcp6:22:
    listening: true
    ip:
    - '::'
service:
  sshd:
    enabled: true
    running: true
user:
  sshd:
    exists: true
    uid: 74
    gid: 74
    groups:
    - sshd
    home: /var/empty/sshd
    shell: /sbin/nologin
group:
  sshd:
    exists: true
    gid: 74
process:
  sshd:
    running: true
```
Now that we have a test suite, we can:

* Run it once
```
goss validate
...............

Total Duration: 0.021s # <- yeah, it's that fast..
Count: 15, Failed: 0

```
* keep running it until the system enters a valid state or we timeout
```
goss validate --retry-timeout 30s --sleep 1s
```
* serve the tests as a health endpoint
```
goss serve &
curl localhost:8080/healthz

# JSON endpoint
goss serve --format json &
curl localhost:8080/healthz
```

### Patterns, matchers and metadata
Goss files can be manually edited to match:
* [Patterns](https://github.com/aelsabbahy/goss/blob/master/docs/manual.md#patterns)
* [Advanced Matchers](https://github.com/aelsabbahy/goss/blob/master/docs/manual.md#advanced-matchers).
* `title` and `meta` (arbitrary data) attributes are persisted when adding other resources with `goss add`

Some examples:
```yaml
user:
  sshd:
    title: UID must be between 50-100, GID doesn't matter. home is flexible
    meta:
      desc: Ensure sshd is enabled and running since it's needed for system management
      sev: 5
    exists: true
    uid:
      # Validate that UID is between 50 and 100
      and:
        gt: 50
        lt: 100
    home:
      # Home can be any of the following
      or:
      - /var/empty/sshd
      - /var/run/sshd

package:
  kernel:
    installed: true
    versions:
      # Must have 3 kernels and none of them can be 4.4.0
      and:
      - have-len: 3
      - not:
          contain-element: 4.4.0
```

## Supported resources
* package - add new package
* file - add new file
* addr - add new remote address:port - ex: google.com:80
* port - add new listening [protocol]:port - ex: 80 or udp:123
* service - add new service
* user - add new user
* group - add new group
* command - add new command
* dns - add new dns
* process - add new process name
* kernel-param - add new kernel-param
* mount - add new mount
* interface - add new network interface
* http - add new network http url
* goss - add new goss file, it will be imported from this one

## Supported output formats
* rspecish **(default)** - Similar to rspec output
* documentation - Verbose test results
* JSON - Detailed test result
* TAP
* JUnit
* nagios - Nagios/Sensu compatible output /w exit code 2 for failures.
* nagios_verbose - nagios output with verbose failure output.

## Community Contributions
* [goss-ansible](https://github.com/indusbox/goss-ansible) - Ansible module for Goss.
* [kitchen-goss](https://github.com/ahelal/kitchen-goss) - A test-kitchen verifier plugin for Goss.
* [goss-fpm-files](https://github.com/deanwilson/unixdaemon-fpm-cookery-recipes) - Might be useful for building goss system packages.
* [molecule](https://github.com/metacloud/molecule) - Automated testing for Ansible roles, with native Goss support.

## Limitations

Currently goss only runs on Linux.

The following tests have limitations.

Package:
  * rpm
  * deb
  * Alpine apk
  * pacman

Service:
  * systemd
  * sysV init
  * OpenRC init
  * Upstart
