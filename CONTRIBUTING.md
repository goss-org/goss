# Contributing

## Development setup

You need `make`, `glide` and `golint` installed:

```bash
#!/usr/bin/env bash
os="linux" # or darwin, or windows
curl -L https://github.com/Masterminds/glide/releases/download/0.10.2/glide-0.10.2-${os}-amd64.zip > glide.zip
unzip glide.zip
export PATH="$PATH:$PWD/${os}-amd64"
go get -u github.com/golang/lint/golint
```

Then:

```bash
#!/usr/bin/env bash
make deps
make
```
