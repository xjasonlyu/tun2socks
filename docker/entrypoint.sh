#!/bin/sh

TUN="${TUN:-tun0}"
ADDR="${ADDR:-198.18.0.1/15}"
LOGLEVEL="${LOGLEVEL:-info}"

# default values
TABLE="${TABLE:-0x22b}"
FWMARK="${FWMARK:-0x22b}"
CLONE_MAIN="${CLONE_MAIN:-1}"

create_tun() {
  # create tun device
  ip tuntap add mode tun dev "$TUN"
  ip addr add "$ADDR" dev "$TUN"
  ip link set dev "$TUN" up
}

create_table() {
  if [ "$CLONE_MAIN" -ne 0 ]; then
    # clone main route table
    ip route show table main |
      while read -r route; do
        ip route add ${route%linkdown*} table "$TABLE"
      done
    # replace default route
    ip route replace default dev "$TUN" table "$TABLE"
  else
    # just add default route
    ip route add default dev "$TUN" table "$TABLE"
  fi
}

config_route() {
  # policy routing
  ip rule add not fwmark "$FWMARK" table "$TABLE"
  ip rule add fwmark "$FWMARK" to "$ADDR" prohibit

  # add tun included routes
  for addr in $(echo "$TUN_INCLUDED_ROUTES" | tr ',' '\n'); do
    ip rule add to "$addr" table "$TABLE"
  done

  # add tun excluded routes
  for addr in $(echo "$TUN_EXCLUDED_ROUTES" | tr ',' '\n'); do
    ip rule add to "$addr" table main
  done
}

run() {
  create_tun
  create_table
  config_route

  # execute extra commands
  if [ -n "$EXTRA_COMMANDS" ]; then
    sh -c "$EXTRA_COMMANDS"
  fi

  if [ -n "$MTU" ]; then
    ARGS="--mtu $MTU"
  fi

  if [ -n "$RESTAPI" ]; then
    ARGS="$ARGS --restapi $RESTAPI"
  fi

  if [ -n "$UDP_TIMEOUT" ]; then
    ARGS="$ARGS --udp-timeout $UDP_TIMEOUT"
  fi

  if [ -n "$TCP_SNDBUF" ]; then
    ARGS="$ARGS --tcp-sndbuf $TCP_SNDBUF"
  fi

  if [ -n "$TCP_RCVBUF" ]; then
    ARGS="$ARGS --tcp-rcvbuf $TCP_RCVBUF"
  fi

  if [ "$TCP_AUTO_TUNING" = 1 ]; then
    ARGS="$ARGS --tcp-auto-tuning"
  fi

  if [ -n "$MULTICAST_GROUPS" ]; then
    ARGS="$ARGS --multicast-groups $MULTICAST_GROUPS"
  fi

  exec tun2socks \
    --loglevel "$LOGLEVEL" \
    --fwmark "$FWMARK" \
    --device "$TUN" \
    --proxy "$PROXY" \
    $ARGS
}

run || exit 1
