# dcgoss

dcgoss is a convenience wrapper around goss that aims to bring the simplicity of goss to docker-compose managed containers. It is based on `dgoss`.

## Usage

`dcgoss [run|edit] <docker_run_params>`


### Run

Run is used to validate a docker container defined in `docker-compose.yml` by default. A custom dockerfile can be provided via the `-f` parameter, ie. `./dcgoss -f custom-docker-compose.yaml [run|edit] <service>`. Container configuration is used from the compose file, for example:

**run:**

`docker-compose up db`

**test:**

`dcgoss run db`

`dcgoss run` will do the following:
* Start the container as defined in `docker-compose.yml`
* Stream the containers log output into the container as `/goss/docker_output.log`
  * This allows writing tests or waits against the docker output
* (optional) Run `goss` with `$GOSS_WAIT_OPTS` if `./goss_wait.yaml` file exists in the current dir
* Run `goss` with `$GOSS_OPTS` using `./goss.yaml`


### Edit

Edit will launch a docker container, install goss, and drop the user into an interactive shell. Once the user quits the interactive shell, any `goss.yaml` or `goss_wait.yaml` are copied out into the current directory. This allows the user to leverage the `goss add|autoadd` commands to write tests as they would on a regular machine.

**Example:**

`dcgoss edit db`

### Environment vars and defaults
The following environment variables can be set to change the behavior of dcgoss.

##### DEBUG
Enables debug output of `dcgoss`. (Default: empty)

When running in debug mode, the tmp dir with the container output will not be cleaned up.

**Example:**

`DEBUG=true dcgoss edit db`

##### GOSS_PATH
Location of the goss binary to use. (Default: `$(which goss)`)

##### GOSS_OPTS
Options to use for the goss test run. (Default: `--color --format documentation`)

##### GOSS_WAIT_OPTS
Options to use for the goss wait run, when `./goss_wait.yaml` exists. (Default: `-r 30s -s 1s > /dev/null`)

##### GOSS_SLEEP
Time to sleep after running container (and optionally `goss_wait.yaml`) and before running tests. (Default: `0.2`)

##### GOSS_FILES_PATH
Location of the goss yaml files. (Default: `.`)

**Example:**

`GOSS_FILES_PATH=db dcgoss edit db`

##### GOSS_FILE
Allows to specify a differing name for `goss.yaml`. Useful when the same image is started for different configurations.

**Example:**

`GOSS_FILE=goss_config1.yaml dcgoss run db`

##### GOSS_VARS
The name of the variables file relative to `GOSS_FILES_PATH` to copy into the
docker container and use for valiation (i.e. `dcgoss run`) and copy out of the
docker container when writing tests (i.e. `dcgoss edit`). If set, the
`--vars` flag is passed to `goss validate` commands inside the container.
If unset (or empty), the `--vars` flag is omitted, which is the normal behavior.
(Default: `''`).

##### GOSS_FILES_STRATEGY
Strategy used for copying goss files into the docker container. If set to `'mount'` a volume with goss files is mounted and log output is streamed into the container as `/goss/docker_output.log` file. Other strategy is `'cp'` which uses `'docker cp'` command to copy goss files into docker container. With the `'cp'` strategy you lose the ability to write tests or waits against the docker output. The `'cp'` strategy is required especially when docker daemon is not on the local machine. 
(Default `'mount'`)

## Debugging test runs

When debugging test execution its beneficual to set both `DEBUG=true` and `GOSS_WAIT_OPTS=-r 60s -s 5s` (without the redirect to `/dev/null`).

**Example:**

`DEBUG=true GOSS_FILES_PATH=db GOSS_WAIT_OPTS="-r 60s -s 5s" dcgoss run db`
