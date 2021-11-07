package remotedns

import (
	"net"
	"time"
)

var _enabled = false
var _ip4net, _ip6net *net.IPNet
var _ip4NextAddress, _ip4BroadcastAddress , _ip6NextAddress, _ip6BroadcastAddress net.IP
var _minTimeout = 30 * time.Second

func IsEnabled() bool {
	return _enabled
}

func SetCacheTimeout(timeout time.Duration) error {
	if timeout < _minTimeout {
		timeout = _minTimeout
	}
	_ttl = uint32(timeout.Seconds())

	// Keep the value a little longer in cache than propagated via DNS
	return _cache.SetTTL(timeout + 10 * time.Second)
}

func SetNetwork(ipnet *net.IPNet) {
	if len(ipnet.IP) == 4 {
		_ip4net = ipnet
	} else {
		_ip6net = ipnet
	}
}

func Enable() {
	_ip4NextAddress = incrementIp(getNetworkAddress(_ip4net))
	_ip4BroadcastAddress = getBroadcastAddress(_ip4net)
	_ip6NextAddress = incrementIp(getNetworkAddress(_ip6net))
	_ip6BroadcastAddress = getBroadcastAddress(_ip6net)
	_enabled = true
}
