![tun2socks](docs/logo.png)

[![GitHub Workflow][1]](https://github.com/xjasonlyu/tun2socks/actions)
[![Go Version][2]](https://github.com/xjasonlyu/tun2socks/blob/main/go.mod)
[![Go Report][3]](https://goreportcard.com/badge/github.com/xjasonlyu/tun2socks)
[![Maintainability][4]](https://qlty.sh/gh/xjasonlyu/projects/tun2socks)
[![GitHub License][5]](https://github.com/xjasonlyu/tun2socks/blob/main/LICENSE)
[![Docker Pulls][6]](https://hub.docker.com/r/xjasonlyu/tun2socks)
[![Releases][7]](https://github.com/xjasonlyu/tun2socks/releases)

## Features

- **Universal Proxying**: Transparently routes all network traffic from any application through a proxy.
- **Multi-Protocol**: Supports HTTP/SOCKS/Shadowsocks/SSH/Relay proxies with optional authentication.
- **Web Interface**: Built-in GitHub-style dashboard for real-time monitoring and configuration.
  - **Status Monitoring**: View service status, PID, uptime, and real-time traffic statistics.
  - **Dynamic Configuration**: Change proxy settings on the fly without restarting.
  - **Route Management**: Add and remove network routes via a user-friendly UI.
  - **Device Status**: Monitor TUN device state, IP address, and MTU.
- **Cross-Platform**: Runs on Linux/macOS/Windows/FreeBSD/OpenBSD with platform-specific optimizations.
- **Gateway Mode**: Acts as a Layer 3 gateway to route traffic from other devices on the same network.
- **Full IPv6 Compatibility**: Natively supports IPv6; seamlessly tunnels IPv4 over IPv6 and vice versa.
- **User-Space Networking**: Leverages the **[gVisor](https://github.com/google/gvisor)** network stack for enhanced performance and flexibility.

## Web Dashboard

tun2socks now includes a modern, built-in web dashboard for easy management.

**Access**: `http://127.0.0.1:7777` (default)

Key capabilities:
- **Service Control**: Monitor engine status and traffic speeds.
- **Proxy Switching**: Switch between SOCKS5/HTTP proxies instantly.
- **Routing**: Manage routing rules with automatic gateway detection.
- **Persistence**: Configuration is automatically saved to `~/.config/tun2socks/config.yaml`.

## Usage

### Command Line

Start tun2socks with the web interface enabled:

```bash
sudo ./tun2socks \
  --device tun://tunsocks \
  --proxy socks5://127.0.0.1:1080 \
  --restapi http://mytoken@127.0.0.1:7777
```

- `--device`: Specifies the TUN device name (will be automatically created and configured).
- `--proxy`: (Optional) Initial proxy server. Can be updated via Web UI later.
- `--restapi`: Enables the Web API and Dashboard. Format: `http://[token]@[ip]:[port]`.

Access the dashboard at `http://127.0.0.1:7777` and log in with the token `mytoken`.

### Configuration File

You can also use a configuration file instead of command-line arguments:

```bash
sudo ./tun2socks -config config.yaml
```

See [config.example.yaml](config.example.yaml) for a complete reference.

### Systemd Service (Linux)

To run tun2socks as a system service:

1. Copy the binary to `/usr/local/bin/`
2. Create configuration directory: `sudo mkdir -p /etc/tun2socks`
3. Copy your config to `/etc/tun2socks/config.yaml`
4. Install the service file:
   ```bash
   sudo cp tun2socks.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable --now tun2socks
   ```

### Build from Source

Requirements:
- Go 1.21+
- Node.js 18+ (for web interface)

**Build with web interface** (default):

```bash
make
```

This will:
1. Build the React/TypeScript frontend
2. Embed the compiled assets into the Go binary
3. Generate a single, self-contained executable

**Manual build**:

```bash
# Build Frontend
cd web
npm install
npm run build
cd ..

# Build Binary
go build -o tun2socks .
```

**Build without embedded web interface**:

```bash
go build -tags="" -o tun2socks .
```

## Benchmarks

![benchmark](docs/benchmark.png)

For all scenarios of usage, tun2socks performs best.
See [benchmarks](https://github.com/xjasonlyu/tun2socks/wiki/Benchmarks) for more details.

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
- [wintun](https://git.zx2c4.com/wintun/) - Layer 3 TUN Driver for Windows

## License

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fxjasonlyu%2Ftun2socks.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fxjasonlyu%2Ftun2socks?ref=badge_large)

## Star History

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
