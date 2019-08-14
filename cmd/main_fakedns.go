// +build fakeDNS

package main

import (
	"flag"

	"github.com/xjasonlyu/tun2socks/common/dns/fakedns"
)

func init() {
	args.EnableFakeDNS = flag.Bool("fakeDNS", false, "Enable fake DNS")
	args.DNSCacheSize = flag.Int("dnsCacheSize", 100, "Size of DNS LRU Cache")
	args.FakeDNSAddr = flag.String("fakeDNSAddr", ":53", "Listen address of fake DNS")
	args.FakeIPRange = flag.String("fakeIPRange", "198.18.0.1/16", "Fake IP CIDR range for DNS")
	args.FakeDNSHosts = flag.String("fakeDNSHosts", "", "DNS hosts mapping, e.g. 'example.com=1.1.1.1,example.net=2.2.2.2'")

	addPostFlagsInitFn(func() {
		if *args.EnableFakeDNS {
			fakeDNSServer, err := fakedns.NewServer(*args.FakeIPRange, *args.FakeDNSHosts, *args.DNSCacheSize)
			if err != nil {
				panic("create fake dns server error")
			}

			fakedns.ServeAddr = *args.FakeDNSAddr
			if err := fakeDNSServer.Start(); err != nil {
				panic("cannot start fake dns server")
			}
			fakeDNS = fakeDNSServer
		} else {
			fakeDNS = nil
		}
	})
}