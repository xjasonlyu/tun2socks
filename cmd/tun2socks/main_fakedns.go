// +build fakedns

package main

import (
	"flag"

	"github.com/xjasonlyu/tun2socks/common/dns/fakedns"
)

func init() {
	args.EnableFakeDns = flag.Bool("fakeDns", false, "Enable fake DNS")
	args.DnsCacheSize = flag.Int("dnsCacheSize", 100, "Size of DNS LRU Cache")
	args.FakeDnsAddr = flag.String("fakeDnsAddr", ":53", "Listen address of fake DNS")
	args.FakeIPRange = flag.String("fakeIPRange", "198.18.0.1/16", "Fake IP CIDR range for DNS")
	args.FakeDnsHosts = flag.String("fakeDnsHosts", "", "DNS hosts mapping, e.g. 'example.com=1.1.1.1,example.net=2.2.2.2'")

	addPostFlagsInitFn(func() {
		if *args.EnableFakeDns {
			fakeDnsServer, err := fakedns.NewServer(*args.FakeIPRange, *args.FakeDnsHosts, *args.DnsCacheSize)
			if err != nil {
				panic("create fake dns server error")
			}
			if err := fakeDnsServer.StartServer(*args.FakeDnsAddr); err != nil {
				panic("cannot start fake dns server")
			}
			fakeDns = fakeDnsServer
		} else {
			fakeDns = nil
		}
	})
}
