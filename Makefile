export GO15VENDOREXPERIMENT=1

exe = github.com/aelsabbahy/goss/cmd/goss
pkgs = $(shell glide novendor)
cmd = goss
TRAVIS_TAG ?= "0.0.0"
GO_FILES = $(shell find . \( -path ./vendor -o -name '_test.go' \) -prune -o -name '*.go' -print)

.PHONY: all build install test coverage deps release bench test-int lint gen centos7 wheezy precise alpine3 arch

all: test-all

install: release/goss-linux-amd64
	$(info INFO: Starting build $@)
	cp release/$(cmd)-linux-amd64 $(GOPATH)/bin/goss

test:
	$(info INFO: Starting build $@)
	go test $(pkgs)

lint:
	$(info INFO: Starting build $@)
	#go tool vet .
	golint $(pkgs) | grep -v 'unexported' || true

bench:
	$(info INFO: Starting build $@)
	go test -bench=.

coverage:
	$(info INFO: Starting build $@)
	go test -cover $(pkgs)
	#go test -coverprofile=/tmp/coverage.out .
	#go tool cover -func=/tmp/coverage.out
	#go tool cover -html=/tmp/coverage.out -o /tmp/coverage.html
	#xdg-open /tmp/coverage.html

release/goss-linux-386: $(GO_FILES)
	$(info INFO: Starting build $@)
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "-X main.version=$(TRAVIS_TAG) -s -w" -o release/$(cmd)-linux-386 $(exe)
release/goss-linux-amd64: $(GO_FILES)
	$(info INFO: Starting build $@)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(TRAVIS_TAG) -s -w" -o release/$(cmd)-linux-amd64 $(exe)

release:
	$(MAKE) clean
	$(MAKE) build

build: release/goss-linux-386 release/goss-linux-amd64

test-int: centos7 wheezy precise alpine3 arch

centos7: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh $@
wheezy: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh $@
precise: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh $@
alpine3: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh $@
arch: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh $@


test-all: lint test test-int

deps:
	$(info INFO: Starting build $@)
	glide install

gen:
	$(info INFO: Starting build $@)
	go generate -tags genny $(pkgs)

clean:
	$(info INFO: Starting build $@)
	rm -rf ./release
