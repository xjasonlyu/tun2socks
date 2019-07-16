#!/bin/sh

set -x

config_route() {
  sleep 2
  sudo ip addr add 10.255.0.1/24 dev tun1
  sudo ip link set dev tun1 up
  sudo ip route del default
  sudo ip route add default via 10.255.0.1
}

config_route &
sudo ./build/tun2socks -tunAddr 10.255.0.2 -tunGw 10.255.0.1 -proxyServer 192.168.1.1:1086 -fakeDns -loglevel info -stats
