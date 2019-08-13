package proxy

import (
	"net"

	"github.com/xjasonlyu/tun2socks/core"
	. "github.com/xjasonlyu/tun2socks/proxy/utils"
)

type udpHandler struct {
	excludeApps []string

	proxyHandler  core.UDPConnHandler
	directHandler core.UDPConnHandler
}

func NewUDPHandler(proxyHandler, directHandler core.UDPConnHandler, excludeApps []string) core.UDPConnHandler {
	return &udpHandler{
		excludeApps:   excludeApps,
		proxyHandler:  proxyHandler,
		directHandler: directHandler,
	}
}

func (h *udpHandler) isExcludeApp(name string) bool {
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

func (h *udpHandler) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	process := GetProcessName(conn.LocalAddr())
	if h.isExcludeApp(process) {
		return h.directHandler.Connect(conn, target)
	} else {
		return h.proxyHandler.Connect(conn, target)
	}
}

func (h *udpHandler) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	process := GetProcessName(conn.LocalAddr())
	if h.isExcludeApp(process) {
		return h.directHandler.ReceiveTo(conn, data, addr)
	} else {
		return h.proxyHandler.ReceiveTo(conn, data, addr)
	}
}
