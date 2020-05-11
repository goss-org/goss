# Platform feature-parity

macOS and Windows binaries are new and considered alpha-quality. Some functionality may be missing, some may be broken. (Enhancements and bug-reports welcome, please see [#551: Multi-OS support](https://github.com/aelsabbahy/goss/issues/551))

This matrix attempts to track parity across platforms.

Legend:

* `f` - full support
* `p` - partial support (with page-link to heading underneath table, with details)
* `n` - no current support (with page-link to heading underneath table, with details)

## Matrix

| Feature                   | `linux` |        `macOS`         |        `Windows`         |
|:--------------------------|:-------:|:----------------------:|:------------------------:|
| Assertion: `addr`         |   `f`   |                        |                          |
| Assertion: `command`      |   `f`   | [`p`](#command-macos)  |                          |
| Assertion: `dns`          |   `f`   |                        |                          |
| Assertion: `file`         |   `f`   |   [`p`](#file-macos)   |   [`p`](#file-windows)   |
| Assertion: `gossfile`     |   `f`   |                        |                          |
| Assertion: `group`        |   `f`   |                        |                          |
| Assertion: `http`         |   `f`   |                        |                          |
| Assertion: `interface`    |   `f`   |                        |                          |
| Assertion: `kernel-param` |   `f`   |                        |                          |
| Assertion: `mount`        |   `f`   |                        |                          |
| Assertion: `matching`     |   `f`   |                        |                          |
| Assertion: `package`      |   `f`   |                        |                          |
| Assertion: `port`         |   `f`   |                        |                          |
| Assertion: `process`      |   `f`   |                        |                          |
| Assertion: `service`      |   `f`   |                        |                          |
| Assertion: `user`         |   `f`   |                        |                          |
| Command: `add`            |   `f`   |                        |                          |
| Command: `autoadd`        |   `f`   |                        |                          |
| Command: `help`           |   `f`   |                        |                          |
| Command: `render`         |   `f`   |                        |                          |
| Command: `serve`          |   `f`   |  [`p`](#serve-macos)   |                          |
| Command: `validate`       |   `f`   | [`p`](#validate-macos) | [`p`](#validate-windows) |

## Details

Please keep this section sorted for ease of navigation.

Template (so that page-links are regular, and people with markdown aware editors can fold headings).

```md
### Attribute|Command `ref`

#### `ref`: platform

{details here}
```

### Attribute: `addr`

Not yet tested.

#### `addr`: macOS

#### `addr`: Windows

### Attribute: `command`

#### `command`: macOS

Manually tested via [`serve`](#serve-macos).

#### `command`: Windows

### Attribute: `dns`

Not yet tested.

#### `dns`: macOS

#### `dns`: Windows

### Attribute: `file`

#### `file`: macOS

`owner` and `group` return `##` rather than the actual owner and group.

#### `file`: Windows

`owner` and `group` are not applicable on Windows.

### Attribute: `gossfile`

Not yet tested.

#### `gossfile`: macOS

#### `gossfile`: Windows

### Attribute: `group`

Not yet tested.

#### `group`: macOS

#### `group`: Windows

### Attribute: `http`

Not yet tested.

#### `http`: macOS

#### `http`: Windows

### Attribute: `interface`

Not yet tested.

#### `interface`: macOS

#### `interface`: Windows

### Attribute: `kernel-param`

Not yet tested.

#### `kernel-param`: macOS

#### `kernel-param`: Windows

### Attribute: `mount`

Not yet tested.

#### `mount`: macOS

#### `mount`: Windows

### Attribute: `matching`

Not yet tested.

#### `matching`: macOS

#### `matching`: Windows

### Attribute: `package`

Not yet tested.

#### `package`: macOS

#### `package`: Windows

### Attribute: `port`

Not yet tested.

#### `port`: macOS

#### `port`: Windows

### Attribute: `process`

Not yet tested.

#### `process`: macOS

#### `process`: Windows

### Attribute: `service`

Not yet tested.

#### `service`: macOS

#### `service`: Windows

### Attribute: `user`

Not yet tested.

#### `user`: macOS

#### `user`: Windows

### Command: `add`

Not yet tested.

#### `add`: macOS

#### `add`: Windows

### Command: `autoadd`

Not yet tested.

#### `autoadd`: macOS

#### `autoadd`: Windows

### Command: `help`

Not yet tested.

#### `help`: macOS

#### `help`: Windows

### Command: `render`

Not yet tested.

#### `render`: macOS

#### `render`: Windows

### Command: `serve`

Not yet tested.

#### `serve`: macOS

Manually tested.

```bash
make build
trap 'killall goss-darwin-amd64' EXIT
release/goss-darwin-amd64 -g integration-tests/goss/goss-serve.yaml serve &
curl http://localhost:9100/healthz | grep 'Count: 2, Failed: 0, Skipped: 0'
```

#### `serve`: Windows

### Command: `validate`

#### `validate`: macOS

```bash
make build
release/goss-darwin-amd64 -g integration-tests/goss/goss-serve.yaml validate | grep 'Count: 2, Failed: 0, Skipped: 0'
```

#### `validate`: Windows

Manually tested `goss validate`; success.
