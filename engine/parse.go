package engine

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/xjasonlyu/tun2socks/device"
	"github.com/xjasonlyu/tun2socks/device/tun"
	"github.com/xjasonlyu/tun2socks/proxy"
)

func parseDevice(s string, mtu uint32) (device.Device, error) {
	const defaultDriver = "tun"
	if !strings.Contains(s, "://") {
		s = defaultDriver + "://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	name := u.Host
	driver := u.Scheme

	var d device.Device

	switch driver {
	case "tun":
		d, err = tun.Open(tun.WithName(name), tun.WithMTU(mtu))
	default:
		err = fmt.Errorf("unsupported driver: %s", driver)
	}
	if err != nil {
		return nil, err
	}

	return d, nil
}

func parseProxy(s string) (proxy.Dialer, error) {
	const defaultProto = "socks5"
	if !strings.Contains(s, "://") {
		s = defaultProto + "://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	addr := u.Host
	user := u.User.Username()
	pass, _ := u.User.Password()
	proto := strings.ToLower(u.Scheme)

	switch proto {
	case "direct":
		return proxy.NewDirect(), nil
	case "socks5":
		return proxy.NewSocks5(addr, user, pass)
	case "ss", "shadowsocks":
		method, password := user, pass
		return proxy.NewShadowSocks(addr, method, password)
	}

	return nil, fmt.Errorf("unsupported protocol: %s", proto)
}
