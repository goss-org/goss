# goss manual

## Table of Contents

* [Table of Contents](#table-of-contents)
* [Usage](#usage)
  * [global options](#global-options)
    * [\-g gossfile](#-g-gossfile)
  * [commands](#commands)
    * [add, a \- Add system resource to test suite](#add-a---add-system-resource-to-test-suite)
    * [autoadd, aa \- Auto add all matching resources to test suite](#autoadd-aa---auto-add-all-matching-resources-to-test-suite)
    * [render, r \- Render gossfile after importing all referenced gossfiles](#render-r---render-gossfile-after-importing-all-referenced-gossfiles)
    * [serve, s \- Serve a health endpoint](#serve-s---serve-a-health-endpoint)
    * [validate, v \- Validate the system](#validate-v---validate-the-system)
* [Goss test creation](#goss-test-creation)
* [Important note about goss file format](#important-note-about-goss-file-format)
* [Available tests](#available-tests)
  * [addr](#addr)
  * [command](#command)
  * [dns](#dns)
  * [file](#file)
  * [gossfile](#gossfile)
  * [group](#group)
  * [http](#http)
  * [interface](#interface)
  * [kernel-param](#kernel-param)
  * [mount](#mount)
  * [matching](#matching)
  * [package](#package)
  * [port](#port)
  * [process](#process)
  * [service](#service)
  * [user](#user)
* [Matchers](#matchers)
* [Advanced Matchers](#advanced-matchers)
* [Templates](#templates)

## Usage

```
NAME:
   goss - Quick and Easy server validation

USAGE:
   goss [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     validate, v  Validate system
     serve, s     Serve a health endpoint
     render, r    render gossfile after imports
     autoadd, aa  automatically add all matching resource to the test suite
     add, a       add a resource to the test suite
     help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --gossfile value, -g value  Goss file to read from / write to (default: "./goss.yaml") [$GOSS_FILE]
   --vars value                json/yaml file containing variables for template [$GOSS_VARS]
   --vars-inline value         json/yaml string containing variables for template (overwrites vars) [$GOSS_VARS_INLINE]
   --package value             Package type to use [rpm, deb, apk, pacman]
   --help, -h                  show help
   --version, -v               print the version
```
**Note:** *Most flags can be set by using environment variables, see `--help` for more info.*


## global options
### -g gossfile
The file to use when reading/writing tests. Use `-g -` to read from `STDIN`.

Valid formats:
* **YAML** (default)
* **JSON**

### --vars
The file to read variables from when rendering gossfile [templates](#templates).

Valid formats:
* **YAML** (default)
* **JSON**

### --package <type>
The package type to check for.

Valid options are:
* `apk`
* `deb`
* `pacman`
* `rpm`


## commands
Commands are the actions goss can run.

* [add](#add-a---add-system-resource-to-test-suite): add a single test for a resource
* [autoadd](#autoadd-aa---auto-add-all-matching-resources-to-test-suite): automatically add multiple tests for a resource
* [render](#render-r---render-gossfile-after-importing-all-referenced-gossfiles): renders and outputs the gossfile, importing all included gossfiles
* [serve](#serve-s---serve-a-health-endpoint): serves the gossfile validation as an HTTP endpoint on a specified address and port, so you can use your gossfile as a health repor for the host
* [validate](#validate-v---validate-the-system): runs the goss test suite on your server


### add, a - Add system resource to test suite

This will add a test for a resource. Non existent resources will add a test to ensure they do not exist on the system. A sub-command *resource type* has to be provided when running `add`.

#### Resource types
* `addr` - can verify if a remote `address:port` is reachable, see [addr](#addr)
* `command` - can run a [command](#command) and validate the exit status and/or output
* `dns` - resolves a [dns](#dns) name and validates the addresses
* `file` - can validate a [file](#file) existence, permissions, stats (size, etc) and contents
* `goss` - allows you to include the contents of another [gossfile](#gossfile)
* `group` - can validate the existence and values of a [group](#group) on the system
* `http` - can validate the HTTP response code, headers, and content of a URI, see [http](#http)
* `interface` - can validate the existence and values (es. the addresses) of a network interface, see [interface](#interface)
* `kernel-param` - can validate kernel parameters (sysctl values), see [kernel-param](#kernel-param)
* `mount` - can validate the existence and options relative to a [mount](#mount) point
* `package` - can validate the status of a [package](#package) using the package manager specified on the commandline with `--package`
* `port` - can validate the status of a local [port](#port), for example `80` or `udp:123`
* `process` - can validate the status of a [process](#process)
* `service` - can validate if a [service](#service) is running and/or enabled at boot
* `user` - can validate the existence and values of a [user](#user) on the system

#### Flags
##### --exclude-attr
Ignore **non-required** attribute(s) matching the provided glob when adding a new resource, may be specified multiple times.

#### Example:
```bash
$ goss a file /etc/passwd
$ goss a user nobody
$ goss a --exclude-attr home --exclude-attr shell user nobody
$ goss a --exclude-attr '*' user nobody
```


### autoadd, aa - Auto add all matching resources to test suite
Automatically [adds](#add-a---add-system-resource-to-test-suite) all **existing** resources matching the provided argument.

Will automatically add the following matching resources:
* `file` - only if argument contains `/`
* `group`
* `package`
* `port`
* `process` - Also adding any ports it's listening to (if run as root)
* `service`
* `user`

Will **NOT** automatically add:
* `addr`
* `command` - for safety
* `dns`
* `http`
* `interface`
* `kernel-param`
* `mount`

#### Example:
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


### render, r - Render gossfile after importing all referenced gossfiles
This command allows you to keep your tests separated and render a single, valid, gossfile, by including them with the `gossfile` directive.

#### Flags
##### --debug
This prints the rendered golang template prior to printing the parsed JSON/YAML gossfile.

#### Example:
```bash

$ cat goss_httpd_package.yaml
package:
  httpd:
    installed: true
    versions:
    - 2.2.15

$ cat goss_httpd_service.yaml
service:
  httpd:
    enabled: true
    running: true

$ cat goss_nginx_service-NO.yaml
service:
  nginx:
    enabled: false
    running: false

$ cat goss.yaml
gossfile:
  goss_httpd_package.yaml: {}
  goss_httpd_service.yaml: {}
  goss_nginx_service-NO.yaml: {}

$ goss -g goss.yaml render
package:
  httpd:
    installed: true
    versions:
    - 2.2.15
service:
  httpd:
    enabled: true
    running: true
  nginx:
    enabled: false
    running: false
```


### serve, s - Serve a health endpoint

`serve` exposes the goss test suite as a health endpoint on your server. The end-point will return the stest results in the format requested and an http status of 200 or 503.

`serve` will look for a test suite in the same order as [validate](#validate-v---validate-the-system)

#### Flags
* `--cache <value>`, `-c <value>` - Time to cache the results (default: 5s)
* `--endpoint <value>`, `-e <value>` - Endpoint to expose (default: `/healthz`)
* `--format`, `-f` - output format, same as [validate](#validate-v---validate-the-system)
* `--listen-addr [ip]:port`, `-l [ip]:port` - Address to listen on (default: `:8080`)
* `--max-concurrent` - Max number of tests to run concurrently

#### Example:

```bash
$ goss serve &
$ curl http://localhost:8080/healthz

# JSON endpoint
$ goss serve --format json &
$ curl localhost:8080/healthz
```


### validate, v - Validate the system

`validate` runs the goss test suite on your server. Prints an rspec-like (by default) output of test results. Exits with status 0 on success, non-0 otherwise.

#### Flags
* `--format`, `-f` (output format)
  * `documentation` - Verbose test results
  * `json` - Detailed test result on a single line (See `pretty` format option)
  * `junit`
  * `nagios` - Nagios/Sensu compatible output /w exit code 2 for failures
  * `rspecish` **(default)** - Similar to rspec output
  * `tap`
  * `silent` - No output. Avoids exposing system information (e.g. when serving tests as a healthcheck endpoint)
* `--format-options`, `-o` (output format option)
  * `perfdata` - Outputs Nagios "performance data". Applies to `nagios` output
  * `verbose` - Gives verbose output. Applies to `nagios` output
  * `pretty` - Pretty printing for the `json` output
* `--max-concurrent` - Max number of tests to run concurrently
* `--no-color` - Disable color
* `--color` - Force enable color
* `--retry-timeout`, `-r` - Retry on failure so long as elapsed + sleep time is less than this (default: 0)
* `--sleep`, `-s` - Time to sleep between retries (default: 1s)

#### Examples:

```bash
$ goss validate --format documentation
File: /etc/hosts: exists: matches expectation: [true]
DNS: localhost: resolvable: matches expectation: [true]
[...]
Total Duration: 0.002s
Count: 10, Failed: 2, Skipped: 0

$ curl -s https://static/or/dynamic/goss.json | goss validate
...F.F
[...]
Total Duration: 0.002s
Count: 6, Failed: 2, Skipped: 0

$ goss render | ssh remote-host 'goss -g - validate'
......

Total Duration: 0.002s
Count: 6, Failed: 0, Skipped: 0

$ goss validate --format nagios -o verbose -o perfdata
GOSS CRITICAL - Count: 76, Failed: 1, Skipped: 0, Duration: 1.009s|total=76 failed=1 skipped=0 duration=1.009s
Fail 1 - DNS: localhost: addrs: doesn't match, expect: [["127.0.0.1","::1"]] found: [["127.0.0.1"]]
$ echo $?
2
```

## Goss test creation
Goss tests can be created by using either of following methods.
1. goss autoadd <resource to test>
2. goss add <resource to test>
3. manually create YAML/JSON test file by hand.

To customize the parameters generated by `goss add` and `goss autoadd` YAML file you need to manually edit it.

`goss add package nginx` will generate below YAML
```
package:
  nginx:
    installed: true
    versions:
    - 1.17.8
```
To test uninstall scenario you would need to manually edit it and set it as below.
```
package:
  nginx:
    installed: false
```

## Important note about goss file format
It is important to note that both YAML and JSON are formats that describe a nested data structure.

**WRONG way to write a goss file**

```yaml
file:
  /etc/httpd/conf/httpd.conf:
    exists: true

service:
  httpd:
    enabled: true
    running: true

file:
  /var/www/html:
    filetype: directory
    exists: true
```

If you try to validate this file, it will **only** run the second `file` test:

```bash
# goss validate --format documentation
File: /var/www/html: exists: matches expectation: [true]
File: /var/www/html: filetype: matches expectation: ["directory"]
Service: httpd: enabled: matches expectation: [true]
Service: httpd: running: matches expectation: [true]

Total Duration: 0.014s
Count: 8, Failed: 0, Skipped: 0
```

As you can see, the first `file` check has not been run because the second `file` entry *overwrites* the previous one.

You need to make sure all the entries of the same type are under the same declaration.

**This is the CORRECT way to write a goss file**

```yaml
file:
  /etc/httpd/conf/httpd.conf:
    exists: true
  /var/www/html:
    filetype: directory
    exists: true

service:
  httpd:
    enabled: true
    running: true
```

Running validate with this configuration will correctly check both files:

```bash
# goss validate --format documentation
File: /var/www/html: exists: matches expectation: [true]
File: /var/www/html: filetype: matches expectation: ["directory"]
File: /etc/httpd/conf/httpd.conf: exists: matches expectation: [true]
Service: httpd: enabled: matches expectation: [true]
Service: httpd: running: matches expectation: [true]

Total Duration: 0.014s
Count: 10, Failed: 0, Skipped: 0
```

Please note that using the `goss add` and `goss autoadd` command will create a valid file, but if you're writing your files by hand you'll save a lot of time by taking this in consideration.

If you want to keep your tests in separate files, the best way to obtain a single, valid, file is to create a main goss file that includes the others with the [gossfile](#gossfile) directive and then [render](#render-r---render-gossfile-after-importing-all-referenced-gossfiles) it.



## Available tests

* [addr](#addr)
* [command](#command)
* [dns](#dns)
* [file](#file)
* [gossfile](#gossfile)
* [group](#group)
* [http](#http)
* [interface](#interface)
* [kernel-param](#kernel-param)
* [matching](#matching)
* [mount](#mount)
* [package](#package)
* [port](#port)
* [process](#process)
* [service](#service)
* [user](#user)


### addr
Validates if a remote `address:port` are accessible.

```yaml
addr:
  tcp://ip-address-or-domain-name:80:
    reachable: true
    timeout: 500
    # optional attributes
    local-address: 127.0.0.1
```


### command
Validates the exit-status and output of a command

```yaml
command:
  version:
    # required attributes
    exit-status: 0
    # defaults to hash key
    exec: "go version"
    # optional attributes
    stdout:
    - go version go1.6 linux/amd64
    stderr: []
    timeout: 10000 # in milliseconds
    skip: false
```

`stdout` and `stderr` can be a string or [pattern](#patterns)

The `exec` attribute is the command to run; this defaults to the name of
the hash for backwards compatibility

### dns
Validates that the provided address is resolvable and the addrs it resolves to.

```yaml
dns:
  localhost:
    # required attributes
    resolvable: true
    # optional attributes
    addrs:
    - 127.0.0.1
    - ::1
    server: 8.8.8.8 # Also supports server:port
    timeout: 500 # in milliseconds (Only used when server attribute is provided)
```

With the server attribute set, it is possible to validate the following types of DNS record:

- A
- AAAA
- CAA
- CNAME
- MX
- NS
- PTR
- SRV
- TXT

To validate specific DNS address types, prepend the hostname with the type and a colon, a few examples:

```yaml
dns:
  # Validate a CNAME record
  CNAME:c.dnstest.io:
    resolvable: true
    server: 208.67.222.222
    addrs:
    - "a.dnstest.io."

  # Validate a PTR record
  PTR:8.8.8.8:
    resolvable: true
    server: 8.8.8.8
    addrs:
    - "dns.google."

  # Validate and SRV record
  SRV:_https._tcp.dnstest.io:
    resolvable: true
    server: 208.67.222.222
    addrs:
    - "0 5 443 a.dnstest.io."
    - "10 10 443 b.dnstest.io."
```

Please note that if you want `localhost` to **only** resolve `127.0.0.1` you'll need to use [Advanced Matchers](#advanced-matchers)

```yaml
dns:
  localhost:
    resolvable: true
    addrs:
      consist-of: [127.0.0.1]
    timeout: 500 # in milliseconds
```

### file
Validates the state of a file, directory, or symbolic link

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
    md5: 7c9bb14b3bf178e82c00c2a4398c93cd # md5 checksum of file
    # A stronger checksum alternative to md5 (recommended)
    sha256: 7f78ce27859049f725936f7b52c6e25d774012947d915e7b394402cfceb70c4c
  /etc/alternatives/mta:
    # required attributes
    exists: true
    # optional attributes
    filetype: symlink # file, symlink, directory
    linked-to: /usr/sbin/sendmail.sendmail
    skip: false
```

`contains` can be a string or a [pattern](#patterns)


### gossfile
Import other gossfiles from this one. This is the best way to maintain a large number of tests, and/or create profiles. See [render](#render-r---render-gossfile-after-importing-all-referenced-gossfiles) for more examples. Glob patterns can be also be used to specify matching gossfiles.

```yaml
gossfile:
  goss_httpd.yaml: {}
  /etc/goss.d/*.yaml: {}
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
    skip: false
```


### http
Validates HTTP response status code and content.

```yaml
http:
  https://www.google.com:
    # required attributes
    status: 200
    # optional attributes
    allow-insecure: false
    no-follow-redirects: false # Setting this to true will NOT follow redirects
    timeout: 1000
    request-headers: # Set request header values
       - "Content-Type: text/html"
    headers: [] # Check http response headers for these patterns (e.g. "Content-Type: text/html")
    body: [] # Check http response content for these patterns
    username: "" # username for basic auth
    password: "" # password for basic auth
    skip: false
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
    mtu: 1500
```


### kernel-param
Validates kernel param (sysctl) value.

```yaml
kernel-param:
  kernel.ostype:
    # required attributes
    value: Linux
```

To see the full list of current values, run `sysctl -a`.


### mount
Validates mount point attributes.

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
    usage: #% of blocks used in this mountpoint
      lt: 95
```

### matching
Validates specified content against a matcher. Best used with [Templates](#templates).

#### With [Templates](#templates):
Let's say we have a `data.json` file that gets generated as part of some testing pipeline:

```json
{
  "instance_count": 14,
  "failures": 3,
  "status": "FAIL"
}
```

This could then be passed into goss: `goss --vars data.json validate`

And then validated against:

```yaml
matching:
  check_instance_count: # Make sure there is at least one instance
    content: {{ .Vars.instance_count }}
    matches:
      gt: 0

  check_failure_count_from_all_instance: # expect no failures
    content: {{ .Vars.failures }}
    matches: 0

  check_status:
    content: {{ .Vars.status }}
    matches:
      - not: FAIL
```

#### Without [Templates](#templates):
```yaml
matching:
  has_substr: # friendly test name
    content: some string
    matches:
      match-regexp: some str
  has_2:
    content:
      - 2
    matches:
      contain-element: 2
  has_foo_bar_and_baz:
    content:
      foo: bar
      baz: bing
    matches:
      and:
        - have-key: baz
```

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
    skip: false
```

**NOTE:** this check uses the `--package <format>` parameter passed on the command line.


### port
Validates the state of a local port.

**Note:** Goss might consider your port to be listening on `tcp6` rather than `tcp`, try running `goss add port ..` to see how goss detects it. ([explanation](https://github.com/aelsabbahy/goss/issues/149))

```yaml
port:
  # {tcp,tcp6,udp,udp6}:port_num
  tcp:22:
    # required attributes
    listening: true
    # optional attributes
    ip: # what IP(s) is it listening on
    - 0.0.0.0
    skip: false
```


### process
Validates if a process is running.

```yaml
process:
  chrome:
    # required attributes
    running: true
    skip: false
```

**NOTE:** This check is inspecting the name of the binary, not the name of the process. For example, a process with the name `nginx: master process /usr/sbin/nginx` would be checked with the process `nginx`. To discover the binary of a pid run `ps -p <PID> -o comm`.

### service
Validates the state of a service.

```yaml
service:
  sshd:
    # required attributes
    enabled: true
    running: true
    skip: false
```

**NOTE:** this will **not** automatically check if the process is alive, it will check the status from `systemd`/`upstart`/`init`.


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
    skip: false
```

**NOTE:** This check is inspecting the contents of local passwd file `/etc/passwd`, this does not validate remote users (e.g. LDAP).


## Matchers

### Default Matchers

Default matchers are determined by the attribute value received from the system.

#### Bool, Strings, Integers

Bool, Strings and integers are compared using equality, for example:

```yaml
matching:
  basic_string:
    content: 'foo'
    matches: 'foo'

user:
  nfsnobody:
    exists: true
    uid: 65534
```

#### Arrays

Arrays are treated as a `contains-elements`(fixme, link) by default, this validates that the expected test is a subset of the returned system state.

```yaml
matching:
  basic_array:
    content:
      - 'group1'
      - 'group2'
      - 'group3'
    matches:
      - 'group1'
      - 'group2'

  # This fails, since the returned result and it's no longer a subset
  basic_array_failing:
    content:
      - 'group1'
      - 'group2'
      - 'group3'
    matches:
      - 'group1'
      - 'group2'
      - 'group2' # this 2nd group2 is not in the returned content
```

#### io.Readers

This is the most magical matcher for goss. It remains a default for historic and performance reasons. Some attributes return an io.Reader that is read line by line (ex. file content, command, http body). This allows goss to validate large files/content efficiently.


Each pattern is checked against the attribute output, the type of patterns are:

* `"foo"` - checks if any line contains `foo`
* `"!foo"` - inverse of above, checks that no line contains `foo`
* `"\\!foo"` - escape sequence, check if any line contains `!string`
* `"/[Rr]egex/"` - verifies that line matches regex
* `"!/[Rr]egex/"` - inverse of above, checks that no line matches regex

**NOTE:** Regex support is based on Golang's regex engine documented [here](https://golang.org/pkg/regexp/syntax/)

**NOTE:** You will **need** the double backslash (`\\`) escape for Regex special entities, for example `\\s` for blank spaces.

Example:

```yaml
file:
  /tmp/test.txt:
    exists: true
    contains:
    - "foo"
    - "!bar"
    - "/[Gg]oss/"
```

The above is roughly similar to:

```yaml
file:
  /tmp/test.txt:
    exists: true
    contains:
      and:
        - contain-element: "foo"
        - not: {contain-element: "bar"}
        - contain-element: {match-regexp: "[Gg]oss"}

```

### Transforms

If the system state type and the expected type don't match, goss will attempt to transform the system state type before matching it.

For example, kernel-param attribute returns a string, however, it can be tested using numeric comparisons:

Example kernel-param test:
```yaml
kernel-param:
  net.core.somaxconn:
      value: "128"
```

Kernel-param (failing) test with transform:
```yaml
kernel-param:
  net.core.somaxconn:
      value: {gt: 200}
```

When a transformed test fails, it will detail the transformers used, the `-o include_raw` option can be used to include the raw, untransformed attribute value:
```
$ goss v
F

Failures/Skipped:

KernelParam: net.core.somaxconn: value:
Expected
    128
to be >
    200
the transform chain was
    [{"to-numeric":{}}]

Total Duration: 0.001s
Count: 1, Failed: 1, Skipped: 0


$ goss v -o include_raw
F

Failures/Skipped:

KernelParam: net.core.somaxconn: value:
Expected
    128
to be >
    200
the transform chain was
    [{"to-numeric":{}}]
the raw value was
    "128"

Total Duration: 0.001s
Count: 1, Failed: 1, Skipped: 0

```

### Advanced Matchers

Goss supports advanced matchers by converting YAML input to [gomega](https://onsi.github.io/gomega/) matchers.

#### String Matchers

These will convert the system attribute to a string prior to matching.

* `'55'` - Checks that the numeric is "55" when converted to string
* `have-prefix: pre` - Checks if string starts with "pre"
* `have-suffix: suf` - Checks if string ends with "suf"
* `match-regexp: '.*'` - Checks if string matches regexp
* `contain-substring: '2'` - Checks if string contains "2"

Example:
```yaml
matching:
  example:
    content: 42
    matches:
      and:
        - '42'
        - have-prefix: '4'
        - have-suffix: '2'
        - match-regexp: '\d{2}'
        - contain-substring: '2'
```

#### Numeric matchers

These will convert the system attribute to a numeric prior to matching.

* `42` - If the expected type is a number
* `gt, ge, lt, le` - Greater than, greater than or equal, less than, etc..

Example:
```yaml
matching:
  example:
    content: "42"
    matches:
      and:
        - 42
        - 42.0
        - gt: 40
        - lt: 45
```

#### Array matchers

These will convert the system attribute to an array prior to matching. Strings are split on "\n"


* `contain-element: matcher` - Checks if the array contains an element that passes the matcher
* `contain-elements: [matcher, ...]` - checks if the array is a superset of the provided matchers
* `[matcher, ...]` - same as above
* `equal: [value, ...]` - Checks if the array is exactly equal to provided array
* `consist-of: [matcher, ...]` - Checks if the array consists of the provided matchers (order does not matter)

Example:
```yaml
matching:
  example:
    content: [foo, bar, moo]
    matches:
      and:
        - contain-elements: [foo, bar]
        - [foo, bar] # same as above
        - equal: [foo, bar, moo] # order matters, exact match
        - consist-of: [foo, have-prefix: m, bar] # order doesn't matter, can use matchers
        - contain-element:
            have-prefix: b
```

#### Misc matchers

These matchers don't really fall into any of the above categories, or span multiple categories.

* `equal` - Useful when needing to override a default matcher
* `have-len: 3` - Checks if the array/string/map has length of 3
* `have-key: "foo"` - Checks if key exists in map, useful with `gjson`
* `not: matcher` - Checks that a matcher does not match
* `and: [matcher, ..]` - Checks that all matchers match
* `or: [matcher, ..]` - Checks that any matchers match
* `semver-constraint: ">1.0.0 <2.0.0"` - Checks that all versions match [semver](https://github.com/blang/semver#ranges) constraint (or range) syntax fixme (custom range)
    * when system returns a string it is converted into a one element array and matched
* `gjson: {gjson_path: matcher, gjson_path2: matcher2}` - Checks extracted [gjson](https://gjson.dev/) passes the matcher

See the following for examples: [link..]fixme

## Templates

Goss test files can leverage golang's [text/template](https://golang.org/pkg/text/template/) to allow for dynamic or conditional tests.

Available variables:
* `{{.Env}}`  - Containing environment variables
* `{{.Vars}}` - Containing the values defined in [--vars](#global-options) file

Available functions beyond text/template [built-in functions](https://golang.org/pkg/text/template/#hdr-Functions):
* `mkSlice "ARG1" "ARG2"` - Retuns a slice of all the arguments. See examples below for usage.
* `getEnv "var" ["default"]` - A more forgiving env var lookup. If key is missing either "" or default (if provided) is returned.
* `readFile "fileName"` - Reads file content into a string, trims whitespace. Useful when a file contains a token.
  * **NOTE:** Goss will error out during during the parsing phase if the file does not exist, no tests will be executed.
* `regexMatch "(some)?reg[eE]xp"` - Tests the piped input against the regular expression argument.
* `toLower` - Changes piped input to lowercase
* `toUpper` - Changes piped input to UPPERCASE

**NOTE:** gossfiles containing text/template `{{}}` controls will no longer work with `goss add/autoadd`. One way to get around this is to split your template and static goss files and use [gossfile](#gossfile) to import.

### Examples

Using [puppetlabs/facter](https://github.com/puppetlabs/facter) or [chef/ohai](https://github.com/chef/ohai) as external tools to provide vars.
```bash
$ goss --vars <(ohai) validate
$ goss --vars <(facter -j) validate
```

Using `mkSlice` to define a loop locally.
```yaml
file:
{{- range mkSlice "/etc/passwd" "/etc/group"}}
  {{.}}:
    exists: true
    mode: "0644"
    owner: root
    group: root
    filetype: file
{{end}}
```

Using Env variables and a vars file:

**vars.yaml:**
```yaml
centos:
  packages:
    kernel:
      - "4.9.11-centos"
      - "4.9.11-centos2"
debian:
  packages:
    kernel:
      - "4.9.11-debian"
      - "4.9.11-debian2"
users:
  - user1
  - user2
```

**goss.yaml:**
```yaml
package:
# Looping over a variables defined in a vars.yaml using $OS environment variable as a lookup key
{{range $name, $vers := index .Vars .Env.OS "packages"}}
  {{$name}}:
    installed: true
    versions:
    {{range $vers}}
      - {{.}}
    {{end}}
{{end}}

# This test is only when the OS environment variable matches the pattern
{{if .Env.OS | regexMatch "[Cc]ent(OS|os)"}}
  libselinux:
    installed: true
{{end}}

# Loop over users
user:
{{range .Vars.users}}
  {{.}}:
    exists: true
    groups:
    - {{.}}
    home: /home/{{.}}
    shell: /bin/bash
{{end}}


package:
{{if eq .Env.OS "centos"}}
  # This test is only when $OS environment variable is set to "centos"
  libselinux:
    installed: true
{{end}}
```

Rendered results:
```bash
# To validate:
$ OS=centos goss --vars vars.yaml validate
# To render:
$ OS=centos goss --vars vars.yaml render
# To render with debugging enabled:
$ OS=centos goss --vars vars.yaml render --debug
```
