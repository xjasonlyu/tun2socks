package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/xjasonlyu/tun2socks/v2/dialer"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy/proto"
	"github.com/xjasonlyu/tun2socks/v2/transport/shadowsocks/core"
	obfs "github.com/xjasonlyu/tun2socks/v2/transport/simple-obfs"
	"github.com/xjasonlyu/tun2socks/v2/transport/socks5"
)

var _ Proxy = (*Shadowsocks)(nil)

type Shadowsocks struct {
	*Base

	cipher core.Cipher

	// simple-obfs plugin
	obfsMode, obfsHost string
}

func NewShadowsocks(addr, method, password, obfsMode, obfsHost string) (*Shadowsocks, error) {
	cipher, err := core.PickCipher(method, nil, password)
	if err != nil {
		return nil, fmt.Errorf("ss initialize: %w", err)
	}

	return &Shadowsocks{
		Base: &Base{
			addr:  addr,
			proto: proto.Shadowsocks,
		},
		cipher:   cipher,
		obfsMode: obfsMode,
		obfsHost: obfsHost,
	}, nil
}

func (ss *Shadowsocks) DialContext(ctx context.Context, metadata *M.Metadata) (c net.Conn, err error) {
	c, err = dialer.DialContext(ctx, "tcp", ss.Addr())
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", ss.Addr(), err)
	}
	setKeepAlive(c)

	defer func(c net.Conn) {
		safeConnClose(c, err)
	}(c)

	switch ss.obfsMode {
	case "tls":
		c = obfs.NewTLSObfs(c, ss.obfsHost)
	case "http":
		_, port, _ := net.SplitHostPort(ss.addr)
		c = obfs.NewHTTPObfs(c, ss.obfsHost, port)
	}

	c = ss.cipher.StreamConn(c)
	_, err = c.Write(serializeSocksAddr(metadata))
	return c, err
}

func (ss *Shadowsocks) DialUDP(*M.Metadata) (net.PacketConn, error) {
	pc, err := dialer.ListenPacket("udp", "")
	if err != nil {
		return nil, fmt.Errorf("listen packet: %w", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", ss.Addr())
	if err != nil {
		return nil, fmt.Errorf("resolve udp address %s: %w", ss.Addr(), err)
	}

	pc = ss.cipher.PacketConn(pc)
	return &ssPacketConn{PacketConn: pc, rAddr: udpAddr}, nil
}

type ssPacketConn struct {
	net.PacketConn

	rAddr net.Addr
}

func (pc *ssPacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	var packet []byte
	if ma, ok := addr.(*M.Addr); ok {
		packet, err = socks5.EncodeUDPPacket(serializeSocksAddr(ma.Metadata()), b)
	} else {
		packet, err = socks5.EncodeUDPPacket(socks5.ParseAddr(addr), b)
	}

	if err != nil {
		return n, err
	}
	return pc.PacketConn.WriteTo(packet[3:], pc.rAddr)
}

func (pc *ssPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, _, err := pc.PacketConn.ReadFrom(b)
	if err != nil {
		return 0, nil, err
	}

	addr := socks5.SplitAddr(b[:n])
	if addr == nil {
		return 0, nil, errors.New("parse addr error")
	}

	udpAddr := addr.UDPAddr()
	if udpAddr == nil {
		return 0, nil, errors.New("parse addr error")
	}

	copy(b, b[len(addr):])
	return n - len(addr), udpAddr, err
}
