package proxy

import "fmt"

const (
	DirectProto Proto = iota
	ShadowsocksProto
	Socks5Proto
)

type Proto uint8

func (proto Proto) String() string {
	switch proto {
	case DirectProto:
		return "direct"
	case ShadowsocksProto:
		return "ss"
	case Socks5Proto:
		return "socks5"
	default:
		return fmt.Sprintf("proto(%d)", proto)
	}
}
