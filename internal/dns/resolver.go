package dns

import (
	"net"

	"github.com/xjasonlyu/clash/component/dialer"
)

func init() { /* use bound dialer to resolve DNS */
	net.DefaultResolver.PreferGo = true
	net.DefaultResolver.Dial = dialer.DialContext
}
