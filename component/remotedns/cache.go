package remotedns

import (
	"net"
	"time"

	"github.com/patrickmn/go-cache"
)

var _cache = cache.New(30 * time.Second, 30 * time.Second)

func copyIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

func incrementIp(ip net.IP) net.IP {
	result := copyIP(ip)
	for i := len(result) - 1; i >= 0; i-- {
		result[i]++
		if result[i] != 0 {
			break
		}
	}
	return result
}

func getBroadcastAddress(ipnet *net.IPNet) net.IP {
	result := copyIP(ipnet.IP)
	for i := 0; i < len(ipnet.IP); i++ {
		result[i] |= ^ipnet.Mask[i]
	}
	return result
}

func getNetworkAddress(ipnet *net.IPNet) net.IP {
	result := copyIP(ipnet.IP)
	for i := 0; i < len(ipnet.IP); i++ {
		result[i] &= ipnet.Mask[i]
	}
	return result
}

func debug() {
	print("ip4net.ip: ", _ip4net.IP.String(), "\n")
	print("ip6net.ip: ", _ip6net.IP.String(), "\n")
	print("ip4next: ", _ip4NextAddress.String(), "\n")
	print("ip6next: ", _ip6NextAddress.String(), "\n")
	print("ip4broad: ", _ip4BroadcastAddress.String(), "\n")
	print("ip6broad: ", _ip6BroadcastAddress.String(), "\n")
}

func insertNameIntoCache(ipVersion int, name string) net.IP {
	var result net.IP = nil
	var ipnet *net.IPNet
	var nextAddress net.IP
	var broadcastAddress net.IP
	if ipVersion == 4 {
		ipnet = _ip4net
		nextAddress = _ip4NextAddress
		broadcastAddress = _ip4BroadcastAddress
	} else {
		ipnet = _ip6net
		nextAddress = _ip6NextAddress
		broadcastAddress = _ip6BroadcastAddress
	}

	// Beginning from the pointer to the next most likely free IP, loop through the IP address space
	// until either a free IP is found or the space is exhausted
	passedBroadcastAddress := false
	for result == nil {
		// We have seen the broadcast address twice during looping
		// This means that our IP address space is exhausted
		if nextAddress.Equal(broadcastAddress) {
			nextAddress = incrementIp(ipnet.IP)
			if passedBroadcastAddress {
				return nil
			}
			passedBroadcastAddress = true
		}

		_, found := getCachedName(nextAddress)
		if !found {
			result = nextAddress
		}
		nextAddress = incrementIp(nextAddress)
	}

	_cache.Set(string(result), name, cache.DefaultExpiration)

	return result
}

func getCachedName(address net.IP) (interface{}, bool) {
	name, found := _cache.Get(string(address))
	return name, found
}
