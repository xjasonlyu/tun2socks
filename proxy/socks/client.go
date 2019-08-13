package socks

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"time"
)

var tcpTimeout = 30 * time.Second

func Dial(proxy, target string) (net.Conn, error) {
	c, err := net.DialTimeout("tcp", proxy, tcpTimeout)
	if err != nil {
		return nil, fmt.Errorf("%s connect error", proxy)
	}

	if _, err := ClientHandshake(c, ParseAddr(target), CmdConnect); err != nil {
		return nil, err
	}
	return c, nil
}

func DialUDP(proxy, target string) (_ net.PacketConn, _ net.Addr, err error) {
	c, err := net.DialTimeout("tcp", proxy, tcpTimeout)
	if err != nil {
		err = fmt.Errorf("%s connect error", proxy)
		return
	}

	// tcp set keepalive
	c.(*net.TCPConn).SetKeepAlive(true)
	c.(*net.TCPConn).SetKeepAlivePeriod(tcpTimeout)

	defer func() {
		if err != nil {
			c.Close()
		}
	}()

	bindAddr, err := ClientHandshake(c, ParseAddr(target), CmdUDPAssociate)
	if err != nil {
		err = fmt.Errorf("%v client hanshake error", err)
		return
	}

	addr, err := net.ResolveUDPAddr("udp", bindAddr.String())
	if err != nil {
		return
	}

	targetAddr, err := net.ResolveUDPAddr("udp", target)
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

type socksUDPConn struct {
	net.PacketConn
	tcpConn    net.Conn
	targetAddr net.Addr
}

func (c *socksUDPConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	packet, err := EncodeUDPPacket(c.targetAddr.String(), b)
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
	addr, payload, err := DecodeUDPPacket(b)
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
