#!/bin/bash

config_route() {
	sleep 2
	sudo route delete default
	sudo route add default 240.0.0.1
}

config_route &
sudo ./build/tun2socks -tunAddr 240.0.0.2 -tunGw 240.0.0.1 -proxyServer 192.168.1.1:1080 -fakeDNS -loglevel info -stats
