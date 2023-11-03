![tun2socks](docs/wordmark.png)

[![GitHub Workflow][1]](https://github.com/xjasonlyu/tun2socks/actions)
[![Go Version][2]](https://github.com/xjasonlyu/tun2socks/blob/main/go.mod)
[![Go Report][3]](https://goreportcard.com/badge/github.com/xjasonlyu/tun2socks)
[![Maintainability][4]](https://codeclimate.com/github/xjasonlyu/tun2socks/maintainability)
[![GitHub License][5]](https://github.com/xjasonlyu/tun2socks/blob/main/LICENSE)
[![Docker Pulls][6]](https://hub.docker.com/r/xjasonlyu/tun2socks)
[![Releases][7]](https://github.com/xjasonlyu/tun2socks/releases)

[1]: https://img.shields.io/github/actions/workflow/status/xjasonlyu/tun2socks/release.yml?logo=github
[2]: https://img.shields.io/github/go-mod/go-version/xjasonlyu/tun2socks?logo=go
[3]: https://goreportcard.com/badge/github.com/xjasonlyu/tun2socks
[4]: https://api.codeclimate.com/v1/badges/b5b30239174fc6603aca/maintainability
[5]: https://img.shields.io/github/license/xjasonlyu/tun2socks
[6]: https://img.shields.io/docker/pulls/xjasonlyu/tun2socks?logo=docker
[7]: https://img.shields.io/github/v/release/xjasonlyu/tun2socks?logo=smartthings

English | [简体中文](README_ZH.md)

## Features

- Proxy Everything: Handle all network traffic of any internet programs sent by the device through a proxy.
- Proxy Protocols: HTTP/Socks4/Socks5/Shadowsocks with authentication support for remote connections.
- Run Everywhere: Linux/macOS/Windows/FreeBSD/OpenBSD multi-platform support with specific optimization.
- Gateway Mode: Act as a layer three gateway to handle network traffic from other devices in the same network.
- Full IPv6 Support: All functions work in IPv6, tunnel IPv4 connections through IPv6 proxy and vice versa.
- Network Stack: Powered by user-space TCP/IP stack from Google container application kernel **[gVisor](https://github.com/google/gvisor)**.

## Benchmarks

For all scenarios of usage, tun2socks performs best. See [here](https://github.com/xjasonlyu/tun2socks/wiki/Benchmarks) for more details.

![benchmark](docs/benchmark.png)

## Documentation

- [Install from Source](https://github.com/xjasonlyu/tun2socks/wiki/Install-from-Source)
- [Quickstart Examples](https://github.com/xjasonlyu/tun2socks/wiki/Examples)
- [Memory Optimization](https://github.com/xjasonlyu/tun2socks/wiki/Memory-Optimization)

Full documentation and technical guides can be found at [Wiki](https://github.com/xjasonlyu/tun2socks/wiki).

## Community

Welcome and feel free to ask any questions at [Discussions](https://github.com/xjasonlyu/tun2socks/discussions).

## Credits

- [google/gvisor](https://github.com/google/gvisor) - Application Kernel for Containers
- [wireguard-go](https://git.zx2c4.com/wireguard-go) - Go Implementation of WireGuard

## License

[GPL-3.0](https://github.com/xjasonlyu/tun2socks/blob/main/LICENSE)

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fxjasonlyu%2Ftun2socks.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fxjasonlyu%2Ftun2socks?ref=badge_large)

## Star History

<a href="https://star-history.com/#xjasonlyu/tun2socks&Date">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=xjasonlyu/tun2socks&type=Date&theme=dark" />
    <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=xjasonlyu/tun2socks&type=Date" />
    <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=xjasonlyu/tun2socks&type=Date" />
  </picture>
</a>
