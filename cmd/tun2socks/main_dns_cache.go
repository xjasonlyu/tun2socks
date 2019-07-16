// +build dnscache

package main

import (
	"flag"

	"github.com/xjasonlyu/tun2socks/common/dns/cache"
)

func init() {
	args.DisableDnsCache = flag.Bool("disableDNSCache", false, "Disable DNS cache")

	addPostFlagsInitFn(func() {
		if *args.DisableDnsCache {
			dnsCache = nil
		} else {
			dnsCache = cache.NewSimpleDnsCache()
		}
	})
}
