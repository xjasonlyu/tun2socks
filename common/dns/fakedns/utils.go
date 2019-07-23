package fakedns

import (
	D "github.com/miekg/dns"
	cache "github.com/xjasonlyu/tun2socks/common/lru-cache"
)

func putMsgToCache(c *cache.Cache, key string, msg *D.Msg) {
	c.Add(key, msg.Copy())
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
