# Platform feature-parity

macOS and Windows binaries are new and considered alpha-quality. Some functionality may be missing, some may be broken. (Enhancements and bug-reports welcome, please see [#551: Multi-OS support](https://github.com/aelsabbahy/goss/issues/551))

This matrix attempts to track parity across platforms.

## How to use this doc

### Legend

* `x`    - full supported/tested
* `w`    - full working/tested, community support
* `w-nt` - works no automated tests
* `n/a`  - Not applicable for this platform
* `#585` - Link to bug(s) for implementation, fixing
* ` `    - (blank) - not yet tried, no data

### Contributing

The current integration test approach is only appropriate for validating `linux` binaries against `linux` OS/arch combinations.

Validating `macOS` and `Windows` binaries requires adding coverage that runs on those platforms within Travis, but since Travis does not support containerised builds for either platform, assertions are limited to assert against the state of the CI hosts, where we're relying on that to predictable.

## Matrix - tests/assertions

| Test                | `linux` | `macOS` | `Windows` |
|:--------------------|---------|---------|-----------|
| **addr**            | x       |         |           |
| reachable           | x       |         |           |
| local-address       | x       |         |           |
| timeout             | x       |         |           |
|                     | x       |         |           |
| **command**         | x       |         |           |
| exit-status         | x       |         |           |
| stdout              | x       |         |           |
| stderr              | x       |         |           |
| timeout             | x       |         |           |
|                     | x       |         |           |
| **dns**             | x       |         |           |
| resolvable          | x       |         |           |
| addrs               | x       |         |           |
| server              | x       |         |           |
| timeout             | x       |         |           |
| **file**            | x       |         |           |
| exists              | x       | w-nt    | w         |
| mode                | x       |         | n/a       |
| size                | x       |         |           |
| owner               | x       | #585    | n/a       |
| group               | x       | #585    | n/a       |
| filetype            | x       |         |           |
| contains            | x       | w-nt    | w-nt      |
| md5                 | x       | w-nt    | w-nt      |
| sha256              | x       | w-nt    | w-nt      |
| linked-to           | x       | w-nt    | w-nt      |
|                     | x       |         |           |
| **gossfile**        | x       |         |           |
|                     | x       |         |           |
| **group**           | x       |         |           |
| exists              | x       |         |           |
| gid                 | x       |         |           |
|                     | x       |         |           |
| **http**            | x       |         |           |
| status              | x       |         |           |
| allow-insecure      | x       |         |           |
| no-follow-redirects | x       |         |           |
| timeout             | x       |         |           |
| request-headers     | x       |         |           |
| headers             | x       |         |           |
| body                | x       |         |           |
| username            | x       |         |           |
| password            | x       |         |           |
| **interface**       | x       |         |           |
| exists              | x       |         |           |
| addrs               | x       |         |           |
| mtu                 | x       |         |           |
|                     | x       |         |           |
| **kernel-param**    | x       |         |           |
| value               | x       |         |           |
|                     | x       |         |           |
| **mount**           | x       |         |           |
| exists              | x       |         |           |
| opts                | x       |         |           |
| source              | x       |         |           |
| filesystem          | x       |         |           |
| usage               | x       |         |           |
|                     | x       |         |           |
| **matching**        | x       |         |           |
|                     | x       |         |           |
| **package**         | x       |         |           |
| installed           | x       |         |           |
| versions            | x       |         |           |
|                     | x       |         |           |
| **port**            | x       |         |           |
| listening           | x       |         |           |
| ip                  | x       |         |           |
|                     | x       |         |           |
| **process**         | x       |         |           |
| running             | x       |         |           |
|                     | x       |         |           |
| **service**         | x       |         |           |
| enabled             | x       |         |           |
| running             | x       |         |           |
|                     | x       |         |           |
| **user**            | x       |         |           |
| exists              | x       |         |           |
| uid                 | x       |         |           |
| gid                 | x       |         |           |
| groups              | x       |         |           |
| home                | x       |         |           |
| shell               | x       |         |           |

## Matrix - `command`s

| Test       | `linux` | `macOS` | `Windows` |
|:-----------|---------|---------|-----------|
| `add`      | x       |         |           |
| `autoadd`  | x       |         |           |
| `help`     | x       |         |           |
| `render`   | x       |         |           |
| `serve`    | x       | w-nt    |           |
| `validate` | x       | w-nt    | w-nt      |

### `command` testing notes

### Command: `add`

Not yet tested.

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
