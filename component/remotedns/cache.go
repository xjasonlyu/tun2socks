package remotedns

import (
	"net"
	"sync"

	"github.com/ReneKroon/ttlcache/v2"
)

var _cache = ttlcache.NewCache()
var _mutex = sync.Mutex{}
var _ttl uint32 = 0

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

func insertNameIntoCache(ipVersion int, name string) net.IP {
	_mutex.Lock()
	defer _mutex.Unlock()
	var result net.IP = nil
	var ipnet *net.IPNet
	var nextAddress *net.IP
	var broadcastAddress net.IP
	if ipVersion == 4 {
		ipnet = _ip4net
		nextAddress = &_ip4NextAddress
		broadcastAddress = _ip4BroadcastAddress
	} else {
		ipnet = _ip6net
		nextAddress = &_ip6NextAddress
		broadcastAddress = _ip6BroadcastAddress
	}

	// Beginning from the pointer to the next most likely free IP, loop through the IP address space
	// until either a free IP is found or the space is exhausted
	passedBroadcastAddress := false
	for result == nil {
		if nextAddress.Equal(broadcastAddress) {
			*nextAddress = getNetworkAddress(ipnet)
			*nextAddress = incrementIp(ipnet.IP)

			// We have seen the broadcast address twice during looping
			// This means that our IP address space is exhausted
			if passedBroadcastAddress {
				return nil
			}
			passedBroadcastAddress = true
		}

		// This method is protected by a mutex, and we are only inserting elements into the cache here.
		_, err := _cache.Get((*nextAddress).String())
		if err == ttlcache.ErrNotFound {
			_ = _cache.Set((*nextAddress).String(), name)
			result = *nextAddress
		} else if err != nil { // Should never happen
			panic(nil)
		}

		*nextAddress = incrementIp(*nextAddress)
	}

	return result
}

func getCachedName(address net.IP) (interface{}, bool) {
	name, err := _cache.Get(address.String())
	return name, err == nil
}

func PostponeCacheExpiry(virtualIP net.IP) {
	if !IsEnabled() || virtualIP == nil {
		return
	}
	_ = _cache.Touch(virtualIP.String())
}
