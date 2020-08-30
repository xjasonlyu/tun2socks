#!/bin/sh

TUN="${TUN:-tun0}"
ETH="${ETH:-eth0}"
TUN_ADDR="${TUN_ADDR:-198.18.0.1}"
TUN_MASK="${TUN_MASK:-255.254.0.0}"
LOGLEVEL="${LOGLEVEL:-warning}"

MONITOR="${MONITOR:-1}"
MONITOR_ADDR="${MONITOR_ADDR:-0.0.0.0:80}"
FAKEDNS="${FAKEDNS:-1}"
BACKEND_DNS="${BACKEND_DNS:-8.8.8.8:53}"
HOSTS="${HOSTS:-localhost=127.0.0.1}"

# create tun device
ip tuntap add mode tun dev "$TUN"
ip addr add "$TUN_ADDR"/"$TUN_MASK" dev "$TUN"
ip link set dev "$TUN" up

# change default gateway
ip route del default > /dev/null
ip route add default via "$TUN_ADDR" dev "$TUN"

# add to ip route
for addr in $(echo "$EXCLUDED" | tr ',' '\n')
do
    ip route add "$addr" via "$ETH_ADDR"
done

if [ -n "$EXTRACMD" ]; then
    sh -c "$EXTRACMD"
fi

if [ "$MONITOR" -ne 0 ]; then
    ARGS="-monitor -monitorAddr $MONITOR_ADDR"
fi

if [ "$FAKEDNS" -ne 0 ]; then
    ARGS="$ARGS -fakeDNS -hosts $HOSTS -backendDNS $BACKEND_DNS"
fi

eval exec /tun2socks -loglevel "$LOGLEVEL" \
    -tunName "$TUN" -proxyServer "$PROXY" "$ARGS"
