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

[English](README.md) | 简体中文

## 为什么使用 tun2socks ？

通过在主机上运行`tun2socks`，可以轻松地接管所有的`TCP/UDP`流量，同时提供诸多专业的功能特性，这包括：

- 强制使不支持代理的程序走代理
- 配合Clash、V2Ray等工具实现全局代理上网
- 配合Burp、Charles等工具进行应用层数据的调试
- 配合DHCP、CoreDNS等工具部署路由模式代理局域网流量

## 特性介绍

- **全面支持：** IPv4/IPv6/ICMP/TCP/UDP
- **代理协议：** Socks5/Shadowsocks
- **游戏加速：** 针对UDP传输的优化
- **纯Go实现：** 无需CGO，稳定性提升
- **路由模式：** 转发代理局域网内所有流量
- **TCP/IP栈：** 由 **[gVisor](https://github.com/google/gvisor)** 强力驱动 
- **高性能：** >2.5Gbps 的带宽吞吐量

## 硬件需求

| 目标 | 最小 | 建议 |
| :--- | :---: | :---: |
| 系统 | Linux MacOS Freebsd OpenBSD Windows | Linux or MacOS |
| 内存 | >20MB | >128MB |
| 架构 | ANY | AMD64 or ARM64 |

## 使用文档

文档以及使用方式，请看 [Github Wiki](https://github.com/xjasonlyu/tun2socks/wiki)。

## 交流讨论

欢迎来讨论区交流提问，[Github Discussions](https://github.com/xjasonlyu/tun2socks/discussions)。

## 注意事项

1. 由于采用了纯Go实现，所以这一版本的`tun2socks`在有大量连接时内存消耗通常较多。如果您的需求对内存消耗极为敏感，请继续使用 [v1](https://github.com/xjasonlyu/tun2socks/tree/v1) 版本。
2. `tun2socks`只应该专注于将网络层的TCP/UDP流量转发给SOCKS服务器，其他的如DNS（DoH）、DHCP等模块功能应该交由第三方应用实现，所以弃用了DNS模块。
3. 因为是通过用户空间的网络栈接管所有流量并处理转发，在高吞吐时CPU的使用量会剧增，所以CPU的性能直接与可以达到的最大带宽挂钩。

## 特别感谢

- [Dreamacro/clash](https://github.com/Dreamacro/clash) - A rule-based tunnel in Go
- [google/gvisor](https://github.com/google/gvisor) - Application Kernel for Containers
- [wireguard-go](https://git.zx2c4.com/wireguard-go) - Go Implementation of WireGuard

## License

[GPL-3.0](https://github.com/xjasonlyu/tun2socks/blob/main/LICENSE)

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fxjasonlyu%2Ftun2socks.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fxjasonlyu%2Ftun2socks?ref=badge_large)
