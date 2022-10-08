export GO15VENDOREXPERIMENT=1

exe = github.com/aelsabbahy/goss/cmd/goss
pkgs = $(shell ./novendor.sh)
cmd = goss
GO111MODULE=on
GO_FILES = $(shell git ls-files -- '*.go' ':!:*vendor*_test.go')

.PHONY: all build install test release bench fmt lint vet test-int-all gen centos7 wheezy trusty alpine3 arch test-int32 centos7-32 wheezy-32 trusty-32 alpine3-32 arch-32

all: test-short-all test-int-all dgoss-sha256

test-short-all: fmt lint vet test

install: release/goss-linux-amd64
	$(info INFO: Starting build $@)
	cp release/$(cmd)-linux-amd64 $(GOPATH)/bin/goss

test:
	$(info INFO: Starting build $@)
	./ci/go-test.sh $(pkgs)

lint:
	$(info INFO: Starting build $@)
	golint $(pkgs) || true

vet:
	$(info INFO: Starting build $@)
	go vet $(pkgs) || true

fmt:
	$(info INFO: Starting build $@)
	./ci/go-fmt.sh

bench:
	$(info INFO: Starting build $@)
	go test -bench=.

alpha-test-%: release/goss-%
	$(info INFO: Starting build $@)
	./integration-tests/run-tests-alpha.sh $*

test-int-serve-%: release/goss-%
	$(info INFO: Starting build $@)
	./integration-tests/run-serve-tests.sh $*
# shim to account for linux being not in alpha
test-int-serve-linux-amd64: test-int-serve-alpha-linux-amd64

release/goss-%: $(GO_FILES)
	./release-build.sh $*

release:
	$(MAKE) clean
	$(MAKE) build

build: release/goss-alpha-darwin-amd64 release/goss-linux-386 release/goss-linux-amd64 release/goss-linux-arm release/goss-linux-arm64 release/goss-alpha-windows-amd64

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

test-darwin-all: test-short-all test-int-darwin-all
# linux _does_ have the docker-style testing, but does _not_ currently have the same style integration tests darwin+windows do, _because_ of the docker-style testing.
test-linux-all: test-short-all test-int-64 test-int-32
test-windows-all: test-short-all test-int-windows-all

test-int-64: centos7 wheezy trusty alpine3 arch test-int-serve-linux-amd64
test-int-32: centos7-32 wheezy-32 trusty-32 alpine3-32 arch-32
test-int-darwin-all: alpha-test-alpha-darwin-amd64 test-int-serve-alpha-darwin-amd64
test-int-windows-all: alpha-test-alpha-windows-amd64 test-int-serve-alpha-windows-amd64
test-int-all: test-int-32 test-int-64

centos7-32: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh centos7 386
wheezy-32: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh wheezy 386
trusty-32: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh trusty 386
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
trusty: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh trusty amd64
alpine3: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh alpine3 amd64
arch: build
	$(info INFO: Starting build $@)
	cd integration-tests/ && ./test.sh arch amd64

dgoss-sha256:
	cd extras/dgoss/ && sha256sum dgoss > dgoss.sha256
