![tun2socks](docs/logo.png)

[![GitHub Workflow][1]](https://github.com/xjasonlyu/tun2socks/actions)
[![Go Version][2]](https://github.com/xjasonlyu/tun2socks/blob/main/go.mod)
[![Go Report][3]](https://goreportcard.com/badge/github.com/xjasonlyu/tun2socks)
[![GitHub License][4]](https://github.com/xjasonlyu/tun2socks/blob/main/LICENSE)
[![Total Lines][5]](https://img.shields.io/tokei/lines/github/xjasonlyu/tun2socks?style=flat-square)
[![Releases][6]](https://github.com/xjasonlyu/tun2socks/releases)

[1]: https://img.shields.io/github/workflow/status/xjasonlyu/tun2socks/Go?style=flat-square
[2]: https://img.shields.io/github/go-mod/go-version/xjasonlyu/tun2socks/main?style=flat-square
[3]: https://goreportcard.com/badge/github.com/xjasonlyu/tun2socks?style=flat-square
[4]: https://img.shields.io/github/license/xjasonlyu/tun2socks?style=flat-square
[5]: https://img.shields.io/tokei/lines/github/xjasonlyu/tun2socks?style=flat-square
[6]: https://img.shields.io/github/v/release/xjasonlyu/tun2socks?include_prereleases&style=flat-square

English | [简体中文](README_ZH.md)

# What is tun2socks?

`tun2socks` is an application used to "socksify" TCP/UDP (IPv4 and IPv6) traffic at the network layer. It implements a TUN virtual network interface which accepts all incoming TCP/UDP packets and forwards them through a SOCKS server.

## Features

- ICMP echoing / IPv6 support / Socks5 & SS proxy
- Optimized UDP transmission for game acceleration
- Pure Go implementation, no more CGO required
- Router mode, routing all the traffic in LAN
- TCP/IP stack powered by [gVisor](https://github.com/google/gvisor)
- More than 2.5Gbps throughput (10x faster than [v1](https://github.com/xjasonlyu/tun2socks/tree/v1))

## Requirements

| Target | Minimum | Recommended |
| :----- | :-----: | :---------: |
| System | Linux MacOS Freebsd OpenBSD Windows | Linux or MacOS |
| Memory | >20MB | >128MB |
| CPU | AMD64(x86_64) ARM64 | AMD64 with AES-NI & AVX2 |

## Documentation

Documentations and quick start guides can be found at [Github Wiki](https://github.com/xjasonlyu/tun2socks/wiki).

## Building from source

Due to the limitation of gVisor, tun2socks only supports x86_64 and ARM64 for now. Other architectures may become available in the future.

### Environments

Make sure the following dependencies are installed:

- Go 1.15+

### Building

Build and install the `tun2socks` binary:

```shell
make tun2socks
sudo cp ./bin/tun2socks /usr/local/bin
```

Build for all architectures:

```shell
make all-arch
```

## Credits

- [Dreamacro/clash](https://github.com/Dreamacro/clash)
- [google/gvisor](https://github.com/google/gvisor)
- [majek/slirpnetstack](https://github.com/majek/slirpnetstack)
- [WireGuard/wireguard-go](https://git.zx2c4.com/wireguard-go)

## TODO

- [x] Windows support
- [x] FreeBSD support
- [x] OpenBSD support
- [ ] Auto route mode
