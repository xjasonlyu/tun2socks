<h1 align="center">tun2socks</h1>
<h3 align="center">A tun2socks powered by gVisor TCP/IP stack.</h3>

<p align="center">
  <a href="https://github.com/xjasonlyu/tun2socks/actions">
    <img src="https://img.shields.io/github/workflow/status/xjasonlyu/tun2socks/Go/master?style=flat-square" alt="GitHub Workflow">
  </a>
  <a href="https://hub.docker.com/r/xjasonlyu/tun2socks">
    <img src="https://img.shields.io/docker/pulls/xjasonlyu/tun2socks?style=flat-square" alt="Docker Pulls">
  </a>
  <a href="https://img.shields.io/github/go-mod/go-version/xjasonlyu/tun2socks">
    <img src="https://img.shields.io/github/go-mod/go-version/xjasonlyu/tun2socks?style=flat-square" alt="Go Version">
  </a>
  <a href="https://goreportcard.com/badge/github.com/xjasonlyu/tun2socks">
    <img src="https://goreportcard.com/badge/github.com/xjasonlyu/tun2socks?style=flat-square" alt="Go Report">
  </a>
  <a href="https://github.com/xjasonlyu/tun2socks/blob/master/LICENSE">
    <img src="https://img.shields.io/github/license/xjasonlyu/tun2socks?style=flat-square" alt="GitHub License">
  </a>
  <a href="https://img.shields.io/tokei/lines/github/xjasonlyu/tun2socks">
    <img src="https://img.shields.io/tokei/lines/github/xjasonlyu/tun2socks?style=flat-square" alt="Total Lines">
  </a>
  <a href="https://github.com/xjasonlyu/tun2socks/releases">
    <img src="https://img.shields.io/github/v/release/xjasonlyu/tun2socks?include_prereleases&style=flat-square" alt="Release">
  </a>
</p>

## Features

