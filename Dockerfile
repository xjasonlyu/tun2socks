FROM golang:alpine as builder

WORKDIR /tun2socks-src
COPY . /tun2socks-src

RUN apk add --update --no-cache \
    gcc git make musl-dev \
    && go mod download \
    && make build \
    && mv ./bin/tun2socks /tun2socks

FROM alpine:latest

COPY ./tun2socks.sh /
COPY --from=builder /tun2socks /usr/local/bin

RUN apk add --update --no-cache \
    curl lsof iptables iproute2 bind-tools \
    && chmod +x /tun2socks.sh

ENV TUN tun0
ENV ETH eth0
ENV ETHGW 172.16.1.1
ENV TUNGW 240.0.0.1
ENV PROXY 172.16.1.2:1080
ENV MONITOR 0.0.0.0:80
ENV EXCLUDED 172.16.1.2/32
ENV LOGLEVEL warning
ENV BACKENDDNS 8.8.8.8:53
ENV HOSTS localhost=127.0.0.1

ENTRYPOINT ["/tun2socks.sh"]
