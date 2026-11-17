package core

import (
	"testing"
	"time"
)

func TestDeferredConfigDefaultsAndCustomValues(t *testing.T) {
	input := DeferredTCPConfig{}
	settings, err := input.normalized()
	if err != nil || settings.FailureCacheTTL != 30*time.Second || settings.FailureCacheCapacity != 4096 || settings.Context == nil {
		t.Fatalf("bad defaults: %+v %v", settings, err)
	}
	if input.Context != nil || input.FailureCacheTTL != 0 || input.FailureCacheCapacity != 0 {
		t.Fatal("normalization mutated caller config")
	}
	settings, err = (DeferredTCPConfig{FailureCacheTTL: 2 * time.Second, FailureCacheCapacity: 2}).normalized()
	if err != nil {
		t.Fatal(err)
	}
	c := newTCPFailureCache(settings.FailureCacheTTL, settings.FailureCacheCapacity)
	now := time.Now()
	c.now = func() time.Time { return now }
	c.add(tcpFailureKey{isn: 1})
	c.add(tcpFailureKey{isn: 2})
	c.add(tcpFailureKey{isn: 3})
	if c.contains(tcpFailureKey{isn: 1}) || len(c.entries) != 2 {
		t.Fatal("custom capacity ignored")
	}
	now = now.Add(time.Second)
	if !c.contains(tcpFailureKey{isn: 2}) {
		t.Fatal("custom TTL expired early")
	}
	c.add(tcpFailureKey{isn: 2})
	now = now.Add(time.Second)
	if c.contains(tcpFailureKey{isn: 2}) || c.contains(tcpFailureKey{isn: 3}) {
		t.Fatal("custom TTL ignored or refreshed")
	}
}

func TestDeferredConfigRejectsNegativeValues(t *testing.T) {
	for _, cfg := range []DeferredTCPConfig{{FailureCacheTTL: -1}, {FailureCacheCapacity: -1}} {
		if _, err := cfg.normalized(); err == nil {
			t.Fatal("negative value accepted")
		}
		if s, err := CreateStack(&Config{DeferredTCP: &cfg}); err == nil || s != nil {
			t.Fatal("invalid settings reached stack setup")
		}
	}
}
