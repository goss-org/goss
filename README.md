# Goss - Quick and Easy server validation
[![Build Status](https://travis-ci.org/aelsabbahy/goss.svg?branch=master)](https://travis-ci.org/aelsabbahy/goss)

## Goss in 45 seconds

**Note:** For an even faster way of doing this, see: [autoadd](https://github.com/aelsabbahy/goss/blob/master/docs/manual.md#autoadd-aa---auto-add-all-matching-resources-to-test-suite)

<a href="https://asciinema.org/a/bxcuduzs3n2zo62rpe1t0s6w8?autoplay=1" target="_blank"><img src="https://cloud.githubusercontent.com/assets/6783261/10236274/b708ff8e-6871-11e5-9d39-70876f5ef8f8.gif" alt="asciicast"></a>

## Introduction

### What is goss?

Goss is a [serverspec](http://serverspec.org/)-like tool for validating a server's configuration. It eases the process of generating tests by assuming the user already has a properly configured machine from which they can derive system state. Once the test suite is generated they can be executed on any other host for the full TDD experience.

### Why use goss?

* Goss is EASY!  - [Goss in 45 seconds](#goss-in-45-seconds)
* Goss is FAST!  - small-medium test suits are near instantaneous, see [benchmarks](https://github.com/aelsabbahy/goss/wiki/Benchmarks)
* Goss is SMALL! - single self-contained ~2MB binary
* Goss is UNIXY! - does one thing and does it well, chainable through pipes

## Installation

```bash
curl -L https://github.com/aelsabbahy/goss/releases/download/v0.1.3/goss-linux-amd64 > /usr/local/bin/goss && chmod +rx /usr/local/bin/goss
```

## Full Documentation

Documentation is available here: https://github.com/aelsabbahy/goss/blob/master/docs/manual.md

## Quick start

### Writing a simple sshd test

An initial set of tests can be derived from the system state by using the [add](https://github.com/aelsabbahy/goss/blob/master/docs/manual.md#add-a---add-system-resource-to-test-suite) or [autoadd](https://github.com/aelsabbahy/goss/blob/master/docs/manual.md#autoadd-aa---auto-add-all-matching-resources-to-test-suite) commands.

Let's write a simple sshd test using autoadd.

```
$ goss autoadd sshd
Adding Group to 'goss.yaml':

sshd:
  exists: true
  gid: 74


Adding Process to 'goss.yaml':

sshd:
  running: true


Adding Service to 'goss.yaml':

sshd:
  enabled: true
  running: true


Adding User to 'goss.yaml':

sshd:
  exists: true
  uid: 74
  gid: 74
  groups:
  - sshd
  home: /var/empty/sshd

```

We can now run our test by using `goss validate`:
```
$ goss validate
..........

Total Duration: 0.016s
Count: 10, Failed: 0

```

As you can see goss tests are extremely fast, we were able to validate our system state in **16ms!**

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
* addr - addr is reachable
* command - command, exit status and outputs
* dns - dns is resolvable
* file - file exists, owner/perm, content
* group - group, uid
* package - package is installed, versions
* port - port is listening, listening ip
* process - process is running
* user - uid, home, etc..

## Supported output formats
* rspecish **(default)** - Similar to rspec output
* documentation - Verbose test results
* JSON - Detailed test result
* TAP
* JUnit
* nagios - Nagios/Sensu compatible output /w exit code 2 for failures.

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
