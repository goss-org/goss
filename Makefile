export GO15VENDOREXPERIMENT=1

exe = github.com/aelsabbahy/goss/cmd/goss
pkgs = $(shell glide novendor)
cmd = goss
TRAVIS_TAG ?= "0.0.0"
GO_FILES = $(shell find . \( -path ./vendor -o -name '_test.go' \) -prune -o -name '*.go' -print)

.PHONY: all build install test coverage deps release bench test-int lint gen centos7 wheezy precise alpine3 arch test-int32 centos7-32 wheezy-32 precise-32 alpine3-32 arch-32

all: test-all test-all-32

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
release/goss-linux-arm: $(GO_FILES)
	$(info INFO: Starting build $@)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags "-X main.version=$(TRAVIS_TAG) -s -w" -o release/$(cmd)-linux-arm $(exe)
release/goss-windows.exe: $(GO_FILES)
	$(info INFO: Starting build $@)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(TRAVIS_TAG) -s -w" -o release/$(cmd)-windows.exe $(exe)

release:
	$(MAKE) clean
	$(MAKE) build

build: release/goss-linux-386 release/goss-linux-amd64 release/goss-linux-arm release/goss-windows.exe

test-int: centos7 wheezy precise alpine3 arch
test-int-32: centos7-32 wheezy-32 precise-32 alpine3-32 arch-32

centos7-32: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh centos7 386
wheezy-32: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh wheezy 386
precise-32: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh precise 386
alpine3-32: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh alpine3 386
arch-32: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh arch 386
centos7: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh centos7 amd64
wheezy: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh wheezy amd64
precise: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh precise amd64
alpine3: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh alpine3 amd64
arch: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh arch amd64


test-all-32: lint test test-int-32
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

build-images:
	$(info INFO: Starting build $@)
	development/build_images.sh

push-images:
	$(info INFO: Starting build $@)
	development/push_images.sh
