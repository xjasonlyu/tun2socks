FROM golang:alpine AS builder

WORKDIR /tun2socks-src
COPY . /tun2socks-src

RUN apk add --no-cache make git \
    && go mod download \
    && make docker \
    && mv ./bin/tun2socks-docker /tun2socks

FROM alpine:latest

COPY ./scripts/entrypoint.sh /entrypoint.sh
COPY --from=builder /tun2socks /usr/bin/tun2socks

RUN apk add --update --no-cache iptables iproute2 \
    && chmod +x /entrypoint.sh

ENV TUN tun0
ENV ETH eth0
ENV TUN_ADDR=
ENV TUN_MASK=
ENV LOGLEVEL=
ENV EXCLUDED=
ENV EXTRACMD=

ENV API=
ENV DNS=
ENV HOSTS=
ENV PROXY=

ENTRYPOINT ["/entrypoint.sh"]
