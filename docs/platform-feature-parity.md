# Platform feature-parity

macOS, Windows and FreeBSD binaries are new and considered alpha-quality. Some functionality may be missing, some may be broken. (Enhancements and bug-reports welcome, please see [#551: Multi-OS support](https://github.com/aelsabbahy/goss/issues/551)).

To clearly signal that, goss emits a log message on every invocation saying so, linking here, then exits with a clear error.

To try out the alpha functionality, you must do one of
* pass `--use-alpha=1` to the root command - e.g. `goss --use-alpha=1 validate`.
* set an environment variable `GOSS_USE_ALPHA=1`.

The macOS and Windows support is community driven; there is no commitment to adding features / fixing bugs for those platforms. [See thread](https://github.com/aelsabbahy/goss/pull/585#discussion_r429968540).

This matrix attempts to track parity across platforms.

## How to use this doc

### Legend

* `x`    - full supported/tested
* `w`    - full working/tested, community support
* `wp-pt` - works partially / partially tested
  * this is ambiguous; where you see this, check into the test coverage within `integration-tests/goss/{darwin|windows}/{test}.goss.yaml` for more detail. It might be that not all features work as on `linux`, it might be that not all features are covered by automated tests.
* `w-nt` - works no automated tests
* `ni`   - not implemented / needs implementation
* `n/a`  - Not applicable for this platform
* `#585` - Link to bug(s) for implementation, fixing
* ` `    - (blank) - not yet tried, no data

### Contributing

The current integration test approach is only appropriate for validating `linux` binaries against `linux` OS/arch combinations.

Validating `macOS` and `Windows` binaries requires adding coverage that runs on those platforms within Travis, but since Travis does not support containerised builds for either platform, assertions are limited to assert against the state of the CI hosts, where we're relying on that to predictable.

You can find goss-files that are used to populate this matrix within `integration-tests/goss/{darwin|windows}/{test}.goss.yaml`. Where a feature does note work the same as linux, it is commented. The intent is to end up with a set of running-and-passing tests.

## Matrix - tests/assertions

| Test                | `linux` | `macOS` | `Windows` | `FreeBSD` |
|---------------------|---------|---------|-----------|-----------|
| **addr**            | x       | wp-pt   | wp-pt     | wp-pt     |
| reachable           | x       | wp-pt   | wp-pt     | w         |
| local-address       | x       |         | wp-pt     |           |
| timeout             | x       | w-nt    | w-nt      | w         |
|                     | x       |         |           |           |
| **command**         | x       | wp-pt   | wp-pt     | w         |
| exit-status         | x       | wp-pt   | wp-pt     | w         |
| stdout              | x       | wp-pt   | wp-pt     | w         |
| stderr              | x       | w-nt    | w-nt      | w         |
| timeout             | x       | w-nt    | w-nt      | w         |
|                     | x       |         |           |           |
| **dns**             | x       | wp-pt   | wp-pt     | w         |
| resolvable          | x       | wp-pt   | wp-pt     | w         |
| addrs               | x       | wp-pt   | wp-pt     | w         |
| server              | x       |         | wp-pt     | w         |
| timeout             | x       | w-nt    | wp-pt     | w         |
|                     |         |         |           |           |
| **file**            | x       | wp-pt   | wp-pt     | wp-pt     |
| exists              | x       | wp-pt   | w         | w         |
| mode                | x       | wp-pt   | n/a       | w         |
| size                | x       | wp-pt   | wp-pt     | w         |
| owner               | x       | broken  | n/a       | ni        |
| group               | x       | broken  | n/a       | ni        |
| filetype            | x       | wp-pt   | wp-pt     | w         |
| contains            | x       | wp-pt   | wp-pt     |           |
| md5                 | x       | wp-pt   | wp-pt     | w         |
| sha256              | x       | wp-pt   | wp-pt     | w         |
| linked-to           | x       |         |           |           |
|                     | x       |         |           |           |
| **gossfile**        | x       | wp-pt   | wp-pt     | w         |
|                     | x       |         |           |           |
| **group**           | x       | ni      | ni        | w         |
| exists              | x       | ni      | ni        | w         |
| gid                 | x       | ni      | n/a       | w         |
|                     | x       |         |           |           |
| **http**            | x       | wp-pt   | wp-pt     | w         |
| status              | x       | wp-pt   | wp-pt     | w         |
| allow-insecure      | x       | wp-pt   | wp-pt     | w         |
| no-follow-redirects | x       | wp-pt   | wp-pt     | w         |
| timeout             | x       | w-nt    | wp-pt     | w         |
| request-headers     | x       | wp-pt   | wp-pt     | w         |
| headers             | x       | wp-pt   | wp-pt     | w         |
| body                | x       | wp-pt   | wp-pt     | w         |
| username            | x       | w-nt    | wp-pt     | w         |
| password            | x       | w-nt    | wp-pt     | w         |
|                     |         |         |           |           |
| **interface**       | x       | ni      | ni        | w         |
| exists              | x       | ni      | ni        | w         |
| addrs               | x       | ni      | ni        | w         |
| mtu                 | x       | ni      | ni        | w         |
|                     | x       |         |           |           |
| **kernel-param**    | x       | n/a     | n/a       | n/a       |
| value               | x       | n/a     | n/a       | n/a       |
|                     | x       |         |           |           |
| **mount**           | x       | ni      | ni        | n/a       |
| exists              | x       | ni      | ni        | n/a       |
| opts                | x       | ni      | n/a       | n/a       |
| source              | x       | ni      | n/a       | n/a       |
| filesystem          | x       | ni      | ni        | n/a       |
| usage               | x       | ni      | ni        | n/a       |
|                     | x       |         |           |           |
| **matching**        | x       |         |           | w         |
|                     | x       |         |           |           |
| **package**         | x       | ni      | ni        | w         |
| installed           | x       | ni      | ni        | w         |
| versions            | x       | ni      | ni        | w         |
|                     | x       |         |           |           |
| **port**            | x       | ni      | ni        | n/a       |
| listening           | x       | ni      | ni        | n/a       |
| ip                  | x       |         |           | n/a       |
|                     | x       |         |           |           |
| **process**         | x       | wp-pt   | wp-pt     | w         |
| running             | x       | wp-pt   | wp-pt     | w         |
|                     | x       |         |           |           |
| **service**         | x       | ni      | ni        | w         |
| enabled             | x       | ni      | ni        | w         |
| running             | x       | ni      | ni        | w         |
|                     | x       |         |           |           |
| **user**            | x       | ni      | ni        | w         |
| exists              | x       | ni      | ni        | w         |
| uid                 | x       | ni      | n/a       | w         |
| gid                 | x       | ni      | n/a       | w         |
| groups              | x       | ni      | ni        | w         |
| home                | x       | ni      | ni        | w         |
| shell               | x       | ni      | ni        | w         |

## Matrix - `command`s

| Test       | `linux` | `macOS` | `Windows` | `FreeBSD` |
|------------|---------|---------|-----------|-----------|
| `add`      | x       |         | wp-pt     | w         |
| `autoadd`  | x       |         |           | wp-pt     |
| `help`     | x       |         | wp-pt     | w         |
| `render`   | x       |         |           | w         |
| `serve`    | x       | w-nt    |           | w         |
| `validate` | x       | w-nt    | wp-pt     | w         |

### `command` testing notes

Run all of the `darwin`/`windows` integration tests:

```bash
make alpha-test-alpha-darwin-amd64
make alpha-test-alpha-windows-amd64
```

The script finds all goss spec files within `integration-tests` then filters to just ones matching the passed OS-name, then runs `validate` against them.

### Command: `serve`

This is a special-case test since it requires a persistent process, then to make the http request, then to tear down the process.

#### macOS `serve`

```bash
make "test-serve-alpha-darwin-amd64"
```

#### Windows `serve`

```bash
make "test-serve-alpha-windows-amd64"
```
