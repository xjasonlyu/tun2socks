#!/bin/sh

TUN="${TUN:-tun0}"
ETH="${ETH:-eth0}"
ETHGW="${ETHGW:-172.16.1.1}"
TUNGW="${TUNGW:-240.0.0.1}"
PROXY="${PROXY:-172.16.1.2:1080}"
MONITOR="${MONITOR:-0.0.0.0:80}"
EXCLUDED="${EXCLUDED:-172.16.1.2/32}"
LOGLEVEL="${LOGLEVEL:-warning}"
BACKENDDNS="${BACKENDDNS:-8.8.8.8:53}"
HOSTS="${HOSTS:-localhost=127.0.0.1}"

# create tun device
ip tuntap add mode tun dev $TUN
ip addr add $TUNGW/24 dev $TUN
ip link set dev $TUN up

# change default gateway
ip route del default &> /dev/null
ip route add default via $TUNGW dev $TUN

# add to ip route
for ip in $(echo $EXCLUDED | tr ',' '\n')
do
    ip route add $ip via $ETHGW
done

# DNS settings
echo "nameserver $TUNGW" > /etc/resolv.conf

tun2socks -loglevel $LOGLEVEL \
    -tunName $TUN -proxyServer $PROXY \
    -monitor -monitorAddr $MONITOR \
    -fakeDNS -hosts $HOSTS \
    -backendDNS $BACKENDDNS
