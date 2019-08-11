package socks

import (
	"net"
	"time"

	"github.com/xjasonlyu/tun2socks/common/log"
)

type duplexConn interface {
	net.Conn
	CloseRead() error
	CloseWrite() error
}

func tcpCloseRead(conn net.Conn) {
	if c, ok := conn.(duplexConn); ok {
		log.Warnf("ok!----")
		c.CloseRead()
	}
}

func tcpKeepAlive(conn net.Conn) {
	if tcp, ok := conn.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(30 * time.Second)
	}
}
