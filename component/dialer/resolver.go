package dialer

import "net"

func init() {
	net.DefaultResolver.PreferGo = true
	net.DefaultResolver.Dial = DialContext
}
