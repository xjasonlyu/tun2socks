package engine

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/xjasonlyu/tun2socks/core/device"
	"github.com/xjasonlyu/tun2socks/core/device/tun"
	"github.com/xjasonlyu/tun2socks/proxy"
	"github.com/xjasonlyu/tun2socks/proxy/proto"
)

func parseDevice(s string, mtu uint32) (device.Device, error) {
	if !strings.Contains(s, "://") {
		s = tun.Driver + "://" + s /* default driver */
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	name := u.Host
	driver := strings.ToLower(u.Scheme)

	switch driver {
	case tun.Driver:
		return tun.Open(tun.WithName(name), tun.WithMTU(mtu))
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}

func parseProxy(s string) (proxy.Proxy, error) {
	if !strings.Contains(s, "://") {
		s = proto.Socks5.String() + "://" + s /* default protocol */
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
	case proto.Socks5.String():
		return proxy.NewSocks5(parseSocks(u))
	case proto.Shadowsocks.String():
		return proxy.NewShadowsocks(parseShadowsocks(u))
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

func parseSocks(u *url.URL) (address, username, password string) {
	address = u.Host
	username = u.User.Username()
	password, _ = u.User.Password()
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
