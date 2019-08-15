// +build fakeDNS

package main

import (
	"flag"
	"fmt"

	"github.com/xjasonlyu/tun2socks/common/dns/fakedns"
)

func init() {
	args.EnableFakeDNS = flag.Bool("fakeDNS", false, "Enable fake DNS")
	args.DNSCacheSize = flag.Int("dnsCacheSize", 100, "Size of DNS LRU Cache")
	args.FakeDNSAddr = flag.String("fakeDNSAddr", ":53", "Listen address of fake DNS")
	args.FakeIPRange = flag.String("fakeIPRange", "198.18.0.0/15", "Fake IP CIDR range for DNS")
	args.FakeDNSHosts = flag.String("fakeDNSHosts", "", "DNS hosts mapping, e.g. 'example.com=1.1.1.1,example.net=2.2.2.2'")

	addPostFlagsInitFn(func() {
		if *args.EnableFakeDNS {
			var err error
			fakeDNS, err = fakedns.NewServer(*args.FakeIPRange, *args.FakeDNSHosts, *args.DNSCacheSize)
			if err != nil {
				panic(fmt.Errorf("create fake dns server failed: %v", err))
			}

			// Set fakeDNS variables
			fakedns.ServeAddr = *args.FakeDNSAddr

			// Start fakeDNS server
			fakeDNS.Start()
		} else {
			fakeDNS = nil
		}
	})
}
