FROM alpine:3.23

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/goss /usr/bin/

RUN mkdir /goss
VOLUME /goss
