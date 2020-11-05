# tun2socks

![build](https://img.shields.io/github/workflow/status/xjasonlyu/tun2socks/Go?style=flat-square)
![go report](https://goreportcard.com/badge/github.com/xjasonlyu/tun2socks?style=flat-square)
![license](https://img.shields.io/github/license/xjasonlyu/tun2socks?style=flat-square)
![release](https://img.shields.io/github/v/release/xjasonlyu/tun2socks.svg?include_prereleases&style=flat-square)

A tun2socks implementation written in Go.

## Features

- External RESTful API support
- Fake DNS with manual hosts support
- IPv4/IPv6 support
- Optimized UDP transmission for game acceleration
- Pure Go implementation, no CGO required
- Socks5, Shadowsocks protocol support for remote connections
- TCP/IP stack powered by [gVisor](https://github.com/google/gvisor)
- Up to 2.5Gbps throughput (10x faster than [v1](https://github.com/xjasonlyu/tun2socks/tree/v1))

### Requirements

| Target | Minimum | Recommended |
| --- | --- | --- |
| System | linux darwin | linux |
| Memory | >20MB | +∞ |
| CPU | ANY | amd64 arm64 |

## QuickStart

Download from precompiled [Releases](https://github.com/xjasonlyu/tun2socks/releases).

create tun

```shell script
ip tuntap add mode tun dev tun0
ip addr add 198.18.0.1/15 dev tun0
ip link set dev tun0 up
```

run

```shell script
./tun2socks --loglevel WARN --device tun://tun0 --proxy socks5://server:port --interface eth0
```

or just

```shell script
PROXY=socks5://server:port LOGLEVEL=WARN sh ./scripts/entrypoint.sh
```

## Build from source

```text
$ git clone https://github.com/xjasonlyu/tun2socks.git
$ cd tun2socks
$ make
```

## Issues

Due to the implementation of pure Go, the memory usage is higher than the previous version.
If you are memory sensitive, please go back to [v1](https://github.com/xjasonlyu/tun2socks/tree/v1).

## TODO

- [ ] Windows support

## Credits

- [Dreamacro/clash](https://github.com/Dreamacro/clash)
- [google/gvisor](https://github.com/google/gvisor)
- [majek/slirpnetstack](https://github.com/majek/slirpnetstack)
