FROM golang:alpine AS builder

WORKDIR /tun2socks-src
COPY . /tun2socks-src

RUN apk add --no-cache make git \
    && go mod download \
    && make docker \
    && mv ./bin/tun2socks-docker /tun2socks

FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/xjasonlyu/tun2socks"

COPY docker/entrypoint.sh /entrypoint.sh
COPY --from=builder /tun2socks /usr/bin/tun2socks

RUN apk add --update --no-cache iptables iproute2 \
    && chmod +x /entrypoint.sh

ENV TUN tun0
ENV ETH eth0
ENV TUN_ADDR=198.18.0.1
ENV TUN_MASK=255.254.0.0
ENV LOGLEVEL=INFO
ENV PROXY=direct://
ENV STATS=
ENV TOKEN=
ENV EXTRA_COMMANDS=
ENV TUN_INCLUDED_ROUTES=
ENV TUN_EXCLUDED_ROUTES=

ENTRYPOINT ["/entrypoint.sh"]
