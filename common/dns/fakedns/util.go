package fakedns

import (
	D "github.com/miekg/dns"

	cache "github.com/xjasonlyu/tun2socks/common/lru-cache"
	// "github.com/xjasonlyu/tun2socks/common/cache"
)

/*
func putMsgToCache(c *cache.Cache, key string, msg *D.Msg) {
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
*/

func putMsgToCache(c *cache.Cache, key string, msg *D.Msg) {
	c.Put(key, msg.Copy())
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