- ICMP echoing
- IPv6 support
- Optimized UDP transmission for game acceleration
- Pure Go implementation, no more CGO required
- Router mode, routing all the traffic in LAN
- Socks5, Shadowsocks protocol support for remote connections
- TCP/IP stack powered by [gVisor](https://github.com/google/gvisor)
- Up to 2.5Gbps throughput (10x faster than [v1](https://github.com/xjasonlyu/tun2socks/tree/v1))

## Requirements

| Target | Minimum | Recommended |
| :----- | :-----: | :---------: |
| System | linux darwin | linux |
| Memory | >20MB | >128MB |
| CPU | amd64 arm64 | amd64 |

## Performance

> iPerf3 tested on Debian 10 with i5-10500, 8G RAM

![iPerf3 Test](assets/iperf3.png)

## How to Build

### build from source code

Go compiler version >= 1.15 is required

```text
$ git clone https://github.com/xjasonlyu/tun2socks.git
$ cd tun2socks
$ make
```

### build docker image

```text
$ docker build -t tun2socks .
```

or

```text
$ docker build -t tun2socks -f ./docker/Dockerfile.aarch64 .
```

## QuickStart

Download from precompiled [Releases](https://github.com/xjasonlyu/tun2socks/releases).

<details>
  <summary><b>With Docker</b></summary>

> Since Go 1.12, the runtime now uses MADV_FREE to release unused memory on **linux**. This is more efficient but may result in higher reported RSS. The kernel will reclaim the unused data when it is needed. To revert to the Go 1.11 behavior (MADV_DONTNEED), set the environment variable GODEBUG=madvdontneed=1.

### create docker network (macvlan mode)

```shell script
docker network create -d macvlan \
  --subnet=172.20.1.0/25 \
  --gateway=172.20.1.1 \
  -o parent=eth0 \
  switch
```

### pull `tun2socks` docker image

```shell script
docker pull xjasonlyu/tun2socks:latest
```

### run as gateway

> DNS configuration is required.

```shell script
docker run -d \
  --network switch \
  --name tun2socks \
  --ip 172.20.1.2 \
  --privileged \
  --restart always \
  --sysctl net.ipv4.ip_forward=1 \
  -e PROXY=socks5://server:port \
  -e KEY=VALUE... \
  xjasonlyu/tun2socks:latest
```

### use docker-compose (recommended)

```yaml
version: '2.4'

services:
  tun2socks:
    image: xjasonlyu/tun2socks:latest
    cap_add:
      - NET_ADMIN
    devices:
        - '/dev/net/tun:/dev/net/tun'
    environment:
      # - GODEBUG=madvdontneed=1
      - PROXY=socks5://server:port
      - LOGLEVEL=INFO
      - API=api://:8080
      - DNS=dns://:53
      - HOSTS=localhost=127.0.0.1,router.local=172.20.1.1
      - EXCLUDED=1.1.1.1,1.0.0.1
      - EXTRACMD=
    networks:
      switch:
        ipv4_address: 172.20.1.2
    restart: always
    container_name: tun2socks

networks:
  switch:
    name: switch
    ipam:
      driver: default
      config:
        - subnet: '172.20.1.0/25'
          gateway: 172.20.1.1
    driver: macvlan
    driver_opts:
      parent: eth0
```
</details>

<details>
  <summary><b>With Linux</b></summary>

### create tun

```shell script
ip tuntap add mode tun dev tun0
ip addr add 198.18.0.1/15 dev tun0
ip link set dev tun0 up
```

### config policy routing

```shell script
echo "100 tun2socks" >> /etc/iproute2/rt_tables

ip route add default via 198.18.0.1 dev tun0 table tun2socks
ip route add 172.17.0.0/16 dev eth0 src 172.17.0.3 table tun2socks
ip route add 198.18.0.0/15 dev tun0 src 198.18.0.1 table tun2socks

ip rule add from 172.20.0.3 to 198.18.0.0/15 priority 1000 prohibit
ip rule add from 172.20.0.3 priority 2000 table main
ip rule add from all priority 3000 table tun2socks
```

### run

> bind to a specific interface to prevent traffic looping.

```shell script
./tun2socks --loglevel info --device tun://tun0 --proxy socks5://server:port --interface eth0
```
</details>

<details>
  <summary><b>With MacOS</b></summary>

### start tun2socks
```shell script
./tun2socks --loglevel info --device tun://utun123 --proxy socks5://server:port --interface eth0
```

### config interface

```shell script
sudo ifconfig utun123 198.18.0.1 netmask 255.255.255.255 198.18.0.1 up
```

### config route

```shell script
sudo route del default
sudo route add default 198.18.0.1
sudo route add ${proxy_server_ip} ${your_gateway}
```

### check route table

```shell script
netstat -nr
```
</details>

<details>
  <summary><b>With Script</b></summary>

> entrypoint.sh would take care of tun & routes.

```shell script
PROXY=socks5://server:port LOGLEVEL=INFO sh ./scripts/entrypoint.sh
```
</details>

## Details

<details>
  <summary><b>API Reference</b></summary>

| Path | Methods | Parameters | Description |
| :--- | :------ | :--------: | :---------- |
| `/logs` | GET | `level` | Get real-time logs |
| `/traffic` | GET | / | Get real-time traffic data |
| `/version` | GET | / | Get current version |
| `/connections` | GET | `interval` | Get all connections |
| `/connections` | DELETE | / | Close all connections |
| `/connections/{id}` | DELETE | / | Close connection by `id` |

</details>

<details>
  <summary><b>Help Text</b></summary>

```text
NAME:
   tun2socks - A tun2socks powered by gVisor TCP/IP stack.

USAGE:
   tun2socks [global options] [arguments...]

GLOBAL OPTIONS:
   --api value                  URL of external API to listen
   --device value, -d value     URL of device to open
   --dns value                  URL of fake DNS to listen
   --hosts value                Extra hosts mapping
   --interface value, -i value  Bind interface to dial
   --loglevel value, -l value   Set logging level (default: "INFO")
   --proxy value, -p value      URL of proxy to dial
   --version, -v                Print current version (default: false)
   --help, -h                   show help (default: false)
```

</details>

<details>
  <summary><b>Proxy URL</b></summary>

| Protocol | Scheme | Examples |
| :------- | :----- | :------- |
| direct | `direct` | `direct://` |
| socks5 | `socks5` | `socks5://username:password@server:port` |
| shadowsocks | `ss`, `shadowsocks` | `ss://method:password@server:port` |

</details>

## Credits

- [Dreamacro/clash](https://github.com/Dreamacro/clash)
- [google/gvisor](https://github.com/google/gvisor)
- [majek/slirpnetstack](https://github.com/majek/slirpnetstack)
- [WireGuard/wireguard-go](https://github.com/WireGuard/wireguard-go)

## Known Issues

Due to the implementation of pure Go, the memory usage is higher than the previous version.
If you are sensitive to memory, please go back to [v1](https://github.com/xjasonlyu/tun2socks/tree/v1).

## TODO

- [ ] Windows support
- [ ] FreeBSD support
- [ ] OpenBSD support
