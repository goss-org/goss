---
render_macros: true

fully_supported: :fontawesome-solid-circle-check:{ .green title="Fully supported/tested" }
community_supported: :fontawesome-solid-circle-check:{ .blue title="Fully working/tested, community support" }
not_automated: :fontawesome-solid-circle-pause:{ .green title="Works but is not covered by automated tests" }
work_partially: :fontawesome-solid-circle-minus:{ .orange title="Works partially / partially tested" }
not_implemented: :fontawesome-solid-circle-xmark:{ .red title="Not implemented / needs implementation" }
broken: :material-image-broken-variant:{ .red title="Currently broken" }
n_a: :fontawesome-regular-circle:{ .grey title="Not applicable for this platform" }
no_data: :fontawesome-regular-circle-question:{ .grey title="Not yet tried, no data" }
---
# Platform feature-parity

macOS and Windows binaries are new and considered alpha-quality.
Some functionality may be missing, some may be broken.
(Enhancements and bug-reports welcome, please see #551)

To clearly signal that, goss emits a log message on every invocation saying so, linking here, then exits with a clear error.

To try out the alpha functionality, you must do one of:

* pass `--use-alpha=1` to the root command (e.g. `goss --use-alpha=1 validate`).
* set an environment variable `GOSS_USE_ALPHA=1`.

The macOS and Windows support is community driven;
there is no commitment to adding features / fixing bugs for those platforms.
[See thread](https://github.com/goss-org/goss/pull/585#discussion_r429968540).

This matrix attempts to track parity across platforms.

## Legend

| Symbol                  | Meaning                                |
|:-----------------------:|----------------------------------------|
| {{ fully_supported }}   | Fully supported/tested                 |
| {{community_supported}} | Full working/tested, community support |
| {{ not_automated }}     | Works but without automated tests      |
| {{ work_partially }}    | Works partially / partially tested     |
| {{ not_implemented }}   | Not implemented / needs implementation |
| {{ n_a }}               | Not applicable for this platform       |
| {{ no_data }}           | Not yet tried, no data                 |

!!! note "About partial support"

    This is ambiguous. Where you see this, check into the test coverage within `integration-tests/goss/{darwin|windows}/{test}.goss.yaml` for more detail.
    It might be that not all features work as on `linux`, it might be that not all features are covered by automated tests.

## Tests/assertions support matrix

| Test                | Option              | Linux                   | macOS                  | Windows                 |
|:--------------------|:--------------------|:-----------------------:|:----------------------:|:-----------------------:|
| **addr**            |                     | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | reachable           | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | local-address       | {{ fully_supported }}   | {{ no_data }}          | {{ work_partially }}    |
|                     | timeout             | {{ fully_supported }}   | {{ not_automated }}    | {{ not_automated }}     |
| **command**         |                     | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | exit-status         | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | stdout              | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | stderr              | {{ fully_supported }}   | {{ not_automated }}    | {{ not_automated }}     |
|                     | timeout             | {{ fully_supported }}   | {{ not_automated }}    | {{ not_automated }}     |
| **dns**             |                     | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | resolvable          | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | addrs               | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | server              | {{ fully_supported }}   | {{ no_data }}          | {{ work_partially }}    |
|                     | timeout             | {{ fully_supported }}   | {{ not_automated }}    | {{ work_partially }}    |
| **file**            |                     | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | exists              | {{ fully_supported }}   | {{ work_partially }}   | {{community_supported}} |
|                     | mode                | {{ fully_supported }}   | {{ work_partially }}   | {{ n_a }}               |
|                     | size                | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | owner               | {{ fully_supported }}   | {{ broken }}           | {{ n_a }}               |
|                     | group               | {{ fully_supported }}   | {{ broken }}           | {{ n_a }}               |
|                     | filetype            | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | contains            | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | md5                 | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | sha256              | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | linked-to           | {{ fully_supported }}   | {{ no_data }}          | {{ no_data }}           |
| **gossfile**        |                     | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
| **group**           |                     | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | exists              | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | gid                 | {{ fully_supported }}   | {{ not_implemented }}  | {{ n_a }}               |
| **http**            |                     | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | status              | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | allow-insecure      | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | no-follow-redirects | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | timeout             | {{ fully_supported }}   | {{ not_automated }}    | {{ work_partially }}    |
|                     | request-headers     | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | headers             | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | body                | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | username            | {{ fully_supported }}   | {{ not_automated }}    | {{ work_partially }}    |
|                     | password            | {{ fully_supported }}   | {{ not_automated }}    | {{ work_partially }}    |
| **interface**       |                     | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | exists              | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | addrs               | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | mtu                 | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
| **kernel-param**    |                     | {{ fully_supported }}   | {{ n_a }}              | {{ n_a }}               |
|                     | value               | {{ fully_supported }}   | {{ n_a }}              | {{ n_a }}               |
| **mount**           |                     | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | exists              | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | opts                | {{ fully_supported }}   | {{ not_implemented }}  | {{ n_a }}               |
|                     | source              | {{ fully_supported }}   | {{ not_implemented }}  | {{ n_a }}               |
|                     | filesystem          | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | usage               | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
| **matching**        |                     | {{ fully_supported }}   | {{ no_data }}          | {{ no_data }}           |
| **package**         |                     | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | installed           | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | versions            | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
| **port**            |                     | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | listening           | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | ip                  | {{ fully_supported }}   |  {{ no_data }}         | {{ no_data }}           |
| **process**         |                     | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
|                     | running             | {{ fully_supported }}   | {{ work_partially }}   | {{ work_partially }}    |
| **service**         |                     | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | enabled             | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | running             | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
| **user**            |                     | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | exists              | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | uid                 | {{ fully_supported }}   | {{ not_implemented }}  | {{ n_a }}               |
|                     | gid                 | {{ fully_supported }}   | {{ not_implemented }}  | {{ n_a }}               |
|                     | groups              | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | home                | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |
|                     | shell               | {{ fully_supported }}   | {{ not_implemented }}  | {{ not_implemented }}   |

## Commands support matrix

| Test       | Linux                  | macOS               | Windows              |
|:-----------|------------------------|---------------------|----------------------|
| `add`      | {{ fully_supported }}  | {{ no_data }}       | {{ work_partially }} |
| `autoadd`  | {{ fully_supported }}  | {{ no_data }}       | {{ no_data }}        |
| `help`     | {{ fully_supported }}  | {{ no_data }}       | {{ work_partially }} |
| `render`   | {{ fully_supported }}  | {{ no_data }}       | {{ no_data }}        |
| `serve`    | {{ fully_supported }}  | {{ not_automated }} | {{ no_data }}        |
| `validate` | {{ fully_supported }}  | {{ not_automated }} | {{ work_partially }} |

### `command` testing notes

Run all of the `darwin`/`windows` integration tests:

```bash
make test-int-validate-darwin-amd64
make test-int-validate-windows-amd64
```

The script finds all goss spec files within `integration-tests` then filters to just ones matching the passed OS-name,
then runs `validate` against them.

### Command: `serve`

This is a special-case test since it requires a persistent process,
then to make the http request, then to tear down the process.

#### macOS `serve`

```bash
make "test-int-serve-darwin-amd64"
```

#### Windows `serve`

```bash
make "test-int-serve-windows-amd64"
```

## Contributing

The current integration test approach is only appropriate for validating `linux` binaries against `linux` OS/arch combinations.

Validating `macOS` and `Windows` binaries requires adding coverage that runs on those platforms within Travis,
but since Travis does not support containerised builds for either platform,
assertions are limited to assert against the state of the CI hosts, where we're relying on that to predictable.

You can find goss-files that are used to populate this matrix within `integration-tests/goss/{darwin|windows}/{test}.goss.yaml`.
Where a feature does note work the same as linux, it is commented.
The intent is to end up with a set of running-and-passing tests.
