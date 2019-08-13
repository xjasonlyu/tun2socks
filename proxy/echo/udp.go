package echo

import (
	"net"

	"github.com/xjasonlyu/tun2socks/common/log"
	"github.com/xjasonlyu/tun2socks/core"
)

// An echo server, do nothing but echo back data to the sender.
type udpHandler struct{}

func NewUDPHandler() core.UDPConnHandler {
	return &udpHandler{}
}

func (h *udpHandler) Connect(localConn core.UDPConn, target *net.UDPAddr) error {
	return nil
}

func (h *udpHandler) ReceiveTo(localConn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	// Dispatch to another goroutine, otherwise will result in deadlock.
	payload := append([]byte(nil), data...)
	go func(b []byte) {
		_, err := localConn.WriteFrom(b, addr)
		if err != nil {
			log.Warnf("failed to echo back data: %v", err)
		}
	}(payload)
	return nil
}
