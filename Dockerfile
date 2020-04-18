FROM golang:alpine as builder

WORKDIR /tun2socks-src
COPY . /tun2socks-src

RUN apk add --no-cache \
        git \
		make \
		gcc \
		musl-dev \
    && go mod download \
    && make build \
    && mv ./bin/tun2socks /tun2socks

FROM alpine:latest

COPY --from=builder /tun2socks /
ENTRYPOINT ["/tun2socks"]
