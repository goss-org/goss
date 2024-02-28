# Command Line Interface

## Usage

```console
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

!!! note
    Most flags can be set by using environment variables, see `--help` for more info.

## Global options

`--gossfile/-g <gossfile>`
:   The file to use when reading/writing tests. Use `--gossfile -` or `-g -` to read from `STDIN`.

    Valid formats:
    * `yaml` *(default)*
    * `json`

`--vars <varfile>`
:   The file to read variables from when rendering gossfile [templates](gossfile.md#templates).

    Valid formats:

    * `yaml` *(default)*
    * `json`

`--package <type>`
:   The package type to check for.

    Valid options are:

    * `apk`
    * `deb`
    * `pacman`
    * `rpm`

## Commands

Commands are the actions goss can run.
- [add](#add): add a single test for a resource
- [autoadd](#autoadd): automatically add multiple tests for a resource
- [render](#render): renders and outputs the gossfile, importing all included gossfiles
- [serve](#serve): serves the gossfile validation as an HTTP endpoint on a specified address and port,
    so you can use your gossfile as a health report for the host
- [validate](#validate): runs the goss test suite on your server

### `add`

!!! abstract "Add system resource to test suite"
    ```console
    goss add [--exclude-attr <pattern>] <test> [<test>]
    goss a [--exclude-attr <pattern>] <test> [<test>]
    ```

This will add a test for a resource. Non existent resources will add a test to ensure they do not exist on the system.
A sub-command *resource type* has to be provided when running `add`.

`--exclude-attr`
:   Ignore **non-required** attribute(s) matching the provided glob when adding a new resource,
    may be specified multiple times.

!!! example
    ```console
    goss add file /etc/passwd
    goss a user nobody
    goss add --exclude-attr home --exclude-attr shell user nobody
    goss a --exclude-attr '*' user nobody
    ```

#### Resources types

| Type                                       | Description                                                                                                                     |
|--------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------|
| [`addr`](gossfile.md#addr)                 | Verify if a remote `address:port` is reachable                                                                                  |
| [`command`](gossfile.md#command)           | Run a [command](gossfile.md#command) and validate the exit status and/or output                                                 |
| [`dns`](gossfile.md#dns)                   | Resolves a [dns](gossfile.md#dns) name and validates the addresses                                                              |
| [`file`](gossfile.md#file)                 | Validate a [file](gossfile.md#file) existence, permissions, stats (size, etc) and contents                                      |
| [`goss`](gossfile.md#gossfile)             | Includes the contents of another [gossfile](gossfile.md)                                                                        |
| [`group`](gossfile.md#group)               | can validate the existence and values of a [group](gossfile.md#group) on the system                                             |
| [`http`](gossfile.md#http)                 | Validate the HTTP response code, headers, and content of a URI                                                                  |
| [`interface`](gossfile.md#interface)       | Validate the existence and values (es. the addresses) of a network interface                                                    |
| [`kernel-param`](gossfile.md#kernel-param) | Validate kernel parameters (sysctl values)                                                                                      |
| [`mount`](gossfile.md#mount)               | Validate the existence and options relative to a [mount](gossfile.md#mount) point                                               |
| [`package`](gossfile.md#package)           | Validate the status of a [package](gossfile.md#package) using the package manager specified on the commandline with `--package` |
| [`port`](gossfile.md#port)                 | Validate the status of a local [port](gossfile.md#port), for example `80` or `udp:123`                                          |
| [`process`](gossfile.md#process)           | Validate the status of a [process](gossfile.md#process)                                                                         |
| [`service`](gossfile.md#service)           | Validate if a [service](gossfile.md#service) is running and/or enabled at boot                                                  |
| [`user`](gossfile.md#user)                 | Validate the existence and values of a [user](gossfile.md#user) on the system                                                   |

### `autoadd`

!!! abstract "Auto add all matching resources to test suite"
    ```console
    goss autoadd [arguments...]
    goss aa [arguments...]
    ```

Automatically [adds](#add) all **existing** resources matching the provided argument.

Will automatically add the following matching resources:
- `file` - only if argument contains `/`
- `group`
- `package`
- `port`
- `process` - Also adding any ports it's listening to (if run as root)
- `service`
- `user`

Will **NOT** automatically add:
- `addr`
- `command` - for safety
- `dns`
- `http`
- `interface`
- `kernel-param`
- `mount`

!!! example
    ```console
    goss autoadd sshd
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

### `render`

