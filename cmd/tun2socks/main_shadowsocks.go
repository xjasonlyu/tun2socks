// +build shadowsocks

package main

import (
	"flag"
	"net"
	"strings"

	sscore "github.com/shadowsocks/go-shadowsocks2/core"

	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/core"
	"github.com/xjasonlyu/tun2socks/proxy/shadowsocks"
)

func init() {
	args.addFlag(fProxyServer)
	args.addFlag(fUdpTimeout)

	args.ProxyCipher = flag.String("proxyCipher", "AEAD_CHACHA20_POLY1305", "Cipher used for Shadowsocks proxy, available ciphers: "+strings.Join(sscore.ListCipher(), " "))
	args.ProxyPassword = flag.String("proxyPassword", "", "Password used for Shadowsocks proxy")

	registerHandlerCreator("shadowsocks", func() {
		// Verify proxy server address.
		proxyAddr, err := net.ResolveTCPAddr("tcp", *args.ProxyServer)
		if err != nil {
			log.Fatalf("invalid proxy server address: %v", err)
		}
		proxyHost := proxyAddr.IP.String()
		proxyPort := uint16(proxyAddr.Port)

		if *args.ProxyCipher == "" || *args.ProxyPassword == "" {
			log.Fatalf("invalid cipher or password")
		}
		core.RegisterTCPConnHandler(shadowsocks.NewTCPHandler(core.ParseTCPAddr(proxyHost, proxyPort).String(), *args.ProxyCipher, *args.ProxyPassword, fakeDns))
		core.RegisterUDPConnHandler(shadowsocks.NewUDPHandler(core.ParseUDPAddr(proxyHost, proxyPort).String(), *args.ProxyCipher, *args.ProxyPassword, *args.UdpTimeout, dnsCache, fakeDns))
	})
}
