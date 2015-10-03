export GO15VENDOREXPERIMENT=1

exe = github.com/aelsabbahy/goss/cmd/goss
cmd = goss
TRAVIS_TAG ?= "0.0.0"

.PHONY: all build install test coverage deps release bench test-int lint

all: test-all

install: deps
	go install -v $(exe)

test: deps
	go test $(glide novendor)

lint: deps
	go tool vet .
	golint $(glide novendor) | grep -v 'unexported' || true

bench: deps
	go test -bench=.

coverage: deps
	go test -cover $(glide novendor)
	#go test -coverprofile=/tmp/coverage.out .
	#go tool cover -func=/tmp/coverage.out
	#go tool cover -html=/tmp/coverage.out -o /tmp/coverage.html
	#xdg-open /tmp/coverage.html

build: deps
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(TRAVIS_TAG)" -o release/$(cmd)-linux-amd64 $(exe)

release: build
	upx release/*

test-int: build
	cd integration-tests/ && ./test.sh

test-all: test lint test-int

deps:
	glide up
