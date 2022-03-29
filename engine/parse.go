package engine

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/fdbased"
	"github.com/xjasonlyu/tun2socks/v2/core/device/tun"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/proxy/proto"
)

func parseRestAPI(s string) (*url.URL, error) {
	if !strings.Contains(s, "://") {
		s = fmt.Sprintf("%s://%s", "http", s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	addr, err := net.ResolveTCPAddr("tcp", u.Host)
	if err != nil {
		return nil, err
	}
	if addr.IP == nil {
		addr.IP = net.IPv4zero /* default: 0.0.0.0 */
	}
	u.Host = addr.String()

	switch u.Scheme {
	case "http":
		return u, nil
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
}

func parseDevice(s string, mtu uint32) (device.Device, error) {
	if !strings.Contains(s, "://") {
		s = fmt.Sprintf("%s://%s", tun.Driver /* default driver */, s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	name := u.Host
	driver := strings.ToLower(u.Scheme)

	switch driver {
	case fdbased.Driver:
		return fdbased.Open(name, mtu)
	case tun.Driver:
		return tun.Open(name, mtu)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}

func parseProxy(s string) (proxy.Proxy, error) {
	if !strings.Contains(s, "://") {
		s = fmt.Sprintf("%s://%s", proto.Socks5 /* default protocol */, s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	protocol := strings.ToLower(u.Scheme)

	switch protocol {
	case proto.Direct.String():
		return proxy.NewDirect(), nil
	case proto.Reject.String():
		return proxy.NewReject(), nil
	case proto.HTTP.String():
		return proxy.NewHTTP(parseHTTP(u))
	case proto.Socks4.String():
		return proxy.NewSocks4(parseSocks4(u))
	case proto.Socks5.String():
		return proxy.NewSocks5(parseSocks5(u))
	case proto.Shadowsocks.String():
		return proxy.NewShadowsocks(parseShadowsocks(u))
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

func parseHTTP(u *url.URL) (address, username, password string) {
	address, username = u.Host, u.User.Username()
	password, _ = u.User.Password()
	return
}

func parseSocks4(u *url.URL) (address, username string) {
	address, username = u.Host, u.User.Username()
	return
}

func parseSocks5(u *url.URL) (address, username, password string) {
	address, username = u.Host, u.User.Username()
	password, _ = u.User.Password()

	// Socks5 over UDS
	if address == "" {
		address = u.Path
	}
	return
}

func parseShadowsocks(u *url.URL) (address, method, password, obfsMode, obfsHost string) {
	address = u.Host

	if pass, set := u.User.Password(); set {
		method = u.User.Username()
		password = pass
	} else {
		data, _ := base64.RawURLEncoding.DecodeString(u.User.String())
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

	return
}
