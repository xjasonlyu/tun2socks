package socks5

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"

	"github.com/xjasonlyu/tun2socks/v2/dialer"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/proxy/internal/utils"
	"github.com/xjasonlyu/tun2socks/v2/transport/socks5"
)

var _ proxy.Proxy = (*Socks5)(nil)

type Socks5 struct {
	addr string

	user string
	pass string

	// unix indicates if socks5 over UDS is enabled.
	unix bool
}

func New(addr, user, pass string) (*Socks5, error) {
	unix := len(addr) > 0 && addr[0] == '/'

	// For support Linux abstract namespace
	if len(addr) > 2 && addr[1] == '@' || addr[1] == 0x00 {
		addr = addr[1:]
	}

	return &Socks5{
		addr: addr,
		user: user,
		pass: pass,
		unix: unix,
	}, nil
}

func (ss *Socks5) DialContext(ctx context.Context, metadata *M.Metadata) (c net.Conn, err error) {
	network := "tcp"
	if ss.unix {
		network = "unix"
	}

	c, err = dialer.DialContext(ctx, network, ss.addr)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", ss.addr, err)
	}
	utils.SetKeepAlive(c)

	defer func(c net.Conn) {
		utils.SafeConnClose(c, err)
	}(c)

	var user *socks5.User
	if ss.user != "" {
		user = &socks5.User{
			Username: ss.user,
			Password: ss.pass,
		}
	}

	_, err = socks5.ClientHandshake(c, utils.SerializeSocksAddr(metadata), socks5.CmdConnect, user)
	return c, err
}

func (ss *Socks5) DialUDP(*M.Metadata) (_ net.PacketConn, err error) {
	if ss.unix {
		return nil, fmt.Errorf("%w when unix domain socket is enabled", errors.ErrUnsupported)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		utils.TCPConnectTimeout,
	)
	defer cancel()

	c, err := dialer.DialContext(ctx, "tcp", ss.addr)
	if err != nil {
		err = fmt.Errorf("connect to %s: %w", ss.addr, err)
		return
	}
	utils.SetKeepAlive(c)

	defer func() {
		if err != nil && c != nil {
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

	addr, err := socks5.ClientHandshake(c, targetAddr, socks5.CmdUDPAssociate, user)
	if err != nil {
		return nil, fmt.Errorf("client handshake: %w", err)
	}

	pc, err := dialer.ListenPacket("udp", "")
	if err != nil {
		return nil, fmt.Errorf("listen packet: %w", err)
	}

	go func() {
		io.Copy(io.Discard, c)
		c.Close()
		// A UDP association terminates when the TCP connection that the UDP
		// ASSOCIATE request arrived on terminates. RFC1928
		pc.Close()
	}()

	bindAddr := addr.UDPAddr()
	if bindAddr == nil {
		return nil, fmt.Errorf("invalid UDP binding address: %#v", addr)
	}

	if bindAddr.IP.IsUnspecified() { /* e.g. "0.0.0.0" or "::" */
		udpAddr, err := net.ResolveUDPAddr("udp", ss.addr)
		if err != nil {
			return nil, fmt.Errorf("resolve udp address %s: %w", ss.addr, err)
		}
		bindAddr.IP = udpAddr.IP
	}

	return &socksPacketConn{PacketConn: pc, rAddr: bindAddr, tcpConn: c}, nil
}

type socksPacketConn struct {
	net.PacketConn

	rAddr   net.Addr
	tcpConn net.Conn
}

func (pc *socksPacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	var packet []byte
	if ma, ok := addr.(*M.Addr); ok {
		packet, err = socks5.EncodeUDPPacket(utils.SerializeSocksAddr(ma.Metadata()), b)
	} else {
		packet, err = socks5.EncodeUDPPacket(socks5.ParseAddr(addr), b)
	}

	if err != nil {
		return n, err
	}
	return pc.PacketConn.WriteTo(packet, pc.rAddr)
}

func (pc *socksPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, _, err := pc.PacketConn.ReadFrom(b)
	if err != nil {
		return 0, nil, err
	}

	addr, payload, err := socks5.DecodeUDPPacket(b)
	if err != nil {
		return 0, nil, err
	}

	udpAddr := addr.UDPAddr()
	if udpAddr == nil {
		return 0, nil, fmt.Errorf("convert %s to UDPAddr is nil", addr)
	}

	// due to DecodeUDPPacket is mutable, record addr length
	copy(b, payload)
	return n - len(addr) - 3, udpAddr, nil
}

func (pc *socksPacketConn) Close() error {
	pc.tcpConn.Close()
	return pc.PacketConn.Close()
}

func Parse(u *url.URL) (proxy.Proxy, error) {
	address, username := u.Host, u.User.Username()
	password, _ := u.User.Password()

	// Socks5 over UDS
	if address == "" {
		address = u.Path
	}
	return New(address, username, password)
}

func init() {
	proxy.RegisterProtocol("socks5", Parse)
}
