ARG GO_VERSION=1.21

FROM docker.io/golang:${GO_VERSION}-alpine AS base

ARG GOSS_VERSION=v0.0.0
WORKDIR /build

RUN --mount=target=. \
    CGO_ENABLED=0 go build \
    -ldflags "-X github.com/goss-org/goss/util.Version=${GOSS_VERSION} -s -w" \
    -o "/release/goss" \
    ./cmd/goss

FROM alpine:3.19

COPY --from=base /release/* /usr/bin/

RUN mkdir /goss
VOLUME /goss
