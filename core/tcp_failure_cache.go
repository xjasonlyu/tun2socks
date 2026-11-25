package core

import (
	"container/list"
	"sync"
	"time"

	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

const (
	tcpFailureTTL      = 30 * time.Second
	tcpFailureCapacity = 4096
)

type tcpFailureKey struct {
	id  stack.TransportEndpointID
	isn uint32
}

type tcpFailureEntry struct {
	key     tcpFailureKey
	expires time.Time
}

// Per-stack FIFO, not LRU: hits never extend the lifetime or move an entry.
// Access-time expiry keeps idle stacks free of timer goroutines. Stop releases
// all entries and prevents late canceled dials from repopulating the cache.
type tcpFailureCache struct {
	mu       sync.Mutex
	entries  map[tcpFailureKey]*list.Element
	order    list.List
	now      func() time.Time
	stopped  bool
	ttl      time.Duration
	capacity int
}

func newTCPFailureCache(ttl time.Duration, capacity int) *tcpFailureCache {
	return &tcpFailureCache{entries: make(map[tcpFailureKey]*list.Element), now: time.Now, ttl: ttl, capacity: capacity}
}

func (c *tcpFailureCache) prune(now time.Time) {
	for e := c.order.Front(); e != nil; e = c.order.Front() {
		if now.Before(e.Value.(tcpFailureEntry).expires) {
			break
		}
		c.remove(e)
	}
}

func (c *tcpFailureCache) remove(e *list.Element) {
	delete(c.entries, e.Value.(tcpFailureEntry).key)
	c.order.Remove(e)
}

func (c *tcpFailureCache) contains(key tcpFailureKey) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prune(c.now())
	_, found := c.entries[key]
	return found
}

func (c *tcpFailureCache) add(key tcpFailureKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stopped {
		return
	}
	now := c.now()
	c.prune(now)
	if _, found := c.entries[key]; found {
		return
	}
	if len(c.entries) >= c.capacity {
		c.remove(c.order.Front())
	}
	c.entries[key] = c.order.PushBack(tcpFailureEntry{key: key, expires: now.Add(c.ttl)})
}

func (c *tcpFailureCache) stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stopped = true
	clear(c.entries)
	c.order.Init()
}
