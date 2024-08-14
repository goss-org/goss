# dgoss

dgoss is a convenience wrapper around goss that aims to bring the simplicity of goss to containers.

## Examples and Tutorials

* [blog tutorial](https://medium.com/@aelsabbahy/tutorial-how-to-test-your-docker-image-in-half-a-second-bbd13e06a4a9) -
Introduction to dgoss tutorial
* [video tutorial](https://youtu.be/PEHz5EnZ-FM) - Same as above, but in video format
* [dgoss-examples](https://github.com/aelsabbahy/dgoss-examples) - Repo containing examples of using dgoss to validate
container images

## Installation

### Linux

Follow the goss [installation instructions](https://github.com/goss-org/goss#installation)

### Mac OSX

Since goss runs on the target container, dgoss can be used on a Mac OSX system by doing the following:

```shell
# Install dgoss
curl -L https://raw.githubusercontent.com/goss-org/goss/master/extras/dgoss/dgoss -o /usr/local/bin/dgoss
chmod +rx /usr/local/bin/dgoss

# Download desired goss version to your preferred location (e.g. v0.4.8)
curl -L https://github.com/goss-org/goss/releases/download/v0.4.8/goss-linux-amd64 -o ~/Downloads/goss-linux-amd64

# Set your GOSS_PATH to the above location
export GOSS_PATH=~/Downloads/goss-linux-amd64

# Set DGOSS_TEMP_DIR to the tmp directory in your home, since /tmp is private on Mac OSX
export DGOSS_TEMP_DIR=~/tmp

# Use dgoss
dgoss edit ...
dgoss run ...
```

## Usage

`dgoss [run|edit] <docker_run_params>`

### Run

Run is used to validate a container.
It expects a `./goss.yaml` file to exist in the directory it was invoked from.
In most cases one can just substitute the runtime command (`docker` or `podman`)
for the dgoss command, for example:

**run:**

`docker run -e JENKINS_OPTS="--httpPort=8080 --httpsPort=-1" -e JAVA_OPTS="-Xmx1048m" jenkins:alpine`

**test:**

`dgoss run -e JENKINS_OPTS="--httpPort=8080 --httpsPort=-1" -e JAVA_OPTS="-Xmx1048m" jenkins:alpine`

`dgoss run` will do the following:

* Run the container with the flags you specified.
* Stream the containers log output into the container as `/goss/docker_output.log`
    * This allows writing tests or waits against the container output
* (optional) Run `goss` with `$GOSS_WAIT_OPTS` if `./goss_wait.yaml` file exists in the current dir
* Run `goss` with `$GOSS_OPTS` using `./goss.yaml`

### Edit

Edit will launch a container, install goss, and drop the user into an interactive shell.
Once the user quits the interactive shell, any `goss.yaml` or `goss_wait.yaml` are copied out into the current directory.
This allows the user to leverage the `goss add|autoadd` commands to write tests as they would on a regular machine.

**Example:**

`dgoss edit -e JENKINS_OPTS="--httpPort=8080 --httpsPort=-1" -e JAVA_OPTS="-Xmx1048m" jenkins:alpine`

### Environment vars and defaults

The following environment variables can be set to change the behavior of dgoss.

#### GOSS_PATH

Location of the goss binary to use. (Default: `$(which goss)`)

#### GOSS_FILE

Name of the goss file to use. (Default: `goss.yaml`)

#### GOSS_OPTS

Options to use for the goss test run. (Default: `--color --format documentation`)

#### GOSS_WAIT_OPTS

Options to use for the goss wait run, when `./goss_wait.yaml` exists. (Default: `-r 30s -s 1s > /dev/null`)

#### GOSS_SLEEP

Time to sleep after running container (and optionally `goss_wait.yaml`) and before running tests. (Default: `0.2`)

#### GOSS_FILES_PATH

Location of the goss yaml files. (Default: `.`)

#### GOSS_ADDITIONAL_COPY_PATH

Colon-seperated list of additional directories to copy to container.

By default dgoss copies `goss.yaml` from the current working directory and
nothing else. You may need other files like scripts and configurations copied
as well. Specify `GOSS_ADDITIONAL_COPY_PATH` similar to `$PATH` as colon seperated
list of directories for each additional directory you'd like to recursively copy.
These will be copied as directories next to `goss.yaml` in the temporary
directory `DGOSS_TEMP_DIR`. (Default: `''`)

#### GOSS_VARS

The name of the variables file relative to `GOSS_FILES_PATH` to copy into the
container and use for valiation (i.e. `dgoss run`) and copy out of the
container when writing tests (i.e. `dgoss edit`). If set, the
`--vars` flag is passed to `goss validate` commands inside the container.
If unset (or empty), the `--vars` flag is omitted, which is the normal behavior.
(Default: `''`).

#### GOSS_FILES_STRATEGY

Strategy used for copying goss files into the container. If set to `'mount'` a volume with goss files is mounted
and log output is streamed into the container as `/goss/docker_output.log` file. Other strategy is `'cp'` which uses
`'docker cp'` command to copy goss files into container. With the `'cp'` strategy you lose the ability to write
tests or waits against the container output. The `'cp'` strategy is required especially when container daemon is not on the
local machine.
(Default `'mount'`)

#### CONTAINER_LOG_OUTPUT

Location of the file that contains tested container logs. Logs are retained only if the variable is set to a non-empty
string. (Default `''`)

#### DGOSS_TEMP_DIR

Location of the temporary directory used by dgoss. (Default `'$(mktemp -d /tmp/tmp.XXXXXXXXXX)'`)

#### CONTAINER_RUNTIME

Container runtime to use - `docker` or `podman`. Defaults to `docker`. Note that `podman` requires a run command to keep
the container running. This defaults to `sleep infinity` in case only an image is passed to `dgoss` commands.
