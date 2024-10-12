package engine

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"runtime"
	"strings"

	"github.com/gorilla/schema"

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

	driver := strings.ToLower(u.Scheme)

	switch driver {
	case fdbased.Driver:
		return parseFD(u, mtu)
	case tun.Driver:
		return parseTUN(u, mtu)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}

func parseFD(u *url.URL, mtu uint32) (device.Device, error) {
	offset := 0
	// fd offset in ios
	// https://stackoverflow.com/questions/69260852/ios-network-extension-packet-parsing/69487795#69487795
	if runtime.GOOS == "ios" {
		offset = 4
	}
	return fdbased.Open(u.Host, mtu, offset)
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
		return parseHTTP(u)
	case proto.Socks4.String():
		return parseSocks4(u)
	case proto.Socks5.String():
		return parseSocks5(u)
	case proto.Shadowsocks.String():
		return parseShadowsocks(u)
	case proto.Relay.String():
		return parseRelay(u)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

func parseHTTP(u *url.URL) (proxy.Proxy, error) {
	address, username := u.Host, u.User.Username()
	password, _ := u.User.Password()
	return proxy.NewHTTP(address, username, password)
}

func parseSocks4(u *url.URL) (proxy.Proxy, error) {
	address, userID := u.Host, u.User.Username()
	return proxy.NewSocks4(address, userID)
}

func parseSocks5(u *url.URL) (proxy.Proxy, error) {
	address, username := u.Host, u.User.Username()
	password, _ := u.User.Password()

	// Socks5 over UDS
	if address == "" {
		address = u.Path
	}
	return proxy.NewSocks5(address, username, password)
}

func parseShadowsocks(u *url.URL) (proxy.Proxy, error) {
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

	return proxy.NewShadowsocks(address, method, password, obfsMode, obfsHost)
}

func parseRelay(u *url.URL) (proxy.Proxy, error) {
	address, username := u.Host, u.User.Username()
	password, _ := u.User.Password()

	opts := struct {
		NoDelay bool
	}{}
	if err := schema.NewDecoder().Decode(&opts, u.Query()); err != nil {
		return nil, err
	}

	return proxy.NewRelay(address, username, password, opts.NoDelay)
}

func parseMulticastGroups(s string) (multicastGroups []netip.Addr, _ error) {
	for _, ip := range strings.Split(s, ",") {
		if ip = strings.TrimSpace(ip); ip == "" {
			continue
		}
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			return nil, err
		}
		if !addr.IsMulticast() {
			return nil, fmt.Errorf("invalid multicast IP: %s", addr)
		}
		multicastGroups = append(multicastGroups, addr)
	}
	return
}
