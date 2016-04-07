export GO15VENDOREXPERIMENT=1

exe = github.com/aelsabbahy/goss/cmd/goss
pkgs = $(shell glide novendor)
cmd = goss
TRAVIS_TAG ?= "0.0.0"

.PHONY: all build install test coverage deps release bench test-int lint gen centos6 wheezy precise alpine3 arch

all: test-all

install: release/goss-linux-amd64
	cp release/$(cmd)-linux-amd64 $(GOPATH)/bin/goss

test:
	go test $(pkgs)

lint:
	#go tool vet .
	golint $(pkgs) | grep -v 'unexported' || true

bench:
	go test -bench=.

coverage:
	go test -cover $(pkgs)
	#go test -coverprofile=/tmp/coverage.out .
	#go tool cover -func=/tmp/coverage.out
	#go tool cover -html=/tmp/coverage.out -o /tmp/coverage.html
	#xdg-open /tmp/coverage.html

release/goss-linux-386:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "-X main.version=$(TRAVIS_TAG)" -o release/$(cmd)-linux-386 $(exe)
	goupx $@
release/goss-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(TRAVIS_TAG)" -o release/$(cmd)-linux-amd64 $(exe)
	goupx $@
build: release/goss-linux-386 release/goss-linux-amd64

release: build
	#goupx release/*

test-int: centos6 wheezy precise alpine3 arch

centos6: build test
	cd integration-tests/ && ./test.sh $@
wheezy: build test
	cd integration-tests/ && ./test.sh $@
precise: build test
	cd integration-tests/ && ./test.sh $@
alpine3: build test
	cd integration-tests/ && ./test.sh $@
arch: build test
	cd integration-tests/ && ./test.sh $@


test-all: test lint test-int

deps:
	glide up

gen:
	go generate -tags genny $(pkgs)

clean:
	rm -rf ./release
