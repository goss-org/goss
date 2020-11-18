FROM gliderlabs/alpine:3.3
MAINTAINER Ahmed

# install apache2 and remove un-needed services
RUN apk update && \
  apk add openrc apache2 bash ca-certificates && \
  rc-update add apache2 && \
  rm -rf /etc/init.d/networking /etc/init.d/hwdrivers /var/cache/apk/* /tmp/*
RUN mkfifo /pipe
