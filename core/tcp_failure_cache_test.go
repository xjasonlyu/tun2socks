package core

import (
	"sync"
	"testing"
	"time"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func TestTCPFailureCacheFixedExpiry(t *testing.T) {
	c := newTCPFailureCache(tcpFailureTTL, tcpFailureCapacity)
	now := time.Now()
	c.now = func() time.Time { return now }
	k := tcpFailureKey{isn: 42}
	c.add(k)
	now = now.Add(29 * time.Second)
	if !c.contains(k) {
		t.Fatal("expired before 30 seconds")
	}
	c.add(k) // Neither hits nor duplicate insertion refresh the expiry.
	now = now.Add(time.Second)
	if c.contains(k) || len(c.entries) != 0 {
		t.Fatal("entry survived its fixed 30-second TTL")
	}
	c.add(k)
	if !c.contains(k) {
		t.Fatal("expired key cannot be admitted again")
	}
}

func TestTCPFailureCacheCapacityFIFOAndExpiry(t *testing.T) {
	c := newTCPFailureCache(tcpFailureTTL, tcpFailureCapacity)
	now := time.Now()
	c.now = func() time.Time { return now }
	for i := 0; i < 4096; i++ {
		c.add(tcpFailureKey{isn: uint32(i)})
	}
	if len(c.entries) != 4096 {
		t.Fatal("incorrect capacity")
	}
	if !c.contains(tcpFailureKey{}) {
		t.Fatal("missing oldest key")
	}
	c.add(tcpFailureKey{isn: 4096})
	if len(c.entries) != 4096 || c.contains(tcpFailureKey{}) || !c.contains(tcpFailureKey{isn: 1}) {
		t.Fatal("cache must evict oldest insertion, not least recently accessed")
	}
	now = now.Add(30 * time.Second)
	c.add(tcpFailureKey{isn: 9000})
	if len(c.entries) != 1 || c.order.Len() != 1 {
		t.Fatal("expired entries not pruned before admission")
	}
}

func TestTCPFailureCacheKeysAndStop(t *testing.T) {
	c := newTCPFailureCache(tcpFailureTTL, tcpFailureCapacity)
	k := tcpFailureKey{id: stack.TransportEndpointID{
		LocalAddress: tcpip.AddrFrom4([4]byte{1, 2, 3, 4}), LocalPort: 5,
		RemoteAddress: tcpip.AddrFrom4([4]byte{10, 60, 173, 33}), RemotePort: 63293,
	}, isn: 3361995830}
	c.add(k)
	variants := []tcpFailureKey{k, k, k, k, k}
	variants[0].isn++
	variants[1].id.LocalPort++
	variants[2].id.RemotePort++
	variants[3].id.LocalAddress = tcpip.AddrFrom4([4]byte{1, 2, 3, 5})
	variants[4].id.RemoteAddress = tcpip.AddrFrom4([4]byte{10, 60, 173, 34})
	for _, other := range variants {
		if c.contains(other) {
			t.Fatal("unrelated SYN rejected")
		}
	}
	if newTCPFailureCache(tcpFailureTTL, tcpFailureCapacity).contains(k) {
		t.Fatal("cache leaked across stacks")
	}
	c.stop()
	c.add(k)
	if c.contains(k) || c.order.Len() != 0 {
		t.Fatal("stopped cache repopulated by late dial")
	}
}

func TestTCPFailureCacheConcurrentStop(t *testing.T) {
	c := newTCPFailureCache(tcpFailureTTL, tcpFailureCapacity)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for n := 0; n < 1000; n++ {
				key := tcpFailureKey{isn: uint32(i*1000 + n)}
				c.add(key)
				c.contains(key)
				if n == 500 {
					c.stop()
				}
			}
		}(i)
	}
	wg.Wait()
	if len(c.entries) != 0 || c.order.Len() != 0 {
		t.Fatal("stop did not clear all entries")
	}
}
