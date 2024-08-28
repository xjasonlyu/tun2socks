package internal

import (
	"net"
	"time"

	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/transport/socks5"
)

const tcpKeepAlivePeriod = 30 * time.Second

// SetKeepAlive sets the tcp keepalive option for the tcp connection.
func SetKeepAlive(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(tcpKeepAlivePeriod)
	}
}

// SafeConnClose closes the given tcp connection safely.
func SafeConnClose(c net.Conn, err error) {
	if c != nil && err != nil {
		c.Close()
	}
}

// SerializeSocksAddr serializes *metadata.Metadata to socks5.Addr.
func SerializeSocksAddr(m *M.Metadata) socks5.Addr {
	return socks5.SerializeAddr("", m.DstIP, m.DstPort)
}
