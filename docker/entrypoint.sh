#!/bin/sh

TUN="${TUN:-tun0}"
TUN_ADDR="${TUN_ADDR:-198.18.0.1/15}"
LOGLEVEL="${LOGLEVEL:-info}"

# default values
TABLE="${TABLE:-0x22b}"
FWMARK="${FWMARK:-0x22b}"

create_tun() {
  # create tun device
  ip tuntap add mode tun dev "$TUN"
  ip addr add "$TUN_ADDR" dev "$TUN"
  ip link set dev "$TUN" up
}

config_route() {
  # clone main route table
  ip route show table main |
    while read -r route; do
      ip route add ${route%linkdown*} table "$TABLE"
    done

  # replace default route
  ip route replace default dev "$TUN" table "$TABLE"

  # policy routing
  ip rule add not fwmark "$FWMARK" table "$TABLE"
  ip rule add fwmark "$FWMARK" to "$TUN_ADDR" prohibit

  # add tun included routes
  for addr in $(echo "$TUN_INCLUDED_ROUTES" | tr ',' '\n'); do
    ip rule add to "$addr" table "$TABLE"
  done

  # add tun excluded routes
  for addr in $(echo "$TUN_EXCLUDED_ROUTES" | tr ',' '\n'); do
    ip rule add to "$addr" table main
  done
}

main() {
  create_tun
  config_route

  # execute extra commands
  if [ -n "$EXTRA_COMMANDS" ]; then
    sh -c "$EXTRA_COMMANDS"
  fi

  if [ -n "$MTU" ]; then
    ARGS="--mtu $MTU"
  fi

  if [ -n "$STATS" ]; then
    ARGS="$ARGS --stats $STATS"
  fi

  if [ -n "$TOKEN" ]; then
    ARGS="$ARGS --token $TOKEN"
  fi

  exec tun2socks \
    --loglevel "$LOGLEVEL" \
    --fwmark "$FWMARK" \
    --device "$TUN" \
    --proxy "$PROXY" \
    $ARGS
}

main || exit 1
