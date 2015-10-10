export GO15VENDOREXPERIMENT=1

exe = github.com/aelsabbahy/goss/cmd/goss
pkgs = $(shell glide novendor)
cmd = goss
TRAVIS_TAG ?= "0.0.0"

.PHONY: all build install test coverage deps release bench test-int lint gen

all: test-all

install:
	go install -v $(exe)

test:
	go test $(pkgs)

lint:
	go tool vet .
	golint $(pkgs) | grep -v 'unexported' || true

bench:
	go test -bench=.

coverage:
	go test -cover $(pkgs)
	#go test -coverprofile=/tmp/coverage.out .
	#go tool cover -func=/tmp/coverage.out
	#go tool cover -html=/tmp/coverage.out -o /tmp/coverage.html
	#xdg-open /tmp/coverage.html

build:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(TRAVIS_TAG)" -o release/$(cmd)-linux-amd64 $(exe)

release: build
	#upx release/*

test-int: build
	cd integration-tests/ && ./test.sh

test-all: test lint test-int

deps:
	glide up

gen:
	go generate -tags genny $(pkgs)
