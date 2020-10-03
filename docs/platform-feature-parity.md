# Platform feature-parity

macOS and Windows binaries are new and considered alpha-quality. Some functionality may be missing, some may be broken. (Enhancements and bug-reports welcome, please see [#551: Multi-OS support](https://github.com/aelsabbahy/goss/issues/551)).

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

| Test                | `linux` | `macOS` | `Windows` |
|:--------------------|---------|---------|-----------|
| **addr**            | x       | wp-pt   | wp-pt     |
| reachable           | x       | wp-pt   | wp-pt     |
| local-address       | x       |         | wp-pt     |
| timeout             | x       | w-nt    | w-nt      |
|                     | x       |         |           |
| **command**         | x       | wp-pt   | wp-pt     |
| exit-status         | x       | wp-pt   | wp-pt     |
| stdout              | x       | wp-pt   | wp-pt     |
| stderr              | x       | w-nt    | w-nt      |
| timeout             | x       | w-nt    | w-nt      |
|                     | x       |         |           |
| **dns**             | x       | wp-pt   | wp-pt     |
| resolvable          | x       | wp-pt   | wp-pt     |
| addrs               | x       | wp-pt   | wp-pt     |
| server              | x       |         | wp-pt     |
| timeout             | x       | w-nt    | wp-pt     |
|                     |         |         |           |
| **file**            | x       | wp-pt   | wp-pt     |
| exists              | x       | wp-pt   | w         |
| mode                | x       | wp-pt   | n/a       |
| size                | x       | wp-pt   | wp-pt     |
| owner               | x       | broken  | n/a       |
| group               | x       | broken  | n/a       |
| filetype            | x       | wp-pt   | wp-pt     |
| contains            | x       | wp-pt   | wp-pt     |
| md5                 | x       | wp-pt   | wp-pt     |
| sha256              | x       | wp-pt   | wp-pt     |
| linked-to           | x       |         |           |
|                     | x       |         |           |
| **gossfile**        | x       | wp-pt   | wp-pt     |
|                     | x       |         |           |
| **group**           | x       | ni      | ni        |
| exists              | x       | ni      | ni        |
| gid                 | x       | ni      | n/a       |
|                     | x       |         |           |
| **http**            | x       | wp-pt   | wp-pt     |
| status              | x       | wp-pt   | wp-pt     |
| allow-insecure      | x       | wp-pt   | wp-pt     |
| no-follow-redirects | x       | wp-pt   | wp-pt     |
| timeout             | x       | w-nt    | wp-pt     |
| request-headers     | x       | wp-pt   | wp-pt     |
| headers             | x       | wp-pt   | wp-pt     |
| body                | x       | wp-pt   | wp-pt     |
| username            | x       | w-nt    | wp-pt     |
| password            | x       | w-nt    | wp-pt     |
|                     |         |         |           |
| **interface**       | x       | ni      | ni        |
| exists              | x       | ni      | ni        |
| addrs               | x       | ni      | ni        |
| mtu                 | x       | ni      | ni        |
|                     | x       |         |           |
| **kernel-param**    | x       | n/a     | n/a       |
| value               | x       | n/a     | n/a       |
|                     | x       |         |           |
| **mount**           | x       | ni      | ni        |
| exists              | x       | ni      | ni        |
| opts                | x       | ni      | n/a       |
| source              | x       | ni      | n/a       |
| filesystem          | x       | ni      | ni        |
| usage               | x       | ni      | ni        |
|                     | x       |         |           |
| **matching**        | x       |         |           |
|                     | x       |         |           |
| **package**         | x       | ni      | ni        |
| installed           | x       | ni      | ni        |
| versions            | x       | ni      | ni        |
|                     | x       |         |           |
| **port**            | x       | ni      | ni        |
| listening           | x       | ni      | ni        |
| ip                  | x       |         |           |
|                     | x       |         |           |
| **process**         | x       | wp-pt   | wp-pt     |
| running             | x       | wp-pt   | wp-pt     |
|                     | x       |         |           |
| **service**         | x       | ni      | ni        |
| enabled             | x       | ni      | ni        |
| running             | x       | ni      | ni        |
|                     | x       |         |           |
| **user**            | x       | ni      | ni        |
| exists              | x       | ni      | ni        |
| uid                 | x       | ni      | n/a       |
| gid                 | x       | ni      | n/a       |
| groups              | x       | ni      | ni        |
| home                | x       | ni      | ni        |
| shell               | x       | ni      | ni        |

## Matrix - `command`s

| Test       | `linux` | `macOS` | `Windows` |
|:-----------|---------|---------|-----------|
| `add`      | x       |         | wp-pt     |
| `autoadd`  | x       |         |           |
| `help`     | x       |         | wp-pt     |
| `render`   | x       |         |           |
| `serve`    | x       | w-nt    |           |
| `validate` | x       | w-nt    | wp-pt     |

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
