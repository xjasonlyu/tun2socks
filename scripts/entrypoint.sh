#!/bin/sh

TUN="${TUN:-tun0}"
ETH="${ETH:-eth0}"
TUN_ADDR="${TUN_ADDR:-198.18.0.1}"
TUN_MASK="${TUN_MASK:-255.254.0.0}"
LOGLEVEL="${LOGLEVEL:-INFO}"

mk_tun() {
  # params
  NAME="$1"
  ADDR="$2"
  MASK="$3"
  # create tun device
  ip tuntap add mode tun dev "$NAME"
  ip addr add "$ADDR/$MASK" dev "$NAME"
  ip link set dev "$NAME" up
}

config_route() {
  # params
  TABLE="$1"
  TUN_IF="$2"
  ETH_IF="$3"
  TUN_EXCLUDED="$4"

  # add custom table
  printf "%s\t%s\n" 100 "$TABLE" >>/etc/iproute2/rt_tables

  # clone main route
  ip route show table main |
    while read -r route; do
      ip route add ${route%linkdown*} table "$TABLE"
    done

  # config default route
  ip route del default table "$TABLE"
  ip route add default dev "$TUN_IF" table "$TABLE"

  # policy routing
  tun=$(ip -4 addr show "$TUN_IF" | awk 'NR==2 {print $2}')
  eth=$(ip -4 addr show "$ETH_IF" | awk 'NR==2 {split($2,a,"/");print a[1]}')
  ip rule add from "$eth" to "$tun" priority 1000 prohibit
  ip rule add from "$eth" priority 2000 table main
  ip rule add from all priority 3000 table "$TABLE"

  # add tun excluded to route
  for addr in $(echo "$TUN_EXCLUDED" | tr ',' '\n'); do
    ip rule add to "$addr" table main
  done
}

main() {
  mk_tun "$TUN" "$TUN_ADDR" "$TUN_MASK"
  config_route "tun2socks" "$TUN" "$ETH" "$EXCLUDED"

  # execute extra commands
  if [ -n "$EXTRACMD" ]; then
    sh -c "$EXTRACMD"
  fi

  if [ -n "$API" ]; then
    ARGS="--api $API"
  fi

  if [ -n "$DNS" ]; then
    ARGS="$ARGS --dns $DNS"
  fi

  for item in $(echo "$HOSTS" | tr ',' '\n'); do
    ARGS="$ARGS --hosts $item"
  done

  exec tun2socks \
    --loglevel "$LOGLEVEL" \
    --interface "$ETH" \
    --device "$TUN" \
    --proxy "$PROXY" \
    $ARGS
}

main || exit 1
