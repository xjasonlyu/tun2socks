package proxy

import (
	"net"

	"github.com/xjasonlyu/tun2socks/core"
	. "github.com/xjasonlyu/tun2socks/proxy/utils"
)

type tcpHandler struct {
	excludeApps []string

	proxyHandler  core.TCPConnHandler
	directHandler core.TCPConnHandler
}

func NewTCPHandler(proxyHandler, directHandler core.TCPConnHandler, excludeApps []string) core.TCPConnHandler {
	return &tcpHandler{
		excludeApps:   excludeApps,
		proxyHandler:  proxyHandler,
		directHandler: directHandler,
	}
}

func (h *tcpHandler) isExcludeApp(name string) bool {
	if name == "" {
		return false
	}
	for _, app := range h.excludeApps {
		if app == name {
			return true
		}
	}
	return false
}

func (h *tcpHandler) Handle(localConn net.Conn, target *net.TCPAddr) error {
	process := GetProcessName(localConn.LocalAddr())
	if h.isExcludeApp(process) {
		return h.directHandler.Handle(localConn, target)
	} else {
		return h.proxyHandler.Handle(localConn, target)
	}
}
