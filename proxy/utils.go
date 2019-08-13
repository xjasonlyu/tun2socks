package proxy

import (
	"net"
	"time"
)

type duplexConn interface {
	net.Conn
	CloseRead() error
	CloseWrite() error
}

func TCPCloseRead(conn net.Conn) {
	if c, ok := conn.(duplexConn); ok {
		c.CloseRead()
	}
}

func TCPCloseWrite(conn net.Conn) {
	if c, ok := conn.(duplexConn); ok {
		c.CloseWrite()
	}
}

func TCPKeepAlive(conn net.Conn) {
	if tcp, ok := conn.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(30 * time.Second)
	}
}
