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

[English](README.md) | 简体中文

## 什么是 tun2socks ？

`tun2socks`是一个用来将网络层的`TCP/UDP`（包括`IPv4`和`IPv6`）流量套接字化的应用。它实现了一个`TUN`虚拟网卡接口，可以把所有到来的`TCP/UDP`包处理并转发给SOCKS服务器。

## 为什么使用 tun2socks ？

通过在主机上运行`tun2socks`，可以轻松地接管所有的`TCP/UDP`流量，同时提供诸多专业的功能特性，这包括：

- 强制使不支持代理的程序走代理
- 结合Clash、V2Ray等代理后端实现全局科学上网
- 结合Burp、Charles等工具进行应用层数据的调试
- 结合DHCP、CoreDNS等应用部署路由模式代理局域网流量

## 特性介绍

- ICMP 回应 / IPv6 支持 / Socks5 和 SS 代理支持
- 适用于游戏加速，针对UDP流量传输的专门优化
- 纯 Go 实现，不再需要 CGO，提升了稳定性
- 路由模式，可以用来转发及代理局域网内所有流量
- 核心由 [gVisor](https://github.com/google/gvisor) 强力驱动的 TCP/IP 网络栈
- 超过 2.5Gbps 的带宽吞吐量（[v1](https://github.com/xjasonlyu/tun2socks/tree/v1) 版本的10x倍以上）

## 硬件需求

| 目标 | 最小 | 建议 |
| :--- | :---: | :---: |
| 系统 | Linux MacOS Freebsd OpenBSD Windows | Linux or MacOS |
| 内存 | >20MB | >128MB |
| 架构 | AMD64(x86_64) ARM64 | AMD64 with AES-NI & AVX2 |

## 使用文档

文档以及使用方式，请看 [Github Wiki](https://github.com/xjasonlyu/tun2socks/wiki)。

## 源码编译

由于 tun2socks 是基于 gVisor 的 TCP/IP 栈，所以目前只支持 x86_64 和 ARM64 架构。以后其他的架构的支持取决于 gVisor。

### 环境依赖

确保安装了以下环境:

- Go 1.15+

### 开始编译

编译以及安装 `tun2socks` 二进制文件:

```shell
make tun2socks
sudo cp ./bin/tun2socks /usr/local/bin
```

编译所有架构的二进制文件:

```shell
make all-arch
```

## 特别感谢

- [Dreamacro/clash](https://github.com/Dreamacro/clash)
- [google/gvisor](https://github.com/google/gvisor)
- [majek/slirpnetstack](https://github.com/majek/slirpnetstack)
- [WireGuard/wireguard-go](https://git.zx2c4.com/wireguard-go)

## 注意事项

1. 由于采用了纯Go实现，所以这一版本的`tun2socks`在有大量连接时内存消耗通常较多。如果您的需求对内存消耗极为敏感，请继续使用 [v1](https://github.com/xjasonlyu/tun2socks/tree/v1) 版本。
2. `tun2socks`只应该专注于将网络层的TCP/UDP流量转发给SOCKS服务器，其他的如DNS（DoH）、DHCP等模块功能应该交由第三方应用实现，所以弃用了DNS模块。
3. 因为是通过用户空间的网络栈接管所有流量并处理转发，在高吞吐时CPU的使用量会剧增，所以CPU的性能直接与可以达到的最大带宽挂钩。

## TODO

- [x] Windows 支持
- [x] FreeBSD 支持
- [x] OpenBSD 支持
- [ ] 自动路由模式
