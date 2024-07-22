package engine

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/fdbased"
	"github.com/xjasonlyu/tun2socks/v2/core/device/tun"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
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
	return fdbased.Open(u.Host, mtu, 0)
}

func parseProxy(s string) (proxy.Proxy, error) {
	return proxy.ParseFromURL(s)
}

func parseMulticastGroups(s string) (multicastGroups []net.IP, _ error) {
	ipStrings := strings.Split(s, ",")
	for _, ipString := range ipStrings {
		if strings.TrimSpace(ipString) == "" {
			continue
		}
		ip := net.ParseIP(ipString)
		if ip == nil {
			return nil, fmt.Errorf("invalid IP format: %s", ipString)
		}
		if !ip.IsMulticast() {
			return nil, fmt.Errorf("invalid multicast IP address: %s", ipString)
		}
		multicastGroups = append(multicastGroups, ip)
	}
	return
}