!!! abstract "Render gossfile after importing all referenced gossfiles"
    ```
    goss render
    goss r
    ```

This command allows you to keep your tests separated and render a single, valid, gossfile,
by including them with the `gossfile` directive.

`--debug`
:   This prints the rendered golang template prior to printing the parsed JSON/YAML gossfile.

!!! example
    ```console
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

### `serve`

!!! abstract "Serve a health endpoint"
    ```console
    goss serve [<opts>...]
    goss s [<opts>...]
    ```

`serve` exposes the goss test suite as a health endpoint on your server.
The end-point will return the stest results in the format requested and an http status of 200 or 503.

`serve` will look for a test suite in the same order as [validate](#validate)

`--cache <duration>`, `-c <duration>`
:   Time to cache the results (default: 5s)

`--endpoint <endpoint>`, `-e <endpoint>`
:   Endpoint to expose (default: `/healthz`)

`--format <format>`, `-f <format>`
:   Output format, same as [validate](#validate)

`--listen-addr [ip]:port`, `-l [ip]:port`
:   Address to listen on (default: `:8080`)

`--loglevel <level>`, `-L <level>`
:   Goss logging verbosity level (default: `INFO`).
    Lower levels of tracing include all upper levels traces also (ie. `INFO` include `WARN` and `ERROR`).
    `level` can be one of:
    - `ERROR` - Critical errors that halt goss or significantly affect its functionality, requiring immediate intervention.
    - `WARN` - Non-critical issues that may require attention, such as overwritten keys or deprecated features.
    - `INFO` - General operational messages, useful for tasks where a more structured output is needed (e.g. goss serve).
    - `DEBUG` - Information useful for the goss user to debug.
    - `TRACE` - Detailed internal system activities useful for goss developers to debug.

`--max-concurrent <num>`
:   Max number of tests to run concurrently

!!! example
    ```console
    $ goss serve &
    $ curl http://localhost:8080/healthz
    # JSON endpoint
    $ goss serve --format json &
    $ curl localhost:8080/healthz
    # rspecish output format in response via content negotiation
    goss serve --format json &
    curl -H "Accept: application/vnd.goss-rspecish" localhost:8080/healthz
    ```

The `application/vnd.goss-{output format}` media type can be used in the `Accept` request header
to determine the response's content-type.
You can also `Accept: application/json` to get back `application/json`.

### `validate`

!!! abstract "Validate the system"
    ```console
    goss validate [<opts>...]
    goss v [<opts>...]
    ```

`validate` runs the goss test suite on your server. Prints an rspec-like (by default) output of test results.
Exits with status 0 on success, non-0 otherwise.

`--format <format>`, `-f <format>`
:   Output format. Can be one of:
    - `documentation` - Verbose test results
    - `json` - Detailed test result on a single line (See `pretty` format option)
    - `junit`
    - `nagios` - Nagios/Sensu compatible output /w exit code 2 for failures
    - `rspecish` **(default)** - Similar to rspec output
    - `tap`
    - `prometheus` - Prometheus compatible output.
    - `silent` - No output. Avoids exposing system information (e.g. when serving tests as a healthcheck endpoint)

`--format-options`, `-o`
:   Output format option:
    - `perfdata` - Outputs Nagios "performance data". Applies to `nagios` output
    - `verbose`  - Gives verbose output. Applies to `nagios` and `prometheus` output
    - `pretty`   - Pretty printing for the `json` output
    - `sort`     - Sorts the results

`--loglevel <level>`, `-L <level>`
:   Goss logging verbosity level (default: `INFO`).
    Lower levels of tracing include all upper levels traces also (ie. `INFO` includes `WARN`, `ERROR` and `FATAL` outputs).
    `level` can be one of :
    - `TRACE` - Print details for each check, successful or not and all incoming healthchecks
    - `DEBUG` - Print details of summary response to healthchecks including remote IP address, return code and full body
    - `INFO` - Print summary when all checks run OK
    - `WARN` - Print summary and corresponding checks when encountering some failures
    - `ERROR` - Not used for now (will not print anything)
    - `FATAL` - Not used for now (will not print anything)

`--max-concurrent <num>`
:   Max number of tests to run concurrently

`--color`/`--no-color`
:   Force color or disable color

`--retry-timeout <timeout>`, `-r <timeout>`
:   Retry on failure so long as elapsed + sleep time is less than this

    *default: `0`*

`--sleep <duration>`, `-s <duration>`
:   Time to sleep between retries

    *default: `1s`*

!!! example

    ```console
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
