package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/Dreamacro/go-shadowsocks2/core"

	"github.com/xjasonlyu/clash/component/dialer"
	"github.com/xjasonlyu/clash/component/socks5"
	"github.com/xjasonlyu/tun2socks/internal/adapter"
)

type ShadowSocks struct {
	*Base

	cipher core.Cipher
}

func NewShadowSocks(url *url.URL, method, password string) (*ShadowSocks, error) {
	cipher, err := core.PickCipher(method, nil, password)
	if err != nil {
		return nil, fmt.Errorf("ss initialize: %w", err)
	}

	return &ShadowSocks{
		Base: &Base{
			url: url,
		},
		cipher: cipher,
	}, nil
}

func (ss *ShadowSocks) DialContext(ctx context.Context, metadata *adapter.Metadata) (c net.Conn, err error) {
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

	c = ss.cipher.StreamConn(c)
	_, err = c.Write(metadata.SerializesSocksAddr())
	return
}

func (ss *ShadowSocks) DialUDP(_ *adapter.Metadata) (net.PacketConn, error) {
	pc, err := dialer.ListenPacket("udp", "")
	if err != nil {
		return nil, err
	}

	udpAddr, err := resolveUDPAddr("udp", ss.Addr())
	if err != nil {
		return nil, fmt.Errorf("resolve UDPAddr %s failed", ss.Addr())
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
	if m, ok := addr.(*adapter.Metadata); ok {
		packet, err = socks5.EncodeUDPPacket(m.SerializesSocksAddr(), b)
	} else {
		packet, err = socks5.EncodeUDPPacket(socks5.ParseAddrToSocksAddr(addr), b)
	}

	if err != nil {
		return
	}
	return pc.PacketConn.WriteTo(packet[3:], pc.rAddr)
}

func (pc *ssPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, _, e := pc.PacketConn.ReadFrom(b)
	if e != nil {
		return 0, nil, e
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
	return n - len(addr), udpAddr, e
}
