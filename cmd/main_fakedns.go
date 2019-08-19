// +build fakedns

package main

import (
	"flag"

	"github.com/xjasonlyu/tun2socks/component/fakedns"
	"github.com/xjasonlyu/tun2socks/log"
)

func init() {
	args.EnableFakeDNS = flag.Bool("fakeDNS", false, "Enable fake DNS")
	args.Hosts = flag.String("hosts", "", "DNS hosts mapping, e.g. 'example.com=1.1.1.1,example.net=2.2.2.2'")
	args.HijackDNS = flag.String("hijackDNS", "", "Hijack the specific DNS query to get a fake ip, e.g. '*:53', '8.8.8.8:53,8.8.4.4:53'")
	args.BackendDNS = flag.String("backendDNS", "8.8.8.8:53,8.8.4.4:53", "Backend DNS to resolve !TypeA or !ClassINET query (must support tcp)")

	registerInitFn(func() {
		if *args.EnableFakeDNS {
			resolver, err := fakedns.NewResolver(*args.Hosts, *args.BackendDNS)
			if err != nil {
				log.Fatalf("Create fake DNS server failed: %v", err)
			}
			fakeDNS = resolver
		} else {
			fakeDNS = nil
		}
	})
}
