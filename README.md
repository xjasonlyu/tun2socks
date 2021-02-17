![tun2socks](docs/logo.png)

[![GitHub Workflow][1]](https://github.com/xjasonlyu/tun2socks/actions)
[![Go Version][2]](https://github.com/xjasonlyu/tun2socks/blob/main/go.mod)
[![Go Report][3]](https://goreportcard.com/badge/github.com/xjasonlyu/tun2socks)
[![GitHub License][4]](https://github.com/xjasonlyu/tun2socks/blob/main/LICENSE)
[![Releases][5]](https://github.com/xjasonlyu/tun2socks/releases)

[1]: https://img.shields.io/github/workflow/status/xjasonlyu/tun2socks/Go?style=flat-square
[2]: https://img.shields.io/github/go-mod/go-version/xjasonlyu/tun2socks/main?style=flat-square
[3]: https://goreportcard.com/badge/github.com/xjasonlyu/tun2socks?style=flat-square
[4]: https://img.shields.io/github/license/xjasonlyu/tun2socks?style=flat-square
[5]: https://img.shields.io/github/v/release/xjasonlyu/tun2socks?include_prereleases&style=flat-square

English | [简体中文](README_ZH.md)

## Features

- **Fully support:** IPv4/IPv6/ICMP/TCP/UDP
- **Proxy protocol:** Socks5/Shadowsocks
- **Game ready:** optimized UDP transmission
- **Pure Go:** no CGO required, stability improved
- **Router mode:** forwarding packets in LAN
- **TCP/IP stack:** powered by **[gVisor](https://github.com/google/gvisor)**
- **High performance:** >2.5Gbps throughput

## Requirements

| Target | Minimum | Recommended |
| :----- | :-----: | :---------: |
| System | Linux MacOS Freebsd OpenBSD Windows | Linux or MacOS |
| Memory | >20MB | >128MB |
| CPU | ANY | AMD64 or ARM64 |

## Documentation

Documentations and quick start guides can be found at [Github Wiki](https://github.com/xjasonlyu/tun2socks/wiki).

## Credits

- [Dreamacro/clash](https://github.com/Dreamacro/clash) - A rule-based tunnel in Go
- [google/gvisor](https://github.com/google/gvisor) - Application Kernel for Containers
- [wireguard-go](https://git.zx2c4.com/wireguard-go) - Go Implementation of WireGuard

## License

[GPL-3.0](https://github.com/xjasonlyu/tun2socks/blob/main/LICENSE)

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fxjasonlyu%2Ftun2socks.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fxjasonlyu%2Ftun2socks?ref=badge_large)
