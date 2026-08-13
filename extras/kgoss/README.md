# kgoss

kgoss is a wrapper for goss that aims to bring the simplicity of testing
with goss to containers running in pods in Kubernetes.

kgoss is a script which when invoked copies and runs goss (the binary) within a
Linux container. goss itself is only supported on Linux, but since it need only
run in the target container, the kgoss script can be used from any
bash-compatible shell, including Terminal on Mac and git-bash on Windows. On
Windows, [winpty][] is used for interactive connections to the pod under test.

[winpty]: https://github.com/rprichard/winpty

## Install

Installing kgoss requires copying the kgoss file to a directory in your PATH
and copying the goss file to your home folder (or a path set as `GOSS_PATH`),
as follows.

### Manual / UI

You can manually install kgoss and goss by going through the Web UI, getting
the files and putting them in the right path. To get each of them:

* **kgoss**: Run `curl -sSLO
  https://raw.githubusercontent.com/goss-org/goss/master/extras/kgoss/kgoss`.
* **goss**: Download the `goss_<VERSION>_linux_x86_64.tar.gz` archive from
  <https://github.com/goss-org/goss/releases> and extract the binary from it
  with `tar xzf goss_<VERSION>_linux_x86_64.tar.gz goss`. Place it in your
  HOME directory, e.g. `C:\Users\<username>` on Windows; or set the
  environment variable `GOSS_PATH` to its path.

### Automatic / CLI

To install from the command line or automatically, use the following commands.
Set `GOSS_DST` to a directory in your `PATH` env var.

```shell
GOSS_VER=v0.4.10
GOSS_DST=$HOME/bin

curl -L "https://github.com/goss-org/goss/releases/download/$GOSS_VER/dgoss" -o $GOSS_DST/kgoss
chmod a+rx "$GOSS_DST/kgoss"
curl -fsSL https://goss.rocks/install | sh

# If `goss` is not in your path, export a GOSS_PATH variable:
export GOSS_PATH="$GOSS_DST/goss"

# Now you can use kgoss as described below:
# kgoss edit ...
# kgoss run ...
```

## Use

```sh
kgoss [run|edit] -i <image_url> [-p | -c "command to run" | -a "args to pass"] [-d "directory to include"]* [-e "k=v"]*
```

If none of `-p|-c|-a` are specified the container is run with its configured entry point.

`-d` and `-e` can be specified multiple (or zero) times to add additional
directories and env vars.

By default kgoss copies `goss.yaml` from the current working directory and
nothing else. You may need other files like scripts and configurations copied
as well. Specify `-d <path_to_dir>` for each additional directory you'd like
to recursively copy. These will be copied as directories next to `goss.yaml`
in the target container's `GOSS_CONTAINER_PATH`.

To find `goss.yaml` in another directory specify that directory's path in `GOSS_FILES_PATH`.

### Run

The `run` command is used to validate a container. It expects a
`./goss.yaml` file to exist in the directory it was invoked from.

If the file `./goss_wait.yaml` exists in the current directory, goss regularly
checks whether the conditions in the file are met. Only then does goss start the
actual check with the file `./goss.yaml`. This is used, for example, to wait
until a certain port is open before executing the tests.

**Example:**

```sh
kgoss run -e JENKINS_OPTS="--httpPort=8080 --httpsPort=-1" -e JAVA_OPTS="-Xmx1048m" -i jenkins:alpine
```

`kgoss run` will do the following:

* Run the container with the start commands specified by `-c`, `-a`, or `-p`.
* Run `goss` with `$GOSS_WAIT_OPTS` if `./goss_wait.yaml` file exists in the current dir.
* Run `goss` with `$GOSS_OPTS` using `./goss.yaml` from `GOSS_FILES_PATH`.

### Edit

Edit will launch a container, install goss, and drop the user into an
interactive shell. Once the user quits the interactive shell, any `goss.yaml`
or `goss_wait.yaml` are copied out into the current directory. This allows the
user to leverage the `goss add|autoadd` commands to write tests as they would
on a regular machine.

**Example:**

```sh
kgoss edit -e JENKINS_OPTS="--httpPort=8080 --httpsPort=-1" -e JAVA_OPTS="-Xmx1048m" -i jenkins:alpine
```

## Environment variables

The following environment variables effect the behavior of kgoss.

| Variable | Description | Default |
| -------- | ----------- | ------- |
| GOSS\_PATH | Local location of a compatible goss binary to use in container | `$(which goss)` |
| GOSS\_FILES\_PATH | Location of the goss yaml files | `.` |
| GOSS\_KUBECTL\_BIN | Kubenetes client tool to use | `$(which kubectl)` |
| GOSS\_KUBECTL\_OPTS | Options to inject more options such as "--namespace=default" | "" |
| GOSS\_KUBECTL\_TIMEOUT | Kubectl wait timout for the tester pod to reach ready condition | "60s" |
| GOSS\_KUBECTL\_RUN\_OPTS | Options only for kubectl run (e.g. --overrides for serviceAccount) | "" |
| GOSS\_OPTS | Options to use for the goss test run. | `--color --format documentation` |
| GOSS\_WAIT\_OPTS | Options to use for the goss wait run, when `./goss_wait.yaml` exists. | `-r 30s -s 1s > /dev/null` |
| GOSS\_VARS | Variables file relative to `GOSS_FILES_PATH` to copy and use | "" |
| GOSS\_CONTAINER\_PATH | Path within container to put goss binary and YAML files | `/tmp/goss` |
