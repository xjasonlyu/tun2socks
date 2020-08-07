#!/bin/sh

TUN="${TUN:-tun0}"
ETH="${ETH:-eth0}"
ETHGW="${ETHGW:-172.16.1.1}"
TUNGW="${TUNGW:-198.18.0.1}"
PROXY="${PROXY:-172.16.1.2:1080}"
LOGLEVEL="${LOGLEVEL:-warning}"
EXCLUDED="${EXCLUDED:-172.16.1.2/32}"

MONITOR="${MONITOR:-0.0.0.0:80}"
FAKEDNS="${FAKEDNS:-1}"
BACKENDDNS="${BACKENDDNS:-8.8.8.8:53}"
HOSTS="${HOSTS:-localhost=127.0.0.1}"

# create tun device
ip tuntap add mode tun dev "$TUN"
ip addr add "$TUNGW"/24 dev "$TUN"
ip link set dev "$TUN" up

# change default gateway
ip route del default > /dev/null
ip route add default via "$TUNGW" dev "$TUN"

# add to ip route
for ip in $(echo "$EXCLUDED" | tr ',' '\n')
do
    ip route add "$ip" via "$ETHGW"
done

if [ -n "$EXTRACMD" ]; then
    sh -c "$EXTRACMD"
fi

if [ -n "$MONITOR" ]; then
    ARGS="-monitor -monitorAddr $MONITOR"
fi

if [ "$FAKEDNS" -ne 0 ]; then
    ARGS="$ARGS -fakeDNS -hosts $HOSTS -backendDNS $BACKENDDNS"
fi

exec /tun2socks -loglevel "$LOGLEVEL" \
    -tunName "$TUN" -proxyServer "$PROXY" "$ARGS"
