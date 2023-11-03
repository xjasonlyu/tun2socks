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

[English](README.md) | 简体中文

## 特性介绍

- 全局代理: 处理来自本设备的任意网络应用的所有网络流量并通过代理转发。
- 代理协议: 通过 HTTP/Socks4/Socks5/Shadowsocks 远程连接且支持鉴权。
- 跨平台性: 具有 Linux/macOS/Windows/FreeBSD/OpenBSD 特定优化的多平台支持。
- 网关模式: 作为第三层网关处理来自同一网络中其他设备的所有网络流量。
- IPv6 支持: 所有功能都可以在 IPv6 中工作，允许通过 IPv6 代理转发 IPv4 连接，反之亦然。
- TCP/IP 栈: 由来自 Google 容器应用程序内核 **[gVisor](https://github.com/google/gvisor)** 的用户空间 TCP/IP 网络栈强力驱动。

## 性能测试

对于任意的使用场景，tun2socks 表现最佳。更多细节看[这里](https://github.com/xjasonlyu/tun2socks/wiki/Benchmarks)。

![benchmark](docs/benchmark.png)

## 使用文档

- [源码安装](https://github.com/xjasonlyu/tun2socks/wiki/Install-from-Source)
- [使用例子](https://github.com/xjasonlyu/tun2socks/wiki/Examples)
- [内存优化](https://github.com/xjasonlyu/tun2socks/wiki/Memory-Optimization)

文档以及使用方式可以在 [Wiki](https://github.com/xjasonlyu/tun2socks/wiki) 里找到。

## 交流讨论

欢迎来讨论区 [Discussions](https://github.com/xjasonlyu/tun2socks/discussions) 交流提问。

## 特别感谢

- [google/gvisor](https://github.com/google/gvisor) - Application Kernel for Containers
- [wireguard-go](https://git.zx2c4.com/wireguard-go) - Go Implementation of WireGuard

## 许可协议

[GPL-3.0](https://github.com/xjasonlyu/tun2socks/blob/main/LICENSE)

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fxjasonlyu%2Ftun2socks.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fxjasonlyu%2Ftun2socks?ref=badge_large)

## 星星历史

<a href="https://star-history.com/#xjasonlyu/tun2socks&Date">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=xjasonlyu/tun2socks&type=Date&theme=dark" />
    <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=xjasonlyu/tun2socks&type=Date" />
    <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=xjasonlyu/tun2socks&type=Date" />
  </picture>
</a>
