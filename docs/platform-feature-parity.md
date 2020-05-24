# Platform feature-parity

macOS and Windows binaries are new and considered alpha-quality. Some functionality may be missing, some may be broken. (Enhancements and bug-reports welcome, please see [#551: Multi-OS support](https://github.com/aelsabbahy/goss/issues/551))

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
| **addr**            | x       |         | wp-pt     |
| reachable           | x       |         | wp-pt     |
| local-address       | x       |         | wp-pt     |
| timeout             | x       |         | wp-pt     |
|                     | x       |         |           |
| **command**         | x       |         | wp-pt     |
| exit-status         | x       |         | wp-pt     |
| stdout              | x       |         | wp-pt     |
| stderr              | x       |         | wp-pt     |
| timeout             | x       |         | wp-pt     |
|                     | x       |         |           |
| **dns**             | x       |         | wp-pt     |
| resolvable          | x       |         | wp-pt     |
| addrs               | x       |         | wp-pt     |
| server              | x       |         | wp-pt     |
| timeout             | x       |         | wp-pt     |
|                     |         |         |           |
| **file**            | x       |         | wp-pt     |
| exists              | x       |         | w         |
| mode                | x       |         | n/a       |
| size                | x       |         | wp-pt     |
| owner               | x       |         | n/a       |
| group               | x       |         | n/a       |
| filetype            | x       |         | wp-pt     |
| contains            | x       |         | wp-pt     |
| md5                 | x       |         | wp-pt     |
| sha256              | x       |         | wp-pt     |
| linked-to           | x       |         |           |
|                     | x       |         |           |
| **gossfile**        | x       |         | wp-pt     |
|                     | x       |         |           |
| **group**           | x       |         | ni        |
| exists              | x       |         | ni        |
| gid                 | x       |         | n/a       |
|                     | x       |         |           |
| **http**            | x       |         | wp-pt     |
| status              | x       |         | wp-pt     |
| allow-insecure      | x       |         | wp-pt     |
| no-follow-redirects | x       |         | wp-pt     |
| timeout             | x       |         | wp-pt     |
| request-headers     | x       |         | wp-pt     |
| headers             | x       |         | wp-pt     |
| body                | x       |         | wp-pt     |
| username            | x       |         | wp-pt     |
| password            | x       |         | wp-pt     |
|                     |         |         |           |
| **interface**       | x       |         | ni        |
| exists              | x       |         | ni        |
| addrs               | x       |         | ni        |
| mtu                 | x       |         | ni        |
|                     | x       |         |           |
| **kernel-param**    | x       |         | n/a       |
| value               | x       |         | n/a       |
|                     | x       |         |           |
| **mount**           | x       |         | ni        |
| exists              | x       |         | ni        |
| opts                | x       |         | n/a       |
| source              | x       |         | n/a       |
| filesystem          | x       |         | ni        |
| usage               | x       |         | ni        |
|                     | x       |         |           |
| **matching**        | x       |         |           |
|                     | x       |         |           |
| **package**         | x       |         | ni        |
| installed           | x       |         | ni        |
| versions            | x       |         | ni        |
|                     | x       |         |           |
| **port**            | x       |         | ni        |
| listening           | x       |         | ni        |
| ip                  | x       |         |           |
|                     | x       |         |           |
| **process**         | x       |         | wp-pt     |
| running             | x       |         | wp-pt     |
|                     | x       |         |           |
| **service**         | x       |         | ni        |
| enabled             | x       |         | ni        |
| running             | x       |         | ni        |
|                     | x       |         |           |
| **user**            | x       |         | ni        |
| exists              | x       |         | ni        |
| uid                 | x       |         | n/a       |
| gid                 | x       |         | n/a       |
| groups              | x       |         | ni        |
| home                | x       |         | ni        |
| shell               | x       |         | ni        |

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

### Command: `add`

#### Windows `add`

```powershell
.\release\goss-alpha-windows-amd64.exe add command 'echo hello'
exec: "sh": executable file not found in %PATH%
```

### Command: `autoadd`

Not yet tested.

### Command: `help`

Not yet tested.

### Command: `render`

Not yet tested.

### Command: `serve`

macOS:

```bash
make build
trap 'killall goss-alpha-darwin-amd64' EXIT
release/goss-alpha-darwin-amd64 -g integration-tests/goss/goss-serve.yaml serve &
curl http://localhost:9100/healthz | grep 'Count: 2, Failed: 0, Skipped: 0'
```

### Command: `validate`

macOS:

```bash
make build
release/goss-alpha-darwin-amd64 -g integration-tests/goss/goss-serve.yaml validate | grep 'Count: 2, Failed: 0, Skipped: 0'
```
