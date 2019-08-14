package proxy

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/xjasonlyu/tun2socks/common/dns"
	"github.com/xjasonlyu/tun2socks/proxy/socks"
)

// DNS lookup
func lookupHost(fakeDNS dns.FakeDNS, target net.Addr) (targetHost string, err error) {
	var targetIP net.IP
	switch addr := target.(type) {
	case *net.TCPAddr:
		targetIP = addr.IP
	case *net.UDPAddr:
		targetIP = addr.IP
	default:
		err = errors.New("invalid target type")
		return
	}

	targetHost = targetIP.String()
	// Replace with a domain name if target address IP is a fake IP
	if fakeDNS != nil {
		if host, exist := fakeDNS.IPToHost(targetIP); exist {
			targetHost = host
		}
	}
	return
}

// TCP functions
type duplexConn interface {
	net.Conn
	CloseRead() error
	CloseWrite() error
}

func tcpCloseRead(conn net.Conn) {
	if c, ok := conn.(duplexConn); ok {
		c.CloseRead()
	}
}

func tcpCloseWrite(conn net.Conn) {
	if c, ok := conn.(duplexConn); ok {
		c.CloseWrite()
	}
}

func tcpKeepAlive(conn net.Conn) {
	if tcp, ok := conn.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(30 * time.Second)
	}
}

// Socks dialer
func dial(proxy, target string) (net.Conn, error) {
	c, err := net.DialTimeout("tcp", proxy, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("%s connect error", proxy)
	}

	targetAddr := socks.ParseAddr(target)
	if targetAddr == nil {
		return nil, fmt.Errorf("target address parse error")
	}

	if _, err := socks.ClientHandshake(c, targetAddr, socks.CmdConnect); err != nil {
		return nil, err
	}
	return c, nil
}

func dialUDP(proxy, target string) (_ net.PacketConn, _ net.Addr, err error) {
	c, err := net.DialTimeout("tcp", proxy, 30*time.Second)
	if err != nil {
		err = fmt.Errorf("%s connect error", proxy)
		return
	}

	// tcp set keepalive
	tcpKeepAlive(c)

	defer func() {
		if err != nil {
			c.Close()
		}
	}()

	targetAddr := socks.ParseAddr(target)
	if targetAddr == nil {
		err = fmt.Errorf("target address parse error")
		return
	}

	bindAddr, err := socks.ClientHandshake(c, targetAddr, socks.CmdUDPAssociate)
	if err != nil {
		err = fmt.Errorf("%v client hanshake error", err)
		return
	}

	addr, err := net.ResolveUDPAddr("udp", bindAddr.String())
	if err != nil {
		return
	}

	pc, err := net.ListenPacket("udp", "")
	if err != nil {
		return
	}

	go func() {
		io.Copy(ioutil.Discard, c)
		c.Close()
		// A UDP association terminates when the TCP connection that the UDP
		// ASSOCIATE request arrived on terminates. RFC1928
		pc.Close()
	}()

	return &socksUDPConn{PacketConn: pc, tcpConn: c, targetAddr: targetAddr}, addr, nil
}

// Socks wrapped UDPConn
type socksUDPConn struct {
	net.PacketConn
	tcpConn    net.Conn
	targetAddr socks.Addr
}

func (c *socksUDPConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	packet, err := socks.EncodeUDPPacket(c.targetAddr, b)
	if err != nil {
		return
	}
	return c.PacketConn.WriteTo(packet, addr)
}

func (c *socksUDPConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, a, e := c.PacketConn.ReadFrom(b)
	if e != nil {
		return 0, nil, e
	}
	addr, payload, err := socks.DecodeUDPPacket(b)
	if err != nil {
		return 0, nil, err
	}
	copy(b, payload)
	return n - len(addr) - 3, a, nil
}

func (c *socksUDPConn) Close() error {
	c.tcpConn.Close()
	return c.PacketConn.Close()
}
