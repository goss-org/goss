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

#### Manual / UI

You can manually install kgoss and goss by going through the Web UI, getting
the files and putting them in the right path. To get each of them:

* **kgoss**: Run `curl -sSLO
	https://raw.githubusercontent.com/aelsabbahy/goss/master/extras/kgoss/kgoss`.
* **goss**: Download the `goss-linux-amd64` asset from
  <https://github.com/aelsabbahy/goss/releases> and rename it `goss`. Place it
  in your HOME directory, e.g. C:\\Users\\<username> on Windows; or set the
  environment variable `GOSS_PATH` to its path.

#### Automatic / CLI

To install from the command line or automatically, use the following commands.
[jq][] is required to parse the API response and find the release asset's
download URL.

[jq]: https://stedolan.github.io/jq

First get a GitHub personal access token for accessing the GitHub API from
<https://github.com/settings/tokens>. Input it in the first
line below. Set `dest_dir` to a directory in your `PATH` env var.

```
token=<personal_access_token>
username=$(whoami)
dest_dir=${HOME}/bin

host=raw.githubusercontent.com
repo=aelsabbahy/goss
# for private repos, replace:
# host=github.yourcompany.com
# repo=org-name/goss

## install kgoss
curl -sSL -u "${username}:${token}" -H 'Accept: application/vnd.github.v3.raw' -o "${dest_dir}/kgoss" \
  https://${host}/api/v3/repos/${repo}/contents/extras/kgoss/kgoss
chmod a+rx "${dest_dir}/kgoss"

## install goss
if [[ ! $(which jq) ]]; then echo "jq is required, get from https://stedolan.github.io/jq"; fi
version=v0.3.8
arch=amd64
host=github.com
# for private repos, leave `host` blank or same as above:
# host=github.yourcompany.com
dl_url=$(curl -sSL -u "${username}:${token}" https://${host}/api/v3/repos/${repo}/releases \
  | jq -r ".[] | select (.name == \"${version}\") | .assets[] | select (.name == \"goss-linux-${arch}\") | .url")
curl -sSL -u "${username}:${token}" -H 'Accept: application/octet-stream' -o "${dest_dir}/goss" $dl_url
chmod a+rx "${dest_dir}/goss"

# If `goss` is not in your path, export a GOSS_PATH variable:
export GOSS_PATH=${dest_dir}/goss

# Now you can use kgoss as described below:
# kgoss edit ...
# kgoss run ...
```

## Use

`kgoss [run|edit] -i <image_url> [-p | -c "command to run" | -a "args to pass"] [-d "directory to include"]* [-e "k=v"]*`

If none of `-p|-c|-a` are specified the container is run with its configured entry point.

`-d` and `-e` can be specified multiple (or zero) times to add additional
directories and env vars.

By default kgoss copies `goss.yaml` from the current working directory and
nothing else. You may need other files like scripts and configurations copied
as well. Specify `-d <path_to_dir>` for each additional directory you'd like
to recursively copy. These will be copied as directories next to `goss.yaml`
in the target container's `GOSS_CONTAINER_PATH`.

To find `goss.yaml` in another directory specify that directory's path in `GOSS_FILES_PATH`.

#### Run

The `run` command is used to validate a docker container. It expects a
`./goss.yaml` file to exist in the directory it was invoked from.

**Example:**

`kgoss run -e JENKINS_OPTS="--httpPort=8080 --httpsPort=-1" -e JAVA_OPTS="-Xmx1048m" -i jenkins:alpine`

`kgoss run` will do the following:
* Run the container with the start commands specified by `-c`, `-a`, or `-p`.
* Run `goss` with `$GOSS_WAIT_OPTS` if `./goss_wait.yaml` file exists in the current dir.
* Run `goss` with `$GOSS_OPTS` using `./goss.yaml` from `GOSS_FILES_PATH`.

#### Edit

Edit will launch a docker container, install goss, and drop the user into an
interactive shell. Once the user quits the interactive shell, any `goss.yaml`
or `goss_wait.yaml` are copied out into the current directory. This allows the
user to leverage the `goss add|autoadd` commands to write tests as they would
on a regular machine.

**Example:**

`kgoss edit -e JENKINS_OPTS="--httpPort=8080 --httpsPort=-1" -e JAVA_OPTS="-Xmx1048m" -i jenkins:alpine`

## Environment variables

The following environment variables effect the behavior of kgoss.

Variable | Description | Default
---------|-------------|--------
GOSS\_PATH | Local location of a compatible goss binary to use in container | `$(which goss)`
GOSS\_FILES\_PATH | Location of the goss yaml files | `.`
GOSS\_KUBECTL\_BIN | Kubenetes client tool to use | `$(which kubectl)`
GOSS\_OPTS | Options to use for the goss test run. | `--color --format documentation`
GOSS\_WAIT\_OPTS | Options to use for the goss wait run, when `./goss_wait.yaml` exists. | `-r 30s -s 1s > /dev/null`
GOSS\_VARS | Variables file relative to `GOSS_FILES_PATH` to copy and use | ""
GOSS\_CONTAINER\_PATH | Path within container to put goss binary and YAML files | `/tmp/goss`
