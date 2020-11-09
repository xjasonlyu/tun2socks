package dns

import (
	"net"

	"github.com/xjasonlyu/clash/component/dialer"
	"github.com/xjasonlyu/clash/component/resolver"
)

func init() {

	// enable ipv6
	resolver.DisableIPv6 = false

	// use bound dialer to resolve DNS
	net.DefaultResolver.PreferGo = true
	net.DefaultResolver.Dial = dialer.DialContext
}
