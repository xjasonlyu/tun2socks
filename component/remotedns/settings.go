package remotedns

import (
	"errors"
	"net"
	"time"
)

const (
	minTimeout = 30 * time.Second
)

var (
	enabled             = false
	ip4net              *net.IPNet
	ip6net              *net.IPNet
	ip4NextAddress      net.IP
	ip6NextAddress      net.IP
	ip4BroadcastAddress net.IP
	ip6BroadcastAddress net.IP
)

func IsEnabled() bool {
	return enabled
}

func SetCacheTimeout(timeout time.Duration) error {
	if timeout < minTimeout {
		timeout = minTimeout
	}
	ttl = uint32(timeout.Seconds())

	// Keep the value a little longer in cache than propagated via DNS
	return cache.SetTTL(timeout + 10*time.Second)
}

func SetNetwork(ipnet *net.IPNet) error {
	leadingOnes, _ := ipnet.Mask.Size()
	if len(ipnet.IP) == 4 {
		if leadingOnes > 30 {
			return errors.New("IPv4 remote DNS subnet too small")
		}
		ip4net = ipnet
	} else {
		if leadingOnes > 126 {
			return errors.New("IPv6 remote DNS subnet too small")
		}
		ip6net = ipnet
	}
	return nil
}

func Enable() {
	ip4NextAddress = incrementIp(getNetworkAddress(ip4net))
	ip4BroadcastAddress = getBroadcastAddress(ip4net)
	ip6NextAddress = incrementIp(getNetworkAddress(ip6net))
	ip6BroadcastAddress = getBroadcastAddress(ip6net)
	enabled = true
}
