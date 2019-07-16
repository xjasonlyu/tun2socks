config_route() {
	sleep 2
	sudo route delete default
	sudo route add default 10.255.0.1
}

config_route &
sudo ./build/tun2socks -tunAddr 10.255.0.2 -tunGw 10.255.0.1 -proxyServer 192.168.1.1:1086 -fakeDns -loglevel info -stats
