package remotedns

import (
	"net"
	"sync"

	"github.com/ReneKroon/ttlcache/v2"
)

var (
	cache        = ttlcache.NewCache()
	mutex        = sync.Mutex{}
	ttl   uint32 = 0
)

func PostponeCacheExpiry(virtualIP net.IP) {
	if !IsEnabled() || virtualIP == nil {
		return
	}
	_ = cache.Touch(virtualIP.String())
}

func insertNameIntoCache(ipVersion int, name string) net.IP {
	mutex.Lock()
	defer mutex.Unlock()
	var result net.IP = nil
	var ipnet *net.IPNet
	var nextAddress *net.IP
	var broadcastAddress net.IP
	if ipVersion == 4 {
		ipnet = ip4net
		nextAddress = &ip4NextAddress
		broadcastAddress = ip4BroadcastAddress
	} else {
		ipnet = ip6net
		nextAddress = &ip6NextAddress
		broadcastAddress = ip6BroadcastAddress
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
		_, err := cache.Get((*nextAddress).String())
		if err == ttlcache.ErrNotFound {
			_ = cache.Set((*nextAddress).String(), name)
			result = *nextAddress
		} else if err != nil { // Should never happen
			panic(nil)
		}

		*nextAddress = incrementIp(*nextAddress)
	}

	return result
}

func getCachedName(address net.IP) (interface{}, bool) {
	name, err := cache.Get(address.String())
	return name, err == nil
}
