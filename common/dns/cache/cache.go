// This file is copied from https://github.com/yinghuocho/gotun2socks/blob/master/udp.go

package cache

import (
	"sync"
	"time"

	"github.com/miekg/dns"

	cdns "github.com/xjasonlyu/tun2socks/common/dns"
	"github.com/xjasonlyu/tun2socks/common/log"
)

const minCleanupInterval = 5 * time.Minute

type dnsCacheEntry struct {
	msg []byte
	exp time.Time
}

type simpleDnsCache struct {
	mutex       sync.Mutex
	storage     map[string]*dnsCacheEntry
	lastCleanup time.Time
}

func NewSimpleDnsCache() cdns.DnsCache {
	return &simpleDnsCache{
		storage:     make(map[string]*dnsCacheEntry),
		lastCleanup: time.Now(),
	}
}

func packUint16(i uint16) []byte { return []byte{byte(i >> 8), byte(i)} }

func cacheKey(q dns.Question) string {
	return string(append([]byte(q.Name), packUint16(q.Qtype)...))
}

func (c *simpleDnsCache) cleanup() {
	newStorage := make(map[string]*dnsCacheEntry)
	log.Debugf("cleaning up dns %v cache entries", len(c.storage))
	for key, entry := range c.storage {
		if time.Now().Before(entry.exp) {
			newStorage[key] = entry
		}
	}
	c.storage = newStorage
	log.Debugf("cleanup done, remaining %v entries", len(c.storage))
}

func (c *simpleDnsCache) Query(payload []byte) []byte {
	request := new(dns.Msg)
	e := request.Unpack(payload)
	if e != nil {
		return nil
	}
	if len(request.Question) == 0 {
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	key := cacheKey(request.Question[0])
	entry := c.storage[key]
	if entry == nil {
		return nil
	}
	if time.Now().After(entry.exp) {
		delete(c.storage, key)
		return nil
	}

	resp := new(dns.Msg)
	resp.Unpack(entry.msg)
	resp.Id = request.Id
	var buf [1024]byte
	dnsAnswer, err := resp.PackBuffer(buf[:])
	if err != nil {
		return nil
	}
	log.Debugf("got dns answer from cache with key: %v", key)
	return append([]byte(nil), dnsAnswer...)
}

func (c *simpleDnsCache) Store(payload []byte) {
	resp := new(dns.Msg)
	e := resp.Unpack(payload)
	if e != nil {
		return
	}
	if resp.Rcode != dns.RcodeSuccess {
		return
	}
	if len(resp.Question) == 0 || len(resp.Answer) == 0 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	key := cacheKey(resp.Question[0])
	c.storage[key] = &dnsCacheEntry{
		msg: payload,
		exp: time.Now().Add(time.Duration(resp.Answer[0].Header().Ttl) * time.Second),
	}
	log.Debugf("stored dns answer with key: %v, ttl: %v sec", key, resp.Answer[0].Header().Ttl)

	now := time.Now()
	if now.Sub(c.lastCleanup) > minCleanupInterval {
		c.cleanup()
		c.lastCleanup = now
	}
}
