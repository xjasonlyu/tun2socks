FROM golang:alpine AS builder

WORKDIR /tun2socks-src
COPY . /tun2socks-src

RUN apk add --no-cache make git \
    && go mod download \
    && make tun2socks \
    && mv ./build/tun2socks /tun2socks

FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/xjasonlyu/tun2socks"

COPY docker/entrypoint.sh /entrypoint.sh
COPY --from=builder /tun2socks /usr/bin/tun2socks

RUN apk add --update --no-cache iptables iproute2 \
    && chmod +x /entrypoint.sh

ENV TUN=tun0
ENV ADDR=198.18.0.1/15
ENV LOGLEVEL=info
ENV PROXY=direct://
ENV MTU=9000
ENV STATS=
ENV TOKEN=
ENV EXTRA_COMMANDS=
ENV TUN_INCLUDED_ROUTES=
ENV TUN_EXCLUDED_ROUTES=

ENTRYPOINT ["/entrypoint.sh"]
