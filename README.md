# Goss - Quick and Easy server validation

[![Build Status](https://travis-ci.org/goss-org/goss.svg?branch=master)](https://travis-ci.org/goss-org/goss)
[![Github All Releases](https://img.shields.io/github/downloads/goss-org/goss/total.svg?maxAge=604800)](https://github.com/goss-org/goss/releases)
[![Documentation Status](https://readthedocs.org/projects/goss/badge/)](https://goss.rocks/)
**
[![Blog](https://img.shields.io/badge/follow-blog-brightgreen.svg)](https://medium.com/@aelsabbahy)

<!-- --8<-- [start:intro] -->

## Goss in 45 seconds

<!-- markdownlint-disable line-length no-inline-html-->
<a href="https://asciinema.org/a/4suhr8p42qcn6r7crfzt6cc3e?autoplay=1" target="_blank"><img src="https://cloud.githubusercontent.com/assets/6783261/17330426/ce7ad066-5894-11e6-84ea-29fd4207af58.gif" alt="asciicast"></a>
<!-- markdownlint-enable line-length no-inline-html -->

**Note:** For testing containers see the [dgoss](https://github.com/goss-org/goss/tree/master/extras/dgoss) wrapper.
Also, user submitted wrapper scripts for Kubernetes [kgoss](https://github.com/goss-org/goss/tree/master/extras/kgoss)
and Docker Compose [dcgoss](https://github.com/goss-org/goss/tree/master/extras/dcgoss).

**Note:** For some Container/Kubernetes healthcheck, health endpoint, and
container ordering examples, see my blog post
[here][kubernetes-simplified-health-checks].

## Introduction

### What is Goss?

Goss is a YAML based [serverspec](http://serverspec.org/) alternative tool for validating a server's configuration.
It eases the process of writing tests by allowing the user to generate tests from the current system state.
Once the test suite is written they can be executed, waited-on, or served as a health endpoint.

### Why use Goss?

* Goss is EASY! - [Goss in 45 seconds](#goss-in-45-seconds)
* Goss is FAST! - small-medium test suites are near instantaneous, see [benchmarks](https://github.com/goss-org/goss/wiki/Benchmarks)
* Goss is SMALL! - <10MB single self-contained binary

## Installation

**Note:** For macOS and Windows, see: [platform-feature-parity].

This will install goss and [dgoss](https://github.com/goss-org/goss/tree/master/extras/dgoss).

**Note:** Using `curl | sh` is not recommended for production systems, use manual installation below.

```bash
# Install latest version to /usr/local/bin
curl -fsSL https://goss.rocks/install | sh

# Install v0.3.16 version to ~/bin
curl -fsSL https://goss.rocks/install | GOSS_VER=v0.3.16 GOSS_DST=~/bin sh
```

<!-- --8<-- [end:intro] -->
<!-- --8<-- [start:install] -->

### Manual installation

#### Latest

```bash
curl -L https://github.com/goss-org/goss/releases/latest/download/goss-linux-amd64 -o /usr/local/bin/goss
chmod +rx /usr/local/bin/goss

curl -L https://github.com/goss-org/goss/releases/latest/download/dgoss -o /usr/local/bin/dgoss
# Alternatively, using the latest master
# curl -L https://raw.githubusercontent.com/goss-org/goss/master/extras/dgoss/dgoss -o /usr/local/bin/dgoss
chmod +rx /usr/local/bin/dgoss
```

#### Specific Version

```bash
# See https://github.com/goss-org/goss/releases for release versions
VERSION=v0.3.10
curl -L "https://github.com/goss-org/goss/releases/download/${VERSION}/goss-linux-amd64" -o /usr/local/bin/goss
chmod +rx /usr/local/bin/goss

# (optional) dgoss docker wrapper (use 'master' for latest version)
VERSION=v0.3.10
curl -L "https://github.com/goss-org/goss/releases/download/${VERSION}/dgoss" -o /usr/local/bin/dgoss
chmod +rx /usr/local/bin/dgoss
```

### Build it yourself

```bash
make build
```

<!-- --8<-- [end:install] -->

## Full Documentation

[Full Documentation](https://github.com/goss-org/goss/blob/e73553f9c3065ac297499dafb4f8abef6acb24ad/docs/manual.md)

## Quick start

<!-- --8<-- [start:quickstart] -->

### Writing a simple sshd test

An initial set of tests can be derived from the system state by using the [add](https://goss.rocks/cli/#add)
or [autoadd](https://goss.rocks/cli/#autoadd) commands.

Let's write a simple sshd test using autoadd.

```txt
# Running it as root will allow it to also detect ports
$ sudo goss autoadd sshd
```

Generated `goss.yaml`:

```yaml
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

```console
$ goss validate
...............

Total Duration: 0.021s # <- yeah, it's that fast..
Count: 15, Failed: 0
```

* Edit it to use [templates](https://goss.rocks/gossfile/#templates), and run with a vars file

```console
goss --vars vars.yaml validate
```

* keep running it until the system enters a valid state or we timeout

```console
goss validate --retry-timeout 30s --sleep 1s
```

* serve the tests as a health endpoint

```console
$ goss serve &
$ curl localhost:8080/healthz

# JSON endpoint
$ goss serve --format json &
$ curl localhost:8080/healthz

# rspecish response via content negotiation
$ goss serve --format json &
$ curl -H "Accept: application/vnd.goss-rspecish" localhost:8080/healthz
```

### Manually editing Goss files

Goss files can be manually edited to improve readability and expressiveness of tests.

A [Json draft 7 schema](https://github.com/json-schema-org/json-schema-spec/blob/draft-07/schema.json) available
at <https://goss.rocks/schema.yaml> makes it easier to edit simple goss.yaml files in IDEs,
providing usual coding assistance such as inline documentation, completion and static analysis.
See #793 for screenshots.

For example, to configure the Json schema in JetBrains intellij IDEA,
follow [documented instructions](https://www.jetbrains.com/help/idea/json.html#ws_json_schema_add_custom),
with arguments such as:
* `schema url=https://goss.rocks/schema.yaml`
* `schema version=Json schema version 7`
* `file path pattern=*/goss.yaml`

In addition, Goss files can also be further manually edited (without yet full json support) to use:

* [Patterns](https://goss.rocks/gossfile/#patterns)
* [Advanced Matchers](https://goss.rocks/gossfile/#advanced-matchers)
* [Templates](https://goss.rocks/gossfile/#templates)
* `title` and `meta` (arbitrary data) attributes are persisted when adding other resources with `goss add`

Some examples:

```yaml+jinja
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

  # Loaded from --vars YAML/JSON file
  {{.Vars.package}}:
    installed: true

{{if eq .Env.OS "centos"}}
  # This test is only when $OS environment variable is set to "centos"
  libselinux:
    installed: true
{{end}}
```

Goss.yaml files with templates can still be validated through the Json schema after being rendered
using the `goss render` command. See example below

```console
$ cd docs
$ goss --vars ./vars.yaml render > rendered_goss.yaml
# proceed with json schema validation of rendered_goss.yaml in your favorite IDE
# or in one of the Json schema validator listed in https://json-schema.org/implementations.html
# The following example is for a Linux AMD64 host
$ curl -LO https://github.com/neilpa/yajsv/releases/download/v1.4.1/yajsv.linux.amd64
$ chmod a+x yajsv.linux.amd64
$ sudo mv yajsv.linux.amd64 /usr/sbin/yajsv

$ yajsv -s goss-json-schema.yaml rendered_goss.yaml

rendered_goss.yaml: fail: process.chrome: skip is required
rendered_goss.yaml: fail: service.sshd: skip is required
1 of 1 failed validation
rendered_goss.yaml: fail: process.chrome: skip is required
rendered_goss.yaml: fail: service.sshd: skip is required
```

Full list of available Json schema validators can be found in <https://json-schema.org/implementations.html#validator-command%20line>

<!-- --8<-- [end:quickstart] -->
<!-- --8<-- [start:about] -->

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
* http - add new network http url with proxy support
* goss - add new goss file, it will be imported from this one
* matching - test for matches in supplied content

## Supported output formats

* rspecish - **(default)** Similar to rspec output
* documentation - Verbose test results
* json - JSON, detailed test result
* tap - TAP style
* junit - JUnit style
* nagios - Nagios/Sensu compatible output /w exit code 2 for failures.
* prometheus - Prometheus compatible output.
* silent - No output. Avoids exposing system information (e.g. when serving tests as a healthcheck endpoint).

## Community Contributions

<!-- markdownlint-disable line-length -->
* [goss-ansible](https://github.com/indusbox/goss-ansible) - Ansible module for Goss.
* [degoss](https://github.com/naftulikay/ansible-role-degoss) - Ansible role for installing, running, and removing Goss in a single go.
* [kitchen-goss](https://github.com/ahelal/kitchen-goss) - A test-kitchen verifier plugin for Goss.
* [goss-fpm-files](https://github.com/deanwilson/unixdaemon-fpm-cookery-recipes) - Might be useful for building goss system packages.
* [packer-provisioner-goss](https://github.com/YaleUniversity/packer-provisioner-goss) - A packer plugin to run Goss as a provision step.
* [gossboss](https://github.com/mdb/gossboss) - Collect and view aggregated Goss test results from multiple remote Goss servers.
<!-- markdownlint-enable line-length -->

## Limitations

`goss` works well on Linux, but support on Windows & macOS is alpha. See [platform-feature-parity].

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

[kubernetes-simplified-health-checks]: https://medium.com/@aelsabbahy/docker-1-12-kubernetes-simplified-health-checks-and-container-ordering-with-goss-fa8debbe676c
[platform-feature-parity]: https://goss.rocks/platforms

<!-- --8<-- [end:about] -->
