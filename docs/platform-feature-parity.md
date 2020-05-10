# Platform feature-parity

macOS and Windows binaries are new and considered alpha-quality. Some functionality may be missing, some may be broken.

This matrix attempts to track parity across platforms.

Legend:

* `√` - implemented, passing in CI
* an issue number - refer to that issue for details

## Matrix

| Feature | `linux` implementation + unit-tests | `linux` integration-tests | `macOS` implementation + unit-tests | `macOS` integration-tests | `Windows` implementation + unit-tests | `Windows` integration-tests |
|:---------------|:--------|:-----------------:|:----------------:|:-----------------:|:----------------:|:-------------------:|:------------------:|
| Assertion: `addr` | √ | √ | | | | |
| Assertion: `command` | √ | √ | | | | |
| Assertion: `dns` | √ | √ | | | | |
| Assertion: `file` | √ | √ | | | | |
| Assertion: `gossfile` | √ | √ | | | | |
| Assertion: `group` | √ | √ | | | | |
| Assertion: `http` | √ | √ | | | | |
| Assertion: `interface` | √ | √ | | | | |
| Assertion: `kernel-param` | √ | √ | | | | |
| Assertion: `mount` | √ | √ | | | | |
| Assertion: `matching` | √ | √ | | | | |
| Assertion: `package` | √ | √ | | | | |
| Assertion: `port` | √ | √ | | | | |
| Assertion: `process` | √ | √ | | | | |
| Assertion: `service` | √ | √ | | | | |
| Assertion: `user` | √ | √ | | | | |
| Command: `add` | √ | √ | | | | |
| Command: `autoadd` | √ | √ | | | | |
| Command: `help` | √ | √ | | | | |
| Command: `render` | √ | √ | | | | |
| Command: `serve` | √ | √ | | | | |
| Command: `validate` | √ | √ | | | | |
