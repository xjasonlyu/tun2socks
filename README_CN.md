![tun2socks](docs/logo.png)

[![GitHub Workflow][1]](https://github.com/xjasonlyu/tun2socks/actions)
[![Go Version][2]](https://github.com/xjasonlyu/tun2socks/blob/main/go.mod)
[![Go Report][3]](https://goreportcard.com/badge/github.com/xjasonlyu/tun2socks)
[![Maintainability][4]](https://qlty.sh/gh/xjasonlyu/projects/tun2socks)
[![GitHub License][5]](https://github.com/xjasonlyu/tun2socks/blob/main/LICENSE)
[![Docker Pulls][6]](https://hub.docker.com/r/xjasonlyu/tun2socks)
[![Releases][7]](https://github.com/xjasonlyu/tun2socks/releases)

## 功能特性

- **通用代理**：透明地将任何应用程序的网络流量路由到代理服务器。
- **多协议支持**：支持 HTTP/SOCKS/Shadowsocks/SSH/Relay 代理，支持身份验证。
- **Web 管理界面**：内置 GitHub 风格的现代化仪表盘，用于实时监控和配置。
  - **状态监控**：查看服务运行状态、PID、运行时间以及实时的上传/下载流量统计。
  - **动态配置**：无需重启即可实时切换代理服务器配置。
  - **路由管理**：通过友好的用户界面添加、删除和管理网络路由规则。
  - **设备状态**：监控 TUN 设备的连接状态、IP 地址和 MTU 设置。
- **跨平台**：支持 Linux/macOS/Windows/FreeBSD/OpenBSD，并针对各平台进行了优化。
- **网关模式**：可作为三层（Layer 3）网关，路由同一网络中其他设备的流量。
- **完整 IPv6 支持**：原生支持 IPv6；可在 IPv6 上隧道传输 IPv4，反之亦然。
- **用户态网络栈**：利用 **[gVisor](https://github.com/google/gvisor)** 网络栈，提供卓越的性能和灵活性。

## Web 仪表盘

tun2socks 现在包含一个内置的现代化 Web 管理界面。

**访问地址**：`http://127.0.0.1:7777`（默认）

主要功能：
- **服务控制**：监控引擎状态和实时流量速度。
- **代理切换**：即时在不同的 SOCKS5/HTTP 代理之间切换。
- **路由管理**：管理路由规则，自动检测网关地址。
- **持久化配置**：配置自动保存至 `~/.config/tun2socks/config.yaml`。

## 使用说明

### 命令行启动

启动 tun2socks 并开启 Web 界面：

```bash
sudo ./tun2socks \
  --device tun://tunsocks \
  --proxy socks5://127.0.0.1:1080 \
  --restapi http://mytoken@127.0.0.1:7777
```

参数说明：
- `--device`：指定 TUN 设备名称（程序会自动创建并配置 IP）。
- `--proxy`：（可选）初始代理服务器地址。稍后可通过 Web 界面动态更新。
- `--restapi`：启用 Web API 和仪表盘。格式：`http://[token]@[ip]:[port]`。

启动后，在浏览器访问 `http://127.0.0.1:7777` 并输入 token `mytoken` 即可登录。

### 配置文件

您也可以使用配置文件启动：

```bash
sudo ./tun2socks -config config.yaml
```

完整配置参考请见 [config.example.yaml](config.example.yaml)。

### Systemd 系统服务 (Linux)

将 tun2socks 作为系统服务运行：

1. 复制二进制文件到 `/usr/local/bin/`
2. 创建配置目录：`sudo mkdir -p /etc/tun2socks`
3. 复制配置文件到 `/etc/tun2socks/config.yaml`
4. 安装服务文件：
   ```bash
   sudo cp tun2socks.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable --now tun2socks
   ```

### 源码编译

编译要求：
- Go 1.21+
- Node.js 18+ (用于编译前端)

```bash
# 1. 编译前端
cd web
npm install
npm run build
cd ..

# 2. 编译二进制文件
go build -o tun2socks .
```

## 性能基准

![benchmark](docs/benchmark.png)

在各种使用场景下，tun2socks 都表现出色。
更多详情请参阅 [Benchmarks](https://github.com/xjasonlyu/tun2socks/wiki/Benchmarks)。

## 文档

- [源码安装](https://github.com/xjasonlyu/tun2socks/wiki/Install-from-Source)
- [快速开始示例](https://github.com/xjasonlyu/tun2socks/wiki/Examples)
- [内存优化](https://github.com/xjasonlyu/tun2socks/wiki/Memory-Optimization)

完整文档和技术指南请访问 [Wiki](https://github.com/xjasonlyu/tun2socks/wiki)。

## 社区

欢迎在 [Discussions](https://github.com/xjasonlyu/tun2socks/discussions) 中提出问题或参与讨论。

## 致谢

- [google/gvisor](https://github.com/google/gvisor) - 容器应用内核
- [wireguard-go](https://git.zx2c4.com/wireguard-go) - WireGuard 的 Go 实现
- [wintun](https://git.zx2c4.com/wintun/) - Windows 下的三层 TUN 驱动

## 许可证

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fxjasonlyu%2Ftun2socks.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fxjasonlyu%2Ftun2socks?ref=badge_large)

## Star 历史

<a href="https://star-history.com/#xjasonlyu/tun2socks&Date">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=xjasonlyu/tun2socks&type=Date&theme=dark" />
    <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=xjasonlyu/tun2socks&type=Date" />
    <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=xjasonlyu/tun2socks&type=Date" />
  </picture>
</a>

[1]: https://img.shields.io/github/actions/workflow/status/xjasonlyu/tun2socks/docker.yml?logo=github
[2]: https://img.shields.io/github/go-mod/go-version/xjasonlyu/tun2socks?logo=go
[3]: https://goreportcard.com/badge/github.com/xjasonlyu/tun2socks
[4]: https://qlty.sh/gh/xjasonlyu/projects/tun2socks/maintainability.svg
[5]: https://img.shields.io/github/license/xjasonlyu/tun2socks
[6]: https://img.shields.io/docker/pulls/xjasonlyu/tun2socks?logo=docker
[7]: https://img.shields.io/github/v/release/xjasonlyu/tun2socks?logo=smartthings
