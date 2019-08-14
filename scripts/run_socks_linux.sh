#!/bin/bash

set -x

config_route() {
  sleep 2
  sudo ip addr add 240.0.0.1/24 dev tun0
  sudo ip link set dev tun0 up
  sudo ip route del default
  sudo ip route add default via 240.0.0.1
}

config_route &
sudo ./build/tun2socks -tunAddr 240.0.0.2 -tunGw 240.0.0.1 -proxyServer 192.168.1.1:1080 -fakeDNS -loglevel info -stats
