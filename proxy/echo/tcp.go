package echo

import (
	"io"
	"net"

	"github.com/xjasonlyu/tun2socks/core"
)

// An echo proxy, do nothing but echo back data to the sender, the handler was
// created for testing purposes, it may causes issues when more than one clients
// are connecting the handler simultaneously.
type tcpHandler struct{}

func NewTCPHandler() core.TCPConnHandler {
	return &tcpHandler{}
}

func (h *tcpHandler) echoBack(localConn net.Conn) {
	io.Copy(localConn, localConn)
}

func (h *tcpHandler) Handle(localConn net.Conn, target *net.TCPAddr) error {
	go h.echoBack(localConn)
	return nil
}
