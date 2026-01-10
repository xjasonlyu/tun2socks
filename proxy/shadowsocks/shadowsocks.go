package shadowsocks

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/xjasonlyu/tun2socks/v2/dialer"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/proxy/internal/utils"
	"github.com/xjasonlyu/tun2socks/v2/transport/shadowsocks/core"
	obfs "github.com/xjasonlyu/tun2socks/v2/transport/simple-obfs"
	"github.com/xjasonlyu/tun2socks/v2/transport/socks5"
)

var _ proxy.Proxy = (*Shadowsocks)(nil)

type Shadowsocks struct {
	addr   string
	cipher core.Cipher

	// simple-obfs plugin
	obfsMode, obfsHost string
}

func New(addr, method, password, obfsMode, obfsHost string) (*Shadowsocks, error) {
	cipher, err := core.PickCipher(method, nil, password)
	if err != nil {
		return nil, fmt.Errorf("ss initialize: %w", err)
	}

	return &Shadowsocks{
		addr:     addr,
		cipher:   cipher,
		obfsMode: obfsMode,
		obfsHost: obfsHost,
	}, nil
}

func (ss *Shadowsocks) DialContext(ctx context.Context, metadata *M.Metadata) (c net.Conn, err error) {
	c, err = dialer.DialContext(ctx, "tcp", ss.addr)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", ss.addr, err)
	}
	utils.SetKeepAlive(c)

	defer func(c net.Conn) {
		utils.SafeConnClose(c, err)
	}(c)

	switch ss.obfsMode {
	case "tls":
		c = obfs.NewTLSObfs(c, ss.obfsHost)
	case "http":
		_, port, _ := net.SplitHostPort(ss.addr)
		c = obfs.NewHTTPObfs(c, ss.obfsHost, port)
	}

	c = ss.cipher.StreamConn(c)
	_, err = c.Write(utils.SerializeSocksAddr(metadata))
	return c, err
}

func (ss *Shadowsocks) DialUDP(*M.Metadata) (net.PacketConn, error) {
	pc, err := dialer.ListenPacket("udp", "")
	if err != nil {
		return nil, fmt.Errorf("listen packet: %w", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", ss.addr)
	if err != nil {
		return nil, fmt.Errorf("resolve udp address %s: %w", ss.addr, err)
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
		packet, err = socks5.EncodeUDPPacket(utils.SerializeSocksAddr(ma.Metadata()), b)
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

func Parse(u *url.URL) (proxy.Proxy, error) {
	var (
		address            = u.Host
		method, password   string
		obfsMode, obfsHost string
	)

	if ss := u.User.String(); ss == "" {
		method = "dummy" // none cipher mode
	} else if pass, set := u.User.Password(); set {
		method = u.User.Username()
		password = pass
	} else {
		data, _ := base64.RawURLEncoding.DecodeString(ss)
		userInfo := strings.SplitN(string(data), ":", 2)
		if len(userInfo) == 2 {
			method = userInfo[0]
			password = userInfo[1]
		}
	}

	rawQuery, _ := url.QueryUnescape(u.RawQuery)
	for _, s := range strings.Split(rawQuery, ";") {
		data := strings.SplitN(s, "=", 2)
		if len(data) != 2 {
			continue
		}
		key := data[0]
		value := data[1]

		switch key {
		case "obfs":
			obfsMode = value
		case "obfs-host":
			obfsHost = value
		}
	}

	return New(address, method, password, obfsMode, obfsHost)
}

func init() {
	proxy.RegisterProtocol("ss", Parse)
}
