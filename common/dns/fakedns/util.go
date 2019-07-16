package fakedns

import (
	"encoding/binary"
	"net"
	"time"

	D "github.com/miekg/dns"
)

func putMsgToCache(c *Cache, key string, msg *D.Msg) {
	var ttl time.Duration
	if len(msg.Answer) != 0 {
		ttl = time.Duration(msg.Answer[0].Header().Ttl) * time.Second
	} else if len(msg.Ns) != 0 {
		ttl = time.Duration(msg.Ns[0].Header().Ttl) * time.Second
	} else if len(msg.Extra) != 0 {
		ttl = time.Duration(msg.Extra[0].Header().Ttl) * time.Second
	} else {
		return
	}

	c.Put(key, msg.Copy(), ttl)
}

func setMsgTTL(msg *D.Msg, ttl uint32) {
	for _, answer := range msg.Answer {
		answer.Header().Ttl = ttl
	}

	for _, ns := range msg.Ns {
		ns.Header().Ttl = ttl
	}

	for _, extra := range msg.Extra {
		extra.Header().Ttl = ttl
	}
}

/*
func uint322ip(n uint32) net.IP {
	return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
*/

func ip2uint32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32([]byte(ip)[net.IPv6len-net.IPv4len:])
}
