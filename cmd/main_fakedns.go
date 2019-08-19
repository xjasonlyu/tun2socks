// +build fakedns

package main

import (
	"flag"

	"github.com/xjasonlyu/tun2socks/component/fakedns"
	"github.com/xjasonlyu/tun2socks/log"
)

func init() {
	args.EnableFakeDNS = flag.Bool("fakeDNS", false, "Enable fake DNS")
	args.FakeDNSAddr = flag.String("fakeDNSAddr", ":53", "Listen address of fake DNS")
	args.Hosts = flag.String("hosts", "", "DNS hosts mapping, e.g. 'example.com=1.1.1.1,example.net=2.2.2.2'")
	args.HijackDNS = flag.String("hijackDNS", "", "Hijack the specific DNS query to get a fake ip, e.g. '*:53', '8.8.8.8:53,8.8.4.4:53'")
	args.BackendDNS = flag.String("backendDNS", "8.8.8.8:53,8.8.4.4:53", "Backend DNS to resolve !TypeA or !ClassINET query (must support tcp)")

	registerInitFn(func() {
		if *args.EnableFakeDNS {
			var err error
			fakeDNS, err = fakedns.NewResolver(*args.FakeDNSAddr, *args.Hosts)
			if err != nil {
				log.Fatalf("Create fake DNS server failed: %v", err)
			}

			// Register backend DNS
			fakedns.RegisterDNS(*args.BackendDNS)

			// Start FakeDNS
			if err := fakeDNS.Start(); err != nil {
				log.Fatalf("Start fake DNS failed: %v", err)
			}
		} else {
			fakeDNS = nil
		}
	})
}
