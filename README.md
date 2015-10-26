# Goss - Quick and Easy server validation
[![Build Status](https://travis-ci.org/aelsabbahy/goss.svg?branch=master)](https://travis-ci.org/aelsabbahy/goss)

## Goss in 45 seconds

**Note:** For an even faster way of doing this, see: [autoadd](#autoadd-aa---auto-add-all-matching-resources-to-test-suite)

<a href="https://asciinema.org/a/bxcuduzs3n2zo62rpe1t0s6w8?autoplay=1" target="_blank"><img src="https://cloud.githubusercontent.com/assets/6783261/10236274/b708ff8e-6871-11e5-9d39-70876f5ef8f8.gif" alt="asciicast"></a>

## Table of Contents

  * [Introduction](#introduction)
    * [What is goss?](#what-is-goss)
    * [Why use goss?](#why-use-goss)
    * [Limitations](#limitations)
  * [Installation](#installation)
  * [Usage](#usage)
    * [global options](#global-options)
      * [-f gossfile](#-f-gossfile)
    * [validate, v - Validate the system](#validate-v---validate-the-system)
      * [Flags](#flags)
    * [add, a - Add system resource to test suite](#add-a---add-system-resource-to-test-suite)
      * [package - Add a package](#package---add-a-package)
        * [Attributes](#attributes)
      * [file - Add a file](#file---add-a-file)
        * [Attributes](#attributes-1)
      * [port - Add a port](#port---add-a-port)
        * [Attributes](#attributes-2)
      * [service - Add a service](#service---add-a-service)
        * [Attributes](#attributes-3)
      * [user - Add a user](#user---add-a-user)
        * [Attributes](#attributes-4)
      * [group - Add a group](#group---add-a-group)
        * [Attributes](#attributes-5)
      * [command - Add a command](#command---add-a-command)
        * [Attributes](#attributes-6)
      * [dns - Add a dns lookup](#dns---add-a-dns-lookup)
        * [Attributes](#attributes-7)
      * [process - Add a process running check](#process---add-a-process-running-check)
        * [Attributes](#attributes-8)
      * [goss - Add a goss file import](#goss---add-a-goss-file-import)
        * [Attributes](#attributes-9)
    * [render, r - Render gossfile after importing all referenced gossfiles](#render-r---render-gossfile-after-importing-all-referenced-gossfiles)
    * [Patterns](#patterns)


## Introduction

### What is goss?

Goss is a [serverspec](http://serverspec.org/)-like tool for validating a server's configuration. It eases the process of generating tests by assuming the user already has a properly configured machine from which they can derive system state. Once the test suite is generated they can be executed on any other host for the full TDD experience.

### Why use goss?

* Goss is EASY!  - [Goss in 45 seconds](#goss-in-45-seconds)
* Goss is FAST!  - small-medium tests suits run in less than a second
* Goss is SMALL! - single self-contained binary
* Goss is UNIXY! - small, does one thing and does it well, chainable through pipes

### Limitations

Currently goss only runs on Linux.

The following tests have limitations.

Package:
  * rpm
  * deb

Service:
  * systemd
  * sysV init

## Installation

```bash
curl -L https://github.com/aelsabbahy/goss/releases/download/v0.0.7/goss-linux-amd64 > /usr/local/bin/goss && chmod +x /usr/local/bin/goss
```

## Usage
```
NAME:
   goss - Quick and Easy server validation

USAGE:
   goss [global options] command [command options] [arguments...]

VERSION:
   0.0.2

COMMANDS:
   validate, v  Validate system
   render, r    render gossfile after imports
   autoadd, aa  automatically add all matching resource to the test suite
   add, a       add a resource to the test suite
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --gossfile, -g "./goss.json" Goss file to read from / write to [$GOSS_FILE]
   --package                    Package type to use [rpm, deb]
   --help, -h                   show help
   --generate-bash-completion
   --version, -v                print the version

```

### global options
#### -g gossfile
The gossfile file to use when reading/writing tests.

Example (default: ./goss.json):
```bash
$ goss validate
........

Count: 8 failed: 0
```

To run a different file (ex. goss_httpd.json):
```bash
$ goss -g goss_httpd.json validate
......

Count: 6 failed: 0
```

### validate, v - Validate the system

`validate` runs the goss test suite on your server. Prints an rspec-like output of test results. Exists with status 0 on success, non-0 otherwise.

`validate` will look for a test suite in the following order:
* stdin
* -g flag (if provided)
* ./goss.json

Success:
```bash
$ goss validate
..

Count: 2 failed: 0

```

Failure:
```bash
$ goss validate
.F
tcp:22: ip doesn't match, expect: 127.0.0.1 found: 0.0.0.0


Count: 2 failed: 1
$ echo $?
1
```

Pipe examples:
```bash

$ cat goss.json | goss validate
$ goss render | ssh remote-host 'goss validate'
$ curl -s https://static/or/dynamic/goss.json | goss validate
```

#### Flags
* --format (output format)
* --no-color (disable color)

### autoadd, aa - Auto add all matching resources to test suite
automatically adds all **existing** resources matching the provided argument.

```bash
$ goss aa httpd
Adding to './goss.json':

{
    "name": "httpd",
    "installed": true,
    "versions": [
        "2.4.16"
    ]
}

Adding to './goss.json':

{
    "executable": "httpd",
    "running": true
}

Adding to './goss.json':

{
    "port": "tcp6:80",
    "listening": true,
    "ip": "0000:0000:0000:0000:0000:0000:0000:0000"
}

Adding to './goss.json':

{
    "service": "httpd",
    "enabled": true,
    "running": true
}

$ goss aa foobar
# no output
```

Will automatically add the following resources:
* file - only if argument contains "/"
* user
* group
* package
* port
* process - Also adding any ports it's listening to (if run as root)
* service

Will **NOT** automatically add:
* commands - for safety
* dns
* addr


### add, a - Add system resource to test suite

#### package - Add a package
Adds the current state of a package to the goss file.

```bash
$ goss a package httpd
Adding to './goss.json':

{
    "name": "httpd",
    "installed": true,
    "versions": [
        "2.4.10"
    ]
}

$ goss a package foobar
Adding to './goss.json':

{
    "name": "foobar",
    "installed": false
}
```
##### Attributes
* name **(required)**- Package name
* installed **(required)** - Is it installed?
* versions - Checks if defined versions are installed.

#### file - Add a file
```bash
$ goss a file /etc/passwd
Adding to './goss.json':

{
    "path": "/etc/passwd",
    "exists": true,
    "mode": "0644",
    "owner": "root",
    "group": "root",
    "filetype": "file",
    "contains": []
}

$ goss a file /etc/system-release
Adding to './goss.json':

{
    "path": "/etc/system-release",
    "exists": true,
    "mode": "0777",
    "owner": "root",
    "group": "root",
    "linked-to": "fedora-release",
    "filetype": "symlink",
    "contains": []
}
```
##### Attributes
* path **(required)** - file/dir/symlink path
* exits **(required)** - does it exists?
* mode - file mode (ex 0644)
* owner - name of owner
* group - group that ownes the file
* linked-to - symlink target
* filetype - file, symlink, directory
* contains - checks if file contents contains these [patterns](#patterns)

#### port - Add a port
Checks if a port is listening
```bash
$ goss a port 22
Adding to './goss.json':

{
    "port": "tcp:22",
    "listening": true,
    "ip": "0.0.0.0"
}

$ goss a port udp:999
Adding to './goss.json':

{
    "port": "udp:999",
    "listening": false
}

```
##### Attributes
* port **(required)** - network:port_num
* listening **(required)** - is the network:port_num listening?
* ip - what IP is it listening on

#### service - Add a service
Currently only supports init and systemd.

```bash
$ goss a service sshd
Adding to './goss.json':

{
    "service": "sshd",
    "enabled": true,
    "running": true
}

```
##### Attributes
* service **(required)** - name of service
* enabled - will start on startup
* running - is currently running

#### user - Add a user
```bash
$ goss a user nfsnobody
Adding to './goss.json':

{
    "username": "nfsnobody",
    "exists": true,
    "uid": "65534",
    "gid": "65534",
    "groups": [
        "nfsnobody"
    ],
    "home": "/var/lib/nfs"
}

$ goss a user foobar
Adding to './goss.json':

{
    "username": "foobar",
    "exists": false
}
```
##### Attributes
* username **(required)** - name of user
* exists **(required)** - user exists
* uid - uid of user
* gid - gid of user
* groups - Checks if user is a member of the defined groups.
* home - user home directory

#### group - Add a group
```bash
$ goss a group nfsnobody
Adding to './goss.json':

{
    "groupname": "nfsnobody",
    "exists": true,
    "gid": "65534"
}

$ goss a group foobar
Adding to './goss.json':

{
    "groupname": "foobar",
    "exists": false
}
```
##### Attributes
* groupname **(required)** - name of group
* exists **(required)** - does the group exist
* gid - gid of group

#### command - Add a command
records the output and exit status for a command
```bash
$ goss a command go version
Adding to './goss.json':

{
    "command": "go version",
    "exit-status": "0",
    "stdout": [
        "go version go1.5 linux/amd64"
    ],
    "stderr": []
}

$ goss a command lksdjflksad
Adding to './goss.json':

{
    "command": "lksdjflksad",
    "exit-status": "127",
    "stdout": [],
    "stderr": [
        "sh: lksdjflksad: command not found"
    ]
}
```
##### Attributes
* command **(required)** - command to execute
* exit-status - exit status
* stdout - checks if stdout contains these [patterns](#patterns)
* stderr - checks if stderr contains these [patterns](#patterns)

#### dns - Add a dns lookup
Validates that the provided address is resolveable and the addrs it resolves to.
```bash
goss a dns localhost
Adding to './goss.json':

{
    "host": "localhost",
    "resolveable": true,
    "addrs": [
        "127.0.0.1",
        "::1"
    ]
}

Adding to './goss.json':

{
    "host": "foobar",
    "resolveable": false
}
```
##### Attributes
* host **(required)** - hostname to lookup
* resolveable **(required)** - is it resolvable
* addrs - checks if resolved addresses contains these entries

#### process - Add a process running check
Checks if a process by this name is running.
```bash
$ goss a process chrome
Adding to './goss.json':

{
    "executable": "chrome",
    "running": true
}

$ goss a process foobar
Adding to './goss.json':

{
    "executable": "foobar",
    "running": false
}
```
##### Attributes
* executable **(required)** - executable name
* running **(required)** - is it currently running

#### goss - Add a goss file import
Allows you to import another goss file from this one.
```bash
$ goss -g goss_httpd.json a package httpd
Adding to 'goss_httpd.json':

{
    "name": "httpd",
    "installed": true,
    "versions": [
        "2.4.10"
    ]
}

$ goss a goss goss_httpd.json
Adding to './goss.json':

{
    "path": "goss_httpd.json"
}
```
##### Attributes
* path **(required)** - path of goss file

### render, r - Render gossfile after importing all referenced gossfiles
```bash
$ cat goss_httpd.json
{
    "packages": [
        {
            "name": "httpd",
            "installed": true,
            "versions": [
                "2.4.10"
            ]
        }
    ]
}

$ cat goss.json
{
    "gossfiles": [
        {
            "path": "goss_httpd.json"
        }
    ]
}

$ goss -g goss.json render
{
    "packages": [
        {
            "name": "httpd",
            "installed": true,
            "versions": [
                "2.4.10"
            ]
        }
    ]
}
```

### Patterns
For the attributes that use patterns (ex. file, command output), each pattern is checked against the attribute string, the type of patterns are:
* "string" - checks if any line contain string.
* "!string" - inverse of above, checks that no line contains string
* "/regex/" - verifies that line contains regex
* "!/regex/" - inverse of above, checks that no line contains regex

```bash
$ cat /tmp/test.txt
foo
!foo
/foo


$ cat goss.json
{
    "files": [
        {
            "path": "/tmp/test.txt",
            "exists": true,
            "contains": [
                "foo",
                "/fo./",
                "!foo",
                "!/fo./",
                "\\!foo",
                "!lksdajflka",
                "!/lksdajflka/"
            ]
        }
    ]
}

$ goss validate
.F
/tmp/test.txt: contains: patterns not found: [!foo, !/fo./]


Count: 2 failed: 1
```
