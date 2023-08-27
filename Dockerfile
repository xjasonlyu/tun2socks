FROM golang:alpine AS builder

WORKDIR /src
COPY . /src

RUN apk add --update --no-cache make git \
    && make tun2socks

FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/xjasonlyu/tun2socks"

COPY docker/entrypoint.sh /entrypoint.sh
COPY --from=builder /src/build/tun2socks /usr/bin/tun2socks

RUN apk add --update --no-cache iptables iproute2 tzdata \
    && chmod +x /entrypoint.sh

ENV TUN=tun0
ENV ADDR=198.18.0.1/15
ENV LOGLEVEL=info
ENV PROXY=direct://
ENV MTU=9000
ENV RESTAPI=
ENV UDP_TIMEOUT=
ENV TCP_SNDBUF=
ENV TCP_RCVBUF=
ENV TCP_AUTO_TUNING=
ENV MULTICAST_GROUPS=
ENV EXTRA_COMMANDS=
ENV TUN_INCLUDED_ROUTES=
ENV TUN_EXCLUDED_ROUTES=

ENTRYPOINT ["/entrypoint.sh"]
