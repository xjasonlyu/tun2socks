// +build echo

package main

import (
	"github.com/xjasonlyu/tun2socks/core"
	"github.com/xjasonlyu/tun2socks/proxy/echo"
)

func init() {
	registerHandlerCreator("echo", func() {
		core.RegisterTCPConnHandler(echo.NewTCPHandler())
		core.RegisterUDPConnHandler(echo.NewUDPHandler())
	})
}
