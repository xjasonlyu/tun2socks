package relay

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/url"
	"sync"

	"github.com/go-gost/relay"
	"github.com/gorilla/schema"

	"github.com/xjasonlyu/tun2socks/v2/buffer"
	"github.com/xjasonlyu/tun2socks/v2/dialer"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/proxy/internal/utils"
)

var _ proxy.Proxy = (*Relay)(nil)

type Relay struct {
	addr string
	user string
	pass string

	noDelay bool
}

func New(addr, user, pass string, noDelay bool) (*Relay, error) {
	return &Relay{
		addr:    addr,
		user:    user,
		pass:    pass,
		noDelay: noDelay,
	}, nil
}

func (rl *Relay) DialContext(ctx context.Context, metadata *M.Metadata) (c net.Conn, err error) {
	return rl.dialContext(ctx, metadata)
}

func (rl *Relay) DialUDP(metadata *M.Metadata) (net.PacketConn, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		utils.TCPConnectTimeout,
	)
	defer cancel()

	return rl.dialContext(ctx, metadata)
}

func (rl *Relay) dialContext(ctx context.Context, metadata *M.Metadata) (rc *relayConn, err error) {
	var c net.Conn

	c, err = dialer.DialContext(ctx, "tcp", rl.addr)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", rl.addr, err)
	}
	utils.SetKeepAlive(c)

	defer func(c net.Conn) {
		utils.SafeConnClose(c, err)
	}(c)

	req := relay.Request{
		Version: relay.Version1,
		Cmd:     relay.CmdConnect,
	}

	if metadata.Network == M.UDP {
		req.Cmd |= relay.FUDP
		req.Features = append(req.Features, &relay.NetworkFeature{
			Network: relay.NetworkUDP,
		})
	}

	if rl.user != "" {
		req.Features = append(req.Features, &relay.UserAuthFeature{
			Username: rl.user,
			Password: rl.pass,
		})
	}

	req.Features = append(req.Features, serializeRelayAddr(metadata))

	if rl.noDelay {
		if _, err = req.WriteTo(c); err != nil {
			return rc, err
		}
		if err = readRelayResponse(c); err != nil {
			return rc, err
		}
	}

	switch metadata.Network {
	case M.TCP:
		rc = newRelayConn(c, metadata.Addr(), rl.noDelay, false)
		if !rl.noDelay {
			if _, err = req.WriteTo(rc.wbuf); err != nil {
				return rc, err
			}
		}
	case M.UDP:
		rc = newRelayConn(c, metadata.Addr(), rl.noDelay, true)
		if !rl.noDelay {
			if _, err = req.WriteTo(rc.wbuf); err != nil {
				return rc, err
			}
		}
	default:
		err = fmt.Errorf("network %s is unsupported", metadata.Network)
		return rc, err
	}

	return rc, err
}

type relayConn struct {
	net.Conn
	udp  bool
	addr net.Addr
	once sync.Once
	wbuf *bytes.Buffer
}

func newRelayConn(c net.Conn, addr net.Addr, noDelay, udp bool) *relayConn {
	rc := &relayConn{
		Conn: c,
		addr: addr,
		udp:  udp,
	}
	if !noDelay {
		rc.wbuf = &bytes.Buffer{}
	}
	return rc
}

func (rc *relayConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, err := rc.Read(b)
	return n, rc.addr, err
}

func (rc *relayConn) Read(b []byte) (n int, err error) {
	rc.once.Do(func() {
		if rc.wbuf != nil {
			err = readRelayResponse(rc.Conn)
		}
	})
	if err != nil {
		return n, err
	}

	if !rc.udp {
		return rc.Conn.Read(b)
	}

	var bb [2]byte
	_, err = io.ReadFull(rc.Conn, bb[:])
	if err != nil {
		return n, err
	}

	dLen := int(binary.BigEndian.Uint16(bb[:]))
	if len(b) >= dLen {
		return io.ReadFull(rc.Conn, b[:dLen])
	}

	buf := buffer.Get(dLen)
	defer buffer.Put(buf)
	_, err = io.ReadFull(rc.Conn, buf)
	n = copy(b, buf)

	return n, err
}

func (rc *relayConn) WriteTo(b []byte, _ net.Addr) (int, error) {
	return rc.Write(b)
}

func (rc *relayConn) Write(b []byte) (int, error) {
	if rc.udp {
		return rc.udpWrite(b)
	}
	return rc.tcpWrite(b)
}

func (rc *relayConn) tcpWrite(b []byte) (n int, err error) {
	if rc.wbuf != nil && rc.wbuf.Len() > 0 {
		n = len(b)
		rc.wbuf.Write(b)
		_, err = rc.Conn.Write(rc.wbuf.Bytes())
		rc.wbuf.Reset()
		return n, err
	}
	return rc.Conn.Write(b)
}

func (rc *relayConn) udpWrite(b []byte) (n int, err error) {
	if len(b) > math.MaxUint16 {
		err = errors.New("write: data maximum exceeded")
		return n, err
	}

	n = len(b)
	if rc.wbuf != nil && rc.wbuf.Len() > 0 {
		var bb [2]byte
		binary.BigEndian.PutUint16(bb[:], uint16(len(b)))
		rc.wbuf.Write(bb[:])
		rc.wbuf.Write(b)
		_, err = rc.wbuf.WriteTo(rc.Conn)
		return n, err
	}

	var bb [2]byte
	binary.BigEndian.PutUint16(bb[:], uint16(len(b)))
	_, err = rc.Conn.Write(bb[:])
	if err != nil {
		return n, err
	}
	return rc.Conn.Write(b)
}

func readRelayResponse(r io.Reader) error {
	resp := relay.Response{}
	if _, err := resp.ReadFrom(r); err != nil {
		return err
	}
	if resp.Version != relay.Version1 {
		return relay.ErrBadVersion
	}
	if resp.Status != relay.StatusOK {
		return fmt.Errorf("status %d", resp.Status)
	}
	return nil
}

func serializeRelayAddr(m *M.Metadata) *relay.AddrFeature {
	af := &relay.AddrFeature{
		Host: m.DstIP.String(),
		Port: m.DstPort,
	}
	if m.DstIP.Is4() {
		af.AType = relay.AddrIPv4
	} else {
		af.AType = relay.AddrIPv6
	}
	return af
}

func Parse(u *url.URL) (proxy.Proxy, error) {
	address, username := u.Host, u.User.Username()
	password, _ := u.User.Password()

	opts := struct{ NoDelay bool }{}
	if err := schema.NewDecoder().
		Decode(&opts, u.Query()); err != nil {
		return nil, err
	}

	return New(address, username, password, opts.NoDelay)
}

func init() {
	proxy.RegisterProtocol("relay", Parse)
}
