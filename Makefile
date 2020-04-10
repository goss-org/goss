export GO15VENDOREXPERIMENT=1

exe = github.com/aelsabbahy/goss/cmd/goss
pkgs = $(shell ./novendor.sh)
cmd = goss
TRAVIS_TAG ?= "0.0.0"
GO_FILES = $(shell find . \( -path ./vendor -o -name '_test.go' \) -prune -o -name '*.go' -print)
GO111MODULE=on

.PHONY: all build install test release bench fmt lint vet test-int-all gen centos7 wheezy precise alpine3 arch test-int32 centos7-32 wheezy-32 precise-32 alpine3-32 arch-32

all: test-all dgoss-sha256

test-all: fmt lint vet test test-int-all

install: release/goss-linux-amd64
	$(info INFO: Starting build $@)
	cp release/$(cmd)-linux-amd64 $(GOPATH)/bin/goss

test:
	$(info INFO: Starting build $@)
	{ \
set -e ;\
go test -coverprofile=c.out ${pkgs} ;\
cat c.out | sed 's|github.com/aelsabbahy/goss/||' > c.out.tmp ;\
mv c.out.tmp c.out ;\
}

lint:
	$(info INFO: Starting build $@)
	golint $(pkgs) || true

vet:
	$(info INFO: Starting build $@)
	go vet $(pkgs) || true

fmt:
	$(info INFO: Starting build $@)
	{ \
set -e ;\
fmt=$$(gofmt -l ${GO_FILES}) ;\
[ -z "$$fmt" ] && echo "valid gofmt" || (echo -e "invalid gofmt\n$$fmt"; exit 1)\
}

bench:
	$(info INFO: Starting build $@)
	go test -bench=.



# Pattern rule for platform builds.
# `subst` substitutes space for -, thus making an array
# firstword, and word select indexes from said array.
release/goss-%: $(GO_FILES)
	CGO_ENABLED=0 GOOS=$(firstword $(subst -, ,$*)) GOARCH=$(word 2, $(subst -, ,$*)) go build -ldflags "-X main.version=$(TRAVIS_TAG) -s -w" -o $@ $(exe)
	sha256sum $@ > $@.sha256

release:
	$(MAKE) clean
	$(MAKE) build

build: release/goss-linux-386 release/goss-linux-amd64 release/goss-linux-arm

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

test-int-64: centos7 wheezy precise alpine3 arch
test-int-32: centos7-32 wheezy-32 precise-32 alpine3-32 arch-32
test-int-all: test-int-32 test-int-64

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

dgoss-sha256:
	cd extras/dgoss/ && sha256sum dgoss > dgoss.sha256
