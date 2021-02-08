![tun2socks](docs/logo.png)

[![GitHub Workflow][1]](https://github.com/xjasonlyu/tun2socks/actions)
[![Go Version][2]](https://github.com/xjasonlyu/tun2socks/blob/main/go.mod)
[![Go Report][3]](https://goreportcard.com/badge/github.com/xjasonlyu/tun2socks)
[![GitHub License][4]](https://github.com/xjasonlyu/tun2socks/blob/main/LICENSE)
[![Releases][5]](https://github.com/xjasonlyu/tun2socks/releases)

[1]: https://img.shields.io/github/workflow/status/xjasonlyu/tun2socks/Go/master?style=flat-square
[2]: https://img.shields.io/github/go-mod/go-version/xjasonlyu/tun2socks?style=flat-square
[3]: https://goreportcard.com/badge/github.com/xjasonlyu/tun2socks?style=flat-square
[4]: https://img.shields.io/github/license/xjasonlyu/tun2socks?style=flat-square
[5]: https://img.shields.io/github/v/release/xjasonlyu/tun2socks?include_prereleases&style=flat-square

[English](README.md)

## 特性

- ICMP 回应 / IPv6 支持 / Socks5 和 SS 代理
- 适用于游戏加速，由专门优化过UDP流量的传输
- 纯 Go 实现，编译不再需要 CGO 的加持
- 路由模式，可以用来转发代理局域网内所有流量
- 由 [gVisor](https://github.com/google/gvisor) 强力驱动的 TCP/IP 网络栈
- 实测超过 2.5Gbps 的带宽吞吐量（ [v1](https://github.com/xjasonlyu/tun2socks/tree/v1) 版本的10x倍以上）

## 硬件需求

| 目标 | 最小 | 建议 |
| :----- | :-----: | :---------: |
| 系统 | Linux MacOS Freebsd OpenBSD Windows | Linux MacOS |
| 内存 | >20MB | >128MB |
| 架构 | AMD64(x86_64) ARM64 | AMD64 with AES-NI & AVX2 |

## 文档

文档以及使用方式，请看 [Github Wiki](https://github.com/xjasonlyu/tun2socks/wiki) 。

## 从源码编译

由于 tun2socks 是基于 gVisor的，所以目前只支持 x86_64 和 ARM64 架构。以后可能会支持其他的架构（取决于gVisor）。

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

## 感谢

- [Dreamacro/clash](https://github.com/Dreamacro/clash)
- [google/gvisor](https://github.com/google/gvisor)
- [majek/slirpnetstack](https://github.com/majek/slirpnetstack)
- [WireGuard/wireguard-go](https://git.zx2c4.com/wireguard-go)

## 已知问题

由于采用了纯Go实现，所以这一版本的内存消耗通常会高于上一个版本（在有大量连接时更为明显）。
如果您的需求对内存消耗极为敏感，请继续使用 [v1](https://github.com/xjasonlyu/tun2socks/tree/v1) 版本。

## TODO

- [x] Windows 支持
- [x] FreeBSD 支持
- [x] OpenBSD 支持
- [ ] 自动路由模式
