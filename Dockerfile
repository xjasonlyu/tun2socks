FROM golang:alpine as builder

WORKDIR /tun2socks-src
COPY . /tun2socks-src

RUN apk add --update --no-cache \
    gcc git make musl-dev \
    && go mod download \
    && go get -u github.com/gobuffalo/packr/packr \
    && packr \
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

ENTRYPOINT ["/tun2socks.sh"]
