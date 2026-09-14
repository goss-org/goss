FROM alpine:3.24.1

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/goss /usr/bin/

RUN mkdir /goss
VOLUME /goss
