FROM base/archlinux
MAINTAINER @siddharthist

RUN pacman -Sy --noconfirm systemd-sysvcompat
RUN ln -s /does_not_exist /foo && \
    chmod 700 ~root
RUN mkfifo /pipe