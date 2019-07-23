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
		ttl = 6 * time.Duration(dnsDefaultTTL) * time.Second
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
