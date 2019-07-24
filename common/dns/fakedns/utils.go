package fakedns

import (
	"strings"
	"time"

	D "github.com/miekg/dns"
	"github.com/xjasonlyu/tun2socks/common/cache"
)

func putMsgToCache(c *cache.Cache, key string, msg *D.Msg) {
	var ttl time.Duration
	if strings.HasPrefix(key, "fakeip:") {
		ttl = time.Duration(dnsDefaultTTL) * time.Second
	} else {
		ttl = 3 * time.Duration(dnsDefaultTTL) * time.Second
	}
	c.Put(key, msg.Copy(), ttl)
}

func getMsgFromCache(c *cache.Cache, key string) (msg *D.Msg) {
	item := c.Get(key)
	if item == nil {
		return
	}
	msg = item.(*D.Msg).Copy()
	putMsgToCache(c, key, msg)
	return
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
