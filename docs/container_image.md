# Goss container image

## Dockerfiles

* [latest](https://github.com/goss-org/goss/blob/master/Dockerfile)

## Using the base image

This is a simple alpine image with Goss preinstalled on it.
Can be used as a base image for your projects to allow for easy health checking.

### Mount example

Create the container

```sh
docker run --name goss ghcr.io/goss-org/goss goss
```

Create your container and mount goss

```sh
docker run --rm -it --volumes-from goss --name weby nginx
```

Run goss inside your container

```sh
docker exec weby /goss/goss autoadd nginx
```

### HEALTHCHECK example

```dockerfile
FROM ghcr.io/goss-org/goss:latest

COPY goss/ /goss/
HEALTHCHECK --interval=1s --timeout=6s CMD goss -g /goss/goss.yaml validate

# your stuff..
```

### Startup delay example

```dockerfile
FROM ghcr.io/goss-org/goss:latest

COPY goss/ /goss/

# Alternatively, the -r option can be set
# using the GOSS_RETRY_TIMEOUT env variable
CMD goss -g /goss/goss.yaml validate -r 5m && exec real_comand..
```
