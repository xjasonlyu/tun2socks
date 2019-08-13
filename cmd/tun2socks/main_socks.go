// +build socks

package main

import (
	"flag"
	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/core"
	"github.com/xjasonlyu/tun2socks/proxy/socks"
	"net"
	"time"
)

func init() {
	args.ProxyServer = flag.String("proxyServer", "1.2.3.4:1087", "Proxy server address")
	args.UdpTimeout = flag.Duration("udpTimeout", 60*time.Second, "UDP session timeout")

	registerHandlerCreator(func() {
		// Verify proxy server address.
		proxyAddr, err := net.ResolveTCPAddr("tcp", *args.ProxyServer)
		if err != nil {
			log.Fatalf("invalid proxy server address: %v", err)
		}
		proxyHost := proxyAddr.IP.String()
		proxyPort := uint16(proxyAddr.Port)

		core.RegisterTCPConnHandler(socks.NewTCPHandler(proxyHost, proxyPort, fakeDns, sessionStater))
		core.RegisterUDPConnHandler(socks.NewUDPHandler(proxyHost, proxyPort, *args.UdpTimeout, fakeDns, sessionStater))
	})
}
