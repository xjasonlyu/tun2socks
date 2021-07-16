package proxy

import (
	"net"
	"time"
)

const (
	tcpKeepAlivePeriod = 30 * time.Second
)

// setKeepAlive sets tcp keepalive option for tcp connection.
func setKeepAlive(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(tcpKeepAlivePeriod)
	}
}

// safeConnClose closes tcp connection safely.
func safeConnClose(c net.Conn, err error) {
	if c != nil && err != nil {
		c.Close()
	}
}
