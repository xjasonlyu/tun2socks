// +build redirect

package main

import (
	"github.com/xjasonlyu/tun2socks/core"
	"github.com/xjasonlyu/tun2socks/proxy/redirect"
)

func init() {
	registerHandlerCreator("redirect", func() {
		core.RegisterTCPConnHandler(redirect.NewTCPHandler(*args.ProxyServer))
		core.RegisterUDPConnHandler(redirect.NewUDPHandler(*args.ProxyServer, *args.UdpTimeout))
	})
}
