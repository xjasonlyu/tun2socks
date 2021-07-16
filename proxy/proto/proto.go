package proto

import "fmt"

const (
	Direct Proto = iota
	Reject
	HTTP
	Socks4
	Socks5
	Shadowsocks
)

type Proto uint8

func (proto Proto) String() string {
	switch proto {
	case Direct:
		return "direct"
	case Reject:
		return "reject"
	case HTTP:
		return "http"
	case Socks4:
		return "socks4"
	case Socks5:
		return "socks5"
	case Shadowsocks:
		return "ss"
	default:
		return fmt.Sprintf("proto(%d)", proto)
	}
}
