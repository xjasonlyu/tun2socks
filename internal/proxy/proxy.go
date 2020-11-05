package proxy

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/xjasonlyu/tun2socks/internal/adapter"
)

type Dialer interface {
	Addr() string
	Type() string
	String() string

	DialContext(context.Context, *adapter.Metadata) (net.Conn, error)
	DialUDP(*adapter.Metadata) (net.PacketConn, error)
}

var _defaultDialer Dialer = &Base{}

// New returns proxy dialer.
func New(proxyURL string) (Dialer, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	proto := strings.ToLower(u.Scheme)
	user := u.User.Username()
	pass, _ := u.User.Password()

	switch proto {
	case "direct":
		return NewDirect(u)
	case "socks5":
		return NewSocks5(u, user, pass)
	case "ss", "shadowsocks":
		method, password := user, pass
		return NewShadowSocks(u, method, password)
	}

	return nil, fmt.Errorf("unsupported protocol: %s", proto)
}

// Register updates the _defaultDialer.
func Register(proxyURL string) error {
	dialer, err := New(proxyURL)
	if err != nil {
		return err
	}

	_defaultDialer = dialer
	return nil
}

// Dial uses _defaultDialer to dial TCP.
func Dial(metadata *adapter.Metadata) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), tcpConnectTimeout)
	defer cancel()
	return _defaultDialer.DialContext(ctx, metadata)
}

// DialUDP uses _defaultDialer to dial UDP.
func DialUDP(metadata *adapter.Metadata) (net.PacketConn, error) {
	return _defaultDialer.DialUDP(metadata)
}

// Addr returns _defaultDialer addr.
func Addr() string {
	return _defaultDialer.Addr()
}

// Type returns _defaultDialer type.
func Type() string {
	return _defaultDialer.Type()
}

// String returns _defaultDialer URL.
func String() string {
	return _defaultDialer.String()
}
