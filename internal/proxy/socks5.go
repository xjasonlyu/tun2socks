package proxy

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"

	"github.com/xjasonlyu/clash/component/dialer"
	"github.com/xjasonlyu/clash/component/socks5"
	"github.com/xjasonlyu/tun2socks/internal/adapter"
)

type Socks5 struct {
	*Base

	user string
	pass string
}

func NewSocks5(url *url.URL, user, pass string) (*Socks5, error) {
	return &Socks5{
		Base: &Base{
			url: url,
		},
		user: user,
		pass: pass,
	}, nil
}

func (ss *Socks5) DialContext(ctx context.Context, metadata *adapter.Metadata) (c net.Conn, err error) {
	c, err = dialer.DialContext(ctx, "tcp", ss.Addr())
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", ss.Addr(), err)
	}
	tcpKeepAlive(c)

	defer func() {
		if err != nil {
			c.Close()
		}
	}()

	var user *socks5.User
	if ss.user != "" {
		user = &socks5.User{
			Username: ss.user,
			Password: ss.pass,
		}
	}

	if _, err := socks5.ClientHandshake(c, metadata.SerializesSocksAddr(), socks5.CmdConnect, user); err != nil {
		return nil, fmt.Errorf("client hanshake: %w", err)
	}

	return
}

func (ss *Socks5) DialUDP(_ *adapter.Metadata) (_ net.PacketConn, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), tcpConnectTimeout)
	defer cancel()
	c, err := dialer.DialContext(ctx, "tcp", ss.Addr())
	if err != nil {
		err = fmt.Errorf("connect to %s: %w", ss.Addr(), err)
		return
	}
	tcpKeepAlive(c)

	defer func() {
		if err != nil {
			c.Close()
		}
	}()

	var user *socks5.User
	if ss.user != "" {
		user = &socks5.User{
			Username: ss.user,
			Password: ss.pass,
		}
	}

	// The UDP ASSOCIATE request is used to establish an association within
	// the UDP relay process to handle UDP datagrams.  The DST.ADDR and
	// DST.PORT fields contain the address and port that the client expects
	// to use to send UDP datagrams on for the association.  The server MAY
	// use this information to limit access to the association.  If the
	// client is not in possession of the information at the time of the UDP
	// ASSOCIATE, the client MUST use a port number and address of all
	// zeros. RFC1928
	var targetAddr socks5.Addr = []byte{socks5.AtypIPv4, 0, 0, 0, 0, 0, 0}

	bindAddr, err := socks5.ClientHandshake(c, targetAddr, socks5.CmdUDPAssociate, user)
	if err != nil {
		return nil, fmt.Errorf("client hanshake: %w", err)
	}

	pc, err := dialer.ListenPacket("udp", "")
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

	return &socksPacketConn{PacketConn: pc, rAddr: bindAddr.UDPAddr(), tcpConn: c}, nil
}

type socksPacketConn struct {
	net.PacketConn

	rAddr   net.Addr
	tcpConn net.Conn
}

func (pc *socksPacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	var packet []byte
	if m, ok := addr.(*adapter.Metadata); ok {
		packet, err = socks5.EncodeUDPPacket(m.SerializesSocksAddr(), b)
	} else {
		packet, err = socks5.EncodeUDPPacket(socks5.ParseAddrToSocksAddr(addr), b)
	}

	if err != nil {
		return
	}
	return pc.PacketConn.WriteTo(packet, pc.rAddr)
}

func (pc *socksPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, _, e := pc.PacketConn.ReadFrom(b)
	if e != nil {
		return 0, nil, e
	}
	addr, payload, err := socks5.DecodeUDPPacket(b)
	if err != nil {
		return 0, nil, err
	}

	udpAddr := addr.UDPAddr()
	if udpAddr == nil {
		return 0, nil, fmt.Errorf("convert %s to UDPAddr failed", addr)
	}

	// due to DecodeUDPPacket is mutable, record addr length
	copy(b, payload)
	return n - len(addr) - 3, udpAddr, nil
}

func (pc *socksPacketConn) Close() error {
	pc.tcpConn.Close()
	return pc.PacketConn.Close()
}
