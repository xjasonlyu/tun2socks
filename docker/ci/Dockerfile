FROM golang:alpine AS builder

WORKDIR /src
COPY . /src

RUN apk add --update --no-cache make git \
    && make tun2socks

FROM alpine:latest

COPY --from=builder /src/build/tun2socks /usr/bin/tun2socks

RUN apk add --update --no-cache \
      iptables iproute2 tzdata \
      curl iperf3 bind-tools tcpdump
