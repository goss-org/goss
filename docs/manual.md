# Table of Contents

  * [Table of Contents](#table-of-contents)
  * [Usage](#usage)
    * [global options](#global-options)
      * [\-g gossfile](#-g-gossfile)
    * [validate, v \- Validate the system](#validate-v---validate-the-system)
      * [Flags](#flags)
      * [Example:](#example)
    * [serve, s \- Serve a health endpoint](#serve-s---serve-a-health-endpoint)
      * [Flags](#flags)
      * [Example:](#example)
    * [autoadd, aa \- Auto add all matching resources to test suite](#autoadd-aa---auto-add-all-matching-resources-to-test-suite)
      * [Example:](#example-1)
    * [add, a \- Add system resource to test suite](#add-a---add-system-resource-to-test-suite)
      * [Resource types](#resource-types)
      * [Flags](#flags-1)
        * [\-\-exclude\-attr](#--exclude-attr)
      * [Example:](#example-2)
    * [render, r \- Render gossfile after importing all referenced gossfiles](#render-r---render-gossfile-after-importing-all-referenced-gossfiles)
      * [Example:](#example-3)
    * [Available tests](#available-tests)
      * [package](#package)
      * [file](#file)
      * [port](#port)
      * [service](#service)
      * [user](#user)
      * [group](#group)
      * [command](#command)
      * [dns](#dns)
      * [process](#process)
      * [kernel-param](#kernel-param)
      * [mount](#mount)
      * [interface](#interface)
      * [http](#http)
      * [gossfile](#gossfile)
    * [Patterns](#patterns)
    * [Advanced Matchers](#advanced-matchers)


# Usage

```
NAME:
   goss - Quick and Easy server validation

USAGE:
   goss [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
   validate, v	Validate system
   serve, s	Serve a health endpoint
   render, r	render gossfile after imports
   autoadd, aa	automatically add all matching resource to the test suite
   add, a	add a resource to the test suite
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --gossfile, -g "./goss.yaml"	Goss file to read from / write to [$GOSS_FILE]
   --package 			Package type to use [rpm, deb, apk, pacman]
   --help, -h			show help
   --generate-bash-completion
   --version, -v		print the version

```
**Note:** *Most flags can be set by using environment variables, see `--help` for more info.*

## global options
### -g gossfile
The gossfile file to use when reading/writing tests. Use `-g -` to read from STDIN.

Valid formats:
* YAML (default)
* JSON

## validate, v - Validate the system

`validate` runs the goss test suite on your server. Prints an rspec-like (by default) output of test results. Exits with status 0 on success, non-0 otherwise.

### Flags
* --sleep, -s - Time to sleep between retries (default: 1s)
* --retry-timeout, -r - Retry on failure so long as elapsed + sleep time is less than this (default: 0)
* --format, -f (output format)
  * rspecish **(default)** - Similar to rspec output
  * documentation - Verbose test results
  * JSON - Detailed test result
  * TAP
  * JUnit
  * nagios - Nagios/Sensu compatible output /w exit code 2 for failures.
  * nagios_verbose - nagios output with verbose failure output.
* --no-color (disable color)

### Example:

```bash
$ goss validate --format documentation
$ curl -s https://static/or/dynamic/goss.json | goss validate
$ goss render | ssh remote-host 'goss validate'
```

## serve, s - Serve a health endpoint

`serve` exposes the goss test suite as a health endpoint on your server. The end-point will return the stest results in the format requested and an http status of 200 or 503.

`serve` will look for a test suite in the same order as [validate](#validate-v---validate-the-system)

### Flags
* --cache value, -c value - Time to cache the results (default: 5s)
* --listen-addr value, -l value - Address to listen on [ip]:port (default: ":8080")
* --endpoint value, -e value - Endpoint to expose (default: "/healthz")
* --format, -f (output format) - same as [validate](#validate-v---validate-the-system)

### Example:

```bash
$ goss serve &
$ curl http://localhost:8080/healthz

# JSON endpoint
goss serve --format json &
curl localhost:8080/healthz
```

## autoadd, aa - Auto add all matching resources to test suite
Automatically adds all **existing** resources matching the provided argument.

Will automatically add the following matching resources:
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
* kernel-param
* mount
* interface
* http


### Example:
```bash
$ goss autoadd sshd
```

Generates the following `goss.yaml`
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


## add, a - Add system resource to test suite

This will add a test for a resource. Non existent resources will add a test to ensure they do not exist on the system. A sub-command "resource type" has to be provided when running `add`.

### Resource types
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

### Flags
#### --exclude-attr
Ignore **non-required** attribute(s) matching the provided glob when adding a new resource, may be specified multiple times.

### Example:
```bash
$ goss a file /etc/passwd
$ goss a user nobody
$ goss a --exclude-attr home --exclude-attr shell user nobody
$ goss a --exclude-attr '*' user nobody
```

## render, r - Render gossfile after importing all referenced gossfiles
### Example:
```bash

$ cat goss_httpd.yaml
package:
  httpd:
    installed: true
    versions:
    - 2.2.15

$ cat goss.yaml
gossfile:
  goss_httpd.yaml: {}

$ goss -g goss.yaml render
package:
  httpd:
    installed: true
    versions:
    - 2.2.15
```

## Available tests
### package
Validates the state of a package

```yaml
package:
  httpd:
    # required attributes
    installed: true
    # optional attributes
    versions:
    - 2.2.15
```

### file
Validates the state of a file

```yaml
file:
  /etc/passwd:
    # required attributes
    exists: true
    # optional attributes
    mode: "0644"
    size: 2118 # in bytes
    owner: root
    group: root
    filetype: file # file, symlink, directory
    contains: [] # Check file content for these patterns
```

`contains` can be string or [pattern](#patterns)


### port
Validates the state of a port

```yaml
port:
  # {tcp,tcp6,udp,udp6}:port_num
  tcp:22:
    # required attributes
    listening: true
    # optional attributes
    ip: # what IP(s) is it listening on
    - 0.0.0.0
```

### service
Validates the state of a service

```yaml
service:
  sshd:
    # required attributes
    enabled: true
    running: true
```

### user
Validates the state of a user

```yaml
user:
  nfsnobody:
    # required attributes
    exists: true
    # optional attributes
    uid: 65534
    gid: 65534
    groups:
    - nfsnobody
    home: /var/lib/nfs
    shell: /sbin/nologin
```

### group
Validates the state of a group

```yaml
group:
  nfsnobody:
    # required attributes
    exists: true
    # optional attributes
    gid: 65534
```

### command
Validates the exit-status and output of a command

```yaml
command:
  go version:
    # required attributes
    exit-status: 0
    # optional attributes
    stdout:
    - go version go1.6 linux/amd64
    stderr: []
    timeout: 10000 # in milliseconds
```

`stdout` and `stderr` can be string or [pattern](#patterns)

### dns
Validates that the provided address is resolveable and the addrs it resolves to.

```yaml
dns:
  localhost:
    # required attributes
    resolveable: true
    # optional attributes
    addrs:
    - 127.0.0.1
    - ::1
    timeout: 500 # in milliseconds
```
### process
Validates if a process is running

```yaml
process:
  chrome:
    # required attributes
    running: true
```

### kernel-param
Validates kernel param value

```yaml
kernel-param:
  kernel.ostype:
    # required attributes
    value: Linux
```

### mount
Validates mount point attributes

```yaml
mount:
  /home:
    # required attributes
    exists: true
    # optional attributes
    opts:
    - rw
    - relatime
    source: /dev/mapper/fedora-home
    filesystem: xfs
```

### interface
Validates network interface values

```yaml
interface:
  eth0:
    # required attributes
    exists: true
    # optional attributes
    addrs:
    - 172.17.0.2/16
    - fe80::42:acff:fe11:2/64
```

### http
Validates http code and content
*Note:* Goss will automatically follow redirects

```yaml
http:
  https://www.google.com:
    # required attributes
    status: 200
    # optional attributes
    allow-insecure: false
    timeout: 1000
    body: [] # Check http response content for these patterns
```

### gossfile
Import another goss file from this one.
```yaml
gossfile:
  goss_httpd.yaml: {}
```

## Patterns
For the attributes that use patterns (ex. file, command output), each pattern is checked against the attribute string, the type of patterns are:
* "string" - checks if any line contain string.
* "!string" - inverse of above, checks that no line contains string
* "\\!string" - escape sequence, check if any line contains "!string"
* "/regex/" - verifies that line contains regex
* "!/regex/" - inverse of above, checks that no line contains regex

**NOTE:** Pattern attrubutes do not support [Advanced Matchers](#advanced-matchers)

**NOTE:** Regex support is based on golangs regex engine documented [here](https://golang.org/pkg/regexp/syntax/)

```bash
$ cat /tmp/test.txt
found
!alsofound


$ cat goss.yaml
file:
  /tmp/test.txt:
    exists: true
    contains:
    - found
    - /fou.d/
    - "\\!alsofound"
    - "!missing"
    - "!/mis.ing/"

$ goss validate
..

Total Duration: 0.001s
Count: 2, Failed: 0
```

## Advanced Matchers
Goss supports advanced matchers by converting json input to [gomega](https://onsi.github.io/gomega/) matchers. Here are some examples:

Validate that user "nobody" has a uid that is less than 500 and that they are ONLY a member of the "nobody" group.
```yaml
user:
  nobody:
    exists: true
    uid:
      lt: 500
    groups:
      consist-of: [nobody]
```

Matchers can be nested for more complex logic, Ex:
Ensure that we have 3 kernel versions installed and none of them are "4.1.0":
```yaml
package:
  kernel:
    installed: true
    versions:
      and:
        - have-len: 3
        - not:
            contain-element: "4.1.0"
```

For more information see:
* [gomega_test.go](https://github.com/aelsabbahy/goss/blob/master/resource/gomega_test.go) - For a complete set of supported json -> Gomega mapping
* [gomega](https://onsi.github.io/gomega/) - Gomega matchers reference
