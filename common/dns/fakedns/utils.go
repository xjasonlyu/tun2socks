package fakedns

import (
	"sync"

	D "github.com/miekg/dns"
	cache "github.com/xjasonlyu/tun2socks/common/lru-cache"
)

func putMsgToMap(m sync.Map, key string, msg *D.Msg) {
	m.Store(key, msg.Copy())
}

func putMsgToCache(c *cache.Cache, key string, msg *D.Msg) {
	c.Add(key, msg.Copy())
}

func evictCallback(_ interface{}, value interface{}) {
	msg := value.(*D.Msg).Copy()
	ip := msg.Answer[0].(*D.A).A
	ipToHost.Delete(ip.String())
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
